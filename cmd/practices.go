package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ThalesGroup/besec/lib"
	"github.com/ThalesGroup/besec/store"
)

// practicesCmd is a parent command that has no functionality of its own but maintains state for children
type practicesCmd struct {
	*cobra.Command
	store    store.Store
	parser   lib.PracticeParser
	versions []string // all of the practice version IDs, in lexicographically sorted order
}

const serviceAccountFlagName = "service-account"
const practicesDirFlagName = "practices-dir"

func newPracticesCmd(rc *rootCmd) *practicesCmd {
	mc := &practicesCmd{}

	mc.Command = &cobra.Command{
		Use:   "practices",
		Short: "Manage practices",
		Long:  `Validate local practice definitions and manage published ones`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			rc.PersistentPreRun(cmd, args)

			if cmd.Name() != "prune" {
				practicesDir := viper.GetString(practicesDirFlagName)
				schemaFile := viper.GetString("schema-file")
				if schemaFile == "" {
					schemaFile = filepath.Join(practicesDir, "schema.json")
				}
				mc.parser = lib.NewPracticeParser(practicesDir, schemaFile, nil)
			}

			if cmd.Name() != "check" {
				mc.store = initStore()
				checkEmulator()

				var err error
				mc.versions, err = mc.store.ListPracticesVersions(context.Background())
				if err != nil {
					log.Fatalf("Error retrieving versions: %v", err)
				}
			}
		},
	}

	mc.PersistentFlags().String(practicesDirFlagName, "./practices", "A directory containing practice definitions")
	err := viper.BindPFlag(practicesDirFlagName, mc.PersistentFlags().Lookup(practicesDirFlagName))
	if err != nil {
		log.Fatalf("Error binding viper flag: %v", err)
	}

	mc.PersistentFlags().String("schema-file", "", "The practice schema definitions, default is practicesdir/schema.json")
	err = viper.BindPFlag("schema-file", mc.PersistentFlags().Lookup("schema-file"))
	if err != nil {
		log.Fatalf("Error binding viper flag: %v", err)
	}

	mc.AddCommand(mc.newPracticesCheckCmd())
	mc.AddCommand(mc.newPracticesCompareCmd())
	mc.AddCommand(mc.newPracticesPublishCmd())
	mc.AddCommand(mc.newPracticesDeleteCmd())
	mc.AddCommand(mc.newPracticesPruneCmd())
	return mc
}

func (mc *practicesCmd) newPracticesCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check local practice format is correct",
		Long:  `Check practices are valid yaml and have the correct format. --exclude-practices is ignored.`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			_, err := mc.parser.ParsePracticesDir() // whilst we pass excludedPracticeIds, they are still parsed, and hence checked.
			if err != nil {
				log.Fatal(err)
			}
		},
	}
}

func (mc *practicesCmd) newPracticesCompareCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "compare [version]",
		Short: "Compare local practice definitions with published ones",
		Long:  "If no version is specified, it compares against the latest published definitions",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			version := "latest"
			if len(args) == 1 {
				version = args[0]
			}
			if !mc.match(version) {
				fmt.Printf("Local practices differ from remote practices version '%v'\n", version)
				os.Exit(1)
			}
			fmt.Printf("Local practices match remote practices version '%v'\n", version)
		},
	}
}

func (mc *practicesCmd) match(version string) bool {
	if version == "latest" {
		if len(mc.versions) == 0 {
			log.Info("No existing practice versions found")
			return false
		}
		version = mc.versions[len(mc.versions)-1]
	}

	remotePractices, err := mc.store.GetPractices(context.Background(), version)
	if err != nil {
		log.Fatalf("Error retrieving practices at version %v: %v", version, err)
	}

	localPractices, err := mc.parser.ParsePracticesDir()
	if err != nil {
		log.Fatal(err)
	}

	// They are equal regardless of what order they are in, so sort them first
	sort.Slice(remotePractices, func(i, j int) bool {
		return remotePractices[i].Name > remotePractices[j].Name
	})
	sort.Slice(localPractices, func(i, j int) bool {
		return localPractices[i].Name > localPractices[j].Name
	})
	return reflect.DeepEqual(remotePractices, localPractices)
}

func (mc *practicesCmd) newPracticesPublishCmd() *cobra.Command {
	pc := &cobra.Command{
		Use:   "publish",
		Short: "Publish the local practice definitions",
		Long:  "If they differ from the latest published version, upload the local practice definitions as a new version.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			version, duplicate := mc.getNextVersion()
			manualVersion, err := cmd.Flags().GetString("force-version")
			if (err == nil) && manualVersion != "" {
				log.Infof("Using version %v instead of default %v", manualVersion, version)
				version = manualVersion
			} else if duplicate {
				log.Warnf("There is already at least one version with todays date, creating as %v", version)
			}

			if manualVersion == "" && mc.match("latest") {
				log.Info("The local definitions match the latest published version (after parsing), aborting.")
				return
			}

			practices, err := mc.parser.ParsePracticesDir()
			if err != nil {
				log.Fatal(err)
			}

			err = mc.store.CreatePractices(context.Background(), version, practices)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Published local practices as", version)
		},
	}
	pc.PersistentFlags().String("force-version", "", "Manually specify a particular version identifier")
	return pc
}

// getNextVersion returns either today's date (YYYY-MM-DD), or if there is a version from that date,
// a suffixed form - YYYY-MM-DDrNN - where NN is one higher than the highest version seen, starting at 2.
// If a suffixed form is used, the bool return is True, otherwise it is false
// This is to attempt to avoid re-using previous names, even in the presence of version deletion.
// It isn't foolproof - users need to take care deleting published versions
func (mc *practicesCmd) getNextVersion() (string, bool) {
	date := time.Now().Format("2006-01-02")
	highestRev := 0
	for _, v := range mc.versions {
		if strings.HasPrefix(v, date) {
			if highestRev == 0 {
				highestRev = 1
			}
			split := strings.SplitAfter(v, "r")
			if len(split) == 2 {
				rev, err := strconv.Atoi(split[1])
				if err != nil {
					log.Fatal("Version found which doesn't meeting the pattern 'YYYY-MM-DD[rNN]'")
				}
				if rev > highestRev {
					highestRev = rev
				}
			}
		}
	}
	version := date
	if highestRev > 0 {
		version = fmt.Sprintf("%vr%v", date, highestRev+1)
	}

	return version, version != date
}

func (mc *practicesCmd) newPracticesDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete the specified practices version",
		Long:  "This will break any plans that used the specified version",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			version := args[0]
			found := false
			for _, v := range mc.versions {
				if v == version {
					found = true
				}
			}
			if !found {
				log.Fatalf("Version %v not found", version)
			}

			err := mc.store.DeletePractices(context.Background(), version)
			if err != nil {
				log.Fatal(err)
			}
			log.Infof("Deleted practice definitions version %v. Running serve processes may still have it cached for up to two minutes.", version)
		},
	}
}

func (mc *practicesCmd) newPracticesPruneCmd() *cobra.Command {
	pc := &cobra.Command{
		Use:   "prune",
		Short: "Delete old practice versions that aren't used in any plans",
		Long: `This may be slow, as it has to fetch every plan revision. The latest version is never deleted.

If you've only just published a new version, a user could be in the process of creating a plan using the 
previous version - deleting it underneath them will break their plan.`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			inUse := usedPractices(mc.versions, mc.store)

			// delete unused old ones
			deleted := false
			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				panic(err)
			}
			for i, v := range mc.versions {
				if i == len(mc.versions)-1 {
					// we're done - never want to delete the most recent version
					break
				}
				if !inUse[v] {
					if force || confirm(fmt.Sprintf("Delete unused practices version %v", v)) {
						log.Infof("Deleting unused practice definitions version %v. Running deployments may still have it cached for up to two minutes.", v)
						err := mc.store.DeletePractices(context.Background(), v)
						if err != nil {
							log.Fatal(err)
						}
						delete(inUse, v)
						deleted = true
					}
				}
			}

			if !deleted {
				log.Info("No practice versions were deleted")
			}

			// report which remaining practices are unused (if any)
			unUsed := []string{}
			for v, used := range inUse {
				if !used {
					unUsed = append(unUsed, v)
				}
			}
			sort.Strings(unUsed)
			if len(unUsed) > 0 {
				log.Infof("Remaining unused practices: %v", strings.Join(unUsed, ", "))
			} else {
				log.Infof("There are no unused practice versions.")
			}
		},
	}
	pc.PersistentFlags().Bool("force", false, "Don't prompt for confirmation prior to deletion")
	return pc
}

func usedPractices(versions []string, store store.Store) map[string]bool {
	inUse := make(map[string]bool)
	for _, v := range versions {
		inUse[v] = false
	}

	log.Info("Retrieving all plan revisions")
	projects, err := store.ListProjects(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, project := range projects {
		for _, planID := range project.Plans {
			revisionIDs, err := store.ListPlanRevisionIDs(context.Background(), planID)
			if err != nil {
				log.Fatal(err)
			}
			for _, revisionID := range revisionIDs {
				revision, found, err := store.GetPlanRevision(context.Background(), planID, revisionID)
				if err != nil {
					log.Fatal(err)
				}
				if !found {
					log.Fatalf("Couldn't find plan revision - it could have just been deleted, but that is unlikely. Aborting. Project %v Plan %v Revision %v", project.ID, planID, revisionID)
				}
				inUse[revision.Responses.PracticesVersion] = true
			}
		}
	}

	return inUse
}

// Prompt the user to confirm, defaulting to no
func confirm(prompt string) bool {
	var s string

	fmt.Printf(prompt + " (y/N): ")
	_, err := fmt.Scanln(&s)
	if err != nil {
		return false
	}

	s = strings.TrimSpace(s)
	s = strings.ToLower(s)

	if s == "y" || s == "yes" {
		return true
	}
	return false
}
