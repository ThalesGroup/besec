package cmd

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ThalesGroup/besec/store"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(uiDir fs.FS) {
	UIDir = http.FS(uiDir)
	if err := newRootCmd().Execute(); err != nil {
		log.Fatalln(err)
	}
}

type rootCmd struct {
	*cobra.Command
	baseDir string
	cfgFile string
}

func newRootCmd() *rootCmd {
	rc := &rootCmd{}

	cobra.OnInitialize(rc.initConfig)

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	rc.baseDir = filepath.Dir(ex)

	// the base command when called without any subcommands
	rc.Command = &cobra.Command{
		Use:   "besec",
		Short: "<tagline>",
		Long:  `besec <tagline>?`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if viper.GetBool("verbose") {
				log.SetLevel(log.DebugLevel)
				log.Debug("Debug logging enabled")
			}
		},
		//Run: a function to handle bare calls to the app
	}

	// global persistent flags defined here - these apply to all commands

	rc.PersistentFlags().StringVar(&rc.cfgFile, "config", "", "config file (default is config.yaml in the current directory then the program directory)")

	rc.Command.Version = VERSION + ", git commit: " + GITCOMMIT

	rc.PersistentFlags().BoolP("verbose", "v", false, "Verbose log output")
	err = viper.BindPFlag("verbose", rc.PersistentFlags().Lookup("verbose"))
	if err != nil {
		log.Fatalf("Error binding viper flag: %v", err)
	}

	rc.PersistentFlags().String("gcp-project", "", "Google Cloud Project ID for firestore access. If not set, will attempt to query the instance metadata endpoint to use the current project.")
	err = viper.BindPFlag("gcp-project", rc.PersistentFlags().Lookup("gcp-project"))
	if err != nil {
		log.Fatalf("Error binding viper flag: %v", err)
	}

	rc.PersistentFlags().String(serviceAccountFlagName, "", "The name of a service account to impersonate, e.g. cli-administrator@<project>.iam.gserviceaccount.com. If the principal obtained using default application credentials is not a service account, you must specify this option to manage practices, users, or run the server.")
	err = viper.BindPFlag(serviceAccountFlagName, rc.PersistentFlags().Lookup(serviceAccountFlagName))
	if err != nil {
		log.Fatalf("Error binding viper flag: %v", err)
	}

	rc.PersistentFlags().BoolP("no-emulator", "C", false, "Allow use of the Cloud Firestore rather than a local emulator")
	err = viper.BindPFlag("no-emulator", rc.PersistentFlags().Lookup("no-emulator"))
	if err != nil {
		log.Fatalf("Error binding viper flag: %v", err)
	}

	rc.AddCommand(newPracticesCmd(rc).Command)
	rc.AddCommand(newUsersCmd(rc).Command)
	rc.AddCommand(newDemoCmd().Command)
	rc.AddCommand(newServeCmd())

	return rc
}

func (rc *rootCmd) initConfig() {
	// environment variables like BESEC_PRACTICESDIR will take precedence over a config file but are overruled by flags
	viper.SetEnvPrefix("besec")
	viper.AutomaticEnv()
	viper.AllowEmptyEnv(true) // explicitly set empty environment variables should be used

	// flags that have a dash in their name need to become an underscore when checking the environment
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

	if rc.cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(rc.cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath(rc.baseDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	if err := viper.ReadInConfig(); err != nil {
		if rc.cfgFile != "" {
			log.Fatalf("The specified config file '%v', was not found\n", rc.cfgFile)
		}
		log.Debugf("Primary config file not loaded: %v", err)
	}
	viper.SetConfigName("config.local")
	if err := viper.MergeInConfig(); err != nil {
		log.Debugf("Local config file not loaded: %v", err)
	}
}

// initStore returns a Firestore
func initStore() store.Store {
	return store.NewFireStore(viper.GetString("gcp-project"))
}

func checkEmulator() {
	if !viper.GetBool("no-emulator") {
		if _, ok := os.LookupEnv("FIRESTORE_EMULATOR_HOST"); !ok {
			log.Fatal("FIRESTORE_EMULATOR_HOST is not set.\nTo run against the cloud database, use --no-emulator.\nTo use the emulator, run e.g.\n`gcloud beta emulators firestore start --host-port localhost:8088`")
		}
	}
}
