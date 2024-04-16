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

const IdentityIdentifier = "test-user"
const IdentityID int64 = 1
const IdentityResponseJson = `
{
    "id": 1,
    "identifier": "test-user"
}
`

func TestGetIdentity(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/environments/%s/identities/%d/", EnvironmentAPIKey, IdentityID), req.URL.Path)
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, IdentityResponseJson)
		assert.NoError(t, err)
	}))

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	identity, err := client.GetIdentity(EnvironmentAPIKey, IdentityID)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, IdentityID, *identity.ID)
	assert.Equal(t, IdentityIdentifier, identity.Identifier)
}

func TestCreateIdentity(t *testing.T) {
	// Given
	expectedRequestBody := fmt.Sprintf(`{"identifier":"%s"}`, IdentityIdentifier)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/environments/%s/identities/", EnvironmentAPIKey), req.URL.Path)
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		// Test that we sent the correct body
		rawBody, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequestBody, string(rawBody))

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, IdentityResponseJson)
		assert.NoError(t, err)
	}))

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	identity := &flagsmithapi.Identity{
		Identifier: IdentityIdentifier,
	}
	err := client.CreateIdentity(EnvironmentAPIKey, identity)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, IdentityID, *identity.ID)
	assert.Equal(t, IdentityIdentifier, identity.Identifier)
}
func TestDeleteIdentity(t *testing.T) {
	// Given
	requestReceived := struct {
		mu                sync.Mutex
		isRequestReceived bool
	}{}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestReceived.mu.Lock()
		requestReceived.isRequestReceived = true
		requestReceived.mu.Unlock()

		assert.Equal(t, fmt.Sprintf("/api/v1/environments/%s/identities/%d/", EnvironmentAPIKey, IdentityID), req.URL.Path)
		assert.Equal(t, "DELETE", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

	}))

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	err := client.DeleteIdentity(EnvironmentAPIKey, IdentityID)

	// Then
	requestReceived.mu.Lock()
	assert.True(t, requestReceived.isRequestReceived)
	requestReceived.mu.Unlock()
	assert.NoError(t, err)
}
