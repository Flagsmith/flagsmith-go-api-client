package flagsmithapi_test

import (
	"fmt"
	"io"
	"testing"

	"io/ioutil"
	"net/http"
	"net/http/httptest"

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
		rawBody, err := ioutil.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequestBody, string(rawBody))

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, UpdateFeatureStateResponseJson)
		assert.NoError(t, err)
	}))
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

func TestGetProject(t *testing.T) {
	// Given
	masterAPIKey := "master_api_key"
	projectUUID := "10421b1f-5f29-4da9-abe2-30f88c07c9e8"
	getProjectResponseJson := `
[
    {
        "id": 1,
        "uuid": "10421b1f-5f29-4da9-abe2-30f88c07c9e8",
        "name": "project-1",
        "organisation": 1,
        "hide_disabled_flags": false,
        "enable_dynamo_db": true,
        "migration_status": "NOT_APPLICABLE",
        "use_edge_identities": false
    }
]
`
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/v1/projects/", req.URL.Path)
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "Api-Key "+masterAPIKey, req.Header.Get("Authorization"))

		query := req.URL.Query()
		assert.Equal(t, projectUUID, query.Get("uuid"))

		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, getProjectResponseJson)
		assert.NoError(t, err)
	}))
	client := flagsmithapi.NewClient(masterAPIKey, server.URL+"/api/v1")
	// When
	project, err := client.GetProject(projectUUID)

	// Then
	assert.NoError(t, err)

	assert.Equal(t, int64(1), project.ID)
	assert.Equal(t, projectUUID, project.UUID)
	assert.Equal(t, "project-1", project.Name)
}
