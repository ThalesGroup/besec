package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/go-openapi/runtime"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"

	"github.com/ThalesGroup/besec/api/models"
	"github.com/ThalesGroup/besec/store"
)

// InitAuthClient creates a Firebase admin SDK auth client.
// If useCreds is false, no GCP credentials are required but the client can only validate ID tokens
// If useCreds is true, the environment must have Google application default credentials
// If serviceAccount is non-empty, try and impersonate the named account
func InitAuthClient(projectID string, useCreds bool, serviceAccount string) *auth.Client {
	config := &firebase.Config{
		ProjectID: projectID,
	}

	var err error
	var app *firebase.App
	if !useCreds {
		// We use the firebase SDK to verify ID tokens. The docs state you need credentials, but they are wrong.
		// Initialize the SDK with a dummy token, which won't get used if all we do is verify ID tokens
		ts := oauth2.StaticTokenSource(&oauth2.Token{})
		app, err = firebase.NewApp(context.Background(), config, option.WithTokenSource(ts))
	} else {
		if serviceAccount == "" {
			app, err = firebase.NewApp(context.Background(), config)
		} else {
			tsc := ImpersonatedTokenConfig{
				TargetPrincipal: serviceAccount,
				Lifetime:        100 * time.Second,
				Delegates:       []string{},
				TargetScopes:    []string{"https://www.googleapis.com/auth/cloud-platform"}} // ".../firebase" is insufficient for the auth usage. ".../datastore" might be sufficient for that aspect
			var ts oauth2.TokenSource
			ts, err = ImpersonatedTokenSource(tsc)
			if err == nil {
				app, err = firebase.NewApp(context.Background(), config, option.WithTokenSource(ts))
			}
		}
	}
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("error initializing firebase app")
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Fatal("error getting Auth client")
	}
	return client
}

// MakeKeyAuth returns a function that creates a user from the token in the provided Authorization header
func MakeKeyAuth(rt *Runtime) func(string) (*models.User, error) {
	return func(authHeader string) (*models.User, error) {
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return nil, fmt.Errorf("invalid authorization header, expected format is 'Bearer <id token>'")
		}
		idToken := authHeader[7:]
		token, err := rt.AuthClient.VerifyIDToken(context.Background(), idToken)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Info("Error validating ID token")
			return nil, fmt.Errorf("invalid ID token")
		}
		return NewUser(token, rt.AuthConfig)
	}
}

// MakeDummyKeyAuth returns a function that creates a fixed user, ignoring the provided token
func MakeDummyKeyAuth(rt *Runtime) func(string) (*models.User, error) {
	return func(s string) (*models.User, error) {
		return &models.User{
				UID:                "123",
				Provider:           "example.com",
				Email:              "user@example.com",
				EmailVerified:      true,
				Name:               "Example User",
				PictureURL:         "https://fonts.gstatic.com/s/i/materialicons/person/v1/24px.svg",
				ManuallyAuthorized: true,
				CreationAlertSent:  true,
			},
			nil
	}
}

// MakeAuthorizer returns an authorizer that checks that the authenticated user is allowed to make API requests
func MakeAuthorizer(rt *Runtime) runtime.AuthorizerFunc {
	return runtime.AuthorizerFunc(func(r *http.Request, user interface{}) error {
		// We don't need to check that the request URL is to the API (as opposed to the static site),
		// as this authorizer is only invoked by the API server

		u := user.(*models.User)

		ctx := r.Context()
		logger := log.WithContext(ctx).WithFields(log.Fields{"user": u.UID})

		// Check if use has been individually authorized, or all users from this provider have access
		authorized := u.ManuallyAuthorized || rt.AuthConfig.whitelisted(u.Provider)

		if !authorized && !u.LookedUp {
			// They may be manually authorized, but their token hasn't been refreshed yet - check the database
			// We also want to know if an alert has been sent for them
			err := rt.Store.GetUserData(ctx, u)
			if err != nil {
				logger.Errorf("Authorizer: Failed to look up user data: %v", err)
				return fmt.Errorf("Internal error when checking authorization")
			}
			if u.ManuallyAuthorized {
				logger.Debug("User is manually authorized in the DB but not via their token")
				authorized = true
			}
		}

		if strings.HasSuffix(r.URL.Path, "/auth") {
			// Clients POST to this once on login, to allow us to do these one-off tasks
			if err := loginTasks(ctx, rt, logger, u, authorized); err != nil {
				return err
			}
		}

		if authorized {
			return nil
		}

		// The error response is displayed in the UI
		// An alternative would be to return a json response for the UI to interpret and display an appropriate message,
		// but this is simpler and works fine.
		alertMsg := "Sorry, this account doesn't have access. If you think this is an error, please contact the security team."
		if u.CreationAlertSent {
			alertMsg = "New user request received, an email will be sent when access is granted."
		}
		return fmt.Errorf(alertMsg)
	})
}

func loginTasks(ctx context.Context, rt *Runtime, logger *log.Entry, u *models.User, authorized bool) error {
	if !u.LookedUp {
		err := rt.Store.GetUserData(ctx, u)
		if err != nil {
			logger.Errorf("Authorizer: Failed to look up user data: %v", err)
			return fmt.Errorf("Internal error when checking authorization")
		}
	}

	if u.DetailsNotInFirebase {
		if u.LocalData == nil {
			u.LocalData = &models.LocalUserData{}
		}
		if u.LocalData.Name == "" {
			// Save the user's details, as they aren't available via Firebase Auth when the user isn't logged in
			u.LocalData.Name = u.Name
			u.LocalData.Email = u.Email
			if err := rt.Store.SaveUserData(ctx, u); err != nil {
				logger.Errorf("Authorizer: failed to persist user details from token to store: %v", err)
				return fmt.Errorf("Internal error when checking authorization")
			}
		}
	}

	if !u.CreationAlertSent {
		if authorized {
			logger.Info("First login")
			NewUserAlert(rt, u)
		} else {
			logger.Info("Unauthorized")
			RequestAccessAlert(rt, u)
		}
	}
	return nil
}

// IsSAMLProvider reports whether the given provider ID corresponds to a SAML provider
func IsSAMLProvider(provider string) bool {
	// this is easily determined as Identity Platform has the following constraint:
	return strings.HasPrefix(provider, "saml.")
}

// NewUser creates a User from the already-validated token
func NewUser(token *auth.Token, authConfig ExtendedAuthConfig) (*models.User, error) {
	firebase, providerConfigured := token.Claims["firebase"].(map[string]interface{})
	if !providerConfigured {
		log.Error("Firebase claim missing from validated ID token")
		return nil, fmt.Errorf("invalid ID token")
	}
	provider, providerConfigured := firebase["sign_in_provider"].(string)
	if !providerConfigured {
		log.Error("sign_in_provider missing from validated ID token")
		return nil, fmt.Errorf("invalid ID token")
	}

	var email, name, picture string
	var verified, detailsMissing bool

	providerClaimsMap := authConfig.samlClaimsMap(provider)
	if providerClaimsMap != nil { // This token is from a SAML provider with configured claims
		detailsMissing = true // Users from SAML providers don't have details populated by Firebase

		attrs, ok := firebase["sign_in_attributes"]
		if !ok {
			log.Errorf("sign_in_attributes missing from validated ID token for configured SAML provider '%v' - is this definitely as SAML provider?", provider)
			return nil, fmt.Errorf("internal authentication error")
		}
		attrsMap, ok := attrs.(map[string]interface{})
		if !ok {
			log.Error("sign_in_attributes not of the expected type")
			return nil, fmt.Errorf("internal authentication error")
		}

		email, ok = attrsMap[*providerClaimsMap.Email].(string)
		if !ok {
			log.Error("email missing from SAML claims")
			return nil, fmt.Errorf("invalid ID token")
		}
		verified = true

		if providerClaimsMap.Name != "" {
			name = attrsMap[providerClaimsMap.Name].(string)
		} else {
			name = attrsMap[providerClaimsMap.FirstName].(string) + " " + attrsMap[providerClaimsMap.Surname].(string)
		}

		if providerClaimsMap.PictureURL != "" {
			picture = attrsMap[providerClaimsMap.PictureURL].(string)
		}
	} else if provider == "google.com" {
		email, providerConfigured = token.Claims["email"].(string)
		if !providerConfigured {
			log.Error("email missing from validated ID token")
			return nil, fmt.Errorf("invalid ID token")
		}
		verified, providerConfigured = token.Claims["email_verified"].(bool)
		if !providerConfigured {
			verified = false
		}
		name, providerConfigured = token.Claims["name"].(string)
		if !providerConfigured {
			name = ""
		}
		picture, providerConfigured = token.Claims["picture"].(string)
		if !providerConfigured {
			picture = ""
		}
	} else {
		return nil, fmt.Errorf("Unknown identity provider '%v'", provider)
	}

	manuallyAuthed := false
	if token.Claims != nil {
		claimIf, exists := token.Claims[manuallyAuthorizedClaim]
		if exists {
			manuallyAuthed = claimIf.(bool)
		}
	}

	return &models.User{Token: token,
			UID:                  token.UID,
			Provider:             provider,
			DetailsNotInFirebase: detailsMissing,
			Email:                email,
			EmailVerified:        verified,
			Name:                 name,
			PictureURL:           picture,
			ManuallyAuthorized:   manuallyAuthed,
		},
		nil
}

const manuallyAuthorizedClaim = "ManuallyAuthorized"

func setClaim(ctx context.Context, client *auth.Client, uid string, claim string, value interface{}) error {
	u, err := client.GetUser(ctx, uid)
	if err != nil {
		return fmt.Errorf("Failed to retrieve user %v when setting claims: %v", uid, err)
	}

	if u.CustomClaims == nil {
		u.CustomClaims = make(map[string]interface{})
	}
	u.CustomClaims[claim] = value

	return client.SetCustomUserClaims(ctx, uid, u.CustomClaims)
}

// GetManuallyAuthorized returns whether the user is manually authorized (checking both their claim and the store)
func GetManuallyAuthorized(ctx context.Context, client *auth.Client, store store.Store, uid string) (bool, error) {
	authUser, err := client.GetUser(ctx, uid)
	if err != nil {
		return false, fmt.Errorf("Failed to retrieve user %v when getting claims: %v", uid, err)
	}
	var claim, claimSet bool
	if authUser.CustomClaims != nil {
		var claimIf interface{}
		claimIf, claimSet = authUser.CustomClaims[manuallyAuthorizedClaim]
		if claimSet {
			claim = claimIf.(bool)
		}
	}

	user := models.User{UID: uid}
	err = (store).GetUserData(ctx, &user) // if the user has no local data, user.LocalData will remain nil
	if err != nil {
		return false, fmt.Errorf("Failed to retrieve local data for user %v when getting auth status: %v", uid, err)
	}

	if claimSet {
		if claim && user.LocalData == nil {
			log.Errorf("ManuallyAuthorized claim for user %v is set, but this user has no local data!", uid)
		} else if claim != user.LocalData.ManuallyAuthorized {
			log.Errorf("ManuallyAuthorized claim for user %v is %v, but database entry is %v!", uid, claim, user.LocalData.ManuallyAuthorized)
		}
	}

	return claim || (user.LocalData != nil && user.LocalData.ManuallyAuthorized), nil
}

// SetManuallyAuthorized records whether the user is manually authorized, in the store and as a custom claim for when their token is next created
func SetManuallyAuthorized(ctx context.Context, client *auth.Client, store store.Store, uid string, value bool) error {
	err := setClaim(ctx, client, uid, manuallyAuthorizedClaim, value)
	if err != nil {
		return fmt.Errorf("Failed to set custom claims for user %v: %v", uid, err)
	}

	return store.SetManuallyAuthorized(ctx, uid, value)
}

// NewExtendedAuthConfig validates the AuthConfig and returns an extended type
func NewExtendedAuthConfig(authConfig models.AuthConfig) ExtendedAuthConfig {
	for _, provider := range authConfig.Providers {
		if provider.SamlClaims != nil {
			if provider.SamlClaims.Name == "" {
				if provider.SamlClaims.FirstName == "" || provider.SamlClaims.Surname == "" {
					log.Fatalf("The claims mapping for the SAML provider %s must have either a Name or FirstName+Surname mapping", *provider.ID)
				}
			}
		}
	}
	return ExtendedAuthConfig{authConfig}
}

// ExtendedAuthConfig extends models.AuthConfig with convenience methods
type ExtendedAuthConfig struct {
	models.AuthConfig
}

func (c *ExtendedAuthConfig) whitelisted(provider string) bool {
	for _, p := range c.Providers {
		if *p.ID == provider {
			return p.Whitelisted
		}
	}
	return false
}

func (c *ExtendedAuthConfig) samlClaimsMap(provider string) *models.SamlProviderClaimsMap {
	for _, p := range c.Providers {
		if *p.ID == provider {
			return p.SamlClaims
		}
	}
	return nil
}
