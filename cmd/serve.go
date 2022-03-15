package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof" // for --pprof
	"os"
	"reflect"
	"strings"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"

	"github.com/ThalesGroup/besec/api"
	"github.com/ThalesGroup/besec/api/models"
)

var UIDir http.FileSystem //nolint:gochecknoglobals // we could embed this in a special serve command struct, but it adds a lot of complexity

const stackdriverLogsFlagName = "stackdriver-logs"
const disableAuthFlagName = "disable-auth"
const requestAccessAlertsFlagName = "alert-access-request"
const newUserAlertsFlagName = "alert-first-login"
const apiVersion = "/v1alpha1"
const authConfigKey = "auth"

func newServeCmd() *cobra.Command {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start webserver",
		Long:  `Start the BeSec webapp listening on the configured port`,
		Run: func(cmd *cobra.Command, args []string) {
			checkEmulator()
			serve()
		},
	}

	serveCmd.PersistentFlags().Int16("port", 8080, "Port to listen on")
	err := viper.BindPFlag("port", serveCmd.PersistentFlags().Lookup("port"))
	if err != nil {
		log.Fatalf("Error binding viper flag: %v", err)
	}
	viper.SetDefault("port", 8080)

	serveCmd.PersistentFlags().Bool(requestAccessAlertsFlagName, true, "Send alerts when a user tries to log in for the first time but doesn't have access")
	err = viper.BindPFlag(requestAccessAlertsFlagName, serveCmd.PersistentFlags().Lookup(requestAccessAlertsFlagName))
	if err != nil {
		log.Fatalf("Error binding viper flag: %v", err)
	}
	serveCmd.PersistentFlags().Bool(newUserAlertsFlagName, true, "Send alerts when a user signs in for the first time")
	err = viper.BindPFlag(newUserAlertsFlagName, serveCmd.PersistentFlags().Lookup(newUserAlertsFlagName))
	if err != nil {
		log.Fatalf("Error binding viper flag: %v", err)
	}
	serveCmd.PersistentFlags().String("slack-webhook-name", "", "Name of the webhook (in the database config as slack-webhook-<name>) to use for alerts")
	err = viper.BindPFlag("slack-webhook-name", serveCmd.PersistentFlags().Lookup("slack-webhook-name"))
	if err != nil {
		log.Fatalf("Error binding viper flag: %v", err)
	}
	serveCmd.PersistentFlags().String("slack-webhook-url", "", "URL of the webhook to use for alerts")
	err = viper.BindPFlag("slack-webhook-url", serveCmd.PersistentFlags().Lookup("slack-webhook-url"))
	if err != nil {
		log.Fatalf("Error binding viper flag: %v", err)
	}

	serveCmd.PersistentFlags().Bool("pprof", false, "Enable insecure pprof debug server at /debug/pprof/")
	err = viper.BindPFlag("pprof", serveCmd.PersistentFlags().Lookup("pprof"))
	if err != nil {
		log.Fatalf("Error binding viper flag: %v", err)
	}

	serveCmd.PersistentFlags().Bool(stackdriverLogsFlagName, false, "Output logs in Stackdriver-friendly JSON format")
	err = viper.BindPFlag(stackdriverLogsFlagName, serveCmd.PersistentFlags().Lookup(stackdriverLogsFlagName))
	if err != nil {
		log.Fatalf("Error binding viper flag: %v", err)
	}

	serveCmd.PersistentFlags().Bool(disableAuthFlagName, false, "Don't require authentication; all actions are carried out by a dummy user")
	err = viper.BindPFlag(disableAuthFlagName, serveCmd.PersistentFlags().Lookup(disableAuthFlagName))
	if err != nil {
		log.Fatalf("Error binding viper flag: %v", err)
	}

	return serveCmd
}

// httpsOnlyMiddleware redirects requests that came via http, as indicated by the X-Forwarded-Proto header (as set by Cloud Run)
// If the header is absent or set to https, this middleware passes the unaltered request on.
// If it is set to http, a redirect response is sent to the client and further processing stops.
func httpsOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proto, exists := r.Header["X-Forwarded-Proto"]
		if exists && proto[0] == "http" {
			http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// An IndexOnlyFileSystem ignores the file name requested, and always opens index.html
type indexOnlyFileSystem struct{ http.FileSystem }

// Open ignores the requested filename and opens index.html
func (fs indexOnlyFileSystem) Open(name string) (http.File, error) {
	return fs.FileSystem.Open("index.html")
}

// uiHandler inspects the URL path to locate a file within the UIDir.
// If a file is found, it will be served. If not, the index will be served.
// The cache timeout is set based on the type of file, per https://create-react-app.dev/docs/production-build#static-file-caching
func uiHandler(w http.ResponseWriter, r *http.Request) {
	dir := UIDir

	if _, err := UIDir.Open(r.URL.Path); err != nil {
		log.WithField("path", r.URL.Path).Debug("uiHandler: file does not exist, serving index.html")
		w.Header().Set("Cache-Control", "no-cache")
		dir = indexOnlyFileSystem{UIDir}
	} else if strings.HasPrefix(r.URL.Path, "/static/") {
		// all the files in this directory are named using a hash of their contents, so they never expire
		w.Header().Set("Cache-Control", "max-age=31536000") // 1 year
	}

	http.FileServer(dir).ServeHTTP(w, r)
}

// newServer creates a server encapsulating both the API server and the static file server
func newServer(port int, rt *api.Runtime) *http.Server {
	apiSrv := api.NewAPI(rt)
	defer func() {
		err := apiSrv.Shutdown()
		if err != nil {
			log.Errorf("Failed to cleanly shut server down: %v", err)
		}
	}()

	// Instead of using apiSrv.Serve(), register its handler with our own router so we can serve both the API and other assets
	r := mux.NewRouter()
	r.PathPrefix(apiVersion).Handler(apiSrv.GetHandler())
	if viper.GetBool("pprof") {
		log.Warn("Insecure pprof server listening at /debug/pprof/")
		r.PathPrefix("/debug").Handler(http.HandlerFunc(pprof.Index))
	}
	r.PathPrefix("/").Handler(http.HandlerFunc(uiHandler))

	r.Use(httpsOnlyMiddleware)
	r.Use(XCloudTraceContext)

	hSrv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return hSrv
}

func serve() {
	disableAuth := viper.GetBool(disableAuthFlagName)
	project := viper.GetString("gcp-project")
	if project == "" && !disableAuth {
		log.Fatal("No Google Cloud project ID specified - run with --gcp-project or set gcp-project in the configuration file.\nA project ID is required unless --disable-auth is set, as it is used to validate authentication tokens.")
	}

	var authClient *auth.Client
	if disableAuth {
		authClient = nil
		log.Warn("Authentication has been disabled, anonymous users will be authenticated as a pre-authorized example user")
	} else {
		authClient = api.InitAuthClient(project, true, viper.GetString(serviceAccountFlagName))
	}

	if viper.GetBool(stackdriverLogsFlagName) {
		// Configure log output to work well with Stackdriver Logging, when paired with middleware to extract the trace ID.
		// Use like this in a handler:
		//     logger := log.WithContext(r.Context())
		//     logger.WithFields(log.Fields{..}).Debug("...")
		log.SetFormatter(NewStackdriverFormatter(project))
	}

	authConfig := api.NewExtendedAuthConfig(getAuthConfig())

	st := initStore()

	requestAccessAlerts := viper.GetBool(requestAccessAlertsFlagName)
	newUserAlerts := viper.GetBool(newUserAlertsFlagName)
	webhook := viper.GetString("slack-webhook-url")
	webhookName := viper.GetString("slack-webhook-name")
	if requestAccessAlerts || newUserAlerts {
		if webhook == "" {
			if webhookName == "" {
				log.Error("Configured to send alerts but no webhook url or name specified.")
				return
			}

			var err error
			webhook, err = st.GetConfigString(context.Background(), "slack-webhook-"+webhookName)
			if err != nil {
				log.WithFields(log.Fields{"webhook": webhookName, "error": err}).Error("Couldn't get webhook URL")
				return
			}
		} else if webhookName != "" {
			log.Warn("Both a Slack webhook config name and explicit URL have been provided; the name will be ignored.")
		}
	}

	sc := make(chan api.SlackMessage, 10)

	rt := api.NewRuntime(
		st,
		authClient,
		authConfig,
		requestAccessAlerts,
		newUserAlerts,
		sc,
	)

	port := viper.GetInt("port")
	srv := newServer(port, rt)

	if requestAccessAlerts || newUserAlerts {
		go api.SlackSender(sc, rt, webhook)
	}

	log.WithFields(log.Fields{"port": port}).Print("Listening")
	log.Fatal(srv.ListenAndServe())
}

func getAuthConfig() models.AuthConfig {
	// viper doesn't automatically decode structured environment variables, but we can do it using a decode hook
	var config models.AuthConfig
	opt := viper.DecodeHook(
		yamlStringToStruct(config),
	)
	err := viper.UnmarshalKey(authConfigKey, &config, opt)
	if err != nil {
		log.Fatalf("Failed to parse auth config: %v", err)
	}

	if host, ok := os.LookupEnv("FIREBASE_AUTH_EMULATOR_HOST"); ok {
		expected := "http://" + host
		if config.EmulatorURL == "" {
			config.EmulatorURL = expected
			log.WithFields(log.Fields{"EmulatorURL": config.EmulatorURL}).Info("Updated auth config based on FIREBASE_AUTH_EMULATOR_HOST environment variable")
		} else if config.EmulatorURL != expected {
			log.WithFields(log.Fields{"EmulatorURL": config.EmulatorURL, "expected": expected}).Warn("Configured EmulatorUrl didn't match with FIREBASE_AUTH_EMULATOR_HOST environment variable")
		}
	}

	return config
}

func yamlStringToStruct(m interface{}) func(rf reflect.Kind, rt reflect.Kind, data interface{}) (interface{}, error) {
	return func(rf reflect.Kind, rt reflect.Kind, data interface{}) (interface{}, error) {
		if rf != reflect.String || rt != reflect.Struct {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return m, nil
		}

		return m, yaml.UnmarshalStrict([]byte(raw), &m)
	}
}
