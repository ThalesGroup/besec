package api

import (
	"context"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/go-openapi/loads"
	log "github.com/sirupsen/logrus"

	"github.com/ThalesGroup/besec/api/restapi"
	"github.com/ThalesGroup/besec/api/restapi/operations"
	"github.com/ThalesGroup/besec/lib"
	"github.com/ThalesGroup/besec/store"
)

// Runtime captures the shared state & services used by the server
type Runtime struct {
	practicesCache      map[string]practiceCache // keyed on version
	Store               store.Store
	AuthClient          *auth.Client
	AuthConfig          ExtendedAuthConfig
	RequestAccessAlerts bool // Whether to send notifications to admins when a new unauthorized user attempts to login
	NewUserAlerts       bool // Whether to send notifications to admins when an authorized user signs in for the first time
	SlackChan           chan SlackMessage
	PublicPaths         map[string]map[string]bool // map from path to a map from HTTP method to whether it is public
}

type practiceCache struct {
	practices []lib.Practice
	updated   time.Time
}

// In theory practices shouldn't change after they've been published, in practice they will.
// This limits the pain when that happens if we have a long-running process.
const practiceCacheLength = "2m"

// NewRuntime creates a Runtime with the given parameters
func NewRuntime(Store store.Store,
	AuthClient *auth.Client,
	AuthConfig ExtendedAuthConfig,
	RequestAccessAlerts bool,
	NewUserAlerts bool,
	SlackChan chan SlackMessage,
) *Runtime {
	return &Runtime{
		practicesCache:      map[string]practiceCache{},
		Store:               Store,
		AuthClient:          AuthClient,
		AuthConfig:          AuthConfig,
		RequestAccessAlerts: RequestAccessAlerts,
		NewUserAlerts:       NewUserAlerts,
		SlackChan:           SlackChan,
	}
}

// NewAPI creates a new configured BeSec API server
func NewAPI(rt *Runtime) *restapi.Server {
	swaggerSpec, err := loads.Analyzed(restapi.FlatSwaggerJSON, "") // we use FlatSwaggerJSON to avoid references to the file system
	if err != nil {
		log.Fatalln(err)
	}

	API := operations.NewBesecAPI(swaggerSpec)
	API.LoggedInHandler = NewLoggedInHandler(rt)
	API.GetAuthConfigHandler = NewGetAuthConfigHandler(rt)

	API.ListPracticesVersionsHandler = NewListPracticesVersionsHandler(rt)
	API.GetPracticesHandler = NewGetPracticesHandler(rt)

	API.ListProjectsHandler = NewListProjectsHandler(rt)
	API.GetProjectHandler = NewGetProjectHandler(rt)
	API.CreateProjectHandler = NewCreateProjectHandler(rt)
	API.UpdateProjectHandler = NewUpdateProjectHandler(rt)
	API.DeleteProjectHandler = NewDeleteProjectHandler(rt)

	API.GetPlanHandler = NewGetPlanHandler(rt)
	API.CreatePlanHandler = NewCreatePlanHandler(rt)
	API.DeletePlanHandler = NewDeletePlanHandler(rt)
	API.CreatePlanRevisionHandler = NewCreatePlanRevisionHandler(rt)
	API.GetPlanVersionsHandler = NewGetPlanVersionsHandler(rt)
	API.GetPlanRevisionHandler = NewGetPlanRevisionHandler(rt)
	API.GetPlanRevisionPracticeResponsesHandler = NewGetPlanRevisionPracticeResponsesHandler(rt)

	API.Logger = log.Infof
	if rt.AuthClient == nil {
		API.KeyAuth = MakeDummyKeyAuth(rt)
	} else {
		API.KeyAuth = MakeKeyAuth(rt)
	}
	API.APIAuthorizer = MakeAuthorizer(rt)

	apiSrv := restapi.NewServer(API)
	apiSrv.ConfigureAPI()
	return apiSrv
}

// GetPractices is a caching version of Store.GetPractices
func (rt *Runtime) GetPractices(ctx context.Context, version string) ([]lib.Practice, error) {
	cached, ok := rt.practicesCache[version]
	if ok {
		limit, _ := time.ParseDuration(practiceCacheLength)
		if time.Since(cached.updated) < limit {
			return cached.practices, nil
		}
		log.WithContext(ctx).WithFields(log.Fields{"version": version}).Debug("Cache expired")
	}
	practices, err := rt.Store.GetPractices(ctx, version)
	if err != nil {
		return practices, err
	}
	rt.practicesCache[version] = practiceCache{practices: practices, updated: time.Now()}
	return practices, nil
}
