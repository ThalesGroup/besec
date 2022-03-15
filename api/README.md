The API package contains all of the generated code as well as the business logic for responding to web API requests.

This package is limited to the REST API - the UI is not in scope.

The code in this directory is called from `cmd/serve.go`. The auto-generated code is created via `make generate-api`,
and can be removed with `make clean-api`.

-   `swagger.yaml` -- the API spec from which the server code is generated
-   `models/` -- auto-generated except for user.go
-   `restapi/` -- mostly auto-generated
-   `restapi/configure.go` -- originally auto-generated, now static. Used to set up middleware
-   `api.go` -- API server configuration
