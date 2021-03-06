// Impersonation code from https://gist.github.com/salrashid123/05e0e74be3621754cc295a185e08a3c3
// adapted to remove the use of a root token source, falling back to application default credentials instead
// Usage example of the original: https://gist.github.com/salrashid123/7994be642c466844f85e2f2c68a5dec9

// Once https://github.com/googleapis/google-api-go-client/issues/378 is resolved, this can be removed

// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license

package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"google.golang.org/api/iamcredentials/v1"
)

// ImpersonatedTokenConfig prameters to start Credential impersonation exchange.
type ImpersonatedTokenConfig struct {
	TargetPrincipal string
	Lifetime        time.Duration
	Delegates       []string
	TargetScopes    []string
}

// ImpersonatedTokenSource returns a TokenSource issued to a user or
// service account to impersonate another. The source project using
// ImpersonatedTokenSource must enable the "IAMCredentials" API.  Also, the
// target service account must grant the orginating principal the
// "Service Account Token Creator" IAM role:
// https://cloud.google.com/iam/docs/service-accounts#the_service_account_token_creator_role
//
//  rootSource (TokenSource): The root TokenSource
//     used as to acquire the target identity TokenSource.
//     rootSource *must* include scopes that contains
//     "https://www.googleapis.com/auth/iam"
//     or
//     "https://www.googleapis.com/auth/cloud.platform"
//  targetPrincipal (string): The service account to impersonate.
//  targetScopes ([]string): Scopes to request during the
//     authorization grant.
//  delegates ([]string): The chained list of delegates required
//      to grant the final access_token.  If set, the sequence of
//      identities must have "Service Account Token Creator" capability
//      granted to the preceding identity.  For example, if set to
//      [serviceAccountB, serviceAccountC], the source_credential
//      must have the Token Creator role on serviceAccountB.
//      serviceAccountB must have the Token Creator on serviceAccountC.
//      Finally, C must have Token Creator on target_principal.
//      If left unset, source_credential must have that role on
//      target_principal.
//  lifetime (time.Duration): Number of seconds the impersonated credential should
//     be valid for (up to 3600).
//
// Note that this is not a standard OAuth flow, but rather uses Google Cloud
// IAMCredentials API to exchange one oauth token for an impersonated account
// see: https://cloud.google.com/iam/credentials/reference/rest/v1/projects.serviceAccounts/generateAccessToken
func ImpersonatedTokenSource(tokenConfig ImpersonatedTokenConfig) (oauth2.TokenSource, error) {
	if tokenConfig.Lifetime > (3600 * time.Second) {
		return nil, fmt.Errorf("oauth2/google: lifetime must be less than or equal to 3600 seconds")
	}

	return &impersonatedTokenSource{
		refreshMutex:      &sync.Mutex{}, // guards impersonatedToken; held while fetching or updating it.
		impersonatedToken: nil,           // Token representing the impersonated identity.

		targetPrincipal: tokenConfig.TargetPrincipal,
		lifetime:        tokenConfig.Lifetime,
		delegates:       tokenConfig.Delegates,
		targetScopes:    tokenConfig.TargetScopes,
	}, nil
}

type impersonatedTokenSource struct {
	refreshMutex      *sync.Mutex   // guards t; held while fetching or updating t.
	impersonatedToken *oauth2.Token // Token representing the impersonated identity.

	targetPrincipal string
	lifetime        time.Duration
	delegates       []string
	targetScopes    []string
}

func (ts *impersonatedTokenSource) Token() (*oauth2.Token, error) {
	ts.refreshMutex.Lock()
	defer ts.refreshMutex.Unlock()

	if ts.impersonatedToken.Valid() {
		return ts.impersonatedToken, nil
	}

	service, err := iamcredentials.NewService(context.Background())
	if err != nil {
		return nil, fmt.Errorf("oauth2/google: Error creating IAMCredentials: %v", err)
	}
	name := fmt.Sprintf("projects/-/serviceAccounts/%s", ts.targetPrincipal)
	tokenRequest := &iamcredentials.GenerateAccessTokenRequest{
		Lifetime:  fmt.Sprintf("%ds", int(ts.lifetime.Seconds())),
		Delegates: ts.delegates,
		Scope:     ts.targetScopes,
	}
	at, err := service.Projects.ServiceAccounts.GenerateAccessToken(name, tokenRequest).Do()
	if err != nil {
		e := fmt.Errorf("oauth2/google: Error calling iamcredentials.GenerateAccessToken: %v", err)
		log.Warn(e)
		return nil, e
	}

	expireAt, err := time.Parse(time.RFC3339, at.ExpireTime)
	if err != nil {
		e := fmt.Errorf("oauth2/google: Error parsing ExpireTime from iamcredentials: %v", err)
		log.Warn(e)
		return nil, e
	}

	ts.impersonatedToken = &oauth2.Token{
		AccessToken: at.AccessToken,
		Expiry:      expireAt,
	}

	log.Debug("Created impersonation token for service account ", ts.targetPrincipal)
	return ts.impersonatedToken, nil
}
