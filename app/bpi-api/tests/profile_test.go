package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/appinesshq/bpi/app/bpi-api/handlers"
	"github.com/appinesshq/bpi/business/data/profile"
	"github.com/appinesshq/bpi/business/tests"
	"github.com/appinesshq/bpi/foundation/web"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// ProfileTests holds methods for each profile subtest. This type allows
// passing dependencies for tests while still providing a convenient syntax
// when subtests are registered.
type ProfileTests struct {
	app       http.Handler
	userToken string
}

// TestProfiles runs a series of tests to exercise Profile behavior from the
// API level. The subtests all share the same database and application for
// speed and convenience. The downside is the order the tests are ran matters
// and one test may break if other tests are not ran before it. If a particular
// subtest needs a fresh instance of the application it can make it or it
// should be its own Test* function.
func TestProfiles(t *testing.T) {
	test := tests.NewIntegration(t)
	t.Cleanup(test.Teardown)

	shutdown := make(chan os.Signal, 1)
	tests := ProfileTests{
		app:       handlers.API("develop", shutdown, test.Log, test.Auth, test.DB),
		userToken: test.Token(test.KID, "admin@example.com", "gophers"),
	}

	t.Run("postProfile400", tests.postProfile400)
	t.Run("postProfile401", tests.postProfile401)
	t.Run("getProfile404", tests.getProfile404)

	// t.Run("getProfile400", tests.getProfile400)

	t.Run("deleteProfileNotFound", tests.deleteProfileNotFound)
	t.Run("putProfile404", tests.putProfile404)
	t.Run("crudProfiles", tests.crudProfile)
}

// postProfile400 validates a profile can't be created with the endpoint
// unless a valid profile document is submitted.
func (pt *ProfileTests) postProfile400(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/v1/profiles", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)
	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new profile can't be created with an invalid document.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using an incomplete profile value.", testID)
		{
			if w.Code != http.StatusBadRequest {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 400 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 400 for the response.", tests.Success, testID)

			// Inspect the response.
			var got web.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response to an error type : %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to unmarshal the response to an error type.", tests.Success, testID)

			// Define what we want to see.
			exp := web.ErrorResponse{
				Error: "field validation error",
				Fields: []web.FieldError{
					{Field: "name", Error: "name is a required field"},
					{Field: "display_name", Error: "display_name is a required field"},
				},
			}

			// We can't rely on the order of the field errors so they have to be
			// sorted. Tell the cmp package how to sort them.
			sorter := cmpopts.SortSlices(func(a, b web.FieldError) bool {
				return a.Field < b.Field
			})

			if diff := cmp.Diff(got, exp, sorter); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// postProfile401 validates a profile can't be created with the endpoint
// unless the user is authenticated
func (pt *ProfileTests) postProfile401(t *testing.T) {
	np := profile.NewProfile{
		Name:        "test",
		DisplayName: "Test profile",
	}

	body, err := json.Marshal(&np)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/profiles", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Not setting an authorization header.
	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new profile can't be created with an invalid document.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using an incomplete profile value.", testID)
		{
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 401 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 401 for the response.", tests.Success, testID)
		}
	}
}

// getProfile404 validates a profile request for a profile that does not exist with the endpoint.
func (pt *ProfileTests) getProfile404(t *testing.T) {
	name := "t3st"

	r := httptest.NewRequest(http.MethodGet, "/v1/profiles/"+name, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)
	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a profile with an unknown name.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new profile %s.", testID, name)
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

// deleteProfileNotFound validates deleting a profile that does not exist is not a failure.
func (pt *ProfileTests) deleteProfileNotFound(t *testing.T) {
	name := "t3st"

	r := httptest.NewRequest(http.MethodDelete, "/v1/profiles/"+name, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)
	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate deleting a profile that does not exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new profile %s.", testID, name)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Success, testID)
		}
	}
}

// putProfile404 validates updating a profile that does not exist.
func (pt *ProfileTests) putProfile404(t *testing.T) {
	name := "t3st"

	up := profile.UpdateProfile{
		Name: tests.StringPointer("Nonexistent"),
	}
	body, err := json.Marshal(&up)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPut, "/v1/profiles/"+name, bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)
	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate updating a profile that does not exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new profile %s.", testID, name)
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

// crudProfile performs a complete test of CRUD against the api.
func (pt *ProfileTests) crudProfile(t *testing.T) {
	p := pt.postProfile201(t)
	defer pt.deleteProfile204(t, p.Name)

	pt.getProfile200(t, p.Name)
	pt.getProfileByUserID200(t, p.UserID)
	pt.putProfile204(t, p.Name)
}

// postProfile201 validates a profile can be created with the endpoint.
func (pt *ProfileTests) postProfile201(t *testing.T) profile.Info {
	np := profile.NewProfile{
		Name:        "test",
		DisplayName: "Test profile",
	}

	body, err := json.Marshal(&np)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/profiles", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)
	pt.app.ServeHTTP(w, r)

	// This needs to be returned for other tests.
	var got profile.Info

	t.Log("Given the need to create a new profile with the profiles endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the declared profile value.", testID)
		{
			if w.Code != http.StatusCreated {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 201 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 201 for the response.", tests.Success, testID)

			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like Name and Dates so we copy p.
			exp := got
			exp.Name = "test"
			exp.DisplayName = "Test profile"

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}

	return got
}

// deleteProfile200 validates deleting a profile that does exist.
func (pt *ProfileTests) deleteProfile204(t *testing.T, name string) {
	r := httptest.NewRequest(http.MethodDelete, "/v1/profiles/"+name, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)
	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate deleting a profile that does exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new profile %s.", testID, name)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Success, testID)
		}
	}
}

// getProfile200 validates a profile request for an existing name.
func (pt *ProfileTests) getProfile200(t *testing.T, name string) {
	r := httptest.NewRequest(http.MethodGet, "/v1/profiles/"+name, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)
	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a profile that exists by name.")
	{
		testID := 0
		t.Logf("\tTest : %d\tWhen using the new profile %s.", testID, name)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest : %d\tShould receive a status code of 200 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest : %d\tShould receive a status code of 200 for the response.", tests.Success, testID)

			var got profile.Info
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest : %d\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like Dates so we copy p.
			exp := got
			exp.Name = "test"
			exp.DisplayName = "Test profile"

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest : %d\tShould get the expected result. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest : %d\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// getProfile200 validates a profile request for an existing user id.
func (pt *ProfileTests) getProfileByUserID200(t *testing.T, userID string) {
	r := httptest.NewRequest(http.MethodGet, "/v1/users/"+userID+"/profile", nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)
	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a profile that exists by user ID.")
	{
		testID := 0
		t.Logf("\tTest : %d\tWhen using the new profile %s.", testID, userID)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest : %d\tShould receive a status code of 200 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest : %d\tShould receive a status code of 200 for the response.", tests.Success, testID)

			var got profile.Info
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest : %d\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like Dates so we copy p.
			exp := got
			exp.UserID = userID
			exp.Name = "test"
			exp.DisplayName = "Test profile"

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest : %d\tShould get the expected result. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest : %d\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// getProfile200 validates a profile request for an existing id.
// func (pt *ProfileTests) getProfileByUsername200(t *testing.T, uname string) {
// 	r := httptest.NewRequest(http.MethodGet, "/v1/users/u/"+uname, nil)
// 	w := httptest.NewRecorder()

// 	r.Header.Set("Authorization", "Bearer "+pt.userToken)
// 	pt.app.ServeHTTP(w, r)

// 	t.Log("Given the need to validate getting a profile that exists by username.")
// 	{
// 		testID := 0
// 		t.Logf("\tTest : %d\tWhen using the new profile %s.", testID, uname)
// 		{
// 			if w.Code != http.StatusOK {
// 				t.Fatalf("\t%s\tTest : %d\tShould receive a status code of 200 for the response : %v", tests.Failed, testID, w.Code)
// 			}
// 			t.Logf("\t%s\tTest : %d\tShould receive a status code of 200 for the response.", tests.Success, testID)

// 			var got profile.Info
// 			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
// 				t.Fatalf("\t%s\tTest : %d\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
// 			}

// 			// Define what we wanted to receive. We will just trust the generated
// 			// fields like Dates so we copy p.
// 			exp := got
// 			exp.Name = uname
// 			exp.DisplayName = "Test profile"

// 			if diff := cmp.Diff(got, exp); diff != "" {
// 				t.Fatalf("\t%s\tTest : %d\tShould get the expected result. Diff:\n%s", tests.Failed, testID, diff)
// 			}
// 			t.Logf("\t%s\tTest : %d\tShould get the expected result.", tests.Success, testID)
// 		}
// 	}
// }

// putProfile204 validates updating a profile that does exist.
func (pt *ProfileTests) putProfile204(t *testing.T, name string) {
	body := `{"display_name": "My profile"}`
	r := httptest.NewRequest(http.MethodPut, "/v1/profiles/"+name, strings.NewReader(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+pt.userToken)
	pt.app.ServeHTTP(w, r)

	t.Log("Given the need to update a profile with the profiles endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the modified profile value.", testID)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Success, testID)

			r = httptest.NewRequest(http.MethodGet, "/v1/profiles/"+name, nil)
			w = httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+pt.userToken)
			pt.app.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the retrieve : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the retrieve.", tests.Success, testID)

			var ru profile.Info
			if err := json.NewDecoder(w.Body).Decode(&ru); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			if ru.DisplayName != "My profile" {
				t.Fatalf("\t%s\tTest %d:\tShould see an updated Name : got %q want %q", tests.Failed, testID, ru.DisplayName, "My profile")
			}
			t.Logf("\t%s\tTest %d:\tShould see an updated Name.", tests.Success, testID)
		}
	}
}
