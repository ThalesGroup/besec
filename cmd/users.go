package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"firebase.google.com/go/v4/auth"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/iterator"

	"github.com/ThalesGroup/besec/api"
	"github.com/ThalesGroup/besec/api/models"
	"github.com/ThalesGroup/besec/store"
)

type uCmd struct {
	*cobra.Command
	store store.Store
	authc *auth.Client
	w     *tabwriter.Writer
}

const trustedDomainsFlagName = "trusted-domains"

func newUsersCmd(rc *rootCmd) *uCmd {
	uc := &uCmd{}

	uc.Command = &cobra.Command{Use: "users [flags] [UID] [UID]...",
		Short: "View, authorise, and remove users.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			rc.PersistentPreRun(cmd, args)

			tabw := new(tabwriter.Writer)
			tabw.Init(os.Stdout, 0, 8, 1, '\t', 0)
			uc.w = tabw

			uc.store = initStore()
			uc.authc = api.InitAuthClient(viper.GetString("gcp-project"), true, viper.GetString(serviceAccountFlagName))
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			uc.w.Flush()
		},
	}

	uc.PersistentFlags().StringSlice(trustedDomainsFlagName, []string{}, "A list of trusted domain names; warnings will be provided for users not from these domains.")
	err := viper.BindPFlag(trustedDomainsFlagName, uc.PersistentFlags().Lookup(trustedDomainsFlagName))
	if err != nil {
		log.Fatalf("Error binding viper flag: %v", err)
	}

	uc.AddCommand(uc.newListCmd())
	uc.AddCommand(uc.newAuthorizeCmd(true))
	uc.AddCommand(uc.newAuthorizeCmd(false))
	uc.AddCommand(uc.newRemoveCmd())

	return uc
}

func (uc *uCmd) newAuthorizeCmd(authorize bool) *cobra.Command {
	text := "authorize"
	short := "Manually authorize one or more users"
	if !authorize {
		text = "deauthorize"
		short = "Deauthorize one or more manually authorized users"
	}
	ac := &cobra.Command{
		Use:   text + " [UID] [UID]...",
		Short: short,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			force := false
			if authorize {
				var err error
				force, err = cmd.Flags().GetBool("force")
				if err != nil {
					panic(err)
				}
			}

			for _, uid := range args {
				uc.authorizeUser(uid, authorize, force)
			}
		},
	}
	if authorize {
		ac.PersistentFlags().Bool("force", false, "Allow users from untrusted domains")
	}
	return ac
}

func (uc *uCmd) newRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove [UID] [UID]...",
		Short: "Remove one or more users from the system",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			for _, uid := range args {
				uc.removeUser(uid)
			}
		},
	}
}

func (uc *uCmd) newListCmd() *cobra.Command {
	return &cobra.Command{Use: "list [UID] [UID] ...",
		Short: "List all users and their local records",
		Long:  "Show the specified users or - if none are listed - all the users in the system.",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				for _, uid := range args {
					// Look up the user in Firebase
					record, err := uc.authc.GetUser(context.Background(), uid)
					if err != nil {
						log.Fatalf("Error retrieving user: %v", err)
					}
					uc.printUserHeader()
					uc.showUser(record)
				}
			} else {
				users, err := uc.getUsers()
				if err != nil {
					log.Fatalf("Error retrieving users: %v", err)
				}
				uc.printUserHeader()
				for _, user := range users {
					uc.showUser(user)
				}
			}
		},
	}
}

// authorizeUser sets the user's authorization state to the value of the authorize parameter
func (uc *uCmd) authorizeUser(uid string, authorize bool, force bool) {
	authorized, err := api.GetManuallyAuthorized(context.Background(), uc.authc, uc.store, uid)
	if err != nil {
		log.Fatalf("Error checking current authorized status: %v", err)
	}

	if authorize == authorized {
		not := ""
		if !authorize {
			not = "not "
		}
		fmt.Printf("User %s is already %vauthorized, skipping\n", uid, not)
		return
	}

	record, err := uc.authc.GetUser(context.Background(), uid)
	if err != nil {
		log.Fatalf("Error retrieving user: %v", err)
	}
	_, record, err = uc.lookupUser(record)
	if err != nil {
		log.Warnf("No local data available for %v", record.UID)
	}
	if authorize && !userDomainWhitelisted(record.Email) {
		if force {
			fmt.Printf("Authorizing %s from untrusted domain\n", record.Email)
		} else {
			fmt.Printf("%s's email - '%s' - is not on the trusted-domains list, aborting. Use --force to authorize anyway.\n", record.UID, record.Email)
			return
		}
	}

	err = api.SetManuallyAuthorized(context.Background(), uc.authc, uc.store, uid, authorize)
	if err != nil {
		log.Fatalf("Error saving authorized state: %v", err)
	} else {
		action := "Authorized "
		if !authorize {
			action = "De-authorized "
		}
		fmt.Println(action + record.DisplayName)
	}
}

// Delete the user from Firebase Auth
func (uc *uCmd) removeUser(uid string) {
	err := uc.authc.DeleteUser(context.Background(), uid)
	if err != nil {
		fmt.Printf("Error deleting %s from Firebase: %v\n", uid, err)
		return
	}

	user := &models.User{UID: uid, LocalData: nil}
	err = uc.store.SaveUserData(context.Background(), user)
	if err != nil {
		fmt.Printf("Error deleting %s from the store (warning: already deleted from Firebase): %v\n", uid, err)
		return
	}

	fmt.Printf("Deleted %s\n", uid)
}

func (uc *uCmd) getUsers() ([]*auth.UserRecord, error) {
	recs := []*auth.UserRecord{}
	it := uc.authc.Users(context.Background(), "")
	for {
		user, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return recs, err
		}
		recs = append(recs, user.UserRecord)
	}
	return recs, nil
}

func (uc *uCmd) showUser(r *auth.UserRecord) {
	user, r, err := uc.lookupUser(r)
	if err != nil {
		log.Warn(err)
		// carry on with a null User
	}

	verifiedWarning := ""
	if !r.EmailVerified && !api.IsSAMLProvider(r.ProviderUserInfo[0].ProviderID) {
		verifiedWarning = "[WARNING: unverified email]"
	}

	disabled := ""
	if r.Disabled {
		disabled = "[disabled]"
	}

	localText := ""
	if user.ManuallyAuthorized {
		localText = "[manually authorized]"
	}

	provider := ""
	if len(r.ProviderUserInfo) == 1 {
		provider = r.ProviderUserInfo[0].ProviderID
	} else {
		provider = "multiple"
	}

	domainWarning := ""
	if !userDomainWhitelisted(r.Email) {
		domainWarning = "[WARNING: this domain is not on the whitelist]"
	}

	fmt.Fprintf(uc.w, "%s\t%s\t'%s'\t%s\t%s%s%s%s\n", r.UID, r.Email, r.DisplayName, provider, domainWarning, verifiedWarning, disabled, localText)
	if len(r.ProviderUserInfo) != 1 {
		for _, providerInfo := range r.ProviderUserInfo {
			// don't use the tabwriter here, this follows a different alignment
			fmt.Printf("\t\t%s: %s <%s> '%s'\n", providerInfo.ProviderID, providerInfo.UID, providerInfo.Email, providerInfo.DisplayName)
		}
	}
}

func (uc *uCmd) printUserHeader() {
	fmt.Fprint(uc.w, "UID\tEmail\tDisplay name\tProvider\tStatus\n")
}

func userDomainWhitelisted(email string) bool {
	mailComponents := strings.Split(email, "@")
	domain := mailComponents[len(mailComponents)-1]
	for _, whiteListed := range viper.GetStringSlice(trustedDomainsFlagName) {
		if domain == whiteListed {
			return true
		}
	}
	return false
}

// lookupUser returns a User model corresponding to the record, and also a copy of record with any local data applied
// if err is non-nil, the lookup failed, but an empty user model and copy of the record are still returned
func (uc *uCmd) lookupUser(record *auth.UserRecord) (*models.User, *auth.UserRecord, error) {
	newRecord := &auth.UserRecord{}
	*newRecord = *record

	// Look up the user in the database
	user := &models.User{UID: record.UID}
	if err := uc.store.GetUserData(context.Background(), user); err != nil {
		return user, newRecord, fmt.Errorf("Error retrieving local user data: %v", err)
	}

	if user.LocalData != nil {
		if newRecord.Email == "" {
			newRecord.Email = user.LocalData.Email
		}

		if newRecord.DisplayName == "" {
			newRecord.DisplayName = user.LocalData.Name
		}
	}

	return user, newRecord, nil
}
