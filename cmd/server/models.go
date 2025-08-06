package main

import "time"

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	AccessToken string    `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType   string    `json:"token_type" example:"Bearer"`
	ExpiresIn   int       `json:"expires_in" example:"86400"`
	User        UserModel `json:"user"`
}

// UserModel represents a user in the system
type UserModel struct {
	ID                 string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email              string    `json:"email" example:"user@example.com"`
	FirstName          *string   `json:"first_name" example:"John"`
	LastName           *string   `json:"last_name" example:"Doe"`
	Avatar             *string   `json:"avatar" example:"https://example.com/avatar.jpg"`
	Language           *string   `json:"language" example:"en"`
	Theme              *string   `json:"theme" example:"light"`
	Status             *string   `json:"status" example:"active"`
	RoleID             string    `json:"role_id" example:"456e7890-e89b-12d3-a456-426614174001"`
	RoleName           string    `json:"role_name" example:"admin"`
	LastAccess         *string   `json:"last_access" example:"2023-12-01T10:30:00Z"`
	LastPage           *string   `json:"last_page" example:"/dashboard"`
	Provider           *string   `json:"provider" example:"local"`
	ExternalIdentifier *string   `json:"external_identifier"`
	EmailNotifications bool      `json:"email_notifications" example:"true"`
	Tags               *string   `json:"tags" example:"admin,power-user"`
	CreatedAt          time.Time `json:"created_at" example:"2023-01-01T10:30:00Z"`
	UpdatedAt          time.Time `json:"updated_at" example:"2023-12-01T10:30:00Z"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string `json:"status" example:"ok"`
	Service string `json:"service" example:"gorectus"`
	Version string `json:"version" example:"1.0.0"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request payload"`
}

// SuccessMessage represents a generic success response
type SuccessMessage struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

// CollectionModel represents a collection in the system
type CollectionModel struct {
	ID          string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        string    `json:"name" example:"Products"`
	Description *string   `json:"description" example:"Product catalog collection"`
	Schema      *string   `json:"schema" example:"{}"`
	CreatedAt   time.Time `json:"created_at" example:"2023-01-01T10:30:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2023-12-01T10:30:00Z"`
}

// ItemModel represents an item in a collection
type ItemModel struct {
	ID           string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	CollectionID string    `json:"collection_id" example:"456e7890-e89b-12d3-a456-426614174001"`
	Data         ItemData  `json:"data"`
	CreatedAt    time.Time `json:"created_at" example:"2023-01-01T10:30:00Z"`
	UpdatedAt    time.Time `json:"updated_at" example:"2023-12-01T10:30:00Z"`
}

// ItemData represents the flexible data structure of an item
type ItemData struct {
	// Dynamic fields based on collection schema
	// Examples: name, description, price, etc.
}

// FieldModel represents a field definition in a collection
type FieldModel struct {
	ID           string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	CollectionID string    `json:"collection_id" example:"456e7890-e89b-12d3-a456-426614174001"`
	Name         string    `json:"name" example:"product_name"`
	Type         string    `json:"type" example:"string"`
	Required     bool      `json:"required" example:"true"`
	Options      *string   `json:"options" example:"{\"max_length\":255}"`
	CreatedAt    time.Time `json:"created_at" example:"2023-01-01T10:30:00Z"`
	UpdatedAt    time.Time `json:"updated_at" example:"2023-12-01T10:30:00Z"`
}

// RoleModel represents a user role
type RoleModel struct {
	ID          string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        string    `json:"name" example:"admin"`
	Description *string   `json:"description" example:"Administrator role with full access"`
	Permissions *string   `json:"permissions" example:"{\"read\":true,\"write\":true,\"delete\":true}"`
	CreatedAt   time.Time `json:"created_at" example:"2023-01-01T10:30:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2023-12-01T10:30:00Z"`
}

// DashboardStatsResponse represents dashboard statistics
type DashboardStatsResponse struct {
	TotalUsers       int64 `json:"total_users" example:"150"`
	TotalCollections int64 `json:"total_collections" example:"25"`
	TotalItems       int64 `json:"total_items" example:"1500"`
	ActiveUsers      int64 `json:"active_users" example:"75"`
}
