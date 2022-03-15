package api

import (
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/ThalesGroup/besec/api/models"
	"github.com/ThalesGroup/besec/store"
)

func TestMakeAuthorizer(t *testing.T) {
	if _, ok := os.LookupEnv("FIRESTORE_EMULATOR_HOST"); !ok {
		t.Error("TestMakeAuthorizer requires the firestore emulator")
	}

	st := store.NewFireStore("emulator")
	white := "whitelisted.provider"
	some := "some"
	other := "othe"
	rt := &Runtime{Store: st,
		AuthConfig: NewExtendedAuthConfig(models.AuthConfig{Providers: models.AuthProviders{&models.AuthProvider{ID: &some},
			&models.AuthProvider{ID: &white, Whitelisted: true},
			&models.AuthProvider{ID: &other}}})}

	uOther := models.User{UID: "a", Provider: "p"}
	uWhitelist := models.User{UID: "a", Provider: white}
	uDouble := models.User{UID: "Totoro", Provider: white}

	uExplicit := models.User{UID: "Totoro", Provider: "p"}
	uExplicit.LocalData = &models.LocalUserData{ManuallyAuthorized: true}
	err := st.SaveUserData(context.Background(), &uExplicit)
	if err != nil {
		t.Errorf("Failed to save user data: %v", err)
		return
	}

	type tcase = struct {
		user models.User
		want bool
	}
	cases := []tcase{
		{uOther, false},
		{uWhitelist, true},
		{uExplicit, true},
		{uDouble, true},
	}

	auth := MakeAuthorizer(rt)
	for _, c := range cases {
		got := auth.Authorize(&http.Request{URL: &url.URL{Path: "/path"}}, &c.user)
		if got == nil && !c.want {
			t.Errorf("User %v was incorrectly authorized", c.user)
		} else if got != nil && c.want {
			t.Errorf("User %v was incorrectly denied", c.user)
		}
	}
}
