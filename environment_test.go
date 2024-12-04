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

const EnvironmentName = "Development"
const EnvironmentUUID = "4c830509-116d-46b7-804e-98f74d3b000b"

func TestGetEnvironment(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/environments/%s/", EnvironmentAPIKey), req.URL.Path)
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, EnvironmentJson)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	environment, err := client.GetEnvironment(EnvironmentAPIKey)

	// Then
	// assert that we did not receive an error
	assert.NoError(t, err)

	// assert that the environment is as expected
	assert.Equal(t, EnvironmentID, environment.ID)
	assert.Equal(t, "Development", environment.Name)
	assert.Equal(t, EnvironmentAPIKey, environment.APIKey)
	assert.Equal(t, ProjectID, environment.ProjectID)
}
func TestGetEnvironmentByUUID(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/environments/get-by-uuid/%s/", EnvironmentUUID), req.URL.Path)
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, EnvironmentJson)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	environment, err := client.GetEnvironmentByUUID(EnvironmentUUID)

	// Then
	// assert that we did not receive an error
	assert.NoError(t, err)

	// assert that the environment is as expected
	assert.Equal(t, EnvironmentID, environment.ID)
	assert.Equal(t, "Development", environment.Name)
	assert.Equal(t, EnvironmentAPIKey, environment.APIKey)
	assert.Equal(t, ProjectID, environment.ProjectID)
}
func TestCreateEnvironment(t *testing.T) {
	// Given
	expectedRequestBody := fmt.Sprintf(`{
		"name": "%s",
		"description": "This is a test environment",
		"project": %d
	}`, EnvironmentName, ProjectID)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/v1/environments/", req.URL.Path)
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		// Test that we sent the correct body
		rawBody, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, expectedRequestBody, string(rawBody))

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, EnvironmentJson)
		assert.NoError(t, err)
	}))

	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	environment := &flagsmithapi.Environment{
		Name:        EnvironmentName,
		Description: "This is a test environment",
		ProjectID:   ProjectID,
	}
	err := client.CreateEnvironment(environment)

	// Then
	// assert that we did not receive an error
	assert.NoError(t, err)

	// assert that the environment is as expected
	assert.Equal(t, EnvironmentID, environment.ID)
	assert.Equal(t, EnvironmentName, environment.Name)
	assert.Equal(t, "environment_api_key", environment.APIKey)
}

func TestUpdateEnvironment(t *testing.T) {
	// Given
	updatedDescription := "Updated environment description"
	expectedRequestBody := fmt.Sprintf(`{
		"id": %d,
		"name": "%s",
		"description": "%s",
		"project": %d,
		"api_key": "%s"
	}`, EnvironmentID, EnvironmentName, updatedDescription, ProjectID, EnvironmentAPIKey)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/environments/%s/", EnvironmentAPIKey), req.URL.Path)
		assert.Equal(t, "PUT", req.Method)
		assert.Equal(t, "Api-Key "+EnvironmentAPIKey, req.Header.Get("Authorization"))

		// Test that we sent the correct body
		rawBody, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.JSONEq(t, expectedRequestBody, string(rawBody))

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, fmt.Sprintf(`{"id": 100, "name": "%s", "description": "%s", "api_key": "%s"}`, EnvironmentName, updatedDescription, EnvironmentAPIKey))
		assert.NoError(t, err)
	}))

	defer server.Close()

	client := flagsmithapi.NewClient(EnvironmentAPIKey, server.URL+"/api/v1")

	// When
	environment := &flagsmithapi.Environment{
		ID:          EnvironmentID,
		Name:        EnvironmentName,
		Description: "Updated environment description",
		ProjectID:   ProjectID,
		APIKey:      EnvironmentAPIKey,
	}
	err := client.UpdateEnvironment(environment)

	// Then
	// assert that we did not receive an error
	assert.NoError(t, err)

	// assert that the environment is as expected
	assert.Equal(t, EnvironmentID, environment.ID)
	assert.Equal(t, EnvironmentName, environment.Name)
	assert.Equal(t, EnvironmentAPIKey, environment.APIKey)
}

func TestDeleteEnvironment(t *testing.T) {
	// Given
	requestReceived := struct {
		mu                sync.Mutex
		isRequestReceived bool
	}{}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestReceived.mu.Lock()
		requestReceived.isRequestReceived = true
		requestReceived.mu.Unlock()

		assert.Equal(t, fmt.Sprintf("/api/v1/environments/%s/", EnvironmentAPIKey), req.URL.Path)
		assert.Equal(t, "DELETE", req.Method)
		assert.Equal(t, "Api-Key "+EnvironmentAPIKey, req.Header.Get("Authorization"))
	}))

	client := flagsmithapi.NewClient(EnvironmentAPIKey, server.URL+"/api/v1")

	// When
	err := client.DeleteEnvironment(EnvironmentAPIKey)

	// Then
	requestReceived.mu.Lock()
	assert.True(t, requestReceived.isRequestReceived)
	assert.NoError(t, err)
}
