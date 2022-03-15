package api

import (
	"context"
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/ThalesGroup/besec/api/models"
	"github.com/ThalesGroup/besec/api/restapi/operations"
	"github.com/ThalesGroup/besec/lib"
)

// NewGetPlanHandler creates a handler
func NewGetPlanHandler(rt *Runtime) operations.GetPlanHandler {
	return &getPlanHandlerImp{rt: rt}
}

type getPlanHandlerImp struct {
	rt *Runtime
}

func (h *getPlanHandlerImp) Handle(params operations.GetPlanParams, principal *models.User) middleware.Responder {
	fail := func(code int, msg string) middleware.Responder {
		r := operations.CreatePlanDefault{}
		return r.WithStatusCode(code).WithPayload(&models.Error{Message: &msg})
	}

	ctx := params.HTTPRequest.Context()
	logger := log.WithContext(ctx)

	plan, found, err := h.rt.Store.GetPlan(ctx, params.ID)
	if err != nil {
		return fail(500, "error getting plan")
	}
	if !found {
		return fail(404, "plan not found")
	}
	revisions, err := h.rt.Store.ListPlanRevisionIDs(ctx, params.ID)
	if err != nil {
		logger.Warning("GetPlanHandler error whilst retrieving plan revisions for plan: ", params.ID, err)
		return fail(500, "couldn't retrieve revisions for plan")
	}
	if len(revisions) == 0 {
		logger.Warning("GetPlanHandler there are no revisions associated with plan ", params.ID)
		return fail(500, "couldn't retrieve revisions for plan")
	}
	body := operations.GetPlanOKBody{Plan: plan, LatestRevision: &revisions[len(revisions)-1]}
	return &operations.GetPlanOK{Payload: &body}
}

func makePlanFromReq(ctx context.Context, rt *Runtime, details *lib.PlanDetails, responses *lib.PlanResponses) (*lib.Plan, int, string) {
	practices, err := rt.GetPractices(ctx, responses.PracticesVersion)
	if err != nil {
		return nil, 404, "Couldn't find specified practices version '" + responses.PracticesVersion + "'"
	}
	plan := lib.NewPlan(*details, *responses, practices)

	// Now we have practice information for the plan, do validation again
	// The validation that go-swagger does prior to the handlers doesn't have this contextual information
	err = plan.Responses.Validate(nil)
	if err != nil {
		return nil, 400, err.Error()
	}

	return &plan, 0, ""
}

// NewCreatePlanHandler creates a handler
func NewCreatePlanHandler(rt *Runtime) operations.CreatePlanHandler {
	return &createPlanHandlerImp{rt: rt}
}

type createPlanHandlerImp struct {
	rt *Runtime
}

func (h *createPlanHandlerImp) Handle(params operations.CreatePlanParams, principal *models.User) middleware.Responder {
	fail := func(code int, msg string) middleware.Responder {
		r := operations.CreatePlanDefault{}
		return r.WithStatusCode(code).WithPayload(&models.Error{Message: &msg})
	}

	if params.Body.Details.Committed {
		ready, issues := params.Body.Responses.ReadyToCommit()
		if !ready {
			return fail(500, fmt.Sprintf("cannot commit plan: %v", issues))
		}
	}

	ctx := params.HTTPRequest.Context()

	plan, code, msg := makePlanFromReq(ctx, h.rt, params.Body.Details, params.Body.Responses)
	if code != 0 {
		return fail(code, msg)
	}

	id, revID, err := h.rt.Store.CreatePlan(ctx, plan, principal)
	if err != nil {
		return fail(500, err.Error())
	}
	response := operations.CreatePlanCreatedBody{PlanID: &id, RevisionID: &revID}
	return &operations.CreatePlanCreated{Payload: &response}
}

// NewDeletePlanHandler creates a handler
func NewDeletePlanHandler(rt *Runtime) operations.DeletePlanHandler {
	return &deletePlanHandlerImp{rt: rt}
}

type deletePlanHandlerImp struct {
	rt *Runtime
}

func (h *deletePlanHandlerImp) Handle(params operations.DeletePlanParams, principal *models.User) middleware.Responder {
	// We don't need to check if the plan exists first: if there are no associated revisions then the delete will fail
	if err := h.rt.Store.DeletePlan(params.HTTPRequest.Context(), params.ID); err != nil {
		r := operations.DeletePlanDefault{}
		msg := err.Error()
		return r.WithStatusCode(500).WithPayload(&models.Error{Message: &msg})
	}
	return &operations.DeletePlanNoContent{}
}

// NewCreatePlanRevisionHandler creates a handler
func NewCreatePlanRevisionHandler(rt *Runtime) operations.CreatePlanRevisionHandler {
	return &createPlanRevisionHandlerImp{rt: rt}
}

type createPlanRevisionHandlerImp struct {
	rt *Runtime
}

func (h *createPlanRevisionHandlerImp) Handle(params operations.CreatePlanRevisionParams, principal *models.User) middleware.Responder {
	fail := func(code int, msg string) middleware.Responder {
		r := operations.CreatePlanRevisionDefault{}
		return r.WithStatusCode(code).WithPayload(&models.Error{Message: &msg})
	}

	ctx := params.HTTPRequest.Context()

	// Check the plan exists
	_, found, err := h.rt.Store.GetPlan(ctx, params.ID)
	if err != nil {
		return fail(500, "error creating revision for plan "+params.ID)
	}
	if !found {
		return fail(404, "couldn't find plan "+params.ID)
	}

	if params.Body.Details.Committed {
		ready, issues := params.Body.Responses.ReadyToCommit()
		if !ready {
			return fail(500, fmt.Sprintf("cannot commit plan: %v", issues))
		}
	}

	plan, code, msg := makePlanFromReq(ctx, h.rt, params.Body.Details, params.Body.Responses)
	if code != 0 {
		return fail(code, msg)
	}

	revID, err := h.rt.Store.CreatePlanRevision(ctx, params.ID, plan, principal)
	if err != nil {
		return fail(500, err.Error())
	}
	return &operations.CreatePlanRevisionOK{Payload: revID}
}

// NewGetPlanVersionsHandler creates a handler
func NewGetPlanVersionsHandler(rt *Runtime) operations.GetPlanVersionsHandler {
	return &getPlanVersionsHandlerImp{rt: rt}
}

type getPlanVersionsHandlerImp struct {
	rt *Runtime
}

func (h *getPlanVersionsHandlerImp) Handle(params operations.GetPlanVersionsParams, principal *models.User) middleware.Responder {
	fail := func(code int, msg string) middleware.Responder {
		r := operations.CreatePlanDefault{}
		return r.WithStatusCode(code).WithPayload(&models.Error{Message: &msg})
	}

	ctx := params.HTTPRequest.Context()
	logger := log.WithContext(ctx)

	rvs, err := h.rt.Store.GetPlanVersions(ctx, params.ID)
	if err != nil {
		return fail(500, "error retrieving versions for plan")
	}
	if len(rvs) == 0 {
		logger.Debug("GetPlanVersionsHandler there are no revisions associated with plan ", params.ID)
		return fail(404, "plan not found")
	}
	return &operations.GetPlanVersionsOK{Payload: rvs}
}

// NewGetPlanRevisionHandler creates a handler
func NewGetPlanRevisionHandler(rt *Runtime) operations.GetPlanRevisionHandler {
	return &getPlanRevisionHandlerImp{rt: rt}
}

type getPlanRevisionHandlerImp struct {
	rt *Runtime
}

func (h *getPlanRevisionHandlerImp) Handle(params operations.GetPlanRevisionParams, principal *models.User) middleware.Responder {
	fail := func(code int, msg string) middleware.Responder {
		r := operations.CreatePlanDefault{}
		return r.WithStatusCode(code).WithPayload(&models.Error{Message: &msg})
	}
	p, found, err := h.rt.Store.GetPlanRevision(params.HTTPRequest.Context(), params.ID, params.RevID)
	if err != nil {
		return fail(500, "error retrieving plan")
	}
	if !found {
		return fail(404, "plan not found")
	}
	return &(operations.GetPlanRevisionOK{Payload: &p.Details})
}

// NewGetPlanRevisionPracticeResponsesHandler creates a handler
func NewGetPlanRevisionPracticeResponsesHandler(rt *Runtime) operations.GetPlanRevisionPracticeResponsesHandler {
	return &getPlanRevisionPracticeResponsesHandlerImp{rt: rt}
}

type getPlanRevisionPracticeResponsesHandlerImp struct {
	rt *Runtime
}

func (h *getPlanRevisionPracticeResponsesHandlerImp) Handle(params operations.GetPlanRevisionPracticeResponsesParams, principal *models.User) middleware.Responder {
	fail := func(code int, msg string) middleware.Responder {
		r := operations.CreatePlanDefault{}
		return r.WithStatusCode(code).WithPayload(&models.Error{Message: &msg})
	}

	p, found, err := h.rt.Store.GetPlanRevision(params.HTTPRequest.Context(), params.ID, params.RevID)
	if err != nil {
		return fail(500, "error retrieving plan")
	}
	if !found {
		return fail(404, "plan not found")
	}
	return &operations.GetPlanRevisionPracticeResponsesOK{Payload: &p.Responses}
}
