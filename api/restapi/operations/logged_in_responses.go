// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/ThalesGroup/besec/api/models"
)

// LoggedInOKCode is the HTTP code returned for type LoggedInOK
const LoggedInOKCode int = 200

/*LoggedInOK OK

swagger:response loggedInOK
*/
type LoggedInOK struct {
}

// NewLoggedInOK creates LoggedInOK with default headers values
func NewLoggedInOK() *LoggedInOK {

	return &LoggedInOK{}
}

// WriteResponse to the client
func (o *LoggedInOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}

/*LoggedInDefault No access, access requested, invalid ID token, or internal error

swagger:response loggedInDefault
*/
type LoggedInDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewLoggedInDefault creates LoggedInDefault with default headers values
func NewLoggedInDefault(code int) *LoggedInDefault {
	if code <= 0 {
		code = 500
	}

	return &LoggedInDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the logged in default response
func (o *LoggedInDefault) WithStatusCode(code int) *LoggedInDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the logged in default response
func (o *LoggedInDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the logged in default response
func (o *LoggedInDefault) WithPayload(payload *models.Error) *LoggedInDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the logged in default response
func (o *LoggedInDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *LoggedInDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
