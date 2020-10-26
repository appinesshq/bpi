package category_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/appinesshq/bpi/business/auth"
	"github.com/appinesshq/bpi/business/data/category"
	"github.com/appinesshq/bpi/business/tests"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

func TestCategory(t *testing.T) {
	log, db, teardown := tests.NewUnit(t)
	t.Cleanup(teardown)

	p := category.New(log, db)

	t.Log("Given the need to work with Category records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single Category.", testID)
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

			nc := category.NewCategory{
				Slug: "test",
				Name: "Test category",
			}

			cat, err := p.Create(ctx, traceID, claims, nc, now)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create a category : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create a category.", tests.Success, testID)

			saved, err := p.QueryByID(ctx, traceID, cat.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve category by ID: %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve category by ID.", tests.Success, testID)

			if diff := cmp.Diff(cat, saved); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same category. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same category.", tests.Success, testID)

			parentID := ""
			{
				npc := category.NewCategory{
					Slug: "parent",
					Name: "Parent category",
				}

				parent, err := p.Create(ctx, traceID, claims, npc, now)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to create a parent category : %s.", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to create a parent category.", tests.Success, testID)
				parentID = parent.ID
			}

			upd := category.UpdateCategory{
				Name:     tests.StringPointer("Testing"),
				ParentID: tests.StringPointer(parentID),
			}
			updatedTime := time.Date(2019, time.January, 1, 1, 1, 1, 0, time.UTC)

			if err := p.Update(ctx, traceID, claims, cat.ID, upd, updatedTime); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update category : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update category.", tests.Success, testID)

			categories, err := p.Query(ctx, traceID, 2, 1)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve updated category : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve updated category.", tests.Success, testID)

			// Check specified fields were updated. Make a copy of the original category
			// and change just the fields we expect then diff it with what was saved.
			want := cat
			want.Name = *upd.Name
			want.ParentID = sql.NullString{String: parentID, Valid: true} // *upd.ParentIDs
			want.DateUpdated = updatedTime

			if diff := cmp.Diff(want, categories[0]); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same category. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same category.", tests.Success, testID)

			upd = category.UpdateCategory{
				Name: tests.StringPointer("Testing category"),
			}

			if err := p.Update(ctx, traceID, claims, cat.ID, upd, updatedTime); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to update just some fields of category : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to update just some fields of category.", tests.Success, testID)

			saved, err = p.QueryByID(ctx, traceID, cat.ID)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve updated category : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve updated category.", tests.Success, testID)

			if saved.Name != *upd.Name {
				t.Fatalf("\t%s\tTest %d:\tShould be able to see updated Name field : got %q want %q.", tests.Failed, testID, saved.Name, *upd.Name)
			} else {
				t.Logf("\t%s\tTest %d:\tShould be able to see updated Name field.", tests.Success, testID)
			}

			if err := p.Delete(ctx, traceID, cat.ID); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to delete category : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to delete category.", tests.Success, testID)

			_, err = p.QueryByID(ctx, traceID, cat.ID)
			if errors.Cause(err) != category.ErrNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould NOT be able to retrieve deleted category : %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould NOT be able to retrieve deleted category.", tests.Success, testID)
		}
	}
}
