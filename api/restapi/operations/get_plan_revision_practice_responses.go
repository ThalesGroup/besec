// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/ThalesGroup/besec/api/models"
)

// GetPlanRevisionPracticeResponsesHandlerFunc turns a function with the right signature into a get plan revision practice responses handler
type GetPlanRevisionPracticeResponsesHandlerFunc func(GetPlanRevisionPracticeResponsesParams, *models.User) middleware.Responder

// Handle executing the request and returning a response
func (fn GetPlanRevisionPracticeResponsesHandlerFunc) Handle(params GetPlanRevisionPracticeResponsesParams, principal *models.User) middleware.Responder {
	return fn(params, principal)
}

// GetPlanRevisionPracticeResponsesHandler interface for that can handle valid get plan revision practice responses params
type GetPlanRevisionPracticeResponsesHandler interface {
	Handle(GetPlanRevisionPracticeResponsesParams, *models.User) middleware.Responder
}

// NewGetPlanRevisionPracticeResponses creates a new http.Handler for the get plan revision practice responses operation
func NewGetPlanRevisionPracticeResponses(ctx *middleware.Context, handler GetPlanRevisionPracticeResponsesHandler) *GetPlanRevisionPracticeResponses {
	return &GetPlanRevisionPracticeResponses{Context: ctx, Handler: handler}
}

/* GetPlanRevisionPracticeResponses swagger:route GET /plan/{id}/revision/{revId}/responses getPlanRevisionPracticeResponses

GetPlanRevisionPracticeResponses get plan revision practice responses API

*/
type GetPlanRevisionPracticeResponses struct {
	Context *middleware.Context
	Handler GetPlanRevisionPracticeResponsesHandler
}

func (o *GetPlanRevisionPracticeResponses) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewGetPlanRevisionPracticeResponsesParams()
	uprinc, aCtx, err := o.Context.Authorize(r, route)
	if err != nil {
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}
	if aCtx != nil {
		*r = *aCtx
	}
	var principal *models.User
	if uprinc != nil {
		principal = uprinc.(*models.User) // this is really a models.User, I promise
	}

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params, principal) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
