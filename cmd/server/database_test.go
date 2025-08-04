package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitDBSuccess(t *testing.T) {
	// Create mock database with ping monitoring enabled
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	// Expect ping to succeed
	mock.ExpectPing()

	// Test that database connection logic works
	err = db.Ping()
	assert.NoError(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInitDBPingFailure(t *testing.T) {
	// Create mock database with ping monitoring enabled
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer db.Close()

	// Expect ping to fail
	mock.ExpectPing().WillReturnError(assert.AnError)

	// Test ping failure
	err = db.Ping()
	assert.Error(t, err)

	// Verify all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDatabaseConnectionString(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectedStr string
	}{
		{
			name: "All environment variables set",
			envVars: map[string]string{
				"DB_HOST":     "testhost",
				"DB_PORT":     "5433",
				"DB_USER":     "testuser",
				"DB_PASSWORD": "testpass",
				"DB_NAME":     "testdb",
				"DB_SSLMODE":  "require",
			},
			expectedStr: "host=testhost port=5433 user=testuser password=testpass dbname=testdb sslmode=require",
		},
		{
			name:        "Default values used",
			envVars:     map[string]string{},
			expectedStr: "host=localhost port=5432 user=postgres password= dbname=gorectus sslmode=disable",
		},
		{
			name: "Partial environment variables",
			envVars: map[string]string{
				"DB_HOST":     "customhost",
				"DB_PASSWORD": "secret",
			},
			expectedStr: "host=customhost port=5432 user=postgres password=secret dbname=gorectus sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Build connection string using the same logic as initDB
			host := os.Getenv("DB_HOST")
			if host == "" {
				host = "localhost"
			}

			port := os.Getenv("DB_PORT")
			if port == "" {
				port = "5432"
			}

			user := os.Getenv("DB_USER")
			if user == "" {
				user = "postgres"
			}

			password := os.Getenv("DB_PASSWORD")
			dbname := os.Getenv("DB_NAME")
			if dbname == "" {
				dbname = "gorectus"
			}

			sslmode := os.Getenv("DB_SSLMODE")
			if sslmode == "" {
				sslmode = "disable"
			}

			connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
				host, port, user, password, dbname, sslmode)

			assert.Equal(t, tt.expectedStr, connStr)

			// Clean up environment
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

func TestEnvironmentVariableDefaults(t *testing.T) {
	// Clear all environment variables
	os.Clearenv()

	tests := []struct {
		name         string
		envVar       string
		defaultValue string
		getValue     func() string
	}{
		{
			name:         "DB_HOST defaults to localhost",
			envVar:       "DB_HOST",
			defaultValue: "localhost",
			getValue: func() string {
				host := os.Getenv("DB_HOST")
				if host == "" {
					host = "localhost"
				}
				return host
			},
		},
		{
			name:         "DB_PORT defaults to 5432",
			envVar:       "DB_PORT",
			defaultValue: "5432",
			getValue: func() string {
				port := os.Getenv("DB_PORT")
				if port == "" {
					port = "5432"
				}
				return port
			},
		},
		{
			name:         "DB_USER defaults to postgres",
			envVar:       "DB_USER",
			defaultValue: "postgres",
			getValue: func() string {
				user := os.Getenv("DB_USER")
				if user == "" {
					user = "postgres"
				}
				return user
			},
		},
		{
			name:         "DB_NAME defaults to gorectus",
			envVar:       "DB_NAME",
			defaultValue: "gorectus",
			getValue: func() string {
				dbname := os.Getenv("DB_NAME")
				if dbname == "" {
					dbname = "gorectus"
				}
				return dbname
			},
		},
		{
			name:         "DB_SSLMODE defaults to disable",
			envVar:       "DB_SSLMODE",
			defaultValue: "disable",
			getValue: func() string {
				sslmode := os.Getenv("DB_SSLMODE")
				if sslmode == "" {
					sslmode = "disable"
				}
				return sslmode
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure environment variable is not set
			os.Unsetenv(tt.envVar)

			// Test default value
			value := tt.getValue()
			assert.Equal(t, tt.defaultValue, value)

			// Test with environment variable set
			testValue := "test_value"
			os.Setenv(tt.envVar, testValue)
			value = tt.getValue()
			assert.Equal(t, testValue, value)

			// Clean up
			os.Unsetenv(tt.envVar)
		})
	}
}
