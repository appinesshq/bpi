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
	"github.com/appinesshq/bpi/business/data/category"
	"github.com/appinesshq/bpi/business/tests"
	"github.com/appinesshq/bpi/foundation/web"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// CategoryTests holds methods for each category subtest. This type allows
// passing dependencies for tests while still providing a convenient syntax
// when subtests are registered.
type CategoryTests struct {
	app       http.Handler
	userToken string
}

// TestCategorys runs a series of tests to exercise Category behavior from the
// API level. The subtests all share the same database and application for
// speed and convenience. The downside is the order the tests are ran matters
// and one test may break if other tests are not ran before it. If a particular
// subtest needs a fresh instance of the application it can make it or it
// should be its own Test* function.
func TestCategorys(t *testing.T) {
	test := tests.NewIntegration(t)
	t.Cleanup(test.Teardown)

	shutdown := make(chan os.Signal, 1)
	tests := CategoryTests{
		app:       handlers.API("develop", shutdown, test.Log, test.Auth, test.DB),
		userToken: test.Token(test.KID, "admin@example.com", "gophers"),
	}

	t.Run("postCategory400", tests.postCategory400)
	t.Run("postCategory401", tests.postCategory401)
	t.Run("getCategory404", tests.getCategory404)
	t.Run("getCategory400", tests.getCategory400)
	t.Run("deleteCategoryNotFound", tests.deleteCategoryNotFound)
	t.Run("putCategory404", tests.putCategory404)
	t.Run("crudCategorys", tests.crudCategory)
}

// postCategory400 validates a category can't be created with the endpoint
// unless a valid category document is submitted.
func (ct *CategoryTests) postCategory400(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/v1/categories", strings.NewReader(`{}`))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new category can't be created with an invalid document.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using an incomplete category value.", testID)
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
					{Field: "slug", Error: "slug is a required field"},
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

// postCategory401 validates a category can't be created with the endpoint
// unless the user is authenticated
func (ct *CategoryTests) postCategory401(t *testing.T) {
	np := category.NewCategory{
		Slug: "test",
		Name: "Test category",
	}

	body, err := json.Marshal(&np)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/categories", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// Not setting an authorization header.
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to validate a new category can't be created with an invalid document.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using an incomplete category value.", testID)
		{
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 401 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 401 for the response.", tests.Success, testID)
		}
	}
}

// getCategory400 validates a category request for a malformed id.
func (ct *CategoryTests) getCategory400(t *testing.T) {
	id := "12345"

	r := httptest.NewRequest(http.MethodGet, "/v1/categories/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a category with a malformed id.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new category %s.", testID, id)
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

// getCategory404 validates a category request for a category that does not exist with the endpoint.
func (ct *CategoryTests) getCategory404(t *testing.T) {
	id := "a224a8d6-3f9e-4b11-9900-e81a25d80702"

	r := httptest.NewRequest(http.MethodGet, "/v1/categories/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a category with an unknown id.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new category %s.", testID, id)
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

// deleteCategoryNotFound validates deleting a category that does not exist is not a failure.
func (ct *CategoryTests) deleteCategoryNotFound(t *testing.T) {
	id := "112262f1-1a77-4374-9f22-39e575aa6348"

	r := httptest.NewRequest(http.MethodDelete, "/v1/categories/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to validate deleting a category that does not exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new category %s.", testID, id)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Success, testID)
		}
	}
}

// putCategory404 validates updating a category that does not exist.
func (ct *CategoryTests) putCategory404(t *testing.T) {
	id := "9b468f90-1cf1-4377-b3fa-68b450d632a0"

	up := category.UpdateCategory{
		Name: tests.StringPointer("Nonexistent"),
	}
	body, err := json.Marshal(&up)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPut, "/v1/categories/"+id, bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to validate updating a category that does not exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new category %s.", testID, id)
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

// crudCategory performs a complete test of CRUD against the api.
func (ct *CategoryTests) crudCategory(t *testing.T) {
	p := ct.postCategory201(t)
	defer ct.deleteCategory204(t, p.ID)

	ct.getCategory200(t, p.ID)
	ct.putCategory204(t, p.ID)
}

// postCategory201 validates a category can be created with the endpoint.
func (ct *CategoryTests) postCategory201(t *testing.T) category.Info {
	np := category.NewCategory{
		Slug: "test",
		Name: "Test category",
	}

	body, err := json.Marshal(&np)
	if err != nil {
		t.Fatal(err)
	}

	r := httptest.NewRequest(http.MethodPost, "/v1/categories", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	// This needs to be returned for other tests.
	var got category.Info

	t.Log("Given the need to create a new category with the categories endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the declared category value.", testID)
		{
			if w.Code != http.StatusCreated {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 201 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 201 for the response.", tests.Success, testID)

			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like ID and Dates so we copy p.
			exp := got
			exp.Slug = "test"
			exp.Name = "Test category"

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest %d:\tShould get the expected result. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest %d:\tShould get the expected result.", tests.Success, testID)
		}
	}

	return got
}

// deleteCategory200 validates deleting a category that does exist.
func (ct *CategoryTests) deleteCategory204(t *testing.T, id string) {
	r := httptest.NewRequest(http.MethodDelete, "/v1/categories/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to validate deleting a category that does exist.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the new category %s.", testID, id)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Success, testID)
		}
	}
}

// getCategory200 validates a category request for an existing id.
func (ct *CategoryTests) getCategory200(t *testing.T, id string) {
	r := httptest.NewRequest(http.MethodGet, "/v1/categories/"+id, nil)
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to validate getting a category that exists.")
	{
		testID := 0
		t.Logf("\tTest : %d\tWhen using the new category %s.", testID, id)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest : %d\tShould receive a status code of 200 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest : %d\tShould receive a status code of 200 for the response.", tests.Success, testID)

			var got category.Info
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest : %d\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			// Define what we wanted to receive. We will just trust the generated
			// fields like Dates so we copy p.
			exp := got
			exp.ID = id
			exp.Slug = "test"
			exp.Name = "Test category"

			if diff := cmp.Diff(got, exp); diff != "" {
				t.Fatalf("\t%s\tTest : %d\tShould get the expected result. Diff:\n%s", tests.Failed, testID, diff)
			}
			t.Logf("\t%s\tTest : %d\tShould get the expected result.", tests.Success, testID)
		}
	}
}

// putCategory204 validates updating a category that does exist.
func (ct *CategoryTests) putCategory204(t *testing.T, id string) {
	body := `{"name": "Testing", "slug": "testing"}`
	r := httptest.NewRequest(http.MethodPut, "/v1/categories/"+id, strings.NewReader(body))
	w := httptest.NewRecorder()

	r.Header.Set("Authorization", "Bearer "+ct.userToken)
	ct.app.ServeHTTP(w, r)

	t.Log("Given the need to update a category with the categories endpoint.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen using the modified category value.", testID)
		{
			if w.Code != http.StatusNoContent {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 204 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 204 for the response.", tests.Success, testID)

			r = httptest.NewRequest(http.MethodGet, "/v1/categories/"+id, nil)
			w = httptest.NewRecorder()

			r.Header.Set("Authorization", "Bearer "+ct.userToken)
			ct.app.ServeHTTP(w, r)

			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the retrieve : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the retrieve.", tests.Success, testID)

			var ru category.Info
			if err := json.NewDecoder(w.Body).Decode(&ru); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}

			if ru.Name != "Testing" {
				t.Fatalf("\t%s\tTest %d:\tShould see an updated Name : got %q want %q", tests.Failed, testID, ru.Name, "Testing")
			}
			t.Logf("\t%s\tTest %d:\tShould see an updated Name.", tests.Success, testID)
		}
	}
}
