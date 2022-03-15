package store

import (
	"context"

	"github.com/ThalesGroup/besec/api/models"
	"github.com/ThalesGroup/besec/lib"
)

// Store abstracts over different possible persistence implementations
// (there used to be more than one)
type Store interface {
	// ListProjects returns all of the projects
	ListProjects(ctx context.Context) ([]*models.Project, error)
	// GetProject returns the project with the specified ID, or false if it can't be found
	GetProject(ctx context.Context, id string) (*models.Project, bool, error)
	// UpdateProject replaces a project with with the contents of the project struct (ctx context.Contextthe ID in the struct is ignored)
	UpdateProject(ctx context.Context, id string, p *models.ProjectDetails) error
	// CreateProject creates a project and returns its new id
	CreateProject(ctx context.Context, p *models.ProjectDetails) (string, error)
	// DeleteProject deletes the specified project
	DeleteProject(ctx context.Context, id string) error

	// GetPlan returns the plan with the specified ID and true, or (nil,false) if it can't be found
	GetPlan(ctx context.Context, id string) (*models.Plan, bool, error)
	// CreatePlan creates a plan and returns its new id and revision ID
	CreatePlan(ctx context.Context, p *lib.Plan, user *models.User) (id string, revID string, err error)
	// DeletePlanAndRevisions deletes the specified plan and all of its revisions
	DeletePlan(ctx context.Context, id string) error
	// CreatePlanRevision creates a new revision for the plan, returning its ID
	CreatePlanRevision(ctx context.Context, id string, p *lib.Plan, user *models.User) (string, error)
	// GetPlanRevision returns the plan revision with the specified ID or false if it can't be found
	GetPlanRevision(ctx context.Context, planID string, revID string) (*lib.Plan, bool, error)
	// ListPlanRevisionIDs returns the revision ids of the specified plan, in date order earliest to latest or an error if it can't be found
	ListPlanRevisionIDs(ctx context.Context, id string) ([]string, error)
	// GetPlanVersions returns the versions of the specified plan, in date order earliest to latest
	GetPlanVersions(ctx context.Context, id string) ([]*models.RevisionVersion, error)

	// GetUserData extends the referenced user with any additional data recorded in the store
	GetUserData(ctx context.Context, user *models.User) error
	// SaveUserData records the user's LocalData, or removes it if user.LocalData==nil
	SaveUserData(ctx context.Context, user *models.User) error
	// UserCreationAlertSent sets this user's CreationAlertSent to true
	UserCreationAlertSent(ctx context.Context, UID string) error
	// SetManuallyAuthorized sets this user's ManuallyAuthorized attribute
	SetManuallyAuthorized(ctx context.Context, UID string, value bool) error

	// GetConfigString returns the named configuration string
	GetConfigString(ctx context.Context, field string) (string, error)

	// ListPracticesVersions lists all of the recorded versions of practice definitions, in lexicographic order
	ListPracticesVersions(ctx context.Context) ([]string, error)
	// GetPractices retrieves the specified version of the practice definitions
	GetPractices(ctx context.Context, version string) ([]lib.Practice, error)
	// CreatePractices creates or replaces the practices at the specified version
	CreatePractices(ctx context.Context, version string, practices []lib.Practice) error
	// DeletePractices removes the practices at the specified version. This will break any plans that used this version!
	DeletePractices(ctx context.Context, version string) error
}
