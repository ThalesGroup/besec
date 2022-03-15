package store

import (
	"context"
	"fmt"
	"sort"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/go-openapi/strfmt"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ThalesGroup/besec/api/models"
	"github.com/ThalesGroup/besec/lib"
)

// There is a lot of copy/paste repetition here, which I wonder
// if there is a way to avoid despite the absence of generics
// Maybe implementing some interfaces on the model types is a good way to go?
// https://appliedgo.net/generics/

// Some of the type definitions in the API are informed by how easy it is to operate
// on them directly in the database. This isn't ideal - the API shouldn't care about the
// implementation - but it does make implementation easier!

// Where we have firestore-specific data types, they're defined below.
// The only reason to lowercase the serialized names in these structs is
// to be consistent with the public API, which slightly reduces the chance
// of errors from mis-cased field names.

const projectsCollection = "projects"
const plansCollection = "plans"
const planRevisionsSubCollection = "revisions" // a collection within a particular plan
const usersCollection = "users"
const practicesCollection = "practices"

const configDoc = "config/config"

// Keeping the version information as part of a plan revision makes retrieval of the latest
// revision easier than storing versions separately. Note the storage here doesn't directly
// match the API data structures.
type storedPlanRevision struct {
	Plan    *lib.Plan      `json:"plan"`
	Version *storedVersion `json:"version"`
}
type storedVersion struct {
	Author *models.VersionAuthor `json:"author"`
	Time   time.Time             `json:"time"` // this is aliased in models.Version, so doesn't get saved properly
}

// Firestore documents can't directly contain arrays
type storedPractices struct {
	Practices []lib.Practice
}

// To avoid a very expensive lookup (iterating through every plan to get its latest revision)
// we store the plans associated with a project alongside the project
// Operations that affect the latest plan revision need to also update the project's record of
// associated plan IDs
type storedProject struct {
	Details *models.ProjectDetails `json:"details"`
	Plans   []string               `json:"plans"`
}

// FireStore implements the Store interface with Google Firestore
type FireStore struct {
	client *firestore.Client
}

// Check at compile time that FireStore correctly meets the Store interface
var _ Store = (*FireStore)(nil)

type updateOp struct {
	doc     *firestore.DocumentRef
	updates []firestore.Update
}

// NewFireStore initializes a FireStore
func NewFireStore(project string) *FireStore {
	if project == "" {
		log.Fatal("Cannot set up firestore without specifying a GCP project ID")
	}

	// Get a Firestore client.
	options := []option.ClientOption{}
	client, err := firestore.NewClient(context.Background(), project, options...)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("Failed to create firestore client")
	}

	// Close client when done.
	//defer client.Close()
	return &FireStore{client}
}

func planRevisionsPath(planID string) string {
	return plansCollection + "/" + planID + "/" + planRevisionsSubCollection
}

// ListProjects returns all of the projects
func (s *FireStore) ListProjects(ctx context.Context) ([]*models.Project, error) {
	logger := log.WithContext(ctx)

	docs, err := s.client.Collection(projectsCollection).Documents(ctx).GetAll()
	if err != nil {
		logger.Error("Firestore ListProjects: error retrieving projects: ", err)
		return nil, fmt.Errorf("error retrieving projects")
	}
	logger.WithFields(log.Fields{"num-docs": len(docs)}).Debug("Firestore ListProjects: documents retrieved")

	ps := make([]*models.Project, len(docs))
	for i, d := range docs {
		p, err := projectFromDocsnap(d)
		if err != nil {
			logger.Error("FireStore.ListProjects: error coercing retrieved project to models.Project", err)
			return nil, fmt.Errorf("Error retrieving projects")
		}
		ps[i] = p
	}
	return ps, nil
}

func projectFromDocsnap(d *firestore.DocumentSnapshot) (*models.Project, error) {
	sp := new(storedProject)
	err := d.DataTo(&sp)
	if err != nil {
		return nil, err
	}
	// populate the ID as it's not stored as part of the document
	return &models.Project{ID: d.Ref.ID, Attributes: sp.Details, Plans: sp.Plans}, nil
}

// GetProject returns the project with the specified ID and true, or false if it can't be found
func (s *FireStore) GetProject(ctx context.Context, id string) (*models.Project, bool, error) {
	logger := log.WithContext(ctx)

	docsnap, err := s.client.Collection(projectsCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			logger.Debug("Firestore GetProject: not found: ", id)
			return nil, false, nil
		}
		logger.Error("Firestore GetProject: error retrieving project: ", err)
		return nil, false, fmt.Errorf("error retrieving project")
	}

	p, err := projectFromDocsnap(docsnap)
	if err != nil {
		logger.Error("Firestore GetProject: error coercing retrieved project to models.Project: ", err)
		return nil, true, fmt.Errorf("error retrieving project")
	}
	return p, true, nil
}

func docIds(docs []*firestore.DocumentSnapshot) []string {
	ids := make([]string, len(docs))
	for i, d := range docs {
		ids[i] = d.Ref.ID
	}
	return ids
}

// GetPlanRevision returns the plan revision with the specified ID or false if it can't be found
func (s *FireStore) GetPlanRevision(ctx context.Context, planID string, revID string) (*lib.Plan, bool, error) {
	logger := log.WithContext(ctx)

	docsnap, err := s.client.Collection(planRevisionsPath(planID)).Doc(revID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			logger.Debug("Firestore GetPlanRevision: not found: ", planID, revID)
			return nil, true, nil
		}
		logger.Error("Firestore GetPlanRevision: error retrieving plan revision: ", err)
		return nil, false, fmt.Errorf("error retrieving plan revision")
	}

	spr := new(storedPlanRevision)
	err = docsnap.DataTo(&spr)
	if err != nil {
		logger.Error("FireStore.GetPlans: error coercing retrieved plan to storedPlanRevision", err)
		return nil, true, fmt.Errorf("error whilst processing plan")
	}
	return spr.Plan, true, nil
}

// GetPlan returns the plan with the specified ID or false if it can't be found
func (s *FireStore) GetPlan(ctx context.Context, id string) (*models.Plan, bool, error) {
	logger := log.WithContext(ctx)

	docsnaps, err := s.client.Collection(planRevisionsPath(id)).
		OrderBy("Version.Time", firestore.Desc).Limit(1).
		Documents(ctx).GetAll()
	if err != nil || len(docsnaps) == 0 {
		if len(docsnaps) == 0 || (status.Code(err) == codes.NotFound) {
			logger.Debug("Firestore GetPlan: not found: ", id)
			return nil, false, nil
		}
		logger.Error("Firestore GetPlan: error retrieving plan: ", err)
		return nil, false, fmt.Errorf("error retrieving plan")
	}
	docsnap := docsnaps[0]

	spr := new(storedPlanRevision)
	err = docsnap.DataTo(&spr)
	if err != nil {
		logger.Error("FireStore.GetPlan: error coercing retrieved plan to lib.PlanDetails", err)
		return nil, true, fmt.Errorf("error whilst processing plan")
	}
	p := models.Plan{ID: id, Attributes: &spr.Plan.Details}
	return &p, true, nil
}

// delete deletes the specified document in the collection, error messages use the provided name
func (s *FireStore) delete(ctx context.Context, name string, collection string, id string) error {
	logger := log.WithContext(ctx)

	_, err := s.client.Collection(collection).Doc(id).Delete(ctx)
	if err != nil {
		logger.WithFields(log.Fields{name: id, "error": err}).Warnf("Firestore Delete: couldn't delete %v", name)
		return fmt.Errorf("error deleting %v", name)
	}
	logger.WithFields(log.Fields{name: id}).Infof("Deleted %v", name)
	return nil
}

// DeleteProject deletes the specified project
func (s *FireStore) DeleteProject(ctx context.Context, id string) error {
	// What are the semantics around dangling Project IDs in plans, especially if the latest revision has no existing project?
	// Such projectless plans probably need to be tidied up by admins, and the client shouldn't allow a project deletion
	// if the project has any plans.
	return s.delete(ctx, "project", projectsCollection, id)
}

// DeletePlan atomically deletes the specified plan, all of its revisions, and references in any projects
func (s *FireStore) DeletePlan(ctx context.Context, id string) error {
	logger := log.WithContext(ctx)
	op := s.client.Batch()

	// Remove plan from associated projects
	prevRev, found, err := s.GetPlan(ctx, id)
	if err != nil || !found {
		logger.WithFields(log.Fields{"plan": id, "error": err}).Error("FireStore DeletePlan: couldn't retrieve previous revision of plan")
		return fmt.Errorf("error whilst deleting plan - couldn't get projects associated with plan")
	}
	for _, prev := range prevRev.Attributes.Projects {
		uOp := updateOp{}
		err = s.updateProjectPlans(ctx, prev, id, true, &uOp)
		if err != nil {
			logger.WithFields(log.Fields{"plan": id, "project": prev, "error": err}).Error("FireStore DeletePlan: error removing plan from project")
			return fmt.Errorf("error whilst deleting plan - couldn't remove plan from associated projects")
		}
		op.Update(uOp.doc, uOp.updates)
	}

	// Delete the plan revisions
	docs, err := s.client.Collection(planRevisionsPath(id)).Documents(ctx).GetAll()
	if err != nil {
		logger.WithFields(log.Fields{"plan": id, "error": err}).Error("Firestore DeletePlan: couldn't list associated revisions for plan")
		return fmt.Errorf("error deleting plan %v - no revisions found", id)
	}
	for _, rev := range docs {
		op.Delete(rev.Ref)
	}

	// Delete the plan itself
	op.Delete(s.client.Doc(plansCollection + "/" + id))

	_, err = op.Commit(ctx)
	if err != nil {
		logger.WithFields(log.Fields{"error": err}).Error("Firestore DeletePlanAndRevisions: couldn't commit operation")
		return fmt.Errorf("error deleting plan %v", id)
	}
	return nil
}

// create creates a document in the collection and returns its ID
func (s *FireStore) create(ctx context.Context, name string, collection string, data interface{}) (string, error) {
	logger := log.WithContext(ctx)

	ref, _, err := s.client.Collection(collection).Add(ctx, data)
	if err != nil {
		logger.WithField("error", err).Errorf("Firestore: couldn't create %v", name)
		return "", fmt.Errorf("error creating %v", name)
	}
	logger.WithFields(log.Fields{name: ref.ID}).Infof("Created %v", name)
	return ref.ID, nil
}

// CreateProject creates a project and returns its new id
func (s *FireStore) CreateProject(ctx context.Context, p *models.ProjectDetails) (string, error) {
	sp := storedProject{Details: p, Plans: []string{}}
	return s.create(ctx, "project", projectsCollection, sp)
}

// CreatePlanRevision creates a new revision for the plan, returning its ID
func (s *FireStore) CreatePlanRevision(ctx context.Context, id string, p *lib.Plan, user *models.User) (string, error) {
	return s.createPlanRevSyncProjects(ctx, id, p, user, false)
}

// CreatePlan creates a plan from the plan and plan revision, and returns its new id and revision ID
func (s *FireStore) CreatePlan(ctx context.Context, p *lib.Plan, user *models.User) (id string, revID string, err error) {
	logger := log.WithContext(ctx)

	id, err = s.create(ctx, "plan", plansCollection, map[string]interface{}{}) // the plan has no fields
	if err != nil {
		return "", "", err
	}
	revID, err = s.createPlanRevSyncProjects(ctx, id, p, user, true)
	if err != nil {
		// This should ideally be a transaction to avoid this scenario
		logger.WithFields(log.Fields{"plan": id, "error": err}).Error("FireStore CreatePlan: error whilst creating revision for newly created plan")
		err = s.DeletePlan(ctx, id)
		if err != nil {
			logger.WithFields(log.Fields{"plan": id, "error": err}).Error("FireStore CreatePlan: error deleting newly created plan (trying to tidy up after revision creation error)")
		}
		return "", "", fmt.Errorf("Error creating plan")
	}
	return id, revID, err
}

// if field=="" update replaces collection.id with data, otherwise replace collection.id.field with data
func (s *FireStore) update(ctx context.Context, name string, collection string, doc string, field string, data interface{}) error {
	logger := log.WithContext(ctx).WithFields(log.Fields{name: doc, "field": field})

	docref := s.client.Collection(collection).Doc(doc)
	var err error
	if field == "" {
		_, err = docref.Set(ctx, data)
	} else {
		_, err = docref.Update(ctx, []firestore.Update{{Path: field, Value: data}})
	}
	if err != nil {
		logger.WithField("error", err).Warnf("Firestore: couldn't update %v", name)
		return fmt.Errorf("error updating %v", name)
	}
	logger.Infof("Updated %v", name)
	return nil
}

// UpdateProject replaces a project with with the contents of the project struct (ctx context.Contextthe ID in the struct is ignored)
func (s *FireStore) UpdateProject(ctx context.Context, id string, p *models.ProjectDetails) error {
	return s.update(ctx, "project", projectsCollection, id, "Details", p)
}

// createPlanRevSyncProjects creates a new revision for the plan, returning its ID.
func (s *FireStore) createPlanRevSyncProjects(ctx context.Context, id string, p *lib.Plan, user *models.User, firstRev bool) (string, error) {
	logger := log.WithContext(ctx).WithFields(log.Fields{"plan": id})

	// The projects previously associated with this plan may no longer be, we need to keep track
	prevProjects := []string{}
	if !firstRev {
		prevRev, found, err := s.GetPlan(ctx, id)
		if err != nil || !found {
			logger.WithField("error", err).Error("FireStore createPlanRevSyncProjects: couldn't retrieve previous revision of plan")
			return "", fmt.Errorf("error whilst creating plan - couldn't get previous revision")
		}
		prevProjects = prevRev.Attributes.Projects
	}

	// If multiple requests come in simultaneously we can have multiple revisions with the same timestamp.
	// That's OK, because revisions have unique IDs anyway.
	version := storedVersion{Author: &models.VersionAuthor{UID: &user.UID, Name: &user.Name, PictureURL: user.PictureURL}, Time: time.Now().UTC()}
	spr := storedPlanRevision{Version: &version, Plan: p}

	revID, err := s.create(ctx, "plan revision", planRevisionsPath(id), spr)
	if err != nil {
		logger.WithField("error", err).Error("Firestore: couldn't create plan revision")
		return "", fmt.Errorf("error creating plan revision")
	}

	// delete references for removed projects
	for _, prev := range prevProjects {
		removed := true
		for _, curr := range p.Details.Projects {
			removed = removed && (prev != curr)
		}
		if removed {
			err = s.updateProjectPlans(ctx, prev, id, true, nil)
			if err != nil {
				// We swallow the error as otherwise we need to delete the revision
				logger.WithFields(log.Fields{"project": prev, "error": err}).Error("FireStore createPlanRevSyncProjects: error removing plan from project (no rollback - plan revision created anyway)")
			}
		}
	}

	// add references for new projects
	// delete references for removed projects
	for _, curr := range p.Details.Projects {
		added := true
		for _, prev := range prevProjects {
			added = added && (prev != curr)
		}
		if added {
			err = s.updateProjectPlans(ctx, curr, id, false, nil)
			if err != nil {
				// We swallow the error as otherwise we need to delete the revision
				logger.WithFields(log.Fields{"project": curr, "error": err}).Error("FireStore createPlanRevSyncProjects: error adding plan to project (no rollback - plan revision created anyway)")
			}
		}
	}

	return revID, nil
}

// Add or remove planID from projectID's plans. If uOp is non-nil, populate it with the info required for a batch update rather than updating directly.
func (s *FireStore) updateProjectPlans(ctx context.Context, projectID string, planID string, remove bool, uOp *updateOp) error {
	logger := log.WithContext(ctx).WithFields(log.Fields{"project": projectID, "plan": planID})

	if remove {
		logger.Debug("Removing plan from project")
	} else {
		logger.Debug("Adding plan to project")
	}

	doc := s.client.Collection(projectsCollection).Doc(projectID)
	docsnap, err := doc.Get(ctx)
	if err != nil {
		return err
	}
	project, err := projectFromDocsnap(docsnap)
	if err != nil {
		return fmt.Errorf("Firestore updateProjectPlans: Failed to retrieve project: %v", err)
	}

	newPlans := project.Plans
	found := false
	for i, p := range project.Plans {
		if p == planID {
			if remove {
				newPlans = append(project.Plans[:i], project.Plans[i+1:]...)
			}
			found = true
			break
		}
	}
	if !remove && !found {
		newPlans = append(project.Plans, planID)
	}

	updates := []firestore.Update{{Path: "Plans", Value: newPlans}}
	if uOp == nil {
		_, err = doc.Update(ctx, updates)
	} else {
		uOp.doc = doc
		uOp.updates = updates
	}
	return err
}

// ListPlanRevisionIDs returns the revision ids of the specified plan, in date order earliest to latest or an error if it can't be found
func (s *FireStore) ListPlanRevisionIDs(ctx context.Context, id string) ([]string, error) {
	logger := log.WithContext(ctx)

	docs, err := s.client.Collection(planRevisionsPath(id)).
		OrderBy("Version.Time", firestore.Asc).
		Documents(ctx).GetAll()
	if err != nil {
		logger.WithFields(log.Fields{"plan": id, "error": err}).Warn("Firestore GetPlanRevisionIDs: error retrieving plan revisions")
		return nil, fmt.Errorf("error retrieving plan revisions")
	}

	return docIds(docs), nil
}

// GetPlanVersions returns the versions of the specified plan, in date order earliest to latest
func (s *FireStore) GetPlanVersions(ctx context.Context, id string) ([]*models.RevisionVersion, error) {
	logger := log.WithContext(ctx)

	docs, err := s.client.Collection(planRevisionsPath(id)).
		OrderBy("Version.Time", firestore.Asc).
		Select("Version").
		Documents(ctx).GetAll()
	if err != nil {
		logger.WithFields(log.Fields{"plan": id, "error": err}).Warn("Firestore GetPlanVersionIDs: error retrieving plan versions")
		return nil, fmt.Errorf("error retrieving versions for plan")
	}

	rvs := make([]*models.RevisionVersion, len(docs))
	for i, docsnap := range docs {
		sv := new(struct{ Version storedVersion })
		err = docsnap.DataTo(sv) // only the Version field
		if err != nil {
			logger.Error("FireStore.GetPlanVersions: error coercing retrieved doc to models.RevisionVersion", err)
			return nil, fmt.Errorf("error whilst processing plan versions")
		}
		v := &models.Version{Author: sv.Version.Author, Time: strfmt.DateTime(sv.Version.Time)}
		rvs[i] = &models.RevisionVersion{Version: v, PlanID: id, RevID: &docsnap.Ref.ID}
	}

	return rvs, nil
}

// GetUserData extends the referenced user with any additional data recorded in the store
func (s *FireStore) GetUserData(ctx context.Context, user *models.User) error {
	logger := log.WithContext(ctx).WithFields(log.Fields{"UID": user.UID})

	doc, err := s.client.Collection(usersCollection).Doc(user.UID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			logger.Debug("Firestore GetUserData: user not found")
			return nil
		}
		logger.WithField("error", err).Warn("Firestore GetUserData: error retrieving user")
		return fmt.Errorf("error getting local user data")
	}

	user.LookedUp = true
	if !doc.Exists() {
		user.LocalData = nil
		return nil
	}

	l := models.LocalUserData{}
	err = doc.DataTo(&l)
	if err != nil {
		logger.WithFields(log.Fields{"error": err}).Error("Firestore GetUserData: user document not in expected format - failed to coerce to models.LocalUserData")
		return fmt.Errorf("error getting local user data")
	}
	user.LocalData = &l

	// the database value is more current than a value set from a claim, so it doesn't matter what these were previously set to
	user.ManuallyAuthorized = user.LocalData.ManuallyAuthorized
	user.CreationAlertSent = user.LocalData.CreationAlertSent

	return nil
}

// SaveUserData records the user's LocalData, or removes it if user.LocalData==nil
func (s *FireStore) SaveUserData(ctx context.Context, user *models.User) error {
	logger := log.WithContext(ctx).WithFields(log.Fields{"user": user.UID})

	docref := s.client.Collection(usersCollection).Doc(user.UID)
	var err error

	if user.LocalData == nil {
		_, err = docref.Delete(ctx)
	} else {
		_, err = docref.Set(ctx, user.LocalData)
	}

	if err != nil {
		logger.WithFields(log.Fields{"error": err}).Error("Firestore SaveUserData: failed to update/create/delete record")
		return fmt.Errorf("error saving user data")
	}

	return nil
}

// SetManuallyAuthorized sets this user's ManuallyAuthorized attribute
func (s *FireStore) SetManuallyAuthorized(ctx context.Context, UID string, value bool) error {
	logger := log.WithContext(ctx).WithFields(log.Fields{"user": UID})

	docref := s.client.Collection(usersCollection).Doc(UID)
	_, err := docref.Set(ctx, map[string]interface{}{"ManuallyAuthorized": value}, firestore.MergeAll)

	if err != nil {
		logger.WithFields(log.Fields{"error": err}).Error("Firestore ManuallyAuthorized: failed to update/create record")
		return fmt.Errorf("error recording user manually authorized attribute")
	}

	return nil
}

// UserCreationAlertSent sets this user's CreationAlertSent to true
func (s *FireStore) UserCreationAlertSent(ctx context.Context, UID string) error {
	logger := log.WithContext(ctx).WithFields(log.Fields{"user": UID})

	docref := s.client.Collection(usersCollection).Doc(UID)
	_, err := docref.Set(ctx, map[string]interface{}{"CreationAlertSent": true}, firestore.MergeAll)

	if err != nil {
		logger.WithFields(log.Fields{"error": err}).Error("Firestore UserCreationAlertSent: failed to update/create record")
		return fmt.Errorf("error recording user creation alert sent")
	}

	return nil
}

// GetConfigString returns the named configuration string
func (s *FireStore) GetConfigString(ctx context.Context, field string) (string, error) {
	logger := log.WithContext(ctx)

	docsnap, err := s.client.Doc(configDoc).Get(ctx)
	if err != nil {
		logger.Error("Firestore GetConfigString: error retrieving config: ", err)
		return "", fmt.Errorf("error retrieving config")
	}
	v, ok := docsnap.DataAt(field)
	if ok != nil {
		return "", fmt.Errorf("%v not found in config", field)
	}

	vs, isString := v.(string)
	if !isString {
		return "", fmt.Errorf("%v is not a string", field)
	}
	return vs, nil
}

// ListPracticesVersions lists all of the recorded versions of practice definitions, in lexicographic order
func (s *FireStore) ListPracticesVersions(ctx context.Context) ([]string, error) {
	logger := log.WithContext(ctx)

	docrefs, err := s.client.Collection(practicesCollection).DocumentRefs(ctx).GetAll()
	if err != nil {
		logger.Errorf("Firestore ListPracticesVersions: error listing versions: %v", err)
		return nil, fmt.Errorf("error listing versions")
	}

	versions := []string{}
	for _, doc := range docrefs {
		versions = append(versions, doc.ID)
	}
	sort.Strings(versions)

	return versions, nil
}

// GetPractices retrieves the specified version of the practice definitions
func (s *FireStore) GetPractices(ctx context.Context, version string) ([]lib.Practice, error) {
	logger := log.WithContext(ctx)
	docsnap, err := s.client.Collection(practicesCollection).Doc(version).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			logger.Infof("Firestore GetPractices: couldn't find practices version: %v", version)
		} else {
			logger.Errorf("Firestore GetPractices: error retrieving practices: %v", err)
		}
		return nil, fmt.Errorf("error retrieving practices")
	}

	practices := &storedPractices{}
	err = docsnap.DataTo(practices)
	if err != nil {
		logger.Errorf("Firestore GetPractices: error coercing retrieved practices to {practices: []lib.Practice}: %v", err)
		return nil, fmt.Errorf("error retrieving practices")
	}

	return practices.Practices, nil
}

// CreatePractices creates or replaces the practices at the specified version
func (s *FireStore) CreatePractices(ctx context.Context, version string, practices []lib.Practice) error {
	docref := s.client.Collection(practicesCollection).Doc(version)
	_, err := docref.Set(ctx, storedPractices{practices})
	if err != nil {
		return fmt.Errorf("CreatePractices: %v", err)
	}

	return nil
}

// DeletePractices removes the practices at the specified version. This will break any plans that used this version!
func (s *FireStore) DeletePractices(ctx context.Context, version string) error {
	docref := s.client.Collection(practicesCollection).Doc(version)
	_, err := docref.Delete(ctx)
	if err != nil {
		return fmt.Errorf("DeletePractices: %v", err)
	}

	return nil
}
