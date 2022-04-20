package cmd

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"

	"github.com/ThalesGroup/besec/api/client"
	"github.com/ThalesGroup/besec/api/client/operations"
	"github.com/ThalesGroup/besec/api/models"
	"github.com/ThalesGroup/besec/lib"
)

type demoProject struct {
	Project models.ProjectDetails `yaml:"project"`
	Plans   [][]lib.Plan          `yaml:"plans"` // Each plan is a sequence of one or more revisions
}

type demoCmd struct {
	*cobra.Command
	Client        *client.Besec
	AuthInfo      runtime.ClientAuthInfoWriter
	Data          []demoProject
	DemoPractices []lib.Practice
}

func newDemoCmd() *demoCmd {
	dc := &demoCmd{}

	dc.Command = &cobra.Command{
		Use:   "demo [remove]",
		Short: "Load demo data into a running instance",
		Long: `Load data from the demo directory in to the specified instance.
If remove is specified, remove any projects that have the same name and plan dates as demo projects.`,
		Run: func(cmd *cobra.Command, args []string) {
			checkEmulator()
			endpoint := viper.GetString("endpoint")
			if endpoint == "" {
				log.Fatal("A URL endpoint for a running instance of `besec serve` must be provided, for example http://localhost:8080")
			}
			url, err := url.Parse(endpoint)
			if err != nil {
				log.Fatalf("Cannot parse endpoint URL: %v", err)
			}
			cfg := client.TransportConfig{Host: url.Host, Schemes: []string{url.Scheme}, BasePath: client.DefaultBasePath}
			dc.Client = client.NewHTTPClientWithConfig(nil, &cfg)

			token := viper.GetString("access-token")
			if token == "" {
				log.Warn("No access token found. To operate on endpoints that require authentication, set your OAuth token via the BESEC_ACCESS_TOKEN environment variable or access-token config entry")
			}
			dc.AuthInfo = httptransport.BearerToken(token)

			if err = dc.readDemoData(viper.GetString("demo-dir")); err != nil {
				log.Fatalf("Error reading demo data: %v", err)
			}

			if len(args) == 0 {
				dc.Load()
			} else if len(args) == 1 {
				if args[0] == "remove" {
					dc.Remove()
				} else {
					log.Fatalf("Unrecognized argument, '%v'", args[0])
				}
			} else {
				log.Fatalf("Usage: 'demo' or 'demo remove'")
			}
		},
	}

	dc.PersistentFlags().String("demo-dir", "./demo", "Directory to use with demo data in")
	err := viper.BindPFlag("demo-dir", dc.PersistentFlags().Lookup("demo-dir"))
	if err != nil {
		log.Fatalf("Error binding viper flag: %v", err)
	}

	dc.PersistentFlags().StringP("endpoint", "e", "http://localhost:8080", "The URL of the besec serve instance to interact with")
	err = viper.BindPFlag("endpoint", dc.PersistentFlags().Lookup("endpoint"))
	if err != nil {
		log.Fatalf("Error binding viper flag: %v", err)
	}

	return dc
}

// Load creates the projects and plans from dc.Data in the configured instance
func (dc *demoCmd) Load() {
	dc.loadPractices()

	for _, project := range dc.Data {
		log.Debugf("Creating project %v", *project.Project.Name)
		if err := dc.createProject(project.Project); err != nil {
			log.Fatalf("Error creating project %v: %v", *project.Project.Name, err)
		}
		for _, revisions := range project.Plans {
			if err := dc.createPlanHistory(revisions); err != nil {
				log.Fatalf("Error creating plan for project %v: %v", *project.Project.Name, err)
			}
		}
	}
}

func (dc *demoCmd) loadPractices() {
	// Get the published practice versions
	resp, err := dc.Client.Operations.ListPracticesVersions(nil, dc.AuthInfo)
	if err != nil {
		log.Fatalf("Couldn't retrieve practice versions: %v", err)
	}
	versions := resp.Payload
	if len(versions) == 0 {
		log.Fatal("No practices found - try running `practices publish` first")
	}

	// Update all the plan revisions to reference the latest published version
	latest := versions[len(versions)-1]
	for pi := range dc.Data {
		for ri := range dc.Data[pi].Plans {
			for rj, revision := range dc.Data[pi].Plans[ri] {
				switch revision.Responses.PracticesVersion {
				case "current":
					dc.Data[pi].Plans[ri][rj].Responses.PracticesVersion = latest
				case "0-demo": // fine
				default:
					log.Fatalf("Revision %v in project %v has an unexpected practices version '%v' - must be 'current' or '0-demo'",
						rj, *dc.Data[pi].Project.Name, revision.Responses.PracticesVersion)
				}
			}
		}
	}

	// Publish the demo practices if they aren't already
	found := false
	for _, v := range versions {
		if v == "0-demo" {
			found = true
		}
	}
	if !found {
		log.Info("Publishing demo practices")
		practicesDir := filepath.Join(viper.GetString("demo-dir"), "practices") // initializing the root command resets viper to default values - this has to be called first
		noem := fmt.Sprint(viper.GetBool("no-emulator"))
		schemaFile := filepath.Join(viper.GetString("practices-dir"), "schema.json")
		rc := newRootCmd()
		rc.SetArgs([]string{"practices", "publish",
			"--no-emulator=" + noem,
			"--schema-file", schemaFile,
			"--practices-dir", practicesDir,
			"--force-version", "0-demo"})
		if rc.Execute() != nil {
			log.Fatal("Failed to publish demo practices")
		}
	}
}

// Remove removes the projects in d.data and all their associated plans from the configured instance
func (dc *demoCmd) Remove() {
	resp, err := dc.Client.Operations.ListProjects(operations.NewListProjectsParams(), dc.AuthInfo)
	if err != nil {
		ae, ok := err.(*operations.ListProjectsDefault)
		if ok {
			log.Fatalf("Failed to list projects [%d]: %v", ae.Code(), *ae.Payload.Message)
		}
		log.Fatalf("Failed to list projects: %v", err)
	}
	for _, project := range resp.Payload {
		matched := false
		for _, demoProject := range dc.Data {
			if *project.Attributes.Name == *demoProject.Project.Name {
				matched = true
				log.Infof("Deleting project %s with ID %v and all of its plans", *project.Attributes.Name, project.ID)
				dc.deleteProject(project)
				break
			}
		}
		if !matched {
			log.Infof("Project '%s' doesn't match any demo projects, ignoring", *project.Attributes.Name)
		}
	}
}

func (dc *demoCmd) deleteProject(project *models.Project) {
	// Delete all of the project's plans
	for _, planID := range project.Plans {
		_, err := dc.Client.Operations.DeletePlan(operations.NewDeletePlanParams().WithID(planID), dc.AuthInfo)
		if err != nil {
			ae, ok := err.(*operations.DeletePlanDefault)
			if ok {
				log.Fatalf("Failed to delete plan %v [%v]: %v", planID, ae.Code(), *ae.Payload.Message)
			}
			log.Fatalf("Failed to delete plan %v: %v", planID, err)
		}
		log.Infof("Deleted plan %v", planID)
	}

	// Delete the project
	_, err := dc.Client.Operations.DeleteProject(operations.NewDeleteProjectParams().WithID(project.ID), dc.AuthInfo)
	if err != nil {
		log.Fatalf("Failed to delete project %v: %v", project.ID, err)
	}
	log.Infof("Deleted project %v", project.ID)
}

func (dc *demoCmd) readDemoData(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return err
	}
	for _, filePath := range files {
		var file []byte
		if err != nil {
			return err
		}

		if file, err = ioutil.ReadFile(filePath); err != nil {
			return err
		}

		project := demoProject{}
		if err := yaml.Unmarshal(file, &project); err != nil { // can't use UnmarshalStrict with yamlv2 due to https://github.com/go-yaml/yaml/issues/410
			return fmt.Errorf("Error unmarshalling demo project %v: %v", filePath, err)
		}
		dc.Data = append(dc.Data, project)
	}
	return nil
}

func (dc *demoCmd) createProject(project models.ProjectDetails) error {
	params := operations.NewCreateProjectParams().WithBody(&project)
	resp, err := dc.Client.Operations.CreateProject(params, dc.AuthInfo)
	if err != nil {
		ae, ok := err.(*operations.CreateProjectDefault)
		if ok {
			return fmt.Errorf("Error creating project [%v]: %v", ae.Code(), *ae.Payload.Message)
		}
		return err
	}
	log.Infof("Created project %s: %s", *project.Name, resp.Payload)

	// Update every plan revision that refers to this project by name to use its new ID
	for _, demoProject := range dc.Data {
		for _, revisions := range demoProject.Plans {
			for _, revision := range revisions {
				for i, p := range revision.Details.Projects {
					if p == *project.Name {
						revision.Details.Projects[i] = resp.Payload
					}
				}
			}
		}
	}
	return nil
}

func (dc *demoCmd) createPlanHistory(revisions []lib.Plan) error {
	planID := ""
	for _, revision := range revisions {
		if planID == "" {
			// create a new plan with the first revision
			params := operations.NewCreatePlanParams().WithBody(operations.CreatePlanBody{Responses: &revision.Responses, Details: &revision.Details})
			resp, err := dc.Client.Operations.CreatePlan(params, dc.AuthInfo)
			if err != nil {
				ae, ok := err.(*operations.CreatePlanDefault)
				if ok {
					return fmt.Errorf("Error creating plan [%v]: %v", ae.Code(), *ae.Payload.Message)
				}
				return err
			}
			log.Infof("Created plan %v with initial revision %v", *resp.Payload.PlanID, *resp.Payload.RevisionID)
			planID = *resp.Payload.PlanID
		} else {
			// already have a plan, add a revision to it
			params := operations.NewCreatePlanRevisionParams().WithBody(operations.CreatePlanRevisionBody{Responses: &revision.Responses, Details: &revision.Details}).WithID(planID)
			resp, err := dc.Client.Operations.CreatePlanRevision(params, dc.AuthInfo)
			if err != nil {
				ae, ok := err.(*operations.CreatePlanRevisionDefault)
				if ok {
					return fmt.Errorf("Error creating plan revision [%v]: %v", ae.Code(), *ae.Payload.Message)
				}
				return err
			}
			log.Infof("Created revision %v for plan %v", resp.Payload, planID)
		}
	}
	return nil
}
