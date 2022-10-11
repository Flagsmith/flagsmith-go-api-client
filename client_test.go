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
      "environment": 100,
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
  "environment": 100,
  "identity": null,
  "feature_segment": null,
  "change_request": null
}
`
const FeatureID int64 = 1
const EnvironmentID int64 = 100
const FeatureUUID = "10421b1f-5f29-4da9-abe2-30f88c07c9e8"
const MasterAPIKey = "master_api_key"

func TestGetFeatureState(t *testing.T) {
	// Given
	environmentKey := "test_env_key"
	featureName := "test_feature"

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/environments/%s/featurestates/", environmentKey), req.URL.Path)
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetFeatureStateJson)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	fs, err := client.GetEnvironmentFeatureState(environmentKey, featureName)

	// Then
	// assert that we did not receive an error
	assert.NoError(t, err)

	var nilIntPointer *int64
	var nilBoolPointer *bool

	// assert that the returned feature state is correct
	assert.Equal(t, int64(1), fs.ID)
	assert.Equal(t, FeatureID, fs.Feature)
	assert.Equal(t, EnvironmentID, fs.Environment)
	assert.Equal(t, false, fs.Enabled)

	assert.Equal(t, "some_value", *fs.FeatureStateValue.StringValue)
	assert.Equal(t, "unicode", fs.FeatureStateValue.Type)
	assert.Equal(t, nilIntPointer, fs.FeatureStateValue.IntegerValue)
	assert.Equal(t, nilBoolPointer, fs.FeatureStateValue.BooleanValue)

}

func TestUpdateFeatureState(t *testing.T) {
	// Given
	newFsValue := "updated_value"

	fsValue := flagsmithapi.FeatureStateValue{
		Type:        "unicode",
		StringValue: &newFsValue,
	}
	fs := flagsmithapi.FeatureState{
		ID:                1,
		FeatureStateValue: &fsValue,
		Enabled:           true,
		Feature:           FeatureID,
		Environment:       EnvironmentID,
	}

	expectedRequestBody := fmt.Sprintf(`{"id":1,"feature_state_value":{"type":"unicode","string_value":"updated_value","integer_value":null,"boolean_value":null},`+
		`"enabled":true,"feature":%d,"environment":%d}`, FeatureID, EnvironmentID)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/v1/features/featurestates/1/", req.URL.Path)
		assert.Equal(t, "PUT", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		// Test that we sent the correct body
		rawBody, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequestBody, string(rawBody))

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, UpdateFeatureStateResponseJson)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	updated_fs, err := client.UpdateFeatureState(&fs)
	assert.NoError(t, err)

	var nilIntPointer *int64
	var nilBoolPointer *bool

	// assert that the returned feature state is correct
	assert.Equal(t, int64(1), fs.ID)
	assert.Equal(t, FeatureID, fs.Feature)
	assert.Equal(t, EnvironmentID, fs.Environment)
	assert.Equal(t, true, fs.Enabled)

	assert.Equal(t, newFsValue, *updated_fs.FeatureStateValue.StringValue)
	assert.Equal(t, "unicode", fs.FeatureStateValue.Type)
	assert.Equal(t, nilIntPointer, fs.FeatureStateValue.IntegerValue)
	assert.Equal(t, nilBoolPointer, fs.FeatureStateValue.BooleanValue)

}

const GetProjectResponseJson = `
    {
        "id": 10,
        "uuid": "cba035f8-d801-416f-a985-ce6e05acbe13",
        "name": "project-1",
        "organisation": 1,
        "hide_disabled_flags": false,
        "enable_dynamo_db": true,
        "migration_status": "NOT_APPLICABLE",
        "use_edge_identities": false
    }
`
const ProjectID int64 = 10
const ProjectUUID = "cba035f8-d801-416f-a985-ce6e05acbe13"

func TestGetProject(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/projects/get-by-uuid/%s/", ProjectUUID), req.URL.Path)
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetProjectResponseJson)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	project, err := client.GetProject(ProjectUUID)

	// Then
	assert.NoError(t, err)

	assert.Equal(t, ProjectID, project.ID)
	assert.Equal(t, ProjectUUID, project.UUID)
	assert.Equal(t, "project-1", project.Name)
}

const CreateFeatureResponseJson = `
{
    "id": 1,
    "name": "test_feature",
    "project": 10,
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
const FeatureName = "test_feature"

func TestCreateFeatureFetchesProjectIfProjectIDIsNil(t *testing.T) {
	// Given
	newFeature := flagsmithapi.Feature{
		Name:        FeatureName,
		ProjectUUID: ProjectUUID,
	}
	mux := http.NewServeMux()
	expectedRequestBody := `{"name":"test_feature"}`

	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/%d/features/", ProjectID), func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		rawBody, err := io.ReadAll(req.Body)
		assert.Equal(t, expectedRequestBody, string(rawBody))
		assert.NoError(t, err)

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, CreateFeatureResponseJson)
		assert.NoError(t, err)
	})

	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/get-by-uuid/%s/", ProjectUUID), func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetProjectResponseJson)
		assert.NoError(t, err)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	err := client.CreateFeature(&newFeature)

	// Then

	assert.NoError(t, err)
	assert.Equal(t, FeatureID, *newFeature.ID)
	assert.Equal(t, FeatureName, newFeature.Name)
	assert.Equal(t, "STANDARD", *newFeature.Type)
	assert.Equal(t, false, newFeature.DefaultEnabled)
	assert.Equal(t, false, newFeature.IsArchived)

	assert.Equal(t, "", newFeature.InitialValue)

	assert.Equal(t, ProjectID, *newFeature.ProjectID)
	assert.Equal(t, ProjectUUID, newFeature.ProjectUUID)

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

func TestDeleteFeature(t *testing.T) {
	// Given

	requestReceived := struct {
		mu                sync.Mutex
		isRequestReceived bool
	}{}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestReceived.mu.Lock()
		requestReceived.isRequestReceived = true
		requestReceived.mu.Unlock()

		assert.Equal(t, fmt.Sprintf("/api/v1/projects/%d/features/%d/", ProjectID, FeatureID), req.URL.Path)
		assert.Equal(t, "DELETE", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

	}))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	err := client.DeleteFeature(ProjectID, FeatureID)

	// Then
	requestReceived.mu.Lock()
	assert.True(t, requestReceived.isRequestReceived)
	assert.NoError(t, err)
}

func TestUpdateFeature(t *testing.T) {
	// Given
	projectID := ProjectID
	featureID := FeatureID

	description := "feature description"

	feature := flagsmithapi.Feature{
		Name:        FeatureName,
		ID:          &featureID,
		ProjectUUID: ProjectUUID,
		ProjectID:   &projectID,
		Description: &description,
	}

	expectedRequestBody := fmt.Sprintf(`{"name":"%s","id":%d,"description":"feature description","project":%d}`, FeatureName, featureID, ProjectID)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/projects/%d/features/%d/", ProjectID, FeatureID), req.URL.Path)
		assert.Equal(t, "PUT", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		// Test that we sent the correct body
		rawBody, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequestBody, string(rawBody))

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, CreateFeatureResponseJson)
		assert.NoError(t, err)

	}))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	err := client.UpdateFeature(&feature)

	// Then
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, FeatureID, *feature.ID)
	assert.Equal(t, FeatureName, feature.Name)
	assert.Equal(t, "STANDARD", *feature.Type)
	assert.Equal(t, false, feature.DefaultEnabled)
	assert.Equal(t, false, feature.IsArchived)

	assert.Equal(t, "", feature.InitialValue)

	assert.Equal(t, ProjectID, *feature.ProjectID)
	assert.Equal(t, ProjectUUID, feature.ProjectUUID)

}

func TestGetFeature(t *testing.T) {
	// Given
	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("/api/v1/features/get-by-uuid/%s/", FeatureUUID), func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

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

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	feature, err := client.GetFeature(FeatureUUID)

	// Then
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, FeatureID, *feature.ID)
	assert.Equal(t, FeatureName, feature.Name)
	assert.Equal(t, "STANDARD", *feature.Type)
	assert.Equal(t, false, feature.DefaultEnabled)
	assert.Equal(t, false, feature.IsArchived)

	assert.Equal(t, "", feature.InitialValue)

	assert.Equal(t, ProjectID, *feature.ProjectID)
	assert.Equal(t, ProjectUUID, feature.ProjectUUID)

}

const GetMVFeatureOptionResponseJson = `
{
    "id": 150,
    "uuid": "8d3512d3-721a-4cae-9855-56c02cb0afe9",
    "type": "unicode",
    "string_value": "option_value_30",
    "boolean_value": null,
    "default_percentage_allocation": 60.0,
    "feature": 1
}
`
const MVFeatureOptionUUID = "8d3512d3-721a-4cae-9855-56c02cb0afe9"
const MVFeatureOptionID int64 = 150

func TestGetFeatureMVOption(t *testing.T) {
	// Given
	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("/api/v1/multivariate/options/get-by-uuid/%s/", MVFeatureOptionUUID), func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetMVFeatureOptionResponseJson)
		assert.NoError(t, err)

	})

	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/%d/", ProjectID), func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetProjectResponseJson)
		assert.NoError(t, err)
	})

	mux.HandleFunc(fmt.Sprintf("/api/v1/features/get-by-uuid/%s/", FeatureUUID), func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, CreateFeatureResponseJson)
		assert.NoError(t, err)

	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	featureMVOption, err := client.GetFeatureMVOption(FeatureUUID, MVFeatureOptionUUID)

	// Then
	var nilIntPointer *int64
	var nilBoolPointer *bool

	assert.NoError(t, err)
	assert.Equal(t, MVFeatureOptionUUID, featureMVOption.UUID)
	assert.Equal(t, MVFeatureOptionID, featureMVOption.ID)

	assert.Equal(t, "unicode", featureMVOption.Type)
	assert.Equal(t, "option_value_30", *featureMVOption.StringValue)
	assert.Equal(t, nilIntPointer, featureMVOption.IntegerValue)
	assert.Equal(t, nilBoolPointer, featureMVOption.BooleanValue)

	assert.Equal(t, float64(60), featureMVOption.DefaultPercentageAllocation)
	assert.Equal(t, FeatureID, *featureMVOption.FeatureID)
	assert.Equal(t, FeatureUUID, featureMVOption.FeatureUUID)

	assert.Equal(t, ProjectID, *featureMVOption.ProjectID)

}

func TestDeleteFeatureMVOption(t *testing.T) {
	// Given
	requestReceived := struct {
		mu                sync.Mutex
		isRequestReceived bool
	}{}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestReceived.mu.Lock()
		requestReceived.isRequestReceived = true
		requestReceived.mu.Unlock()

		assert.Equal(t, fmt.Sprintf("/api/v1/projects/%d/features/%d/mv-options/%d/", ProjectID, FeatureID, MVFeatureOptionID), req.URL.Path)
		assert.Equal(t, "DELETE", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

	}))

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	err := client.DeleteFeatureMVOption(ProjectID, FeatureID, MVFeatureOptionID)

	// Then
	requestReceived.mu.Lock()
	assert.True(t, requestReceived.isRequestReceived)
	assert.NoError(t, err)
}

func TestUpdateFeatureMVOption(t *testing.T) {
	// Given
	featureID := FeatureID
	projectID := ProjectID
	stringValue := "option_value_30"
	defaultPercentageAllocation := float64(60)
	featureMVOption := flagsmithapi.FeatureMultivariateOption{
		ID:                          MVFeatureOptionID,
		Type:                        "unicode",
		UUID:                        "", // avoid setting UUID to test that update refreshes the struct
		FeatureID:                   &featureID,
		StringValue:                 &stringValue,
		DefaultPercentageAllocation: defaultPercentageAllocation,
		FeatureUUID:                 FeatureUUID,
		ProjectID:                   &projectID,
	}

	expectedRequestBody := fmt.Sprintf(`{"id":%d,"type":"unicode","feature":%d,"string_value":"%s","default_percentage_allocation":%.0f}`, MVFeatureOptionID, featureID, stringValue, defaultPercentageAllocation)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/projects/%d/features/%d/mv-options/%d/", ProjectID, FeatureID, MVFeatureOptionID), req.URL.Path)
		assert.Equal(t, "PUT", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		// Test that we sent the correct body
		rawBody, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequestBody, string(rawBody))

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, GetMVFeatureOptionResponseJson)
		assert.NoError(t, err)

	}))

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	err := client.UpdateFeatureMVOption(&featureMVOption)

	// Then
	var nilIntPointer *int64
	var nilBoolPointer *bool

	assert.NoError(t, err)
	assert.Equal(t, MVFeatureOptionUUID, featureMVOption.UUID)

	assert.Equal(t, "unicode", featureMVOption.Type)
	assert.Equal(t, stringValue, *featureMVOption.StringValue)
	assert.Equal(t, nilIntPointer, featureMVOption.IntegerValue)
	assert.Equal(t, nilBoolPointer, featureMVOption.BooleanValue)

	assert.Equal(t, float64(60), featureMVOption.DefaultPercentageAllocation)
	assert.Equal(t, FeatureID, *featureMVOption.FeatureID)
	assert.Equal(t, FeatureUUID, featureMVOption.FeatureUUID)

	assert.Equal(t, ProjectID, *featureMVOption.ProjectID)

}

func TestCreateFeatureMVOption(t *testing.T) {
	// Given
	featureID := FeatureID
	projectID := ProjectID
	stringValue := "option_value_30"
	defaultPercentageAllocation := float64(60)
	featureMVOption := flagsmithapi.FeatureMultivariateOption{
		Type:                        "unicode",
		FeatureID:                   &featureID,
		StringValue:                 &stringValue,
		DefaultPercentageAllocation: defaultPercentageAllocation,
		FeatureUUID:                 FeatureUUID,
		ProjectID:                   &projectID,
	}

	expectedRequestBody := fmt.Sprintf(`{"type":"unicode","feature":%d,"string_value":"%s","default_percentage_allocation":%.0f}`, featureID, stringValue, defaultPercentageAllocation)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/projects/%d/features/%d/mv-options/", ProjectID, FeatureID), req.URL.Path)
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		// Test that we sent the correct body
		rawBody, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequestBody, string(rawBody))

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, GetMVFeatureOptionResponseJson)
		assert.NoError(t, err)

	}))

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	err := client.CreateFeatureMVOption(&featureMVOption)

	// Then
	var nilIntPointer *int64
	var nilBoolPointer *bool

	assert.NoError(t, err)
	assert.Equal(t, MVFeatureOptionUUID, featureMVOption.UUID)

	assert.Equal(t, "unicode", featureMVOption.Type)
	assert.Equal(t, stringValue, *featureMVOption.StringValue)
	assert.Equal(t, nilIntPointer, featureMVOption.IntegerValue)
	assert.Equal(t, nilBoolPointer, featureMVOption.BooleanValue)

	assert.Equal(t, float64(60), featureMVOption.DefaultPercentageAllocation)
	assert.Equal(t, FeatureID, *featureMVOption.FeatureID)
	assert.Equal(t, FeatureUUID, featureMVOption.FeatureUUID)

	assert.Equal(t, ProjectID, *featureMVOption.ProjectID)

}

func TestCreateFeatureMVOptionWithFeatureIDNotSet(t *testing.T) {
	// Given
	stringValue := "option_value_30"
	defaultPercentageAllocation := float64(60)
	featureMVOption := flagsmithapi.FeatureMultivariateOption{
		Type:                        "unicode",
		StringValue:                 &stringValue,
		DefaultPercentageAllocation: defaultPercentageAllocation,
		FeatureUUID:                 FeatureUUID,
	}

	expectedRequestBody := fmt.Sprintf(`{"type":"unicode","feature":%d,"string_value":"%s","default_percentage_allocation":%.0f}`, FeatureID, stringValue, defaultPercentageAllocation)

	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/%d/features/%d/mv-options/", ProjectID, FeatureID), func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		// Test that we sent the correct body
		rawBody, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequestBody, string(rawBody))

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, GetMVFeatureOptionResponseJson)
		assert.NoError(t, err)

	})

	mux.HandleFunc(fmt.Sprintf("/api/v1/projects/%d/", ProjectID), func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetProjectResponseJson)
		assert.NoError(t, err)
	})

	mux.HandleFunc(fmt.Sprintf("/api/v1/features/get-by-uuid/%s/", FeatureUUID), func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, CreateFeatureResponseJson)
		assert.NoError(t, err)

	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	err := client.CreateFeatureMVOption(&featureMVOption)

	// Then
	var nilIntPointer *int64
	var nilBoolPointer *bool

	assert.NoError(t, err)
	assert.Equal(t, MVFeatureOptionUUID, featureMVOption.UUID)

	assert.Equal(t, "unicode", featureMVOption.Type)
	assert.Equal(t, stringValue, *featureMVOption.StringValue)
	assert.Equal(t, nilIntPointer, featureMVOption.IntegerValue)
	assert.Equal(t, nilBoolPointer, featureMVOption.BooleanValue)

	assert.Equal(t, float64(60), featureMVOption.DefaultPercentageAllocation)
	assert.Equal(t, FeatureID, *featureMVOption.FeatureID)
	assert.Equal(t, FeatureUUID, featureMVOption.FeatureUUID)

	assert.Equal(t, ProjectID, *featureMVOption.ProjectID)

}
