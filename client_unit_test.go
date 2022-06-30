package flagsmithapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClientUsesDefaultBaseURLIfNotProvided(t *testing.T) {
	// Given
	masterAPIKey := "test_key"
	// When
	client := NewClient(masterAPIKey, "")
	// Then
	assert.Equal(t, "https://api.flagsmith.com/api/v1", client.baseURL)
}
