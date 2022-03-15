// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
)

// NewGetPlanRevisionPracticeResponsesParams creates a new GetPlanRevisionPracticeResponsesParams object
//
// There are no default values defined in the spec.
func NewGetPlanRevisionPracticeResponsesParams() GetPlanRevisionPracticeResponsesParams {

	return GetPlanRevisionPracticeResponsesParams{}
}

// GetPlanRevisionPracticeResponsesParams contains all the bound params for the get plan revision practice responses operation
// typically these are obtained from a http.Request
//
// swagger:parameters getPlanRevisionPracticeResponses
type GetPlanRevisionPracticeResponsesParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  In: path
	*/
	ID string
	/*
	  Required: true
	  In: path
	*/
	RevID string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetPlanRevisionPracticeResponsesParams() beforehand.
func (o *GetPlanRevisionPracticeResponsesParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	rID, rhkID, _ := route.Params.GetOK("id")
	if err := o.bindID(rID, rhkID, route.Formats); err != nil {
		res = append(res, err)
	}

	rRevID, rhkRevID, _ := route.Params.GetOK("revId")
	if err := o.bindRevID(rRevID, rhkRevID, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindID binds and validates parameter ID from path.
func (o *GetPlanRevisionPracticeResponsesParams) bindID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route
	o.ID = raw

	return nil
}

// bindRevID binds and validates parameter RevID from path.
func (o *GetPlanRevisionPracticeResponsesParams) bindRevID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route
	o.RevID = raw

	return nil
}
