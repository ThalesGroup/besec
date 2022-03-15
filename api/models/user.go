package models

import (
	"fmt"

	"firebase.google.com/go/v4/auth"
)

// User represents an authenticated identity
type User struct {
	Token  *auth.Token
	Claims map[string]interface{} // Firebase Custom Claims, looked up or extracted from the token

	// Standard fields from the token:
	UID                  string
	Provider             string // the identity provider, e.g. "google.com" or another provider ID configured in the Cloud Identity Platform
	DetailsNotInFirebase bool
	Email                string
	EmailVerified        bool
	Name                 string
	PictureURL           string
	Department           string

	// Additional data
	ManuallyAuthorized bool
	CreationAlertSent  bool // whether a notification has been sent about this user requesting access or logging in

	LookedUp  bool // Whether a lookup has been made for this user yet. If true, LocalData==nil means this user has no local data
	LocalData *LocalUserData
}

// LocalUserData captures extra information about a user, beyond what is in their ID token
// We could replicate the token info here, so we have a local copy of it, but in general it's better to use the
// firebase API to lookup that data instead
// Custom claims are duplicated, so that the attributes can be set instantly but in future sessions they don't need to be looked up
type LocalUserData struct {
	ManuallyAuthorized bool
	CreationAlertSent  bool

	// Tokens for users authenticated by SAML IDPs don't populate these fields, so we need to manually capture them from the token for use when we don't have a token to hand (admin operations)
	Name  string
	Email string
}

func (u *User) String() string {
	verified := ""
	if !u.EmailVerified {
		verified = " (unverified!)"
	}

	local := "unknown if any local data"
	if u.LookedUp {
		if u.LocalData == nil {
			local = " no local data"
		} else {
			if u.LocalData.ManuallyAuthorized {
				local = " authorized"
			} else {
				local = " not authorized"
			}
			if u.LocalData.CreationAlertSent {
				local += " creation alert seen"
			}
		}
	}

	return fmt.Sprintf("User{%v: '%v' %v%v authenticated by %v; %v}", u.UID, u.Name, u.Email, verified, u.Provider, local)
}
