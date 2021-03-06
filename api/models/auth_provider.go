// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// AuthProvider auth provider
//
// swagger:model authProvider
type AuthProvider struct {

	// The Google Identity Platform provider ID
	// Example: google.com
	// Required: true
	ID *string `json:"id"`

	// saml claims
	SamlClaims *SamlProviderClaimsMap `json:"samlClaims,omitempty"`

	// sign in options
	SignInOptions *SignInOptions `json:"signInOptions,omitempty"`

	// True if every user from this provider has access
	Whitelisted bool `json:"whitelisted,omitempty"`
}

// Validate validates this auth provider
func (m *AuthProvider) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateSamlClaims(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateSignInOptions(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *AuthProvider) validateID(formats strfmt.Registry) error {

	if err := validate.Required("id", "body", m.ID); err != nil {
		return err
	}

	return nil
}

func (m *AuthProvider) validateSamlClaims(formats strfmt.Registry) error {
	if swag.IsZero(m.SamlClaims) { // not required
		return nil
	}

	if m.SamlClaims != nil {
		if err := m.SamlClaims.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("samlClaims")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("samlClaims")
			}
			return err
		}
	}

	return nil
}

func (m *AuthProvider) validateSignInOptions(formats strfmt.Registry) error {
	if swag.IsZero(m.SignInOptions) { // not required
		return nil
	}

	if m.SignInOptions != nil {
		if err := m.SignInOptions.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("signInOptions")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("signInOptions")
			}
			return err
		}
	}

	return nil
}

// ContextValidate validate this auth provider based on the context it is used
func (m *AuthProvider) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateSamlClaims(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateSignInOptions(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *AuthProvider) contextValidateSamlClaims(ctx context.Context, formats strfmt.Registry) error {

	if m.SamlClaims != nil {
		if err := m.SamlClaims.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("samlClaims")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("samlClaims")
			}
			return err
		}
	}

	return nil
}

func (m *AuthProvider) contextValidateSignInOptions(ctx context.Context, formats strfmt.Registry) error {

	if m.SignInOptions != nil {
		if err := m.SignInOptions.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("signInOptions")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("signInOptions")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *AuthProvider) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *AuthProvider) UnmarshalBinary(b []byte) error {
	var res AuthProvider
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
