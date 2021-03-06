// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/ThalesGroup/besec/api/models"
)

// GetPlanVersionsReader is a Reader for the GetPlanVersions structure.
type GetPlanVersionsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetPlanVersionsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetPlanVersionsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewGetPlanVersionsDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetPlanVersionsOK creates a GetPlanVersionsOK with default headers values
func NewGetPlanVersionsOK() *GetPlanVersionsOK {
	return &GetPlanVersionsOK{}
}

/* GetPlanVersionsOK describes a response with status code 200, with default header values.

OK
*/
type GetPlanVersionsOK struct {
	Payload []*models.RevisionVersion
}

func (o *GetPlanVersionsOK) Error() string {
	return fmt.Sprintf("[GET /plan/{id}/versions][%d] getPlanVersionsOK  %+v", 200, o.Payload)
}
func (o *GetPlanVersionsOK) GetPayload() []*models.RevisionVersion {
	return o.Payload
}

func (o *GetPlanVersionsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetPlanVersionsDefault creates a GetPlanVersionsDefault with default headers values
func NewGetPlanVersionsDefault(code int) *GetPlanVersionsDefault {
	return &GetPlanVersionsDefault{
		_statusCode: code,
	}
}

/* GetPlanVersionsDefault describes a response with status code -1, with default header values.

error
*/
type GetPlanVersionsDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the get plan versions default response
func (o *GetPlanVersionsDefault) Code() int {
	return o._statusCode
}

func (o *GetPlanVersionsDefault) Error() string {
	return fmt.Sprintf("[GET /plan/{id}/versions][%d] getPlanVersions default  %+v", o._statusCode, o.Payload)
}
func (o *GetPlanVersionsDefault) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetPlanVersionsDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
