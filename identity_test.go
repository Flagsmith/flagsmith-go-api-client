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

const GetTraitsJsonResponse = `
{
  "results": [
    {
      "id": 1,
      "trait_key": "trait_key_1",
      "value_type": "unicode",
      "integer_value": null,
      "string_value": "value1",
      "boolean_value": null,
      "float_value": null,
      "created_date": "2022-02-04T03:34:31.329637Z"
    },
    {
      "id": 2,
      "trait_key": "trait_key_2",
      "value_type": "unicode",
      "integer_value": null,
      "string_value": "value2",
      "boolean_value": null,
      "float_value": null,
      "created_date": "2022-09-19T08:41:26.542560Z"
    }
  ]
}`

func TestGetTraits(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/environments/%s/identities/%d/traits/", EnvironmentAPIKey, IdentityID), req.URL.Path)
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		rw.Header().Set("Content-Type", "application/json")
		_, err := rw.Write([]byte(GetTraitsJsonResponse))
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	traits, err := client.GetTraits(EnvironmentAPIKey, IdentityID)

	// Then
	assert.NoError(t, err)
	assert.Len(t, traits, 2)

	// Check the first trait
	assert.Equal(t, int64(1), traits[0].ID)
	assert.Equal(t, "trait_key_1", traits[0].TraitKey)
	assert.Equal(t, "unicode", traits[0].ValueType)
	assert.Equal(t, "value1", *traits[0].StringValue)

	// Check the second trait
	assert.Equal(t, int64(2), traits[1].ID)
	assert.Equal(t, "trait_key_2", traits[1].TraitKey)
	assert.Equal(t, "unicode", traits[1].ValueType)
	assert.Equal(t, "value2", *traits[1].StringValue)
}

const TraitResponseJson = `
{
    "id": 1,
    "trait_key": "key1",
    "value_type": "unicode",
    "integer_value": null,
    "string_value": "value1",
    "boolean_value": null,
    "float_value": null,
    "created_date": "2024-04-17T07:09:04.385701Z"
}
`

func TestCreateTrait(t *testing.T) {
	// Given
	trait_value := "value1"
	trait := &flagsmithapi.Trait{
		TraitKey:    "key1",
		ValueType:   "unicode",
		StringValue: &trait_value,
	}
	expectedRequestBody := `{"trait_key":"key1","value_type":"unicode","string_value":"value1"}`

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/environments/%s/identities/%d/traits/", EnvironmentAPIKey, IdentityID), req.URL.Path)
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		// Test that we sent the correct body
		rawBody, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequestBody, string(rawBody))

		rw.Header().Set("Content-Type", "application/json")
		_, err = rw.Write([]byte(TraitResponseJson))
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	err := client.CreateTrait(EnvironmentAPIKey, IdentityID, trait)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, int64(1), trait.ID)
}

func TestUpdateTrait(t *testing.T) {
	// Given
	trait_value := "updated_value"
	trait := &flagsmithapi.Trait{
		ID:          1,
		TraitKey:    "key1",
		ValueType:   "unicode",
		StringValue: &trait_value,
	}
	expectedRequestBody := `{"id":1,"trait_key":"key1","value_type":"unicode","string_value":"updated_value"}`

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/environments/%s/identities/%d/traits/1/", EnvironmentAPIKey, IdentityID), req.URL.Path)
		assert.Equal(t, "PUT", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		// Test that we sent the correct body
		rawBody, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, expectedRequestBody, string(rawBody))

		rw.Header().Set("Content-Type", "application/json")
		_, err = rw.Write([]byte(TraitResponseJson))
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	err := client.UpdateTrait(EnvironmentAPIKey, IdentityID, trait)

	// Then
	assert.NoError(t, err)
	assert.Equal(t, "updated_value", *trait.StringValue)
}

func TestDeleteTrait(t *testing.T) {
	// Given
	requestReceived := struct {
		mu                sync.Mutex
		isRequestReceived bool
	}{}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestReceived.mu.Lock()
		requestReceived.isRequestReceived = true
		requestReceived.mu.Unlock()

		assert.Equal(t, fmt.Sprintf("/api/v1/environments/%s/identities/%d/traits/1/", EnvironmentAPIKey, IdentityID), req.URL.Path)
		assert.Equal(t, "DELETE", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))
	}))

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	err := client.DeleteTrait(EnvironmentAPIKey, IdentityID, 1)

	// Then
	requestReceived.mu.Lock()
	assert.True(t, requestReceived.isRequestReceived)
	requestReceived.mu.Unlock()
	assert.NoError(t, err)
}
