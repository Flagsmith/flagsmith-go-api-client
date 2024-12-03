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

func TestCreateProjectByUUID(t *testing.T) {
	// Given
	project := flagsmithapi.Project{
		Name:         ProjectName,
		Organisation: OrganisationID,
	}
	expectedRequestBody := fmt.Sprintf(`{"name":"%s","organisation":%d}`, ProjectName, OrganisationID)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/api/v1/projects/", req.URL.Path)
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		// Test that we sent the correct body
		rawBody, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequestBody, string(rawBody))

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, GetProjectResponseJson)
		assert.NoError(t, err)

	}))

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	err := client.CreateProject(&project)

	// Then
	// assert that we did not receive an error
	assert.NoError(t, err)

	assert.Equal(t, ProjectID, project.ID)
	assert.Equal(t, ProjectUUID, project.UUID)
	assert.Equal(t, "project-1", project.Name)

}
func TestUpdateProject(t *testing.T) {
	// Given
	project := flagsmithapi.Project{
		ID:           ProjectID,
		Name:         ProjectName,
		Organisation: OrganisationID,
	}
	expectedRequestBody := fmt.Sprintf(`{"id":%d,"name":"%s","organisation":%d}`, ProjectID, ProjectName, OrganisationID)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/projects/%d/", ProjectID), req.URL.Path)
		assert.Equal(t, "PUT", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		// Test that we sent the correct body
		rawBody, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequestBody, string(rawBody))

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, GetProjectResponseJson)
		assert.NoError(t, err)

	}))

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	err := client.UpdateProject(&project)

	// Then
	// assert that we did not receive an error
	assert.NoError(t, err)

	assert.Equal(t, ProjectID, project.ID)
	assert.Equal(t, ProjectUUID, project.UUID)
	assert.Equal(t, "project-1", project.Name)

}
func TestGetProjectByID(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/projects/%d/", ProjectID), req.URL.Path)
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetProjectResponseJson)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	project, err := client.GetProjectByID(ProjectID)

	// Then
	assert.NoError(t, err)

	assert.Equal(t, ProjectID, project.ID)
	assert.Equal(t, ProjectUUID, project.UUID)
	assert.Equal(t, "project-1", project.Name)

}
func TestDeleteProject(t *testing.T) {
	// Given
	requestReceived := struct {
		mu                sync.Mutex
		isRequestReceived bool
	}{}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestReceived.mu.Lock()
		requestReceived.isRequestReceived = true
		requestReceived.mu.Unlock()

		assert.Equal(t, fmt.Sprintf("/api/v1/projects/%d/", ProjectID), req.URL.Path)
		assert.Equal(t, "DELETE", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

	}))

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	err := client.DeleteProject(ProjectID)

	// Then
	requestReceived.mu.Lock()
	assert.True(t, requestReceived.isRequestReceived)
	assert.NoError(t, err)

}
