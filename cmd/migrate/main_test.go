package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestPasswordHashGeneration(t *testing.T) {
	password := "admin123"

	// Generate hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	// Verify hash
	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	assert.NoError(t, err)

	// Verify wrong password fails
	err = bcrypt.CompareHashAndPassword(hash, []byte("wrongpassword"))
	assert.Error(t, err)
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		expected     string
	}{
		{
			name:         "Returns default when env var not set",
			key:          "NON_EXISTENT_VAR",
			defaultValue: "default_value",
			expected:     "default_value",
		},
		{
			name:         "DB_HOST default",
			key:          "DB_HOST_TEST",
			defaultValue: "localhost",
			expected:     "localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getEnvOrDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMigrationPathConstants(t *testing.T) {
	assert.Equal(t, "file://migrations", migrationsPath)
}
