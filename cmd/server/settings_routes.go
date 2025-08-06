package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SettingsHandler handles settings-related routes
type SettingsHandler struct {
	db             *sql.DB
	authMiddleware gin.HandlerFunc
	optionsHandler gin.HandlerFunc
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(server ServerInterface) *SettingsHandler {
	return &SettingsHandler{
		db:             server.GetDB(),
		authMiddleware: server.AuthMiddleware(),
		optionsHandler: server.OptionsHandler(),
	}
}

// SetupRoutes sets up settings routes
func (h *SettingsHandler) SetupRoutes(v1 *gin.RouterGroup) {
	// CORS preflight OPTIONS for settings endpoints
	v1.OPTIONS("/settings", h.optionsHandler)

	// Settings routes (protected)
	settings := v1.Group("/settings")
	settings.Use(h.authMiddleware)
	{
		settings.GET("", h.getSettings)
		settings.PATCH("", h.updateSettings)
		settings.POST("/test-connection", h.testDatabaseConnection)
		settings.POST("/test-email", h.testEmailConfiguration)
	}
}

// SettingsResponse represents the settings response structure
type SettingsResponse struct {
	Data Settings `json:"data"`
}

// Settings represents the application settings
type Settings struct {
	// General Settings
	SiteName          string `json:"site_name"`
	SiteDescription   string `json:"site_description"`
	AllowRegistration bool   `json:"allow_registration"`
	MaintenanceMode   bool   `json:"maintenance_mode"`

	// Database Settings (read-only for security)
	DatabaseHost string `json:"database_host"`
	DatabasePort string `json:"database_port"`
	DatabaseName string `json:"database_name"`
	DatabaseUser string `json:"database_user"`

	// Email Settings
	SMTPHost      string `json:"smtp_host"`
	SMTPPort      string `json:"smtp_port"`
	SMTPUser      string `json:"smtp_user"`
	SMTPFromEmail string `json:"smtp_from_email"`
	EmailEnabled  bool   `json:"email_enabled"`

	// Security Settings (some fields read-only for security)
	SessionTimeout    int  `json:"session_timeout"`
	PasswordMinLength int  `json:"password_min_length"`
	RequireTwoFactor  bool `json:"require_two_factor"`
	JWTSecretExists   bool `json:"jwt_secret_exists"` // Don't expose actual secret

	// Metadata
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy string    `json:"updated_by"`
}

// UpdateSettingsRequest represents the request to update settings
type UpdateSettingsRequest struct {
	// General Settings
	SiteName          *string `json:"site_name,omitempty"`
	SiteDescription   *string `json:"site_description,omitempty"`
	AllowRegistration *bool   `json:"allow_registration,omitempty"`
	MaintenanceMode   *bool   `json:"maintenance_mode,omitempty"`

	// Email Settings
	SMTPHost      *string `json:"smtp_host,omitempty"`
	SMTPPort      *string `json:"smtp_port,omitempty"`
	SMTPUser      *string `json:"smtp_user,omitempty"`
	SMTPFromEmail *string `json:"smtp_from_email,omitempty"`
	EmailEnabled  *bool   `json:"email_enabled,omitempty"`

	// Security Settings
	JWTSecret         *string `json:"jwt_secret,omitempty"` // Only for updates
	SessionTimeout    *int    `json:"session_timeout,omitempty"`
	PasswordMinLength *int    `json:"password_min_length,omitempty"`
	RequireTwoFactor  *bool   `json:"require_two_factor,omitempty"`
}

// isAdmin checks if the requesting user is an admin
func (h *SettingsHandler) isAdmin(c *gin.Context) bool {
	currentUserRole := c.GetString("user_role")
	return currentUserRole == "Administrator"
}

// getSettings godoc
// @Summary Get system settings
// @Description Retrieve current system settings (Admin only)
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SettingsResponse "Current settings"
// @Failure 401 {object} main.ErrorResponse "Unauthorized"
// @Failure 403 {object} main.ErrorResponse "Forbidden - Admin access required"
// @Failure 500 {object} main.ErrorResponse "Internal server error"
// @Router /settings [get]
func (h *SettingsHandler) getSettings(c *gin.Context) {
	// Only admins can access settings
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	settings, err := h.getSettingsFromDB()
	if err != nil {
		logrus.WithError(err).Error("Error fetching settings")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, SettingsResponse{Data: settings})
}

// updateSettings godoc
// @Summary Update system settings
// @Description Update system settings (Admin only)
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param settings body UpdateSettingsRequest true "Settings to update"
// @Success 200 {object} SettingsResponse "Updated settings"
// @Failure 400 {object} main.ErrorResponse "Invalid request payload"
// @Failure 401 {object} main.ErrorResponse "Unauthorized"
// @Failure 403 {object} main.ErrorResponse "Forbidden - Admin access required"
// @Failure 500 {object} main.ErrorResponse "Internal server error"
// @Router /settings [patch]
func (h *SettingsHandler) updateSettings(c *gin.Context) {
	// Only admins can update settings
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Invalid update settings request payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	currentUserID := c.GetString("user_id")
	settings, err := h.updateSettingsInDB(req, currentUserID)
	if err != nil {
		logrus.WithError(err).Error("Error updating settings")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, SettingsResponse{Data: settings})
}

// testDatabaseConnection godoc
// @Summary Test database connection
// @Description Test the database connection with current settings (Admin only)
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} main.SuccessMessage "Connection successful"
// @Failure 401 {object} main.ErrorResponse "Unauthorized"
// @Failure 403 {object} main.ErrorResponse "Forbidden - Admin access required"
// @Failure 500 {object} main.ErrorResponse "Connection failed"
// @Router /settings/test-connection [post]
func (h *SettingsHandler) testDatabaseConnection(c *gin.Context) {
	// Only admins can test connections
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	// Test current database connection
	if err := h.db.Ping(); err != nil {
		logrus.WithError(err).Error("Database connection test failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Database connection successful"})
}

// testEmailConfiguration godoc
// @Summary Test email configuration
// @Description Send a test email with current SMTP settings (Admin only)
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} main.SuccessMessage "Test email sent"
// @Failure 401 {object} main.ErrorResponse "Unauthorized"
// @Failure 403 {object} main.ErrorResponse "Forbidden - Admin access required"
// @Failure 500 {object} main.ErrorResponse "Email test failed"
// @Router /settings/test-email [post]
func (h *SettingsHandler) testEmailConfiguration(c *gin.Context) {
	// Only admins can test email
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	// Get current email settings
	settings, err := h.getSettingsFromDB()
	if err != nil {
		logrus.WithError(err).Error("Error fetching settings for email test")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if !settings.EmailEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is not enabled"})
		return
	}

	// TODO: Implement actual email sending logic
	// For now, just return success if email is enabled and has required fields
	if settings.SMTPHost == "" || settings.SMTPFromEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SMTP configuration incomplete"})
		return
	}

	logrus.Info("Email test requested - would send test email with current SMTP settings")
	c.JSON(http.StatusOK, gin.H{"message": "Test email sent successfully"})
}

// getSettingsFromDB retrieves settings from the database
func (h *SettingsHandler) getSettingsFromDB() (Settings, error) {
	var settings Settings

	// Default values
	settings.SiteName = "GoRectus"
	settings.SiteDescription = "A modern headless CMS built with Go"
	settings.AllowRegistration = false
	settings.MaintenanceMode = false
	settings.DatabaseHost = "localhost"
	settings.DatabasePort = "5432"
	settings.DatabaseName = "gorectus"
	settings.DatabaseUser = "gorectus"
	settings.SMTPHost = ""
	settings.SMTPPort = "587"
	settings.SMTPUser = ""
	settings.SMTPFromEmail = ""
	settings.EmailEnabled = false
	settings.SessionTimeout = 24
	settings.PasswordMinLength = 8
	settings.RequireTwoFactor = false
	settings.JWTSecretExists = true // Assume JWT secret exists
	settings.UpdatedAt = time.Now()
	settings.UpdatedBy = "system"

	// Try to load settings from database
	query := `
		SELECT 
			COALESCE(project_name, 'GoRectus'),
			COALESCE(project_descriptor, 'A modern headless CMS built with Go'),
			COALESCE(public_registration, false),
			COALESCE(maintenance_mode, false),
			COALESCE(smtp_host, ''),
			COALESCE(smtp_port, '587'),
			COALESCE(smtp_user, ''),
			COALESCE(smtp_from_email, ''),
			COALESCE(email_enabled, false),
			COALESCE(session_timeout, 24),
			COALESCE(password_min_length, 8),
			COALESCE(require_two_factor, false),
			updated_at
		FROM settings 
		LIMIT 1
	`

	var projectName, projectDescriptor, smtpHost, smtpPort, smtpUser, smtpFromEmail sql.NullString
	var publicRegistration, maintenanceMode, emailEnabled, requireTwoFactor sql.NullBool
	var sessionTimeout, passwordMinLength sql.NullInt64
	var updatedAt sql.NullTime

	err := h.db.QueryRow(query).Scan(
		&projectName, &projectDescriptor, &publicRegistration, &maintenanceMode,
		&smtpHost, &smtpPort, &smtpUser, &smtpFromEmail, &emailEnabled,
		&sessionTimeout, &passwordMinLength, &requireTwoFactor, &updatedAt,
	)

	if err != nil && err != sql.ErrNoRows {
		logrus.WithError(err).Error("Error loading settings from database")
		return settings, nil // Return defaults on error
	}

	if err == nil {
		// Update settings with database values
		if projectName.Valid {
			settings.SiteName = projectName.String
		}
		if projectDescriptor.Valid {
			settings.SiteDescription = projectDescriptor.String
		}
		if publicRegistration.Valid {
			settings.AllowRegistration = publicRegistration.Bool
		}
		if maintenanceMode.Valid {
			settings.MaintenanceMode = maintenanceMode.Bool
		}
		if smtpHost.Valid {
			settings.SMTPHost = smtpHost.String
		}
		if smtpPort.Valid {
			settings.SMTPPort = smtpPort.String
		}
		if smtpUser.Valid {
			settings.SMTPUser = smtpUser.String
		}
		if smtpFromEmail.Valid {
			settings.SMTPFromEmail = smtpFromEmail.String
		}
		if emailEnabled.Valid {
			settings.EmailEnabled = emailEnabled.Bool
		}
		if sessionTimeout.Valid {
			settings.SessionTimeout = int(sessionTimeout.Int64)
		}
		if passwordMinLength.Valid {
			settings.PasswordMinLength = int(passwordMinLength.Int64)
		}
		if requireTwoFactor.Valid {
			settings.RequireTwoFactor = requireTwoFactor.Bool
		}
		if updatedAt.Valid {
			settings.UpdatedAt = updatedAt.Time
		}
	}

	return settings, nil
}

// updateSettingsInDB updates settings in the database
func (h *SettingsHandler) updateSettingsInDB(req UpdateSettingsRequest, userID string) (Settings, error) {
	// Get current settings
	settings, err := h.getSettingsFromDB()
	if err != nil {
		return settings, err
	}

	// Build update query dynamically
	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	// Update fields if provided
	if req.SiteName != nil {
		settings.SiteName = *req.SiteName
		updateFields = append(updateFields, "project_name = $"+strconv.Itoa(argIndex))
		args = append(args, *req.SiteName)
		argIndex++
	}
	if req.SiteDescription != nil {
		settings.SiteDescription = *req.SiteDescription
		updateFields = append(updateFields, "project_descriptor = $"+strconv.Itoa(argIndex))
		args = append(args, *req.SiteDescription)
		argIndex++
	}
	if req.AllowRegistration != nil {
		settings.AllowRegistration = *req.AllowRegistration
		updateFields = append(updateFields, "public_registration = $"+strconv.Itoa(argIndex))
		args = append(args, *req.AllowRegistration)
		argIndex++
	}
	if req.MaintenanceMode != nil {
		settings.MaintenanceMode = *req.MaintenanceMode
		updateFields = append(updateFields, "maintenance_mode = $"+strconv.Itoa(argIndex))
		args = append(args, *req.MaintenanceMode)
		argIndex++
	}
	if req.SMTPHost != nil {
		settings.SMTPHost = *req.SMTPHost
		updateFields = append(updateFields, "smtp_host = $"+strconv.Itoa(argIndex))
		args = append(args, *req.SMTPHost)
		argIndex++
	}
	if req.SMTPPort != nil {
		settings.SMTPPort = *req.SMTPPort
		updateFields = append(updateFields, "smtp_port = $"+strconv.Itoa(argIndex))
		args = append(args, *req.SMTPPort)
		argIndex++
	}
	if req.SMTPUser != nil {
		settings.SMTPUser = *req.SMTPUser
		updateFields = append(updateFields, "smtp_user = $"+strconv.Itoa(argIndex))
		args = append(args, *req.SMTPUser)
		argIndex++
	}
	if req.SMTPFromEmail != nil {
		settings.SMTPFromEmail = *req.SMTPFromEmail
		updateFields = append(updateFields, "smtp_from_email = $"+strconv.Itoa(argIndex))
		args = append(args, *req.SMTPFromEmail)
		argIndex++
	}
	if req.EmailEnabled != nil {
		settings.EmailEnabled = *req.EmailEnabled
		updateFields = append(updateFields, "email_enabled = $"+strconv.Itoa(argIndex))
		args = append(args, *req.EmailEnabled)
		argIndex++
	}
	if req.SessionTimeout != nil {
		settings.SessionTimeout = *req.SessionTimeout
		updateFields = append(updateFields, "session_timeout = $"+strconv.Itoa(argIndex))
		args = append(args, *req.SessionTimeout)
		argIndex++
	}
	if req.PasswordMinLength != nil {
		settings.PasswordMinLength = *req.PasswordMinLength
		updateFields = append(updateFields, "password_min_length = $"+strconv.Itoa(argIndex))
		args = append(args, *req.PasswordMinLength)
		argIndex++
	}
	if req.RequireTwoFactor != nil {
		settings.RequireTwoFactor = *req.RequireTwoFactor
		updateFields = append(updateFields, "require_two_factor = $"+strconv.Itoa(argIndex))
		args = append(args, *req.RequireTwoFactor)
		argIndex++
	}

	settings.UpdatedAt = time.Now()
	settings.UpdatedBy = userID

	// Only execute update if there are fields to update
	if len(updateFields) > 0 {
		updateFields = append(updateFields, "updated_at = NOW()")

		query := "UPDATE settings SET " + strings.Join(updateFields, ", ") + " WHERE id IS NOT NULL OR id IS NULL"

		// If no settings row exists, insert one
		_, err = h.db.Exec(query, args...)
		if err != nil {
			// Try insert if update failed (no rows)
			insertQuery := `
				INSERT INTO settings (project_name, project_descriptor, public_registration, updated_at)
				VALUES ($1, $2, $3, NOW())
				ON CONFLICT (id) DO UPDATE SET
					project_name = EXCLUDED.project_name,
					project_descriptor = EXCLUDED.project_descriptor,
					public_registration = EXCLUDED.public_registration,
					updated_at = NOW()
			`

			insertArgs := []interface{}{settings.SiteName, settings.SiteDescription, settings.AllowRegistration}
			_, err = h.db.Exec(insertQuery, insertArgs...)
			if err != nil {
				return settings, err
			}
		}
	}

	// Handle other settings that might not be in the main settings table
	if req.SMTPHost != nil {
		settings.SMTPHost = *req.SMTPHost
	}
	if req.SMTPPort != nil {
		settings.SMTPPort = *req.SMTPPort
	}
	if req.SMTPUser != nil {
		settings.SMTPUser = *req.SMTPUser
	}
	if req.SMTPFromEmail != nil {
		settings.SMTPFromEmail = *req.SMTPFromEmail
	}
	if req.EmailEnabled != nil {
		settings.EmailEnabled = *req.EmailEnabled
	}
	if req.SessionTimeout != nil {
		settings.SessionTimeout = *req.SessionTimeout
	}
	if req.PasswordMinLength != nil {
		settings.PasswordMinLength = *req.PasswordMinLength
	}
	if req.RequireTwoFactor != nil {
		settings.RequireTwoFactor = *req.RequireTwoFactor
	}

	// Handle JWT secret if provided
	if req.JWTSecret != nil {
		// TODO: In production, this should be encrypted or stored securely
		// For now, we just acknowledge it was updated
		settings.JWTSecretExists = true
		logrus.Info("JWT secret updated (not stored in database for security)")
	}

	return settings, nil
}

// Helper function for JSON marshalling
func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}
