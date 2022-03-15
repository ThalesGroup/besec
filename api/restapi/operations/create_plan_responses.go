// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/ThalesGroup/besec/api/models"
)

// CreatePlanCreatedCode is the HTTP code returned for type CreatePlanCreated
const CreatePlanCreatedCode int = 201

/*CreatePlanCreated Created

swagger:response createPlanCreated
*/
type CreatePlanCreated struct {

	/*
	  In: Body
	*/
	Payload *CreatePlanCreatedBody `json:"body,omitempty"`
}

// NewCreatePlanCreated creates CreatePlanCreated with default headers values
func NewCreatePlanCreated() *CreatePlanCreated {

	return &CreatePlanCreated{}
}

// WithPayload adds the payload to the create plan created response
func (o *CreatePlanCreated) WithPayload(payload *CreatePlanCreatedBody) *CreatePlanCreated {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create plan created response
func (o *CreatePlanCreated) SetPayload(payload *CreatePlanCreatedBody) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreatePlanCreated) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(201)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*CreatePlanDefault error

swagger:response createPlanDefault
*/
type CreatePlanDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewCreatePlanDefault creates CreatePlanDefault with default headers values
func NewCreatePlanDefault(code int) *CreatePlanDefault {
	if code <= 0 {
		code = 500
	}

	return &CreatePlanDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the create plan default response
func (o *CreatePlanDefault) WithStatusCode(code int) *CreatePlanDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the create plan default response
func (o *CreatePlanDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the create plan default response
func (o *CreatePlanDefault) WithPayload(payload *models.Error) *CreatePlanDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create plan default response
func (o *CreatePlanDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreatePlanDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
