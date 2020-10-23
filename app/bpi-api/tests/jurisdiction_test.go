package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/appinesshq/bpi/app/bpi-api/handlers"
	"github.com/appinesshq/bpi/business/data/jurisdiction"
	"github.com/appinesshq/bpi/business/tests"
	"github.com/google/go-cmp/cmp"
)

// JurisdictionTests holds methods for each jurisdiction subtest. This type allows
// passing dependencies for tests while still providing a convenient syntax
// when subtests are registered.
type JurisdictionTests struct {
	app       http.Handler
	userToken string
}

// TestJurisdictions runs a series of tests to exercise Jurisdiction behavior from the
// API level. The subtests all share the same database and application for
// speed and convenience. The downside is the order the tests are ran matters
// and one test may break if other tests are not ran before it. If a particular
// subtest needs a fresh instance of the application it can make it or it
// should be its own Test* function.
func TestJurisdictions(t *testing.T) {
	test := tests.NewIntegration(t)
	t.Cleanup(test.Teardown)

	shutdown := make(chan os.Signal, 1)
	tests := JurisdictionTests{
		app:       handlers.API("develop", shutdown, test.Log, test.Auth, test.DB),
		userToken: test.Token(test.KID, "admin@example.com", "gophers"),
	}

	t.Run("getJurisdiction404", tests.getJurisdiction404)
	t.Run("getJurisdiction400", tests.getJurisdiction400)
	t.Run("putJurisdiction404", tests.putJurisdiction404)
	t.Run("crudJurisdictions", tests.crudJurisdictions)
}

// getJurisdiction400 validates a jurisdiction request for a malformed id.
func (ct *JurisdictionTests) getJurisdiction400(t *testing.T) {
	id := "QQQ"

	r := httptest.NewRequest(http.MethodGet, "/v1/jurisdictions/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a jurisdiction with a malformed id.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new jurisdiction %s.", testID, id)
		{
			if w.Code != http.StatusBadRequest {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 400 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 400 for the response.", tests.Success, testID)

			got := w.Body.String()
			exp := `{"error":"ID is not in its proper form"}`
			if got != exp {
				t.Logf("\t\tTest %d:\tGot : %v", testID, got)
				t.Logf("\t\tTest %d:\tExp: %v", testID, exp)
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// getJurisdiction404 validates a jurisdiction request for a jurisdiction that does not exist with the endpoint.
func (ct *JurisdictionTests) getJurisdiction404(t *testing.T) {
	id := "QQ.01"

	r := httptest.NewRequest(http.MethodGet, "/v1/jurisdictions/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a jurisdiction with an unknown id.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new jurisdiction %s.", testID, id)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 404 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 404 for the response.", tests.Success, testID)

			got := w.Body.String()
			exp := "not found"
			if !strings.Contains(got, exp) {
				t.Logf("\t\tTest %d:\tGot : %v", testID, got)
				t.Logf("\t\tTest %d:\tExp: %v", testID, exp)
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// putJurisdiction404 validates updating a jurisdiction that does not exist.
func (ct *JurisdictionTests) putJurisdiction404(t *testing.T) {
	id := "QQ.01"

	r := httptest.NewRequest(http.MethodPut, "/v1/jurisdictions/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to validate activating a jurisdiction that does not exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new jurisdiction %s.", testID, id)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 404 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 404 for the response.", tests.Success, testID)

			got := w.Body.String()
			exp := "not found"
			if !strings.Contains(got, exp) {
				t.Logf("\t\tTest %d:\tGot : %v", testID, got)
				t.Logf("\t\tTest %d:\tExp: %v", testID, exp)
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result.", tests.Failed, testID)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// crudJurisdictions performs a complete test of CRUD against the api.
func (ct *JurisdictionTests) crudJurisdictions(t *testing.T) {
	ct.putJurisdiction204(t, "LV.25")
	ct.getJurisdiction200(t, "LV.25")
}

// getJurisdiction200 validates a jurisdiction request for an existing id.
func (ct *JurisdictionTests) getJurisdiction200(t *testing.T, id string) {
	r := httptest.NewRequest(http.MethodGet, "/v1/jurisdictions/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a jurisdiction that exists.")
	{
		testID := 0
		t.Logf("\tTest : %d\tWhen using the new jurisdiction %s.", testID, id)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest : %d\tShould receive a status code of 200 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest : %d\tShould receive a status code of 200 for the response.", tests.Success, testID)

			var got jurisdiction.Info
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest : %d\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like Dates so we copy p.
			// (456173, 'LV.25', 'LV', 'Riga')
			exp := got
			exp.GNID = 456173
			exp.Code = "LV.25"
			exp.CountryCode = "LV"
			exp.Name = "Riga"

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest : %d\tShould get the expected result. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest : %d\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// putJurisdiction204 validates updating a jurisdiction that does exist.
func (ct *JurisdictionTests) putJurisdiction204(t *testing.T, id string) {
	// Activate country for this test
	{
		r := httptest.NewRequest(http.MethodPut, "/v1/countries/LV", nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ct.userToken)
		ct.app.ServeHTTP(w, r)
	}

	r := httptest.NewRequest(http.MethodPut, "/v1/jurisdictions/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to activate a jurisdiction with the jurisdictions endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the modified jurisdiction value.", testID)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Success, testID)

			r = httptest.NewRequest(http.MethodGet, "/v1/jurisdictions/"+id, nil)
			w = httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+ct.userToken)
			ct.app.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the retrieve : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the retrieve.", tests.Success, testID)

			var ci jurisdiction.Info
			if err := json.NewDecoder(w.Body).Decode(&ci); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			if ci.Name != "Riga" {
				t.Fatalf("\t%s\tTest %d:\tShould see activated Jurisdiction : got %q want %q", tests.Failed, testID, ci.Name, "Riga")
			}
			t.Logf("\t%s\tTest %d:\tShould see an activated Jurisdiction.", tests.Success, testID)
		}
	}
}
