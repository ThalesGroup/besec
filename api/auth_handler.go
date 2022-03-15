package api

import (
	"github.com/go-openapi/runtime/middleware"

	"github.com/ThalesGroup/besec/api/models"
	"github.com/ThalesGroup/besec/api/restapi/operations"
)

// NewLoggedInHandler creates a handler
func NewLoggedInHandler(rt *Runtime) operations.LoggedInHandler {
	return &loggedInHandlerImp{rt: rt}
}

type loggedInHandlerImp struct {
	rt *Runtime
}

// Handle does nothing, it just serves as an initial API endpoint to hit to trigger on-login tasks - see the authorizer for details
func (h *loggedInHandlerImp) Handle(params operations.LoggedInParams, principal *models.User) middleware.Responder {
	return &operations.LoggedInOK{}
}

// NewGetAuthConfigHandler creates a handler
func NewGetAuthConfigHandler(rt *Runtime) operations.GetAuthConfigHandler {
	return &getAuthConfigHandlerImp{rt: rt}
}

type getAuthConfigHandlerImp struct {
	rt *Runtime
}

// Handler returns the auth config
func (h *getAuthConfigHandlerImp) Handle(params operations.GetAuthConfigParams) middleware.Responder {
	return &operations.GetAuthConfigOK{Payload: &h.rt.AuthConfig.AuthConfig}
}
