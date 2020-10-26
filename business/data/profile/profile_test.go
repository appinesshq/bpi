package profile_test

import (
	"context"
	"testing"
	"time"

	"github.com/appinesshq/bpi/business/auth"
	"github.com/appinesshq/bpi/business/data/profile"
	"github.com/appinesshq/bpi/business/tests"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

func TestProfile(t *testing.T) {
	log, db, teardown := tests.NewUnit(t)
	t.Cleanup(teardown)

	p := profile.New(log, db)

	t.Log("Given the need to work with Profile records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single Profile.", testID)
		{
			ctx := context.Background()
			now := time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
			traceID := "00000000-0000-0000-0000-000000000000"

			claims := auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "MB Appiness Solutions",
					Subject:   "718ffbea-f4a1-4667-8ae3-b349da52675e",
					Audience:  "users",
					ExpiresAt: now.Add(time.Hour).Unix(),
					IssuedAt:  now.Unix(),
				},
				Roles: []string{auth.RoleAdmin, auth.RoleUser},
			}

			np := profile.NewProfile{
				Name:        "test",
				DisplayName: "Test profile",
			}

			prf, err := p.Create(ctx, traceID, claims, np, now)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create a profile : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create a profile.", tests.Success, testID)

			saved, err := p.QueryByName(ctx, traceID, prf.Name)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve profile by Name: %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve profile by Name.", tests.Success, testID)

			if diff := cmp.Diff(prf, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same profile. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same profile.", tests.Success, testID)

			upd := profile.UpdateProfile{
				DisplayName: tests.StringPointer("My profile"),
			}
			updatedTime := time.Date(2019, time.January, 1, 1, 1, 1, 0, time.UTC)

			if err := p.Update(ctx, traceID, claims, prf.Name, upd, updatedTime); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update profile : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update profile.", tests.Success, testID)

			profiles, err := p.Query(ctx, traceID, 1, 1)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve updated profile : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve updated profile.", tests.Success, testID)

			// Check specified fields were updated. Make a copy of the original profile
			// and change just the fields we expect then diff it with what was saved.
			want := prf
			want.DisplayName = *upd.DisplayName
			want.DateUpdated = updatedTime

			if diff := cmp.Diff(want, profiles[0]); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same profile. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same profile.", tests.Success, testID)

			upd = profile.UpdateProfile{
				DisplayName: tests.StringPointer("My profile"),
			}

			if err := p.Update(ctx, traceID, claims, prf.Name, upd, updatedTime); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update just some fields of profile : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update just some fields of profile.", tests.Success, testID)

			saved, err = p.QueryByName(ctx, traceID, prf.Name)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve updated profile : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve updated profile.", tests.Success, testID)

			if saved.DisplayName != *upd.DisplayName {
				t.Fatalf("\t%s\tTest %d:\tShould be able to see updated DisplayName field : got %q want %q.", tests.Failed, testID, saved.DisplayName, *upd.DisplayName)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updated DisplayName field.", tests.Success, testID)
			}

			if err := p.Delete(ctx, traceID, prf.Name); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to delete profile : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to delete profile.", tests.Success, testID)

			_, err = p.QueryByName(ctx, traceID, prf.Name)
			if errors.Cause(err) != profile.ErrNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould NOT be able to retrieve deleted profile : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould NOT be able to retrieve deleted profile.", tests.Success, testID)
		}
	}
}
