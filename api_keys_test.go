package flagsmithapi_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	flagsmithapi "github.com/Flagsmith/flagsmith-go-api-client"
)

const GetAPIKeysResponseJson = `
[
  {
    "id": 1,
    "key": "ser.UiYoRr6zUjiFBUXaRwo7b5",
    "active": true,
    "created_at": "2022-02-16T12:09:30.349955Z",
    "name": "key1",
    "expires_at": null
  },
  {
    "id": 2,
    "key": "ser.g5N5Q4L8E832cA3iU4u4td",
    "active": true,
    "created_at": "2022-02-16T12:09:21.300028Z",
    "name": "key2",
    "expires_at": null
  }
]
`
const KeyOneID int64 = 1
const KeyTwoID int64 = 2
const KeyOneName = "key1"
const KeyTwoName = "key2"
const KeyOneKey = "ser.UiYoRr6zUjiFBUXaRwo7b5"
const KeyTwoKey = "ser.g5N5Q4L8E832cA3iU4u4td"

const CreateAPIKeyResponseJson = `
{
  "id": 1,
  "key": "ser.UiYoRr6zUjiFBUXaRwo7b5",
  "active": true,
  "created_at": "2024-04-16T07:53:50.808415Z",
  "name": "key1",
  "expires_at": null
}
`

func TestGetServerSideEnvKeys(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/environments/%s/api-keys/", EnvironmentAPIKey), req.URL.Path)
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetAPIKeysResponseJson)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	keys, err := client.GetServerSideEnvKeys(EnvironmentAPIKey)

	// Then
	assert.NoError(t, err)
	assert.Len(t, *keys, 2)
	// Check the first key

	assert.Equal(t, KeyOneID, (*keys)[0].ID)
	assert.Equal(t, KeyOneName, (*keys)[0].Name)
	assert.Equal(t, KeyOneKey, (*keys)[0].Key)
	assert.True(t, (*keys)[0].Active)
	assert.Equal(t, "2022-02-16T12:09:30.349955Z", (*keys)[0].CreatedAt.Format(time.RFC3339Nano))
	assert.Nil(t, (*keys)[0].ExpiresAt)

	// Check the second key
	assert.Equal(t, KeyTwoID, (*keys)[1].ID)
	assert.Equal(t, KeyTwoName, (*keys)[1].Name)
	assert.Equal(t, KeyTwoKey, (*keys)[1].Key)
	assert.True(t, (*keys)[1].Active)
	assert.Equal(t, "2022-02-16T12:09:21.300028Z", (*keys)[1].CreatedAt.Format(time.RFC3339Nano))
	assert.Nil(t, (*keys)[1].ExpiresAt)

}

func TestCreateServerSideEnvKey(t *testing.T) {
	// Given
	expectedRequestBody := `{"active":true,"name":"` + KeyOneName + `"}`

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/environments/%s/api-keys/", EnvironmentAPIKey), req.URL.Path)
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		// Test that we sent the correct body
		rawBody, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequestBody, string(rawBody))

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, CreateAPIKeyResponseJson)
		assert.NoError(t, err)
	}))

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	key := flagsmithapi.ServerSideEnvKey{
		Active: true,
		Name:   KeyOneName,
	}
	err := client.CreateServerSideEnvKey(EnvironmentAPIKey, &key)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, int64(KeyOneID), key.ID)
	assert.Equal(t, true, key.Active)
	assert.Equal(t, KeyOneName, key.Name)
}

func TestUpdateServerSideEnvKey(t *testing.T) {
	// Given
	expectedRequestBody := fmt.Sprintf(`{"id":%d,"active":false,"name":"%s"}`, KeyOneID, KeyOneName)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/environments/%s/api-keys/%d/", EnvironmentAPIKey, KeyOneID), req.URL.Path)
		assert.Equal(t, "PUT", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		// Test that we sent the correct body
		rawBody, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequestBody, string(rawBody))

		rw.Header().Set("Content-Type", "application/json")
		_, err = io.WriteString(rw, fmt.Sprintf(`{"id":%d,"active":false,"name":"%s"}`, KeyOneID, KeyOneName))
		assert.NoError(t, err)
	}))

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	key := flagsmithapi.ServerSideEnvKey{
		ID:     KeyOneID,
		Active: false,
		Name:   KeyOneName,
	}
	err := client.UpdateServerSideEnvKey(EnvironmentAPIKey, &key)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, int64(KeyOneID), key.ID)
	assert.Equal(t, false, key.Active)
	assert.Equal(t, KeyOneName, key.Name)
}

func TestDeleteServerSideEnvKey(t *testing.T) {
	// Given
	requestReceived := struct {
		mu                sync.Mutex
		isRequestReceived bool
	}{}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestReceived.mu.Lock()
		requestReceived.isRequestReceived = true
		requestReceived.mu.Unlock()

		assert.Equal(t, fmt.Sprintf("/api/v1/environments/%s/api-keys/%d/", EnvironmentAPIKey, KeyOneID), req.URL.Path)
		assert.Equal(t, "DELETE", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))
	}))

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	err := client.DeleteServerSideEnvKey(EnvironmentAPIKey, KeyOneID)

	// Then
	requestReceived.mu.Lock()
	assert.True(t, requestReceived.isRequestReceived)
	requestReceived.mu.Unlock()
	assert.NoError(t, err)
}
