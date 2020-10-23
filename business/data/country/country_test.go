package country_test

import (
	"context"
	"testing"
	"time"

	"github.com/appinesshq/bpi/business/auth"
	"github.com/appinesshq/bpi/business/data/country"
	"github.com/appinesshq/bpi/business/tests"
	"github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
)

func seed(db *sqlx.DB) error {
	const q = `
	INSERT INTO countries (gnid, code, name, currency_code, currency_name) VALUES
	(453733, 'EE', 'Estonia', 'EUR', 'Euro'),
	(458258, 'LV', 'Latvia', 'EUR', 'Euro'),
	(597427, 'LT', 'Lithuania', 'EUR', 'Euro')
	`
	if _, err := db.ExecContext(context.Background(), q); err != nil {
		return err
	}

	return nil
}

func TestCountry(t *testing.T) {
	log, db, teardown := tests.NewUnit(t)
	t.Cleanup(teardown)

	if err := seed(db); err != nil {
		t.Fatalf("Couldn't seed database: %v", err)
	}

	c := country.New(log, db)

	t.Log("Given the need to work with Country records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single Country.", testID)
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

			// _, err := c.QueryByID(ctx, traceID, claims, 597427)
			// if err == nil {
			// 	t.Fatalf("\t%s\tTest %d:\tShould not be able to retrieve inactive country by ID.", tests.Failed, testID)
			// }
			// t.Logf("\t%s\tTest %d:\tShould not be able to retrieve inactive country by ID.", tests.Success, testID)

			_, err := c.QueryByCode(ctx, traceID, claims, "LT")
			if err == nil {
				t.Fatalf("\t%s\tTest %d:\tShould not be able to retrieve inactive country by Code.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould not be able to retrieve inactive country by Code.", tests.Success, testID)

			if err := c.ToggleActive(ctx, traceID, claims, "LT"); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to activate inactive country by Code: %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to activate inactive country by Code.", tests.Success, testID)

			// country, err := c.QueryByID(ctx, traceID, claims, 597427)
			// if err != nil {
			// 	t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve active country by ID: %s.", tests.Failed, testID, err)
			// }
			// t.Logf("\t%s\tTest %d:\tShould be able to retrieve active country by ID.", tests.Success, testID)

			_, err = c.QueryByCode(ctx, traceID, claims, "LT")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve active country by Code: %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve active country by Code.", tests.Success, testID)

			countries, err := c.Query(ctx, traceID, 1, 100)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to query countries: %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to query countries.", tests.Success, testID)

			got, expected := len(countries), 1
			if got != expected {
				t.Fatalf("\t%s\tTest %d:\tShould get %d countries back, but got: %d.\n%+v", tests.Failed, testID, expected, got, countries)
			}
			t.Logf("\t%s\tTest %d:\tShould get %d countries back.", tests.Success, testID, expected)
		}
	}
}
