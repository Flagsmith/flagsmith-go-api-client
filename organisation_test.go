package flagsmithapi_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"testing"

	"github.com/stretchr/testify/assert"

	flagsmithapi "github.com/Flagsmith/flagsmith-go-api-client"
)

const GetOrganisationUsersResponseJson = `
[
    {
        "id": 100,
        "email": "john@example.com",
        "first_name": "John",
        "last_name": "Doe",
        "last_login": "2026-03-01T10:00:00Z",
        "uuid": "abc-123",
        "role": "ADMIN",
        "date_joined": "2025-01-01T00:00:00Z"
    },
    {
        "id": 200,
        "email": "jane@example.com",
        "first_name": "Jane",
        "last_name": "Smith",
        "last_login": "2026-03-15T10:00:00Z",
        "uuid": "def-456",
        "role": "USER",
        "date_joined": "2025-06-01T00:00:00Z"
    }
]
`

func TestGetOrganisationUsers(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, fmt.Sprintf("/api/v1/organisations/%d/users/", OrganisationID), req.URL.Path)
		assert.Equal(t, "GET", req.Method)
		assert.Equal(t, "Api-Key "+MasterAPIKey, req.Header.Get("Authorization"))

		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetOrganisationUsersResponseJson)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	users, err := client.GetOrganisationUsers(OrganisationID)

	// Then
	assert.NoError(t, err)
	assert.Len(t, users, 2)

	assert.Equal(t, int64(100), users[0].ID)
	assert.Equal(t, "john@example.com", users[0].Email)
	assert.Equal(t, "John", users[0].FirstName)
	assert.Equal(t, "Doe", users[0].LastName)
	assert.Equal(t, "ADMIN", users[0].Role)

	assert.Equal(t, int64(200), users[1].ID)
	assert.Equal(t, "jane@example.com", users[1].Email)
}

func TestGetOrganisationUserByEmail(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetOrganisationUsersResponseJson)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	user, err := client.GetOrganisationUserByEmail(OrganisationID, "jane@example.com")

	// Then
	assert.NoError(t, err)
	assert.Equal(t, int64(200), user.ID)
	assert.Equal(t, "jane@example.com", user.Email)
	assert.Equal(t, "Jane", user.FirstName)
	assert.Equal(t, "Smith", user.LastName)
	assert.Equal(t, "USER", user.Role)
}

func TestGetOrganisationUserByEmailNotFound(t *testing.T) {
	// Given
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(rw, GetOrganisationUsersResponseJson)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client := flagsmithapi.NewClient(MasterAPIKey, server.URL+"/api/v1")

	// When
	user, err := client.GetOrganisationUserByEmail(OrganisationID, "notfound@example.com")

	// Then
	assert.Nil(t, user)
	assert.Error(t, err)
	assert.IsType(t, flagsmithapi.UserNotFoundError{}, err)
}
