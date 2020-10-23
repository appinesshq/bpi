package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/appinesshq/bpi/app/bpi-api/handlers"
	"github.com/appinesshq/bpi/business/data/country"
	"github.com/appinesshq/bpi/business/tests"
	"github.com/google/go-cmp/cmp"
)

// CountryTests holds methods for each country subtest. This type allows
// passing dependencies for tests while still providing a convenient syntax
// when subtests are registered.
type CountryTests struct {
	app       http.Handler
	userToken string
}

// TestCountries runs a series of tests to exercise Country behavior from the
// API level. The subtests all share the same database and application for
// speed and convenience. The downside is the order the tests are ran matters
// and one test may break if other tests are not ran before it. If a particular
// subtest needs a fresh instance of the application it can make it or it
// should be its own Test* function.
func TestCountries(t *testing.T) {
	test := tests.NewIntegration(t)
	t.Cleanup(test.Teardown)

	shutdown := make(chan os.Signal, 1)
	tests := CountryTests{
		app:       handlers.API("develop", shutdown, test.Log, test.Auth, test.DB),
		userToken: test.Token(test.KID, "admin@example.com", "gophers"),
	}

	t.Run("getCountry404", tests.getCountry404)
	t.Run("getCountry400", tests.getCountry400)
	t.Run("putCountry404", tests.putCountry404)
	t.Run("crudCountries", tests.crudCountry)
}

// getCountry400 validates a country request for a malformed id.
func (ct *CountryTests) getCountry400(t *testing.T) {
	id := "QQQ"

	r := httptest.NewRequest(http.MethodGet, "/v1/countries/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a country with a malformed id.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new country %s.", testID, id)
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

// getCountry404 validates a country request for a country that does not exist with the endpoint.
func (ct *CountryTests) getCountry404(t *testing.T) {
	id := "QQ"

	r := httptest.NewRequest(http.MethodGet, "/v1/countries/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a country with an unknown id.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new country %s.", testID, id)
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

// putCountry404 validates updating a country that does not exist.
func (ct *CountryTests) putCountry404(t *testing.T) {
	id := "QQ"

	r := httptest.NewRequest(http.MethodPut, "/v1/countries/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to validate activating a country that does not exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new country %s.", testID, id)
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

// crudCountry performs a complete test of CRUD against the api.
func (ct *CountryTests) crudCountry(t *testing.T) {
	ct.putCountry204(t, "LT")
	ct.getCountry200(t, "LT")
}

// getCountry200 validates a country request for an existing id.
func (ct *CountryTests) getCountry200(t *testing.T, id string) {
	r := httptest.NewRequest(http.MethodGet, "/v1/countries/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a country that exists.")
	{
		testID := 0
		t.Logf("\tTest : %d\tWhen using the new country %s.", testID, id)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest : %d\tShould receive a status code of 200 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest : %d\tShould receive a status code of 200 for the response.", tests.Success, testID)

			var got country.Info
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest : %d\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like Dates so we copy p.
			// (597427, 'LT', 'Lithuania', 'EUR', 'Euro')
			exp := got
			exp.ID = 597427
			exp.CountryCode = "LT"
			exp.Name = "Lithuania"
			exp.CurrencyCode = "EUR"
			exp.CurrencyName = "Euro"

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest : %d\tShould get the expected result. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest : %d\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// putCountry204 validates updating a country that does exist.
func (ct *CountryTests) putCountry204(t *testing.T, id string) {
	r := httptest.NewRequest(http.MethodPut, "/v1/countries/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to activate a country with the countries endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the modified country value.", testID)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Success, testID)

			r = httptest.NewRequest(http.MethodGet, "/v1/countries/"+id, nil)
			w = httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+ct.userToken)
			ct.app.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the retrieve : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the retrieve.", tests.Success, testID)

			var ci country.Info
			if err := json.NewDecoder(w.Body).Decode(&ci); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			if ci.Name != "Lithuania" {
				t.Fatalf("\t%s\tTest %d:\tShould see an updated Name : got %q want %q", tests.Failed, testID, ci.Name, "Lithuania")
			}
			t.Logf("\t%s\tTest %d:\tShould see an updated Name.", tests.Success, testID)
		}
	}
}
