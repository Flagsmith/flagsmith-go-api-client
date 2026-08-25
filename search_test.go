package flagsmithapi_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	flagsmithapi "github.com/Flagsmith/flagsmith-go-api-client"
)

// searchResultsJson wraps result objects in the paginated envelope the API returns.
// next is emitted verbatim, so pass "null" for the last page.
func searchResultsJson(next string, results ...string) string {
	body := ""
	for i, result := range results {
		if i > 0 {
			body += ","
		}
		body += result
	}
	return fmt.Sprintf(`{"count": %d, "next": %s, "previous": null, "results": [%s]}`,
		len(results), next, body)
}

func featureJson(id int64, name string) string {
	return fmt.Sprintf(`{
		"id": %d,
		"uuid": "%s",
		"name": "%s",
		"project": %d,
		"type": "STANDARD",
		"default_enabled": false,
		"initial_value": null,
		"description": null,
		"tags": [1],
		"is_archived": false,
		"owners": [{"id": 1, "email": "some_user@email.com"}],
		"group_owners": [{"id": 3, "name": "Test Group"}]
	}`, id, FeatureUUID, name, ProjectID)
}

func segmentJson(id int64, name string, feature string) string {
	return fmt.Sprintf(`{
		"id": %d,
		"uuid": "%s",
		"name": "%s",
		"description": null,
		"project": %d,
		"feature": %s,
		"rules": [
			{
				"type": "ALL",
				"rules": [
					{
						"type": "ANY",
						"rules": [],
						"conditions": [{"operator": "EQUAL", "property": "device_type", "value": "mobile"}]
					}
				],
				"conditions": []
			}
		]
	}`, id, SegmentUUID, name, ProjectID, feature)
}

// searchTestServer serves the project lookup and a project sub-resource list endpoint.
// It records the query parameters of every list request, and fails the test if the
// project-by-ID endpoint (which name lookups should never need) is called.
func searchTestServer(t *testing.T, subResource string, pages ...string) (*httptest.Server, *[]url.Values) {
	t.Helper()

	mux := http.NewServeMux()
	var mu sync.Mutex
	queries := []url.Values{}

	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/get-by-uuid/%s/", ProjectUUID), func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetProjectResponseJson)
		assert.NoError(t, err)
	})

	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/%d/", ProjectID), func(rw http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "the project should not be looked up by ID during a lookup by name")
	})

	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/%d/%s/", ProjectID, subResource), func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		mu.Lock()
		page := len(queries)
		queries = append(queries, req.URL.Query())
		mu.Unlock()

		if page >= len(pages) {
			assert.Failf(t, "unexpected request", "requested page %d but only %d were stubbed", page+1, len(pages))
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, pages[page])
		assert.NoError(t, err)
	})

	return httptest.NewServer(mux), &queries
}

func TestGetFeatureByName(t *testing.T) {
	// Given
	server, queries := searchTestServer(t, "features",
		searchResultsJson("null", featureJson(FeatureID, FeatureName)))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	feature, err := client.GetFeatureByName(ProjectUUID, FeatureName)

	// Then
	assert.NoError(t, err)

	assert.Equal(t, FeatureID, *feature.ID)
	assert.Equal(t, FeatureName, feature.Name)
	assert.Equal(t, FeatureUUID, feature.UUID)
	assert.Equal(t, "STANDARD", *feature.Type)
	assert.Equal(t, false, feature.DefaultEnabled)
	assert.Equal(t, false, feature.IsArchived)
	assert.Equal(t, []int64{1}, feature.Tags)
	assert.Equal(t, &[]int64{1}, feature.Owners)
	assert.Equal(t, &[]int64{3}, feature.GroupOwners)

	assert.Equal(t, ProjectID, *feature.ProjectID)
	// The API never returns the project UUID, so the client must set it from the
	// argument, otherwise Terraform sees an inconsistent data source result.
	assert.Equal(t, ProjectUUID, feature.ProjectUUID)

	// Features filter on `search`.
	assert.Equal(t, 1, len(*queries))
	assert.Equal(t, FeatureName, (*queries)[0].Get("search"))
	// CustomPagination.max_page_size, sent so a change to the API's default page size
	// cannot silently truncate results.
	assert.Equal(t, "999", (*queries)[0].Get("page_size"))
}

func TestGetFeatureByNameFiltersPartialMatches(t *testing.T) {
	// Given
	// The API's `search` filter is a "contains" match, so it returns both of these.
	server, _ := searchTestServer(t, "features", searchResultsJson("null",
		featureJson(2, FeatureName+"_suffix"),
		featureJson(FeatureID, FeatureName),
	))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	feature, err := client.GetFeatureByName(ProjectUUID, FeatureName)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, FeatureName, feature.Name)
	assert.Equal(t, FeatureID, *feature.ID)
}

func TestGetFeatureByNameOnlyPartialMatches(t *testing.T) {
	// Given
	server, _ := searchTestServer(t, "features", searchResultsJson("null",
		featureJson(2, FeatureName+"_suffix"),
	))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	feature, err := client.GetFeatureByName(ProjectUUID, FeatureName)

	// Then
	assert.Nil(t, feature)
	assert.IsType(t, flagsmithapi.FeatureNotFoundError{}, err)
	assert.Contains(t, err.Error(), FeatureName)
	assert.Contains(t, err.Error(), ProjectUUID)
}

func TestGetFeatureByNameIsCaseSensitive(t *testing.T) {
	// Given
	server, _ := searchTestServer(t, "features", searchResultsJson("null",
		featureJson(FeatureID, "Test_Feature"),
	))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	feature, err := client.GetFeatureByName(ProjectUUID, "test_feature")

	// Then
	assert.Nil(t, feature)
	assert.IsType(t, flagsmithapi.FeatureNotFoundError{}, err)
}

func TestGetFeatureByNameNotFound(t *testing.T) {
	// Given
	server, _ := searchTestServer(t, "features", searchResultsJson("null"))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	feature, err := client.GetFeatureByName(ProjectUUID, FeatureName)

	// Then
	assert.Nil(t, feature)
	assert.IsType(t, flagsmithapi.FeatureNotFoundError{}, err)
}

// `search` is a "contains" match, so a truncated page means an exact match may have been
// missed. Erroring is better than a false "not found".
func TestGetFeatureByNameErrorsWhenResultsAreTruncated(t *testing.T) {
	// Given
	server, queries := searchTestServer(t, "features",
		searchResultsJson(`"http://api.example.com/api/v1/projects/10/features/?page=2"`,
			featureJson(2, FeatureName+"_suffix")),
	)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	feature, err := client.GetFeatureByName(ProjectUUID, FeatureName)

	// Then
	assert.Nil(t, feature)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too many features")
	// A truncated search must not be reported as "not found".
	_, isNotFound := err.(flagsmithapi.FeatureNotFoundError)
	assert.False(t, isNotFound)

	// Only one request is made; the results are not walked page by page.
	assert.Equal(t, 1, len(*queries))
}

func TestGetFeatureByNameMultipleMatches(t *testing.T) {
	// Given
	server, _ := searchTestServer(t, "features", searchResultsJson("null",
		featureJson(FeatureID, FeatureName),
		featureJson(2, FeatureName),
	))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	feature, err := client.GetFeatureByName(ProjectUUID, FeatureName)

	// Then
	assert.Nil(t, feature)
	assert.IsType(t, flagsmithapi.MultipleFeaturesFoundError{}, err)
	assert.Contains(t, err.Error(), "found 2 features")
}

func TestGetFeatureByNameAPIError(t *testing.T) {
	// Given
	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/get-by-uuid/%s/", ProjectUUID), func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetProjectResponseJson)
		assert.NoError(t, err)
	})

	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/%d/features/", ProjectID), func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	feature, err := client.GetFeatureByName(ProjectUUID, FeatureName)

	// Then
	assert.Nil(t, feature)
	assert.Error(t, err)
	// A server error must not be reported as "not found".
	_, isNotFound := err.(flagsmithapi.FeatureNotFoundError)
	assert.False(t, isNotFound)
	assert.Contains(t, err.Error(), "Error searching features")
}

func TestGetFeatureByNameProjectNotFound(t *testing.T) {
	// Given
	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/get-by-uuid/%s/", ProjectUUID), func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusNotFound)
	})

	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/%d/features/", ProjectID), func(rw http.ResponseWriter, req *http.Request) {
		assert.Fail(t, "features should not be searched when the project cannot be resolved")
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	feature, err := client.GetFeatureByName(ProjectUUID, FeatureName)

	// Then
	assert.Nil(t, feature)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Error getting project")
}

func TestGetSegmentByName(t *testing.T) {
	// Given
	segmentName := "mobile_users"
	server, queries := searchTestServer(t, "segments",
		searchResultsJson("null", segmentJson(SegmentID, segmentName, "null")))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	segment, err := client.GetSegmentByName(ProjectUUID, segmentName)

	// Then
	var nilIntPointer *int64
	var nilStringPointer *string

	assert.NoError(t, err)

	assert.Equal(t, SegmentID, *segment.ID)
	assert.Equal(t, SegmentUUID, segment.UUID)
	assert.Equal(t, segmentName, segment.Name)
	assert.Equal(t, nilStringPointer, segment.Description)
	assert.Equal(t, nilIntPointer, segment.FeatureID)

	assert.Equal(t, ProjectID, *segment.ProjectID)
	assert.Equal(t, ProjectUUID, segment.ProjectUUID)

	// The list endpoint returns the full rules tree, so it round trips exactly as
	// the get-by-uuid response does.
	assert.Equal(t, 1, len(segment.Rules))
	assert.Equal(t, "ALL", segment.Rules[0].Type)
	assert.Equal(t, 0, len(segment.Rules[0].Conditions))
	assert.Equal(t, 1, len(segment.Rules[0].Rules))
	assert.Equal(t, "ANY", segment.Rules[0].Rules[0].Type)
	assert.Equal(t, 1, len(segment.Rules[0].Rules[0].Conditions))
	assert.Equal(t, "EQUAL", segment.Rules[0].Rules[0].Conditions[0].Operator)
	assert.Equal(t, "device_type", segment.Rules[0].Rules[0].Conditions[0].Property)
	assert.Equal(t, "mobile", segment.Rules[0].Rules[0].Conditions[0].Value)

	// Segments filter on `q`, not `search`.
	assert.Equal(t, 1, len(*queries))
	assert.Equal(t, segmentName, (*queries)[0].Get("q"))
	assert.Equal(t, "", (*queries)[0].Get("search"))
	assert.Equal(t, "999", (*queries)[0].Get("page_size"))
}

func TestGetSegmentByNameFiltersPartialMatches(t *testing.T) {
	// Given
	segmentName := "mobile_users"
	server, _ := searchTestServer(t, "segments", searchResultsJson("null",
		segmentJson(2, segmentName+"_eu", "null"),
		segmentJson(SegmentID, segmentName, "null"),
	))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	segment, err := client.GetSegmentByName(ProjectUUID, segmentName)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, SegmentID, *segment.ID)
	assert.Equal(t, segmentName, segment.Name)
}

func TestGetSegmentByNameFeatureSpecific(t *testing.T) {
	// Given
	segmentName := "feature_specific_segment"
	server, _ := searchTestServer(t, "segments", searchResultsJson("null",
		segmentJson(SegmentID, segmentName, "25142"),
	))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	segment, err := client.GetSegmentByName(ProjectUUID, segmentName)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, int64(25142), *segment.FeatureID)
}

func TestGetSegmentByNameNotFound(t *testing.T) {
	// Given
	server, _ := searchTestServer(t, "segments", searchResultsJson("null"))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	segment, err := client.GetSegmentByName(ProjectUUID, "missing_segment")

	// Then
	assert.Nil(t, segment)
	assert.IsType(t, flagsmithapi.SegmentNotFoundError{}, err)
	assert.Contains(t, err.Error(), "missing_segment")
	assert.Contains(t, err.Error(), ProjectUUID)
}

func TestGetSegmentByNameErrorsWhenResultsAreTruncated(t *testing.T) {
	// Given
	segmentName := "mobile_users"
	server, queries := searchTestServer(t, "segments",
		searchResultsJson(`"http://api.example.com/api/v1/projects/10/segments/?page=2"`,
			segmentJson(2, segmentName+"_eu", "null")),
	)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	segment, err := client.GetSegmentByName(ProjectUUID, segmentName)

	// Then
	assert.Nil(t, segment)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too many segments")

	assert.Equal(t, 1, len(*queries))
}

func TestGetSegmentByNameMultipleMatches(t *testing.T) {
	// Given
	// Segment names are not unique within a project.
	segmentName := "mobile_users"
	server, _ := searchTestServer(t, "segments", searchResultsJson("null",
		segmentJson(SegmentID, segmentName, "null"),
		segmentJson(2, segmentName, "null"),
	))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	segment, err := client.GetSegmentByName(ProjectUUID, segmentName)

	// Then
	assert.Nil(t, segment)
	assert.IsType(t, flagsmithapi.MultipleSegmentsFoundError{}, err)
	assert.Contains(t, err.Error(), "found 2 segments")
}
