// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	"github.com/rs/cors"

	"github.com/ThalesGroup/besec/api/restapi/operations"
)

//go:generate swagger generate server --target ../../api --name BeSec --spec ../swagger.yaml --principal models.User --exclude-main

func configureFlags(api *operations.BesecAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.BesecAPI) http.Handler {
	api.ServeError = errors.ServeError
	api.JSONConsumer = runtime.JSONConsumer()
	api.JSONProducer = runtime.JSONProducer()
	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	// Allow cross origin requests from the local server during debugging
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // TODO: this list should live in the config file, which possibly implies not using ConfigureAPI()
		AllowCredentials: false,                             // we do, but just via the Authorization header not cookies
		AllowedHeaders:   []string{"Authorization", "content-type"},
		AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "DELETE"},
		Debug:            false,
	})

	// Insert the middleware
	return c.Handler(handler)
}
