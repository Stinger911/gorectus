package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetMe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Successful get me", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		s := &Server{
			db: db,
		}

		// Mock user ID and role in context
		userID := "user-123"

		// Mock DB query
		rows := sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "avatar", "language", "theme", "status", "role_id", "role_name", "last_access", "last_page", "provider", "external_identifier", "email_notifications", "tags", "created_at", "updated_at"}).
			AddRow(userID, "test@example.com", "Test", "User", nil, "en-US", "dark", "active", "role-123", "User", nil, nil, "local", nil, true, nil, time.Now(), time.Now())
		mock.ExpectQuery(`SELECT u.id, u.email, u.first_name, u.last_name, u.avatar, u.language, u.theme, u.status, u.role_id, r.name as role_name, u.last_access, u.last_page, u.provider, u.external_identifier, u.email_notifications, u.tags, u.created_at, u.updated_at FROM users u JOIN roles r ON u.role_id = r.id WHERE u.id = \$1`).
			WithArgs(userID).
			WillReturnRows(rows)

		// Create a test context and set user ID
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", userID)

		usersHandler := NewUsersHandler(s)
		usersHandler.getMe(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"id":"user-123"`)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("User not found", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		s := &Server{
			db: db,
		}

		userID := "non-existent-user"

		mock.ExpectQuery(`SELECT u.id, u.email, u.first_name, u.last_name, u.avatar, u.language, u.theme, u.status, u.role_id, r.name as role_name, u.last_access, u.last_page, u.provider, u.external_identifier, u.email_notifications, u.tags, u.created_at, u.updated_at FROM users u JOIN roles r ON u.role_id = r.id WHERE u.id = \$1`).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", userID)

		usersHandler := NewUsersHandler(s)
		usersHandler.getMe(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUpdateMe(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Successful update me", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		defer db.Close()

		s := &Server{
			db: db,
		}

		userID := "user-123"
		newFirstName := "Updated"
		updateReq := UpdateUserRequest{
			FirstName: &newFirstName,
		}
		reqBody, _ := json.Marshal(updateReq)

		// Mock update query
		mock.ExpectExec("UPDATE users SET first_name = \\$1, updated_at = CURRENT_TIMESTAMP WHERE id = \\$2").
			WithArgs(newFirstName, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// Mock select query to fetch updated user
		rows := sqlmock.NewRows([]string{"id", "email", "first_name", "last_name", "avatar", "language", "theme", "status", "role_id", "role_name", "last_access", "last_page", "provider", "external_identifier", "email_notifications", "tags", "created_at", "updated_at"}).
			AddRow(userID, "test@example.com", newFirstName, "User", nil, "en-US", "dark", "active", "role-123", "User", nil, nil, "local", nil, true, nil, time.Now(), time.Now())
		mock.ExpectQuery(`SELECT u.id, u.email, u.first_name, u.last_name, u.avatar, u.language, u.theme, u.status, u.role_id, r.name as role_name, u.last_access, u.last_page, u.provider, u.external_identifier, u.email_notifications, u.tags, u.created_at, u.updated_at FROM users u JOIN roles r ON u.role_id = r.id WHERE u.id = \$1`).
			WithArgs(userID).
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPatch, "/me", bytes.NewBuffer(reqBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("user_id", userID)

		usersHandler := NewUsersHandler(s)
		usersHandler.updateMe(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"first_name":"Updated"`)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
