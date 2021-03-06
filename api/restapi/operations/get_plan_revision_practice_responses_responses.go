// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/ThalesGroup/besec/api/models"
	"github.com/ThalesGroup/besec/lib"
)

// GetPlanRevisionPracticeResponsesOKCode is the HTTP code returned for type GetPlanRevisionPracticeResponsesOK
const GetPlanRevisionPracticeResponsesOKCode int = 200

/*GetPlanRevisionPracticeResponsesOK OK

swagger:response getPlanRevisionPracticeResponsesOK
*/
type GetPlanRevisionPracticeResponsesOK struct {

	/*
	  In: Body
	*/
	Payload *lib.PlanResponses `json:"body,omitempty"`
}

// NewGetPlanRevisionPracticeResponsesOK creates GetPlanRevisionPracticeResponsesOK with default headers values
func NewGetPlanRevisionPracticeResponsesOK() *GetPlanRevisionPracticeResponsesOK {

	return &GetPlanRevisionPracticeResponsesOK{}
}

// WithPayload adds the payload to the get plan revision practice responses o k response
func (o *GetPlanRevisionPracticeResponsesOK) WithPayload(payload *lib.PlanResponses) *GetPlanRevisionPracticeResponsesOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get plan revision practice responses o k response
func (o *GetPlanRevisionPracticeResponsesOK) SetPayload(payload *lib.PlanResponses) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetPlanRevisionPracticeResponsesOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*GetPlanRevisionPracticeResponsesDefault error

swagger:response getPlanRevisionPracticeResponsesDefault
*/
type GetPlanRevisionPracticeResponsesDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetPlanRevisionPracticeResponsesDefault creates GetPlanRevisionPracticeResponsesDefault with default headers values
func NewGetPlanRevisionPracticeResponsesDefault(code int) *GetPlanRevisionPracticeResponsesDefault {
	if code <= 0 {
		code = 500
	}

	return &GetPlanRevisionPracticeResponsesDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the get plan revision practice responses default response
func (o *GetPlanRevisionPracticeResponsesDefault) WithStatusCode(code int) *GetPlanRevisionPracticeResponsesDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the get plan revision practice responses default response
func (o *GetPlanRevisionPracticeResponsesDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the get plan revision practice responses default response
func (o *GetPlanRevisionPracticeResponsesDefault) WithPayload(payload *models.Error) *GetPlanRevisionPracticeResponsesDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get plan revision practice responses default response
func (o *GetPlanRevisionPracticeResponsesDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetPlanRevisionPracticeResponsesDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
