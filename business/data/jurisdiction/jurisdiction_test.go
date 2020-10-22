package jurisdiction_test

import (
	"context"
	"testing"
	"time"

	"github.com/appinesshq/bpi/business/auth"
	"github.com/appinesshq/bpi/business/data/jurisdiction"
	"github.com/appinesshq/bpi/business/tests"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
)

func seed(db *sqlx.DB) error {
	const q = `
	INSERT INTO countries (country_id, code, name, currency_code, currency_name, active) VALUES
	(453733, 'EE', 'Estonia', 'EUR', 'Euro', false),
	(458258, 'LV', 'Latvia', 'EUR', 'Euro', false),
	(597427, 'LT', 'Lithuania', 'EUR', 'Euro', true);

	INSERT INTO jurisdictions (jurisdiction_id, code, country_code, name) VALUES
	(587448, 'EE.21', 'EE', 'Võrumaa'),
	(587590, 'EE.20', 'EE', 'Viljandimaa'),
	(587875, 'EE.19', 'EE', 'Valgamaa'),
	(588334, 'EE.18', 'EE', 'Tartu'),
	(588879, 'EE.14', 'EE', 'Saare'),
	(589115, 'EE.13', 'EE', 'Raplamaa'),
	(589373, 'EE.12', 'EE', 'Põlvamaa'),
	(589576, 'EE.11', 'EE', 'Pärnumaa'),
	(590854, 'EE.08', 'EE', 'Lääne-Virumaa'),
	(590856, 'EE.07', 'EE', 'Lääne'),
	(591901, 'EE.05', 'EE', 'Jõgevamaa'),
	(591961, 'EE.04', 'EE', 'Järvamaa'),
	(592075, 'EE.03', 'EE', 'Ida-Virumaa'),
	(592133, 'EE.02', 'EE', 'Hiiumaa'),
	(592170, 'EE.01', 'EE', 'Harjumaa'),
	(454307, 'LV.33', 'LV', 'Ventspils Rajons'),
	(454311, 'LV.32', 'LV', 'Ventspils'),
	(454564, 'LV.31', 'LV', 'Valmiera'),
	(454571, 'LV.30', 'LV', 'Valka'),
	(454771, 'LV.29', 'LV', 'Tukuma novads'),
	(454968, 'LV.28', 'LV', 'Talsi Municipality'),
	(455888, 'LV.27', 'LV', 'Saldus Rajons'),
	(456173, 'LV.25', 'LV', 'Riga'),
	(456197, 'LV.24', 'LV', 'Rēzeknes Novads'),
	(456203, 'LV.23', 'LV', 'Rēzekne'),
	(456528, 'LV.22', 'LV', 'Preiļu novads'),
	(457061, 'LV.21', 'LV', 'Ogre'),
	(457712, 'LV.20', 'LV', 'Madona Municipality'),
	(457773, 'LV.19', 'LV', 'Ludzas novads'),
	(457889, 'LV.18', 'LV', 'Limbažu novads'),
	(457955, 'LV.16', 'LV', 'Liepāja'),
	(458459, 'LV.15', 'LV', 'Kuldīgas novads'),
	(458621, 'LV.14', 'LV', 'Krāslavas novads'),
	(459202, 'LV.13', 'LV', 'Jūrmala'),
	(459278, 'LV.12', 'LV', 'Jelgavas novads'),
	(459281, 'LV.11', 'LV', 'Jelgava'),
	(459282, 'LV.10', 'LV', 'Jēkabpils Municipality'),
	(459664, 'LV.09', 'LV', 'Gulbenes novads'),
	(460311, 'LV.08', 'LV', 'Dobeles novads'),
	(460410, 'LV.07', 'LV', 'Daugavpils municipality'),
	(460414, 'LV.06', 'LV', 'Daugavpils'),
	(460569, 'LV.05', 'LV', 'Cēsu Rajons'),
	(461112, 'LV.04', 'LV', 'Bauskas Rajons'),
	(461160, 'LV.03', 'LV', 'Balvu Novads'),
	(461525, 'LV.02', 'LV', 'Alūksnes Novads'),
	(461613, 'LV.01', 'LV', 'Aizkraukles novads'),
	(7628298, 'LV.60', 'LV', 'Dundaga'),
	(7628299, 'LV.40', 'LV', 'Alsunga'),
	(7628300, 'LV.A5', 'LV', 'Pāvilostas'),
	(7628301, 'LV.99', 'LV', 'Nīca'),
	(7628302, 'LV.B6', 'LV', 'Rucavas'),
	(7628303, 'LV.65', 'LV', 'Grobiņa'),
	(7628304, 'LV.61', 'LV', 'Durbe'),
	(7628305, 'LV.37', 'LV', 'Aizpute'),
	(7628306, 'LV.A8', 'LV', 'Priekule'),
	(7628307, 'LV.D7', 'LV', 'Vaiņode'),
	(7628308, 'LV.C9', 'LV', 'Skrunda'),
	(7628309, 'LV.51', 'LV', 'Brocēni'),
	(7628310, 'LV.B4', 'LV', 'Rojas'),
	(7628311, 'LV.77', 'LV', 'Kandava'),
	(7628312, 'LV.44', 'LV', 'Auces'),
	(7628313, 'LV.73', 'LV', 'Jaunpils'),
	(7628314, 'LV.62', 'LV', 'Engure'),
	(7628315, 'LV.D5', 'LV', 'Tērvete'),
	(7628316, 'LV.A3', 'LV', 'Ozolnieku'),
	(7628317, 'LV.B9', 'LV', 'Rundāles'),
	(7628318, 'LV.45', 'LV', 'Babīte'),
	(7628319, 'LV.95', 'LV', 'Mārupe'),
	(7628320, 'LV.A2', 'LV', 'Olaine'),
	(7628321, 'LV.67', 'LV', 'Lecava'),
	(7628322, 'LV.80', 'LV', 'Ķekava'),
	(7628323, 'LV.C3', 'LV', 'Salaspils'),
	(7628324, 'LV.46', 'LV', 'Baldone'),
	(7628325, 'LV.D2', 'LV', 'Stopiņi'),
	(7628326, 'LV.53', 'LV', 'Carnikava'),
	(7628327, 'LV.34', 'LV', 'Ādaži'),
	(7628328, 'LV.64', 'LV', 'Garkalne'),
	(7628329, 'LV.E4', 'LV', 'Vecumnieki'),
	(7628330, 'LV.79', 'LV', 'Ķegums'),
	(7628331, 'LV.87', 'LV', 'Lielvārde'),
	(7628332, 'LV.C8', 'LV', 'Skrīveri'),
	(7628333, 'LV.71', 'LV', 'Jaunjelgava'),
	(7628334, 'LV.98', 'LV', 'Nereta'),
	(7628335, 'LV.E6', 'LV', 'Viesīte'),
	(7628336, 'LV.C2', 'LV', 'Salas'),
	(7628337, 'LV.74', 'LV', 'Jēkabpils'),
	(7628338, 'LV.38', 'LV', 'Aknīste'),
	(7628339, 'LV.69', 'LV', 'Ilūkste'),
	(7628340, 'LV.E2', 'LV', 'Vārkava'),
	(7628341, 'LV.90', 'LV', 'Līvāni'),
	(7628342, 'LV.E1', 'LV', 'Varakļāni'),
	(7628343, 'LV.E8', 'LV', 'Vilanu'),
	(7628344, 'LV.B3', 'LV', 'Riebiņu'),
	(7628345, 'LV.35', 'LV', 'Aglona'),
	(7628346, 'LV.56', 'LV', 'Cibla'),
	(7628347, 'LV.E9', 'LV', 'Zilupes'),
	(7628348, 'LV.E7', 'LV', 'Viļaka'),
	(7628349, 'LV.47', 'LV', 'Baltinava'),
	(7628350, 'LV.57', 'LV', 'Dagda'),
	(7628351, 'LV.78', 'LV', 'Karsava'),
	(7628352, 'LV.B7', 'LV', 'Rugāju'),
	(7628353, 'LV.55', 'LV', 'Cesvaine'),
	(7628354, 'LV.91', 'LV', 'Lubāna'),
	(7628355, 'LV.85', 'LV', 'Krustpils'),
	(7628356, 'LV.A6', 'LV', 'Pļaviņu'),
	(7628357, 'LV.82', 'LV', 'Koknese'),
	(7628358, 'LV.68', 'LV', 'Ikšķile'),
	(7628359, 'LV.B5', 'LV', 'Ropažu'),
	(7628360, 'LV.70', 'LV', 'Inčukalns'),
	(7628361, 'LV.84', 'LV', 'Krimulda'),
	(7628362, 'LV.C7', 'LV', 'Sigulda'),
	(7628363, 'LV.88', 'LV', 'Līgatne'),
	(7628364, 'LV.94', 'LV', 'Mālpils'),
	(7628365, 'LV.C6', 'LV', 'Sēja'),
	(7628366, 'LV.C5', 'LV', 'Saulkrastu'),
	(7628367, 'LV.C1', 'LV', 'Salacgrīvas'),
	(7628368, 'LV.39', 'LV', 'Aloja'),
	(7628369, 'LV.97', 'LV', 'Naukšēni'),
	(7628370, 'LV.B8', 'LV', 'Rūjienas'),
	(7628371, 'LV.96', 'LV', 'Mazsalaca'),
	(7628372, 'LV.52', 'LV', 'Burtnieki'),
	(7628373, 'LV.A4', 'LV', 'Pārgaujas'),
	(7628374, 'LV.81', 'LV', 'Kocēni'),
	(7628375, 'LV.42', 'LV', 'Amatas'),
	(7628376, 'LV.A9', 'LV', 'Priekuļi'),
	(7628377, 'LV.B1', 'LV', 'Raunas'),
	(7628378, 'LV.D3', 'LV', 'Strenči'),
	(7628379, 'LV.50', 'LV', 'Beverīna'),
	(7628380, 'LV.D1', 'LV', 'Smiltene'),
	(7628381, 'LV.72', 'LV', 'Jaunpiebalga'),
	(7628382, 'LV.63', 'LV', 'Ērgļi'),
	(7628383, 'LV.E3', 'LV', 'Vecpiebalga'),
	(7628384, 'LV.43', 'LV', 'Apes'),
	(8299767, 'LV.F1', 'LV', 'Mesraga'),
	(864389, 'LT.56', 'LT', 'Alytus'),
	(864477, 'LT.57', 'LT', 'Kaunas'),
	(864478, 'LT.58', 'LT', 'Klaipėda County'),
	(864479, 'LT.59', 'LT', 'Marijampolė County'),
	(864480, 'LT.60', 'LT', 'Panevėžys'),
	(864481, 'LT.61', 'LT', 'Siauliai'),
	(864482, 'LT.62', 'LT', 'Tauragė County'),
	(864483, 'LT.63', 'LT', 'Telsiai'),
	(864484, 'LT.64', 'LT', 'Utena'),
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

			_, err := c.QueryByID(ctx, traceID, claims, 864485)
			if err == nil {
				t.Fatalf("\t%s\tTest %d:\tShould not be able to retrieve inactive jurisdiction by ID.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould not be able to retrieve inactive jurisdiction by ID.", tests.Success, testID)

			_, err = c.QueryByCode(ctx, traceID, claims, "LT.65")
			if err == nil {
				t.Fatalf("\t%s\tTest %d:\tShould not be able to retrieve inactive jurisdiction by Code.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould not be able to retrieve inactive jurisdiction by Code.", tests.Success, testID)

			if err := c.ToggleActivate(ctx, traceID, claims, "LT.65"); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to activate inactive jurisdiction by Code: %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to activate inactive jurisdiction by Code.", tests.Success, testID)

			jurisdiction, err := c.QueryByID(ctx, traceID, claims, 864485)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve active jurisdiction by ID: %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve active jurisdiction by ID.", tests.Success, testID)

			jurisdiction2, err := c.QueryByCode(ctx, traceID, claims, "LT.65")
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to retrieve inactive jurisdiction by Code: %s.", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to retrieve inactive jurisdiction by Code.", tests.Success, testID)

			if diff := cmp.Diff(jurisdiction, jurisdiction2); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get back the same jurisdiction. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get back the same jurisdiction.", tests.Success, testID)

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

			if err := c.ToggleActivate(ctx, traceID, claims, "EE.01"); err == nil {
				t.Fatalf("\t%s\tTest %d:\tShould not be able to activate jurisdiction of inactive country.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould not be able to activate jurisdiction of inactive country.", tests.Success, testID)
		}
	}
}
