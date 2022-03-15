//go:build integration
// +build integration

package cmd

import (
	"testing"

	"github.com/ThalesGroup/besec/api/client/operations"
)

func TestLoadRemove(t *testing.T) {
	TestLoad := func(t *testing.T) {
		d := newDemoCmd()
		// Use a non-default port, so tests don't accidentally clobber a locally running instance
		d.SetArgs([]string{"--endpoint", "http://localhost:8081", "--demo-dir", "../demo"})
		d.Execute()

		params := operations.NewListProjectsParams()
		// Cheekily re-use the Demo's client instance
		resp, err := d.Client.Operations.ListProjects(params, d.AuthInfo)
		if err != nil {
			t.Error("Didn't get a response trying to list projects")
		}
		if len(resp.Payload) != 2 {
			t.Errorf("Expected 2 projects, got %v", len(resp.Payload))
		}
		for _, project := range resp.Payload {
			switch *project.Attributes.Name {
			case "Alpha":
				if len(project.Plans) != 2 {
					t.Errorf("Expected Beta project to have two plans, got %v", len(project.Plans))
				}
			case "Beta":
				if len(project.Plans) != 1 {
					t.Errorf("Expected Alpha project to have one plan, got %v", len(project.Plans))
				}
			default:
				t.Error("Unexpected project name")
			}
		}
	}

	TestRemove := func(t *testing.T) {
		d := newDemoCmd()
		d.SetArgs([]string{"remove", "--endpoint", "http://localhost:8081", "--demo-dir", "../demo"})
		d.Execute()

		params := operations.NewListProjectsParams()
		resp, err := d.Client.Operations.ListProjects(params, d.AuthInfo)
		if err != nil {
			t.Error("Didn't get a response trying to list projects")
		}
		if len(resp.Payload) != 0 {
			t.Errorf("Expected to receive no projects, got: %v", resp.Payload)
		}
	}

	// Initialize the practices - if they already exist this will do nothing
	r := newRootCmd()
	r.SetArgs([]string{"practices", "publish", "--practices-dir", "../practices", "--config", "../config.yaml"})
	r.Execute()

	t.Run("Load", TestLoad)
	t.Run("Remove", TestRemove)
}
