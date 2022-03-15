// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"context"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"

	"github.com/ThalesGroup/besec/api/models"
	"github.com/ThalesGroup/besec/lib"
)

// CreatePlanRevisionHandlerFunc turns a function with the right signature into a create plan revision handler
type CreatePlanRevisionHandlerFunc func(CreatePlanRevisionParams, *models.User) middleware.Responder

// Handle executing the request and returning a response
func (fn CreatePlanRevisionHandlerFunc) Handle(params CreatePlanRevisionParams, principal *models.User) middleware.Responder {
	return fn(params, principal)
}

// CreatePlanRevisionHandler interface for that can handle valid create plan revision params
type CreatePlanRevisionHandler interface {
	Handle(CreatePlanRevisionParams, *models.User) middleware.Responder
}

// NewCreatePlanRevision creates a new http.Handler for the create plan revision operation
func NewCreatePlanRevision(ctx *middleware.Context, handler CreatePlanRevisionHandler) *CreatePlanRevision {
	return &CreatePlanRevision{Context: ctx, Handler: handler}
}

/* CreatePlanRevision swagger:route POST /plan/{id} createPlanRevision

CreatePlanRevision create plan revision API

*/
type CreatePlanRevision struct {
	Context *middleware.Context
	Handler CreatePlanRevisionHandler
}

func (o *CreatePlanRevision) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewCreatePlanRevisionParams()
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

// CreatePlanRevisionBody create plan revision body
//
// swagger:model CreatePlanRevisionBody
type CreatePlanRevisionBody struct {

	// details
	// Required: true
	Details *lib.PlanDetails `json:"details"`

	// responses
	// Required: true
	Responses *lib.PlanResponses `json:"responses"`
}

// Validate validates this create plan revision body
func (o *CreatePlanRevisionBody) Validate(formats strfmt.Registry) error {
	var res []error

	if err := o.validateDetails(formats); err != nil {
		res = append(res, err)
	}

	if err := o.validateResponses(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *CreatePlanRevisionBody) validateDetails(formats strfmt.Registry) error {

	if err := validate.Required("body"+"."+"details", "body", o.Details); err != nil {
		return err
	}

	if o.Details != nil {
		if err := o.Details.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("body" + "." + "details")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("body" + "." + "details")
			}
			return err
		}
	}

	return nil
}

func (o *CreatePlanRevisionBody) validateResponses(formats strfmt.Registry) error {

	if err := validate.Required("body"+"."+"responses", "body", o.Responses); err != nil {
		return err
	}

	if o.Responses != nil {
		if err := o.Responses.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("body" + "." + "responses")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("body" + "." + "responses")
			}
			return err
		}
	}

	return nil
}

// ContextValidate validate this create plan revision body based on the context it is used
func (o *CreatePlanRevisionBody) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := o.contextValidateDetails(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := o.contextValidateResponses(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (o *CreatePlanRevisionBody) contextValidateDetails(ctx context.Context, formats strfmt.Registry) error {

	if o.Details != nil {
		if err := o.Details.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("body" + "." + "details")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("body" + "." + "details")
			}
			return err
		}
	}

	return nil
}

func (o *CreatePlanRevisionBody) contextValidateResponses(ctx context.Context, formats strfmt.Registry) error {

	if o.Responses != nil {
		if err := o.Responses.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("body" + "." + "responses")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("body" + "." + "responses")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (o *CreatePlanRevisionBody) MarshalBinary() ([]byte, error) {
	if o == nil {
		return nil, nil
	}
	return swag.WriteJSON(o)
}

// UnmarshalBinary interface implementation
func (o *CreatePlanRevisionBody) UnmarshalBinary(b []byte) error {
	var res CreatePlanRevisionBody
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*o = res
	return nil
}
