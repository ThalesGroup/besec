package api

import (
	"context"

	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/ThalesGroup/besec/api/models"
	"github.com/ThalesGroup/besec/api/restapi/operations"
)

// NewListProjectsHandler creates a handler
func NewListProjectsHandler(rt *Runtime) operations.ListProjectsHandler {
	return &listProjectsHandlerImp{rt: rt}
}

type listProjectsHandlerImp struct {
	rt *Runtime
}

func (h *listProjectsHandlerImp) Handle(params operations.ListProjectsParams, principal *models.User) middleware.Responder {
	projects, err := h.rt.Store.ListProjects(params.HTTPRequest.Context())
	if err != nil {
		r := operations.GetProjectDefault{}
		msg := err.Error()
		return r.WithStatusCode(500).WithPayload(&models.Error{Message: &msg})
	}
	return &operations.ListProjectsOK{Payload: projects}
}

// NewGetProjectHandler creates a handler
func NewGetProjectHandler(rt *Runtime) operations.GetProjectHandler {
	return &getProjectHandlerImp{rt: rt}
}

type getProjectHandlerImp struct {
	rt *Runtime
}

func (h *getProjectHandlerImp) Handle(params operations.GetProjectParams, principal *models.User) middleware.Responder {
	fail := func(code int, msg string) middleware.Responder {
		r := operations.GetProjectDefault{}
		return r.WithStatusCode(code).WithPayload(&models.Error{Message: &msg})
	}
	proj, found, err := h.rt.Store.GetProject(params.HTTPRequest.Context(), params.ID)
	if err != nil {
		return fail(500, "error retrieving project")
	}
	if !found {
		return fail(404, "project not found")
	}
	return &operations.GetProjectOK{Payload: proj}
}

// NewCreateProjectHandler creates a handler
func NewCreateProjectHandler(rt *Runtime) operations.CreateProjectHandler {
	return &createProjectHandlerImp{rt: rt}
}

type createProjectHandlerImp struct {
	rt *Runtime
}

func isUniqueName(ctx context.Context, name *string, rt *Runtime) (bool, error) {
	ps, err := rt.Store.ListProjects(ctx)
	if err != nil {
		return false, err
	}
	for _, p := range ps {
		if *(p.Attributes.Name) == *name {
			return false, nil
		}
	}
	return true, nil
}

func (h *createProjectHandlerImp) Handle(params operations.CreateProjectParams, principal *models.User) middleware.Responder {
	fail := func(code int, msg string) middleware.Responder {
		r := operations.CreateProjectDefault{}
		return r.WithStatusCode(code).WithPayload(&models.Error{Message: &msg})
	}

	ctx := params.HTTPRequest.Context()
	logger := log.WithContext(ctx)

	unique, err := isUniqueName(ctx, params.Body.Name, h.rt)
	if err != nil {
		logger.Error("CreateProject Handler: error testing project uniqueness: ", err)
		return fail(500, "error creating project")
	}
	if !unique {
		return fail(400, "project names must be unique")
	}

	id, err := h.rt.Store.CreateProject(ctx, params.Body)
	if err != nil {
		return fail(500, err.Error())
	}
	return &operations.CreateProjectCreated{Payload: id}
}

// NewUpdateProjectHandler creates a handler
func NewUpdateProjectHandler(rt *Runtime) operations.UpdateProjectHandler {
	return &updateProjectHandlerImp{rt: rt}
}

type updateProjectHandlerImp struct {
	rt *Runtime
}

func (h *updateProjectHandlerImp) Handle(params operations.UpdateProjectParams, principal *models.User) middleware.Responder {
	fail := func(code int, msg string) middleware.Responder {
		r := operations.UpdateProjectDefault{}
		return r.WithStatusCode(code).WithPayload(&models.Error{Message: &msg})
	}

	ctx := params.HTTPRequest.Context()
	logger := log.WithContext(ctx)

	orig, found, err := h.rt.Store.GetProject(ctx, params.ID)
	if err != nil {
		return fail(500, "error retrieving project")
	}
	if !found {
		return fail(404, "project "+params.ID+" doesn't exist")
	}

	if *params.Body.Name != *orig.Attributes.Name {
		if unique, nameErr := isUniqueName(ctx, params.Body.Name, h.rt); nameErr != nil {
			logger.Error("UpdateProject Handler: when testing project name uniqueness: ", nameErr)
			return fail(500, "error whilst creating project")
		} else if !unique {
			return fail(400, "project names must be unique")
		}
	}

	if err = h.rt.Store.UpdateProject(params.HTTPRequest.Context(), params.ID, params.Body); err != nil {
		return fail(500, err.Error())
	}
	return &operations.UpdateProjectOK{}
}

// NewDeleteProjectHandler creates a handler
func NewDeleteProjectHandler(rt *Runtime) operations.DeleteProjectHandler {
	return &deleteProjectHandlerImp{rt: rt}
}

type deleteProjectHandlerImp struct {
	rt *Runtime
}

func (h *deleteProjectHandlerImp) Handle(params operations.DeleteProjectParams, principal *models.User) middleware.Responder {
	fail := func(code int, msg string) middleware.Responder {
		r := operations.DeleteProjectDefault{}
		return r.WithStatusCode(code).WithPayload(&models.Error{Message: &msg})
	}

	// Check it exists first, as delete succeeds even if it doesn't exist
	_, found, err := h.rt.Store.GetProject(params.HTTPRequest.Context(), params.ID)
	if err != nil {
		return fail(500, "error checking project exists")
	}
	if !found {
		return fail(404, "project "+params.ID+" doesn't exist")
	}
	if err := h.rt.Store.DeleteProject(params.HTTPRequest.Context(), params.ID); err != nil {
		return fail(500, err.Error())
	}
	return &operations.DeleteProjectNoContent{}
}
