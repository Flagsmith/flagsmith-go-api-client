package flagsmithapi_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	flagsmithapi "github.com/Flagsmith/flagsmith-go-api-client"
)

const GetFeatureStateJson = `
{
  "count": 1,
  "next": null,
  "previous": null,
  "results": [
    {
      "id": 1,
      "feature_state_value": "some_value",
      "multivariate_feature_state_values": [],
      "identity": null,
      "enabled": false,
      "created_at": "2022-04-02T06:32:07.130623Z",
      "updated_at": "2022-06-24T06:14:43.214447Z",
      "version": 1,
      "live_from": "2022-04-02T06:32:07.161622Z",
      "feature": 1,
      "environment": 1,
      "feature_segment": null,
      "change_request": null
    }
  ]
}
`
const UpdateFeatureStateResponseJson = `
{
  "id": 1,
  "feature_state_value": {
    "type": "unicode",
    "string_value": "updated_value",
    "integer_value": null,
    "boolean_value": null
  },
  "multivariate_feature_state_values": [],
  "enabled": true,
  "created_at": "2022-04-02T06:32:07.130623Z",
  "updated_at": "2022-06-23T06:58:53.519204Z",
  "version": 1,
  "live_from": "2022-04-02T06:32:07.161622Z",
  "feature": 1,
  "environment": 1,
  "identity": null,
  "feature_segment": null,
  "change_request": null
}
`
const FeatureID int64 = 1
const FeatureUUID = "10421b1f-5f29-4da9-abe2-30f88c07c9e8"

func TestGetFeatureState(t *testing.T) {
	// Given
	masterAPIKey := "test_key"
	environmentKey := "test_env_key"
	featureName := "test_feature"
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/environments/%s/featurestates/", environmentKey), req.URL.Path)
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "Api-Key "+masterAPIKey, req.Header.Get("Authorization"))

		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetFeatureStateJson)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := flagsmithapi.NewClient(masterAPIKey, server.URL+"/api/v1")

	// When
	fs, err := client.GetEnvironmentFeatureState(environmentKey, featureName)

	// Then
	// assert that we did not receive an error
	assert.NoError(t, err)

	var nilIntPointer *int64
	var nilBoolPointer *bool

	// assert that the returned feature state is correct
	assert.Equal(t, int64(1), fs.ID)
	assert.Equal(t, int64(1), fs.Feature)
	assert.Equal(t, int64(1), fs.Environment)
	assert.Equal(t, false, fs.Enabled)

	assert.Equal(t, "some_value", *fs.FeatureStateValue.StringValue)
	assert.Equal(t, "unicode", fs.FeatureStateValue.Type)
	assert.Equal(t, nilIntPointer, fs.FeatureStateValue.IntegerValue)
	assert.Equal(t, nilBoolPointer, fs.FeatureStateValue.BooleanValue)

}

func TestUpdateFeatureState(t *testing.T) {
	// Given
	masterAPIKey := "test_key"
	newFsValue := "updated_value"

	fsValue := flagsmithapi.FeatureStateValue{
		Type:        "unicode",
		StringValue: &newFsValue,
	}
	fs := flagsmithapi.FeatureState{
		ID:                1,
		FeatureStateValue: &fsValue,
		Enabled:           true,
		Feature:           1,
		Environment:       1,
	}

	expectedRequestBody := `{"id":1,"feature_state_value":{"type":"unicode","string_value":"updated_value","integer_value":null,"boolean_value":null},` +
		`"enabled":true,"feature":1,"environment":1}`

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/v1/features/featurestates/1/", req.URL.Path)
		assert.Equal(t, "PUT", req.Method)
		assert.Equal(t, "Api-Key "+masterAPIKey, req.Header.Get("Authorization"))

		// Test that we sent the correct body
		rawBody, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequestBody, string(rawBody))

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, UpdateFeatureStateResponseJson)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := flagsmithapi.NewClient(masterAPIKey, server.URL+"/api/v1")

	updated_fs, err := client.UpdateFeatureState(&fs)
	assert.NoError(t, err)

	var nilIntPointer *int64
	var nilBoolPointer *bool

	// assert that the returned feature state is correct
	assert.Equal(t, int64(1), fs.ID)
	assert.Equal(t, int64(1), fs.Feature)
	assert.Equal(t, int64(1), fs.Environment)
	assert.Equal(t, true, fs.Enabled)

	assert.Equal(t, newFsValue, *updated_fs.FeatureStateValue.StringValue)
	assert.Equal(t, "unicode", fs.FeatureStateValue.Type)
	assert.Equal(t, nilIntPointer, fs.FeatureStateValue.IntegerValue)
	assert.Equal(t, nilBoolPointer, fs.FeatureStateValue.BooleanValue)

}

const GetProjectResponseJson = `
    {
        "id": 1,
        "uuid": "cba035f8-d801-416f-a985-ce6e05acbe13",
        "name": "project-1",
        "organisation": 1,
        "hide_disabled_flags": false,
        "enable_dynamo_db": true,
        "migration_status": "NOT_APPLICABLE",
        "use_edge_identities": false
    }
`
const ProjectID int64 = 1
const ProjectUUID = "cba035f8-d801-416f-a985-ce6e05acbe13"

func TestGetProject(t *testing.T) {
	// Given
	masterAPIKey := "master_api_key"
	projectUUID := "cba035f8-d801-416f-a985-ce6e05acbe13"

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/v1/projects/get-by-uuid/cba035f8-d801-416f-a985-ce6e05acbe13/", req.URL.Path)
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "Api-Key "+masterAPIKey, req.Header.Get("Authorization"))

		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetProjectResponseJson)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := flagsmithapi.NewClient(masterAPIKey, server.URL+"/api/v1")

	// When
	project, err := client.GetProject(projectUUID)

	// Then
	assert.NoError(t, err)

	assert.Equal(t, int64(1), project.ID)
	assert.Equal(t, projectUUID, project.UUID)
	assert.Equal(t, "project-1", project.Name)
}

const CreateFeatureResponseJson = `
{
    "id": 1,
    "name": "test_feature",
    "project": 1,
    "type": "STANDARD",
    "default_enabled": false,
    "initial_value": null,
    "created_date": "2022-08-24T03:34:55.862503Z",
    "description": null,
    "tags": [],
    "multivariate_options": [],
    "is_archived": false,
    "owners": []
}
`

func TestCreateFeatureFetchesProjectIfProjectIDIsNil(t *testing.T) {
	// Given
	masterAPIKey := "master_api_key"
	projectUUID := "cba035f8-d801-416f-a985-ce6e05acbe13"
	newFeature := flagsmithapi.Feature{
		Name:        "test_feature",
		ProjectUUID: projectUUID,
	}
	mux := http.NewServeMux()
	expectedRequestBody := `{"name":"test_feature"}`

	mux.HandleFunc("/api/v1/projects/1/features/", func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "Api-Key "+masterAPIKey, req.Header.Get("Authorization"))

		rawBody, err := io.ReadAll(req.Body)
		assert.Equal(t, expectedRequestBody, string(rawBody))
		assert.NoError(t, err)

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, CreateFeatureResponseJson)
		assert.NoError(t, err)
	})

	mux.HandleFunc("/api/v1/projects/get-by-uuid/cba035f8-d801-416f-a985-ce6e05acbe13/", func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetProjectResponseJson)
		assert.NoError(t, err)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := flagsmithapi.NewClient(masterAPIKey, server.URL+"/api/v1")

	// When
	err := client.CreateFeature(&newFeature)

	// Then

	assert.NoError(t, err)
	assert.Equal(t, int64(1), *newFeature.ID)
	assert.Equal(t, "test_feature", newFeature.Name)
	assert.Equal(t, "STANDARD", *newFeature.Type)
	assert.Equal(t, false, newFeature.DefaultEnabled)
	assert.Equal(t, false, newFeature.IsArchived)

	assert.Equal(t, "", newFeature.InitialValue)

	assert.Equal(t, int64(1), *newFeature.ProjectID)
	assert.Equal(t, "cba035f8-d801-416f-a985-ce6e05acbe13", newFeature.ProjectUUID)

}

const CreateMVFeatureResponseJson = `
{
    "id": 1,
    "name": "test_feature",
    "type": "MULTIVARIATE",
    "default_enabled": false,
    "initial_value": null,
    "created_date": "2022-08-26T03:33:41.492354Z",
    "description": null,
    "tags": [],
    "multivariate_options": [
        {
            "id": 1,
            "type": "unicode",
            "integer_value": null,
            "string_value": "value_one",
            "boolean_value": null,
            "default_percentage_allocation": 50.0
        },
        {
            "id": 2,
            "type": "unicode",
            "integer_value": null,
            "string_value": "value_two",
            "boolean_value": null,
            "default_percentage_allocation": 50.0
        }
    ],
    "is_archived": false,
    "owners": []
}
`

func TestCreateMVFeature(t *testing.T) {
	// Given
	masterAPIKey := "master_api_key"
	featureType := "MULTIVARIATE"
	projectID := int64(1)
	projectUUID := "cba035f8-d801-416f-a985-ce6e05acbe13"
	mvValueOne := "value_one"
	mvValueTwo := "value_two"

	mvOptions := []flagsmithapi.MultivariateOption{
		{
			Type:                        "unicode",
			StringValue:                 &mvValueOne,
			DefaultPercentageAllocation: float64(50),
		},
		{
			Type:                        "unicode",
			StringValue:                 &mvValueTwo,
			DefaultPercentageAllocation: float64(50),
		},
	}

	newFeature := flagsmithapi.Feature{
		Name:                "test_feature",
		ProjectUUID:         projectUUID,
		ProjectID:           &projectID,
		Type:                &featureType,
		MultivariateOptions: &mvOptions,
	}

	expectedRequestBody := `{"name":"test_feature","type":"MULTIVARIATE","multivariate_options":[{"type":"unicode","string_value":"value_one","default_percentage_allocation":50},{"type":"unicode","string_value":"value_two","default_percentage_allocation":50}],"project":1}`

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/v1/projects/1/features/", req.URL.Path)
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "Api-Key "+masterAPIKey, req.Header.Get("Authorization"))

		// Test that we sent the correct body
		rawBody, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequestBody, string(rawBody))

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, CreateMVFeatureResponseJson)
		assert.NoError(t, err)

	}))
	defer server.Close()

	client := flagsmithapi.NewClient(masterAPIKey, server.URL+"/api/v1")

	// When
	err := client.CreateFeature(&newFeature)

	// Then
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), *newFeature.ID)
	assert.Equal(t, "test_feature", newFeature.Name)

	assert.Equal(t, "MULTIVARIATE", *newFeature.Type)
	assert.Equal(t, false, newFeature.DefaultEnabled)
	assert.Equal(t, false, newFeature.IsArchived)

	assert.Equal(t, "", newFeature.InitialValue)

	assert.Equal(t, int64(1), *newFeature.ProjectID)
	assert.Equal(t, "cba035f8-d801-416f-a985-ce6e05acbe13", newFeature.ProjectUUID)
	assert.Equal(t, 2, len(*newFeature.MultivariateOptions))

	assert.Equal(t, int64(1), *(*newFeature.MultivariateOptions)[0].ID)
	assert.Equal(t, "value_one", *(*newFeature.MultivariateOptions)[0].StringValue)

	assert.Equal(t, int64(2), *(*newFeature.MultivariateOptions)[1].ID)
	assert.Equal(t, "value_two", *(*newFeature.MultivariateOptions)[1].StringValue)

}

func TestDeleteFeature(t *testing.T) {
	// Given
	masterAPIKey := "master_api_key"

	projectID := int64(1)
	featureID := int64(1)

	requestReceived := struct {
		mu                sync.Mutex
		isRequestReceived bool
	}{}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestReceived.mu.Lock()
		requestReceived.isRequestReceived = true
		requestReceived.mu.Unlock()

		assert.Equal(t, "/api/v1/projects/1/features/1/", req.URL.Path)
		assert.Equal(t, "DELETE", req.Method)
		assert.Equal(t, "Api-Key "+masterAPIKey, req.Header.Get("Authorization"))

	}))
	defer server.Close()

	client := flagsmithapi.NewClient(masterAPIKey, server.URL+"/api/v1")

	// When
	err := client.DeleteFeature(projectID, featureID)

	// Then
	requestReceived.mu.Lock()
	assert.True(t, requestReceived.isRequestReceived)
	assert.NoError(t, err)
}

func TestUpdateFeature(t *testing.T) {
	// Given
	masterAPIKey := "master_api_key"
	projectID := int64(1)
	projectUUID := "cba035f8-d801-416f-a985-ce6e05acbe13"
	featureID := int64(1)

	description := "feature description"

	feature := flagsmithapi.Feature{
		Name:        "test_feature",
		ID:          &featureID,
		ProjectUUID: projectUUID,
		ProjectID:   &projectID,
		Description: &description,
	}

	expectedRequestBody := `{"name":"test_feature","id":1,"description":"feature description","project":1}`
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/v1/projects/1/features/1/", req.URL.Path)
		assert.Equal(t, "PUT", req.Method)
		assert.Equal(t, "Api-Key "+masterAPIKey, req.Header.Get("Authorization"))

		// Test that we sent the correct body
		rawBody, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequestBody, string(rawBody))

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, CreateFeatureResponseJson)
		assert.NoError(t, err)

	}))
	defer server.Close()

	client := flagsmithapi.NewClient(masterAPIKey, server.URL+"/api/v1")

	// When
	err := client.UpdateFeature(&feature)

	// Then
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), *feature.ID)
	assert.Equal(t, "test_feature", feature.Name)
	assert.Equal(t, "STANDARD", *feature.Type)
	assert.Equal(t, false, feature.DefaultEnabled)
	assert.Equal(t, false, feature.IsArchived)

	assert.Equal(t, "", feature.InitialValue)

	assert.Equal(t, int64(1), *feature.ProjectID)
	assert.Equal(t, "cba035f8-d801-416f-a985-ce6e05acbe13", feature.ProjectUUID)

}

func TestGetFeature(t *testing.T) {
	// Given
	masterAPIKey := "master_api_key"
	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("/api/v1/features/get-by-uuid/%s/", FeatureUUID), func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "Api-Key "+masterAPIKey, req.Header.Get("Authorization"))

		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, CreateFeatureResponseJson)
		assert.NoError(t, err)

	})

	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/%d/", ProjectID), func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetProjectResponseJson)
		assert.NoError(t, err)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := flagsmithapi.NewClient(masterAPIKey, server.URL+"/api/v1")

	// When
	feature, err := client.GetFeature(FeatureUUID)

	// Then
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), *feature.ID)
	assert.Equal(t, "test_feature", feature.Name)
	assert.Equal(t, "STANDARD", *feature.Type)
	assert.Equal(t, false, feature.DefaultEnabled)
	assert.Equal(t, false, feature.IsArchived)

	assert.Equal(t, "", feature.InitialValue)

	assert.Equal(t, ProjectID, *feature.ProjectID)
	assert.Equal(t, ProjectUUID, feature.ProjectUUID)

}
