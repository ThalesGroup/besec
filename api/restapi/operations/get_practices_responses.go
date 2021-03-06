// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/ThalesGroup/besec/api/models"
)

// GetPracticesOKCode is the HTTP code returned for type GetPracticesOK
const GetPracticesOKCode int = 200

/*GetPracticesOK OK

swagger:response getPracticesOK
*/
type GetPracticesOK struct {

	/*
	  In: Body
	*/
	Payload *models.GotPractices `json:"body,omitempty"`
}

// NewGetPracticesOK creates GetPracticesOK with default headers values
func NewGetPracticesOK() *GetPracticesOK {

	return &GetPracticesOK{}
}

// WithPayload adds the payload to the get practices o k response
func (o *GetPracticesOK) WithPayload(payload *models.GotPractices) *GetPracticesOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get practices o k response
func (o *GetPracticesOK) SetPayload(payload *models.GotPractices) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetPracticesOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*GetPracticesDefault error

swagger:response getPracticesDefault
*/
type GetPracticesDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetPracticesDefault creates GetPracticesDefault with default headers values
func NewGetPracticesDefault(code int) *GetPracticesDefault {
	if code <= 0 {
		code = 500
	}

	return &GetPracticesDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the get practices default response
func (o *GetPracticesDefault) WithStatusCode(code int) *GetPracticesDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the get practices default response
func (o *GetPracticesDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the get practices default response
func (o *GetPracticesDefault) WithPayload(payload *models.Error) *GetPracticesDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get practices default response
func (o *GetPracticesDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetPracticesDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
