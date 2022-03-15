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

// GetPracticesReader is a Reader for the GetPractices structure.
type GetPracticesReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetPracticesReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetPracticesOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewGetPracticesDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetPracticesOK creates a GetPracticesOK with default headers values
func NewGetPracticesOK() *GetPracticesOK {
	return &GetPracticesOK{}
}

/* GetPracticesOK describes a response with status code 200, with default header values.

OK
*/
type GetPracticesOK struct {
	Payload *models.GotPractices
}

func (o *GetPracticesOK) Error() string {
	return fmt.Sprintf("[GET /practices/{version}][%d] getPracticesOK  %+v", 200, o.Payload)
}
func (o *GetPracticesOK) GetPayload() *models.GotPractices {
	return o.Payload
}

func (o *GetPracticesOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.GotPractices)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetPracticesDefault creates a GetPracticesDefault with default headers values
func NewGetPracticesDefault(code int) *GetPracticesDefault {
	return &GetPracticesDefault{
		_statusCode: code,
	}
}

/* GetPracticesDefault describes a response with status code -1, with default header values.

error
*/
type GetPracticesDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the get practices default response
func (o *GetPracticesDefault) Code() int {
	return o._statusCode
}

func (o *GetPracticesDefault) Error() string {
	return fmt.Sprintf("[GET /practices/{version}][%d] getPractices default  %+v", o._statusCode, o.Payload)
}
func (o *GetPracticesDefault) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetPracticesDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
