package flagsmithapi_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	flagsmithapi "github.com/Flagsmith/flagsmith-go-api-client"
)

// GetProjectResponseJson reports organisation 1, and the metadata endpoints are keyed on
// the organisation ID rather than the project's.
const OrganisationIDForMetadata int64 = 1

// Content type IDs are specific to a Flagsmith installation, so these are deliberately
// arbitrary.
const FeatureContentTypeID int64 = 41
const SegmentContentTypeID int64 = 42
const EnvironmentContentTypeID int64 = 43

const FeatureModelFieldID int64 = 301
const SegmentModelFieldID int64 = 302
const EnvironmentModelFieldID int64 = 303

const SupportedContentTypesJson = `[
	{"id": 41, "app_label": "features", "model": "feature"},
	{"id": 42, "app_label": "segments", "model": "segment"},
	{"id": 43, "app_label": "environments", "model": "environment"}
]`

// A field bound to all three entity types, so a caller has to pick the model field
// matching the entity it is writing to.
const MetadataFieldsJson = `{
	"count": 3,
	"next": null,
	"previous": null,
	"results": [
		{
			"id": 1,
			"name": "Jira Ticket",
			"type": "str",
			"description": null,
			"organisation": 1,
			"project": null,
			"model_fields": [
				{"id": 301, "content_type": 41, "is_required_for": []},
				{"id": 302, "content_type": 42, "is_required_for": []},
				{"id": 303, "content_type": 43, "is_required_for": []}
			]
		},
		{
			"id": 2,
			"name": "Story Points",
			"type": "int",
			"description": null,
			"organisation": 1,
			"project": 10,
			"model_fields": [
				{"id": 304, "content_type": 41, "is_required_for": []}
			]
		},
		{
			"id": 3,
			"name": "Segment Only",
			"type": "str",
			"description": null,
			"organisation": 1,
			"project": null,
			"model_fields": [
				{"id": 305, "content_type": 42, "is_required_for": []}
			]
		}
	]
}`

// requestCounter records how many times each path is requested, so the caching tests can
// assert that a lookup only happens once.
type requestCounter struct {
	mu     sync.Mutex
	counts map[string]int
}

func newRequestCounter() *requestCounter {
	return &requestCounter{counts: map[string]int{}}
}

func (r *requestCounter) record(path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counts[path]++
}

func (r *requestCounter) count(path string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.counts[path]
}

// metadataTestServer serves the three requests a resolver needs.
func metadataTestServer(t *testing.T, fieldsPages ...string) (*httptest.Server, *requestCounter) {
	t.Helper()

	if len(fieldsPages) == 0 {
		fieldsPages = []string{MetadataFieldsJson}
	}

	mux := http.NewServeMux()
	counter := newRequestCounter()

	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/%d/", ProjectID), func(rw http.ResponseWriter, req *http.Request) {
		counter.record("project")
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetProjectResponseJson)
		assert.NoError(t, err)
	})

	mux.HandleFunc(fmt.Sprintf("/api/v1/organisations/%d/metadata-model-fields/supported-content-types/", OrganisationIDForMetadata),
		func(rw http.ResponseWriter, req *http.Request) {
			counter.record("content-types")
			assert.Equal(t, "GET", req.Method)
			assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

			rw.Header().Set("Content-Type", "application/json")
			_, err := io.WriteString(rw, SupportedContentTypesJson)
			assert.NoError(t, err)
		})

	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/%d/metadata/fields/", ProjectID),
		func(rw http.ResponseWriter, req *http.Request) {
			counter.record("fields")

			assert.Equal(t, "GET", req.Method)
			// Without include_organisation, organisation level fields are omitted.
			assert.Equal(t, "true", req.URL.Query().Get("include_organisation"))

			// Keyed on the requested page rather than on a call count, so that
			// repeated fetches each start from the first page.
			page := 1
			if requested := req.URL.Query().Get("page"); requested != "" {
				parsed, err := strconv.Atoi(requested)
				assert.NoError(t, err)
				page = parsed
			}

			if page > len(fieldsPages) {
				assert.Failf(t, "unexpected request", "requested page %d but only %d were stubbed", page, len(fieldsPages))
				return
			}

			rw.Header().Set("Content-Type", "application/json")
			_, err := io.WriteString(rw, fieldsPages[page-1])
			assert.NoError(t, err)
		})

	return httptest.NewServer(mux), counter
}

// Keep the project fixture and OrganisationIDForMetadata in step.
func TestGetProjectResponseOrganisationMatchesConstant(t *testing.T) {
	project := struct {
		Organisation int64 `json:"organisation"`
	}{}
	require.NoError(t, json.Unmarshal([]byte(GetProjectResponseJson), &project))
	assert.Equal(t, project.Organisation, OrganisationIDForMetadata)
}

func TestGetProjectMetadataFields(t *testing.T) {
	// Given
	server, counter := metadataTestServer(t)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	fields, err := client.GetProjectMetadataFields(ProjectID)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, 1, counter.count("fields"))

	require.Equal(t, 3, len(fields))
	assert.Equal(t, "Jira Ticket", fields[0].Name)
	assert.Equal(t, "str", fields[0].Type)
	assert.Equal(t, OrganisationIDForMetadata, fields[0].Organisation)
	assert.Nil(t, fields[0].Project)
	require.Equal(t, 3, len(fields[0].ModelFields))
	assert.Equal(t, FeatureModelFieldID, fields[0].ModelFields[0].ID)
	assert.Equal(t, FeatureContentTypeID, fields[0].ModelFields[0].ContentType)

	assert.Equal(t, "int", fields[1].Type)
	require.NotNil(t, fields[1].Project)
	assert.Equal(t, ProjectID, *fields[1].Project)
}

func TestGetProjectMetadataFieldsPaginated(t *testing.T) {
	// Given
	firstPage := `{"count": 2, "next": "http://api.example.com/api/v1/projects/10/metadata/fields/?page=2", "previous": null, "results": [
		{"id": 1, "name": "Jira Ticket", "type": "str", "organisation": 1, "project": null, "model_fields": [{"id": 301, "content_type": 41}]}
	]}`
	secondPage := `{"count": 2, "next": null, "previous": null, "results": [
		{"id": 2, "name": "Story Points", "type": "int", "organisation": 1, "project": null, "model_fields": [{"id": 304, "content_type": 41}]}
	]}`

	server, counter := metadataTestServer(t, firstPage, secondPage)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	fields, err := client.GetProjectMetadataFields(ProjectID)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, 2, counter.count("fields"))
	require.Equal(t, 2, len(fields))
	assert.Equal(t, "Jira Ticket", fields[0].Name)
	assert.Equal(t, "Story Points", fields[1].Name)
}

func TestGetSupportedMetadataContentTypes(t *testing.T) {
	// Given
	server, _ := metadataTestServer(t)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	contentTypes, err := client.GetSupportedMetadataContentTypes(OrganisationIDForMetadata)

	// Then
	assert.NoError(t, err)
	require.Equal(t, 3, len(contentTypes))
	assert.Equal(t, FeatureContentTypeID, contentTypes[0].ID)
	assert.Equal(t, "features", contentTypes[0].AppLabel)
	assert.Equal(t, "feature", contentTypes[0].Model)
}

// The `entity` query param does not filter each field's nested model_fields, so the
// resolver has to pick the model field whose content type matches the entity. Getting
// this wrong writes the value against the wrong entity type.
func TestResolverPicksModelFieldMatchingEntityContentType(t *testing.T) {
	// Given
	server, _ := metadataTestServer(t)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	testCases := []struct {
		entity               flagsmithapi.MetadataEntity
		expectedModelFieldID int64
	}{
		{flagsmithapi.MetadataEntityFeature, FeatureModelFieldID},
		{flagsmithapi.MetadataEntitySegment, SegmentModelFieldID},
		{flagsmithapi.MetadataEntityEnvironment, EnvironmentModelFieldID},
	}

	for _, testCase := range testCases {
		// When
		resolver, err := client.GetMetadataFieldResolver(ProjectID, testCase.entity)
		require.NoError(t, err)

		modelFieldID, err := resolver.ModelFieldID("Jira Ticket")

		// Then
		assert.NoError(t, err)
		assert.Equal(t, testCase.expectedModelFieldID, modelFieldID, "entity %s", testCase.entity)
	}
}

func TestResolverUnknownFieldName(t *testing.T) {
	// Given
	server, _ := metadataTestServer(t)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")
	resolver, err := client.GetMetadataFieldResolver(ProjectID, flagsmithapi.MetadataEntityFeature)
	require.NoError(t, err)

	// When
	_, err = resolver.ModelFieldID("Not A Field")

	// Then
	require.IsType(t, flagsmithapi.MetadataFieldNotFoundError{}, err)
	notFound := err.(flagsmithapi.MetadataFieldNotFoundError)
	assert.Equal(t, "Not A Field", notFound.Name)
	assert.Equal(t, "feature", notFound.Entity)
	// Sorted, and only the fields bound to features.
	assert.Equal(t, []string{"Jira Ticket", "Story Points"}, notFound.KnownNames)
}

func TestResolverFieldNotBoundToEntity(t *testing.T) {
	// Given
	server, _ := metadataTestServer(t)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")
	resolver, err := client.GetMetadataFieldResolver(ProjectID, flagsmithapi.MetadataEntityFeature)
	require.NoError(t, err)

	// When
	_, err = resolver.ModelFieldID("Segment Only")

	// Then
	require.IsType(t, flagsmithapi.MetadataFieldNotBoundError{}, err)
	assert.Contains(t, err.Error(), "not enabled for features")
}

// Reverse lookups cover every model field in the project, not just the ones bound to the
// entity, so reading back a value never fails spuriously.
func TestResolverReverseLookupCoversAllEntities(t *testing.T) {
	// Given
	server, _ := metadataTestServer(t)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")
	resolver, err := client.GetMetadataFieldResolver(ProjectID, flagsmithapi.MetadataEntityFeature)
	require.NoError(t, err)

	// When / Then
	name, ok := resolver.Name(SegmentModelFieldID)
	assert.True(t, ok)
	assert.Equal(t, "Jira Ticket", name)

	_, ok = resolver.Name(9999)
	assert.False(t, ok)
}

func TestContentTypeIDForEntityFallsBackToModelName(t *testing.T) {
	// Given
	// A content type whose app label does not match the expected one, e.g. a renamed
	// app in an enterprise build.
	contentTypes := `[{"id": 77, "app_label": "flagsmith_features", "model": "feature"}]`

	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/%d/", ProjectID), func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetProjectResponseJson)
		assert.NoError(t, err)
	})
	mux.HandleFunc(fmt.Sprintf("/api/v1/organisations/%d/metadata-model-fields/supported-content-types/", OrganisationIDForMetadata),
		func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("Content-Type", "application/json")
			_, err := io.WriteString(rw, contentTypes)
			assert.NoError(t, err)
		})
	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/%d/metadata/fields/", ProjectID),
		func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("Content-Type", "application/json")
			_, err := io.WriteString(rw, `{"count": 1, "next": null, "results": [
				{"id": 1, "name": "Jira Ticket", "type": "str", "organisation": 1, "project": null, "model_fields": [{"id": 401, "content_type": 77}]}
			]}`)
			assert.NoError(t, err)
		})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	resolver, err := client.GetMetadataFieldResolver(ProjectID, flagsmithapi.MetadataEntityFeature)

	// Then
	require.NoError(t, err)
	modelFieldID, err := resolver.ModelFieldID("Jira Ticket")
	assert.NoError(t, err)
	assert.Equal(t, int64(401), modelFieldID)
}

func TestGetMetadataFieldResolverUnsupportedEntity(t *testing.T) {
	// Given
	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/%d/", ProjectID), func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetProjectResponseJson)
		assert.NoError(t, err)
	})
	mux.HandleFunc(fmt.Sprintf("/api/v1/organisations/%d/metadata-model-fields/supported-content-types/", OrganisationIDForMetadata),
		func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("Content-Type", "application/json")
			_, err := io.WriteString(rw, `[{"id": 41, "app_label": "features", "model": "feature"}]`)
			assert.NoError(t, err)
		})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	_, err := client.GetMetadataFieldResolver(ProjectID, flagsmithapi.MetadataEntitySegment)

	// Then
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no content type found for 'segment'")
}

func TestBuildMetadataAndResolveNamesRoundTrip(t *testing.T) {
	// Given
	server, _ := metadataTestServer(t)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")
	values := map[string]string{"Jira Ticket": "PROD-123", "Story Points": "3"}

	// When
	metadata, err := client.BuildMetadata(ProjectID, flagsmithapi.MetadataEntityFeature, values)

	// Then
	require.NoError(t, err)
	require.NotNil(t, metadata)
	// Sorted by name, so request bodies are deterministic.
	require.Equal(t, 2, len(*metadata))
	assert.Equal(t, FeatureModelFieldID, (*metadata)[0].ModelField)
	assert.Equal(t, "PROD-123", (*metadata)[0].FieldValue)
	assert.Equal(t, int64(304), (*metadata)[1].ModelField)
	assert.Equal(t, "3", (*metadata)[1].FieldValue)

	// And back again
	resolved, unresolved, err := client.ResolveMetadataNames(ProjectID, flagsmithapi.MetadataEntityFeature, *metadata)
	require.NoError(t, err)
	assert.Empty(t, unresolved)
	assert.Equal(t, values, resolved)
}

// A nil or empty map must produce a non-nil empty slice, which serialises as
// `"metadata": []` rather than omitting the key.
func TestBuildMetadataEmptyValues(t *testing.T) {
	// Given
	server, counter := metadataTestServer(t)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	metadata, err := client.BuildMetadata(ProjectID, flagsmithapi.MetadataEntityFeature, nil)

	// Then
	require.NoError(t, err)
	require.NotNil(t, metadata)
	assert.Equal(t, 0, len(*metadata))

	body, err := json.Marshal(flagsmithapi.Feature{Name: "x", Metadata: metadata})
	require.NoError(t, err)
	assert.Contains(t, string(body), `"metadata":[]`)

	// No custom field definitions are needed to send an empty set.
	assert.Equal(t, 0, counter.count("fields"))
}

func TestBuildMetadataRejectsInvalidValue(t *testing.T) {
	// Given
	server, _ := metadataTestServer(t)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	_, err := client.BuildMetadata(ProjectID, flagsmithapi.MetadataEntityFeature,
		map[string]string{"Story Points": "not-a-number"})

	// Then
	require.IsType(t, flagsmithapi.MetadataFieldInvalidValueError{}, err)
	invalid := err.(flagsmithapi.MetadataFieldInvalidValueError)
	assert.Equal(t, "Story Points", invalid.Name)
	assert.Equal(t, "int", invalid.FieldType)
	assert.Equal(t, "not-a-number", invalid.Value)
}

func TestResolveMetadataNamesSkipsUnknownModelField(t *testing.T) {
	// Given
	server, _ := metadataTestServer(t)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	resolved, unresolved, err := client.ResolveMetadataNames(ProjectID, flagsmithapi.MetadataEntityFeature,
		[]flagsmithapi.Metadata{
			{ModelField: FeatureModelFieldID, FieldValue: "PROD-123"},
			{ModelField: 9999, FieldValue: "orphaned"},
		})

	// Then
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"Jira Ticket": "PROD-123"}, resolved)
	assert.Equal(t, []int64{9999}, unresolved)
}

func TestMetadataFieldValidateValue(t *testing.T) {
	testCases := []struct {
		fieldType string
		value     string
		valid     bool
	}{
		{"int", "1", true},
		{"int", "-3", true},
		{"int", "1.5", false},
		{"int", "abc", false},
		{"bool", "true", true},
		{"bool", "TRUE", true},
		{"bool", "false", true},
		{"bool", "1", false},
		{"url", "https://example.com/x", true},
		{"url", "example.com", false},
		{"url", "/foo", false},
		{"str", "anything at all", true},
		{"multiline_str", "line one\nline two", true},
		// Unknown types are accepted, leaving the API authoritative.
		{"something_new", "whatever", true},
		// field_value is blank=False, so empty values are always rejected.
		{"str", "", false},
	}

	for _, testCase := range testCases {
		field := flagsmithapi.MetadataField{Name: "Test Field", Type: testCase.fieldType}
		err := field.ValidateValue(testCase.value)

		if testCase.valid {
			assert.NoError(t, err, "%s: %q should be valid", testCase.fieldType, testCase.value)
		} else {
			assert.Error(t, err, "%s: %q should be invalid", testCase.fieldType, testCase.value)
		}
	}

	// Over the length limit
	tooLong := make([]byte, flagsmithapi.MetadataFieldValueMaxLength+1)
	for i := range tooLong {
		tooLong[i] = 'a'
	}
	field := flagsmithapi.MetadataField{Name: "Test Field", Type: "str"}
	assert.Error(t, field.ValidateValue(string(tooLong)))
	assert.NoError(t, field.ValidateValue(string(tooLong[:flagsmithapi.MetadataFieldValueMaxLength])))
}

// Feature.UnmarshalJSON decodes into an explicit allowlist, so a field added to Feature
// but not to that allowlist is silently dropped. This walks the struct to catch it.
func TestFeatureUnmarshalJSONDoesNotDropFields(t *testing.T) {
	// Given a payload that populates every JSON tagged field on Feature
	payload := `{
		"id": 1,
		"uuid": "10421b1f-5f29-4da9-abe2-30f88c07c9e8",
		"name": "test_feature",
		"type": "STANDARD",
		"description": "a description",
		"initial_value": "some_value",
		"default_enabled": true,
		"is_archived": true,
		"owners": [{"id": 1}],
		"group_owners": [{"id": 3}],
		"tags": [1],
		"project": 10,
		"metadata": [{"id": 5, "model_field": 301, "field_value": "PROD-123"}]
	}`

	// When
	feature := flagsmithapi.Feature{}
	require.NoError(t, json.Unmarshal([]byte(payload), &feature))

	// Then no field is left at its zero value
	value := reflect.ValueOf(feature)
	featureType := value.Type()
	for i := 0; i < featureType.NumField(); i++ {
		field := featureType.Field(i)
		if field.Tag.Get("json") == "-" {
			// Not sent by the API, populated by the client.
			continue
		}
		assert.False(t, value.Field(i).IsZero(),
			"Feature.%s was not populated: add it to the allowlist in Feature.UnmarshalJSON", field.Name)
	}

	require.NotNil(t, feature.Metadata)
	require.Equal(t, 1, len(*feature.Metadata))
	assert.Equal(t, FeatureModelFieldID, (*feature.Metadata)[0].ModelField)
	assert.Equal(t, "PROD-123", (*feature.Metadata)[0].FieldValue)
}

func TestGetFeatureHydratesMetadata(t *testing.T) {
	// Given
	mux := http.NewServeMux()
	counter := newRequestCounter()

	mux.HandleFunc(fmt.Sprintf("/api/v1/features/get-by-uuid/%s/", FeatureUUID), func(rw http.ResponseWriter, req *http.Request) {
		counter.record("get-by-uuid")
		rw.Header().Set("Content-Type", "application/json")
		// CreateFeatureSerializer, i.e. no metadata.
		_, err := io.WriteString(rw, CreateFeatureResponseJson)
		assert.NoError(t, err)
	})

	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/%d/", ProjectID), func(rw http.ResponseWriter, req *http.Request) {
		counter.record("project")
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetProjectResponseJson)
		assert.NoError(t, err)
	})

	// The longer literal path must be registered, otherwise the project subtree pattern
	// above silently swallows this request.
	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/%d/features/%d/", ProjectID, FeatureID),
		func(rw http.ResponseWriter, req *http.Request) {
			counter.record("retrieve")
			rw.Header().Set("Content-Type", "application/json")
			_, err := io.WriteString(rw, `{"id": 1, "name": "test_feature", "project": 10,
				"metadata": [{"id": 5, "model_field": 301, "field_value": "PROD-123"}]}`)
			assert.NoError(t, err)
		})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	feature, err := client.GetFeature(FeatureUUID)

	// Then
	require.NoError(t, err)
	assert.Equal(t, 1, counter.count("retrieve"))

	require.NotNil(t, feature.Metadata)
	require.Equal(t, 1, len(*feature.Metadata))
	assert.Equal(t, FeatureModelFieldID, (*feature.Metadata)[0].ModelField)
	assert.Equal(t, "PROD-123", (*feature.Metadata)[0].FieldValue)

	// The hydration response must not clobber anything the get-by-uuid response set.
	assert.Equal(t, "STANDARD", *feature.Type)
	assert.Equal(t, &[]int64{1, 2}, feature.Owners)
	assert.Equal(t, ProjectUUID, feature.ProjectUUID)
}

func TestCreateFeatureSendsMetadata(t *testing.T) {
	// Given
	newFeature := flagsmithapi.Feature{
		Name:      FeatureName,
		ProjectID: &[]int64{ProjectID}[0],
		Tags:      []int64{},
		Metadata:  &[]flagsmithapi.Metadata{{ModelField: FeatureModelFieldID, FieldValue: "PROD-123"}},
	}

	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/%d/features/", ProjectID), func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)

		rawBody, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(rawBody), `"metadata":[{"model_field":301,"field_value":"PROD-123"}]`)

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, `{"id": 1, "name": "test_feature", "project": 10, "type": "STANDARD",
			"metadata": [{"id": 5, "model_field": 301, "field_value": "PROD-123"}]}`)
		assert.NoError(t, err)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	err := client.CreateFeature(&newFeature)

	// Then
	require.NoError(t, err)
	// The create response carries metadata, so no extra read is needed.
	require.NotNil(t, newFeature.Metadata)
	require.Equal(t, 1, len(*newFeature.Metadata))
	assert.Equal(t, int64(5), (*newFeature.Metadata)[0].ID)
}

func TestFeatureWithNilMetadataOmitsTheKey(t *testing.T) {
	body, err := json.Marshal(flagsmithapi.Feature{Name: "x", Tags: []int64{}})
	require.NoError(t, err)
	assert.NotContains(t, string(body), "metadata")
}

func TestGetSegmentReturnsMetadata(t *testing.T) {
	// Given
	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("/api/v1/segments/get-by-uuid/%s/", SegmentUUID), func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, `{"id": 300, "uuid": "f6c714d3-94e7-4b14-9117-8dd9db91bc19",
			"name": "one_matches_one", "project": 10, "rules": [],
			"metadata": [{"id": 6, "model_field": 302, "field_value": "PROD-123"}]}`)
		assert.NoError(t, err)
	})

	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/%d/", ProjectID), func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetProjectResponseJson)
		assert.NoError(t, err)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	segment, err := client.GetSegment(SegmentUUID)

	// Then
	require.NoError(t, err)
	require.NotNil(t, segment.Metadata)
	require.Equal(t, 1, len(*segment.Metadata))
	assert.Equal(t, SegmentModelFieldID, (*segment.Metadata)[0].ModelField)
	assert.Equal(t, "PROD-123", (*segment.Metadata)[0].FieldValue)
}

func TestGetEnvironmentByUUIDReturnsMetadata(t *testing.T) {
	// Given
	environmentUUID := "b7a0d0f9-2bbd-4b6e-a2c9-a1d2ec0f3a1b"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, fmt.Sprintf(`{"id": 100, "uuid": "%s", "name": "Development",
			"api_key": "%s", "description": "", "project": 10,
			"metadata": [{"id": 7, "model_field": 303, "field_value": "PROD-123"}]}`,
			environmentUUID, EnvironmentAPIKey))
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	environment, err := client.GetEnvironmentByUUID(environmentUUID)

	// Then
	require.NoError(t, err)
	require.NotNil(t, environment.Metadata)
	require.Equal(t, 1, len(*environment.Metadata))
	assert.Equal(t, EnvironmentModelFieldID, (*environment.Metadata)[0].ModelField)
}
