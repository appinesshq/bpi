package jurisdiction_test

import (
	"context"
	"testing"
	"time"

	"github.com/appinesshq/bpi/business/auth"
	"github.com/appinesshq/bpi/business/data/jurisdiction"
	"github.com/appinesshq/bpi/business/tests"
	"github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
)

func seed(db *sqlx.DB) error {
	const q = `
	INSERT INTO countries (gnid, code, name, currency_code, currency_name, active) VALUES
	(453733, 'EE', 'Estonia', 'EUR', 'Euro', false),
	(458258, 'LV', 'Latvia', 'EUR', 'Euro', false),
	(597427, 'LT', 'Lithuania', 'EUR', 'Euro', true);

	INSERT INTO jurisdictions (gnid, code, country_code, name) VALUES
	(592170, 'EE.01', 'EE', 'Harjumaa'),
	(456173, 'LV.25', 'LV', 'Riga'),
	(864485, 'LT.65', 'LT', 'Vilnius');
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

	c := jurisdiction.New(log, db)

	t.Log("Given the need to work with Jurisdiction records.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single Jurisdiction.", testID)
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

			_, err := c.QueryByCode(ctx, traceID, claims, "LT.65")
			if err == nil {
				t.Fatalf("\t%s\tTest %d:\tShould not be able to retrieve inactive jurisdiction by Code.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould not be able to retrieve inactive jurisdiction by Code.", tests.Success, testID)

			if err := c.ToggleActive(ctx, traceID, claims, "LT.65"); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to activate inactive jurisdiction by Code: %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to activate inactive jurisdiction by Code.", tests.Success, testID)

			_, err = c.QueryByCode(ctx, traceID, claims, "LT.65")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve inactive jurisdiction by Code: %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve inactive jurisdiction by Code.", tests.Success, testID)

			jurisdictions, err := c.Query(ctx, traceID, 1, 100)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to query jurisdictions: %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to query jurisdictions.", tests.Success, testID)

			got, expected := len(jurisdictions), 1
			if got != expected {
				t.Fatalf("\t%s\tTest %d:\tShould get %d jurisdictions back, but got: %d.\n%+v", tests.Failed, testID, expected, got, jurisdictions)
			}
			t.Logf("\t%s\tTest %d:\tShould get %d jurisdictions back.", tests.Success, testID, expected)

			if err := c.ToggleActive(ctx, traceID, claims, "EE.01"); err == nil {
				t.Fatalf("\t%s\tTest %d:\tShould not be able to activate jurisdiction of inactive country.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould not be able to activate jurisdiction of inactive country.", tests.Success, testID)
		}
	}
}
