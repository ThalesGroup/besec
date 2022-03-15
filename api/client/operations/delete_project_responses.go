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

// DeleteProjectReader is a Reader for the DeleteProject structure.
type DeleteProjectReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *DeleteProjectReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 204:
		result := NewDeleteProjectNoContent()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewDeleteProjectDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewDeleteProjectNoContent creates a DeleteProjectNoContent with default headers values
func NewDeleteProjectNoContent() *DeleteProjectNoContent {
	return &DeleteProjectNoContent{}
}

/* DeleteProjectNoContent describes a response with status code 204, with default header values.

Deleted
*/
type DeleteProjectNoContent struct {
}

func (o *DeleteProjectNoContent) Error() string {
	return fmt.Sprintf("[DELETE /project/{id}][%d] deleteProjectNoContent ", 204)
}

func (o *DeleteProjectNoContent) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewDeleteProjectDefault creates a DeleteProjectDefault with default headers values
func NewDeleteProjectDefault(code int) *DeleteProjectDefault {
	return &DeleteProjectDefault{
		_statusCode: code,
	}
}

/* DeleteProjectDefault describes a response with status code -1, with default header values.

error
*/
type DeleteProjectDefault struct {
	_statusCode int

	Payload *models.Error
}

// Code gets the status code for the delete project default response
func (o *DeleteProjectDefault) Code() int {
	return o._statusCode
}

func (o *DeleteProjectDefault) Error() string {
	return fmt.Sprintf("[DELETE /project/{id}][%d] deleteProject default  %+v", o._statusCode, o.Payload)
}
func (o *DeleteProjectDefault) GetPayload() *models.Error {
	return o.Payload
}

func (o *DeleteProjectDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
