package api

import (
	"github.com/go-openapi/runtime/middleware"

	"github.com/ThalesGroup/besec/api/models"
	"github.com/ThalesGroup/besec/api/restapi/operations"
	"github.com/ThalesGroup/besec/lib"
)

// NewListPracticesVersionsHandler creates a handler
func NewListPracticesVersionsHandler(rt *Runtime) operations.ListPracticesVersionsHandler {
	return &listPracticesVersionsHandlerImp{rt: rt}
}

type listPracticesVersionsHandlerImp struct {
	rt *Runtime
}

func (h *listPracticesVersionsHandlerImp) Handle(params operations.ListPracticesVersionsParams, principal *models.User) middleware.Responder {
	ctx := params.HTTPRequest.Context()
	versions, err := h.rt.Store.ListPracticesVersions(ctx)
	if err != nil {
		r := operations.ListPracticesVersionsDefault{}
		msg := "Error retrieving versions"
		return r.WithStatusCode(500).WithPayload(&models.Error{Message: &msg})
	}
	return &operations.ListPracticesVersionsOK{Payload: versions}
}

// NewGetPracticesHandler creates a handler
func NewGetPracticesHandler(rt *Runtime) operations.GetPracticesHandler {
	return &getPracticesHandlerImp{rt: rt}
}

type getPracticesHandlerImp struct {
	rt *Runtime
}

func (h *getPracticesHandlerImp) Handle(params operations.GetPracticesParams, principal *models.User) middleware.Responder {
	fail := func(code int, msg string) middleware.Responder {
		r := operations.ListPracticesVersionsDefault{}
		return r.WithStatusCode(code).WithPayload(&models.Error{Message: &msg})
	}

	ctx := params.HTTPRequest.Context()

	version := params.Version
	if version == "latest" {
		versions, err := h.rt.Store.ListPracticesVersions(ctx)
		if err != nil {
			return fail(500, "Error retrieving versions")
		}
		if len(versions) == 0 {
			return fail(404, "No versions found!")
		}
		version = versions[len(versions)-1]
	}

	practices, err := h.rt.GetPractices(ctx, version)
	if err != nil {
		return fail(404, "Practices version not found")
	}
	practiceMap := map[string]lib.Practice{}
	for _, practice := range practices {
		practiceMap[practice.ID] = practice
	}
	return &operations.GetPracticesOK{Payload: &models.GotPractices{Version: &version, Practices: practiceMap}}
}
