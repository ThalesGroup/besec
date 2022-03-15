// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// New creates a new operations API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) ClientService {
	return &Client{transport: transport, formats: formats}
}

/*
Client for operations API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

// ClientOption is the option for Client methods
type ClientOption func(*runtime.ClientOperation)

// ClientService is the interface for Client methods
type ClientService interface {
	CreatePlan(params *CreatePlanParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*CreatePlanCreated, error)

	CreatePlanRevision(params *CreatePlanRevisionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*CreatePlanRevisionOK, error)

	CreateProject(params *CreateProjectParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*CreateProjectCreated, error)

	DeletePlan(params *DeletePlanParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DeletePlanNoContent, error)

	DeleteProject(params *DeleteProjectParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DeleteProjectNoContent, error)

	GetAuthConfig(params *GetAuthConfigParams, opts ...ClientOption) (*GetAuthConfigOK, error)

	GetPlan(params *GetPlanParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*GetPlanOK, error)

	GetPlanRevision(params *GetPlanRevisionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*GetPlanRevisionOK, error)

	GetPlanRevisionPracticeResponses(params *GetPlanRevisionPracticeResponsesParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*GetPlanRevisionPracticeResponsesOK, error)

	GetPlanVersions(params *GetPlanVersionsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*GetPlanVersionsOK, error)

	GetPractices(params *GetPracticesParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*GetPracticesOK, error)

	GetProject(params *GetProjectParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*GetProjectOK, error)

	ListPracticesVersions(params *ListPracticesVersionsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*ListPracticesVersionsOK, error)

	ListProjects(params *ListProjectsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*ListProjectsOK, error)

	LoggedIn(params *LoggedInParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*LoggedInOK, error)

	UpdateProject(params *UpdateProjectParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*UpdateProjectOK, error)

	SetTransport(transport runtime.ClientTransport)
}

/*
  CreatePlan create plan API
*/
func (a *Client) CreatePlan(params *CreatePlanParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*CreatePlanCreated, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCreatePlanParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "createPlan",
		Method:             "POST",
		PathPattern:        "/plan",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &CreatePlanReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CreatePlanCreated)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*CreatePlanDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  CreatePlanRevision create plan revision API
*/
func (a *Client) CreatePlanRevision(params *CreatePlanRevisionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*CreatePlanRevisionOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCreatePlanRevisionParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "createPlanRevision",
		Method:             "POST",
		PathPattern:        "/plan/{id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &CreatePlanRevisionReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CreatePlanRevisionOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*CreatePlanRevisionDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  CreateProject create project API
*/
func (a *Client) CreateProject(params *CreateProjectParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*CreateProjectCreated, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCreateProjectParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "createProject",
		Method:             "POST",
		PathPattern:        "/project",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &CreateProjectReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CreateProjectCreated)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*CreateProjectDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  DeletePlan Delete this plan and all of the revisions associated with it
*/
func (a *Client) DeletePlan(params *DeletePlanParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DeletePlanNoContent, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewDeletePlanParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "deletePlan",
		Method:             "DELETE",
		PathPattern:        "/plan/{id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &DeletePlanReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*DeletePlanNoContent)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*DeletePlanDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  DeleteProject delete project API
*/
func (a *Client) DeleteProject(params *DeleteProjectParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*DeleteProjectNoContent, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewDeleteProjectParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "deleteProject",
		Method:             "DELETE",
		PathPattern:        "/project/{id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &DeleteProjectReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*DeleteProjectNoContent)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*DeleteProjectDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetAuthConfig get auth config API
*/
func (a *Client) GetAuthConfig(params *GetAuthConfigParams, opts ...ClientOption) (*GetAuthConfigOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAuthConfigParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "getAuthConfig",
		Method:             "GET",
		PathPattern:        "/auth",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetAuthConfigReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAuthConfigOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetAuthConfigDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetPlan get plan API
*/
func (a *Client) GetPlan(params *GetPlanParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*GetPlanOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetPlanParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "getPlan",
		Method:             "GET",
		PathPattern:        "/plan/{id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetPlanReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetPlanOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetPlanDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetPlanRevision get plan revision API
*/
func (a *Client) GetPlanRevision(params *GetPlanRevisionParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*GetPlanRevisionOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetPlanRevisionParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "getPlanRevision",
		Method:             "GET",
		PathPattern:        "/plan/{id}/revision/{revId}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetPlanRevisionReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetPlanRevisionOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetPlanRevisionDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetPlanRevisionPracticeResponses get plan revision practice responses API
*/
func (a *Client) GetPlanRevisionPracticeResponses(params *GetPlanRevisionPracticeResponsesParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*GetPlanRevisionPracticeResponsesOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetPlanRevisionPracticeResponsesParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "getPlanRevisionPracticeResponses",
		Method:             "GET",
		PathPattern:        "/plan/{id}/revision/{revId}/responses",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetPlanRevisionPracticeResponsesReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetPlanRevisionPracticeResponsesOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetPlanRevisionPracticeResponsesDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetPlanVersions get plan versions API
*/
func (a *Client) GetPlanVersions(params *GetPlanVersionsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*GetPlanVersionsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetPlanVersionsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "getPlanVersions",
		Method:             "GET",
		PathPattern:        "/plan/{id}/versions",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetPlanVersionsReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetPlanVersionsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetPlanVersionsDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetPractices get practices API
*/
func (a *Client) GetPractices(params *GetPracticesParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*GetPracticesOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetPracticesParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "getPractices",
		Method:             "GET",
		PathPattern:        "/practices/{version}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetPracticesReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetPracticesOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetPracticesDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetProject get project API
*/
func (a *Client) GetProject(params *GetProjectParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*GetProjectOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetProjectParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "getProject",
		Method:             "GET",
		PathPattern:        "/project/{id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetProjectReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetProjectOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetProjectDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  ListPracticesVersions list practices versions API
*/
func (a *Client) ListPracticesVersions(params *ListPracticesVersionsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*ListPracticesVersionsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewListPracticesVersionsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "listPracticesVersions",
		Method:             "GET",
		PathPattern:        "/practices",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &ListPracticesVersionsReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ListPracticesVersionsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*ListPracticesVersionsDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  ListProjects list projects API
*/
func (a *Client) ListProjects(params *ListProjectsParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*ListProjectsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewListProjectsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "listProjects",
		Method:             "GET",
		PathPattern:        "/project",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &ListProjectsReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ListProjectsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*ListProjectsDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  LoggedIn Used to trigger one-time events like requesting access. Clients should hit this once after obtaining an ID token, and can use or ignore the response.
*/
func (a *Client) LoggedIn(params *LoggedInParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*LoggedInOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewLoggedInParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "loggedIn",
		Method:             "POST",
		PathPattern:        "/auth",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &LoggedInReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*LoggedInOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*LoggedInDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  UpdateProject update project API
*/
func (a *Client) UpdateProject(params *UpdateProjectParams, authInfo runtime.ClientAuthInfoWriter, opts ...ClientOption) (*UpdateProjectOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewUpdateProjectParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "updateProject",
		Method:             "PUT",
		PathPattern:        "/project/{id}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &UpdateProjectReader{formats: a.formats},
		AuthInfo:           authInfo,
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*UpdateProjectOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*UpdateProjectDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
