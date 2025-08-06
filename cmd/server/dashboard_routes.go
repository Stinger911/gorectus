package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// DashboardHandler handles dashboard-related routes
type DashboardHandler struct {
	db             *sql.DB
	authMiddleware gin.HandlerFunc
	optionsHandler gin.HandlerFunc
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(server ServerInterface) *DashboardHandler {
	return &DashboardHandler{
		db:             server.GetDB(),
		authMiddleware: server.AuthMiddleware(),
		optionsHandler: server.OptionsHandler(),
	}
}

// SetupRoutes sets up dashboard routes
func (h *DashboardHandler) SetupRoutes(v1 *gin.RouterGroup) {
	// CORS preflight OPTIONS for dashboard endpoints
	v1.OPTIONS("/dashboard", h.optionsHandler)
	v1.OPTIONS("/dashboard/stats", h.optionsHandler)
	v1.OPTIONS("/dashboard/activity", h.optionsHandler)
	v1.OPTIONS("/dashboard/users", h.optionsHandler)
	v1.OPTIONS("/dashboard/collections", h.optionsHandler)

	// Dashboard routes (protected)
	dashboard := v1.Group("/dashboard")
	dashboard.Use(h.authMiddleware)
	{
		dashboard.GET("", h.getDashboardOverview)
		dashboard.GET("/stats", h.getSystemStats)
		dashboard.GET("/activity", h.getRecentActivity)
		dashboard.GET("/users", h.getUserInsights)
		dashboard.GET("/collections", h.getCollectionInsights)
	}
}

// Dashboard data structures
type DashboardOverview struct {
	SystemStats       SystemStats       `json:"system_stats"`
	UserInsights      *UserInsights     `json:"user_insights,omitempty"`
	CollectionMetrics CollectionMetrics `json:"collection_metrics"`
	RecentActivity    []ActivityItem    `json:"recent_activity"`
	SystemHealth      SystemHealth      `json:"system_health"`
}

type SystemStats struct {
	TotalUsers       int `json:"total_users"`
	ActiveUsers      int `json:"active_users"`
	TotalRoles       int `json:"total_roles"`
	TotalCollections int `json:"total_collections"`
	TotalSessions    int `json:"total_sessions"`
}

type UserInsights struct {
	UsersByStatus       map[string]int `json:"users_by_status"`
	UsersByRole         map[string]int `json:"users_by_role"`
	NewUsersThisWeek    int            `json:"new_users_this_week"`
	NewUsersThisMonth   int            `json:"new_users_this_month"`
	RecentRegistrations []UserSummary  `json:"recent_registrations"`
	MostActiveUsers     []UserSummary  `json:"most_active_users"`
}

type CollectionMetrics struct {
	TotalCollections      int                 `json:"total_collections"`
	CollectionsByType     map[string]int      `json:"collections_by_type"`
	RecentCollections     []CollectionSummary `json:"recent_collections"`
	MostActiveCollections []CollectionSummary `json:"most_active_collections"`
}

type ActivityItem struct {
	ID         string    `json:"id"`
	Action     string    `json:"action"`
	UserID     *string   `json:"user_id"`
	UserName   *string   `json:"user_name"`
	Collection *string   `json:"collection"`
	Item       *string   `json:"item"`
	Comment    *string   `json:"comment"`
	Timestamp  time.Time `json:"timestamp"`
	IP         *string   `json:"ip"`
}

type UserSummary struct {
	ID         string     `json:"id"`
	Email      string     `json:"email"`
	FirstName  string     `json:"first_name"`
	LastName   string     `json:"last_name"`
	Role       string     `json:"role"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	LastAccess *time.Time `json:"last_access"`
}

type CollectionSummary struct {
	Collection string    `json:"collection"`
	Icon       *string   `json:"icon"`
	Note       *string   `json:"note"`
	Hidden     bool      `json:"hidden"`
	Singleton  bool      `json:"singleton"`
	ItemCount  int       `json:"item_count"`
	CreatedAt  time.Time `json:"created_at"`
}

type SystemHealth struct {
	DatabaseConnected bool       `json:"database_connected"`
	ServerUptime      string     `json:"server_uptime"`
	Version           string     `json:"version"`
	LastBackup        *time.Time `json:"last_backup"`
}

// isAdmin checks if the requesting user is an admin
func (h *DashboardHandler) isAdmin(c *gin.Context) bool {
	currentUserRole := c.GetString("user_role")
	return currentUserRole == "Administrator"
}

// GetDashboardOverview retrieves comprehensive dashboard overview data
//
//	@Summary		Get dashboard overview
//	@Description	Retrieve comprehensive dashboard overview including system stats, collection metrics, recent activity, system health, and user insights (user insights only available for admins)
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]DashboardOverview	"Dashboard overview data"
//	@Failure		401	{object}	map[string]string				"Unauthorized"
//	@Failure		500	{object}	map[string]string				"Internal server error"
//	@Router			/dashboard [get]
func (h *DashboardHandler) getDashboardOverview(c *gin.Context) {
	// Get all dashboard data
	systemStats, err := h.getSystemStatsData()
	if err != nil {
		logrus.WithError(err).Error("Error fetching system stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Only fetch user insights for admins
	var userInsights *UserInsights
	if h.isAdmin(c) {
		insights, err := h.getUserInsightsData()
		if err != nil {
			logrus.WithError(err).Error("Error fetching user insights")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		userInsights = &insights
	}

	collectionMetrics, err := h.getCollectionMetricsData()
	if err != nil {
		logrus.WithError(err).Error("Error fetching collection metrics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	recentActivity, err := h.getRecentActivityData(10)
	if err != nil {
		logrus.WithError(err).Error("Error fetching recent activity")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	systemHealth := h.getSystemHealthData()

	overview := DashboardOverview{
		SystemStats:       systemStats,
		UserInsights:      userInsights,
		CollectionMetrics: collectionMetrics,
		RecentActivity:    recentActivity,
		SystemHealth:      systemHealth,
	}

	c.JSON(http.StatusOK, gin.H{"data": overview})
}

// GetSystemStats retrieves system statistics
//
//	@Summary		Get system statistics
//	@Description	Retrieve system statistics including user counts, role counts, collection counts, and active sessions
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]SystemStats	"System statistics"
//	@Failure		401	{object}	map[string]string		"Unauthorized"
//	@Failure		500	{object}	map[string]string		"Internal server error"
//	@Router			/dashboard/stats [get]
func (h *DashboardHandler) getSystemStats(c *gin.Context) {
	stats, err := h.getSystemStatsData()
	if err != nil {
		logrus.WithError(err).Error("Error fetching system stats")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetRecentActivity retrieves recent system activity
//
//	@Summary		Get recent activity
//	@Description	Retrieve recent system activity with optional limit parameter
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query		int						false	"Maximum number of activity items to return (1-100, default: 20)"
//	@Success		200		{object}	map[string][]ActivityItem	"Recent activity data"
//	@Failure		401		{object}	map[string]string			"Unauthorized"
//	@Failure		500		{object}	map[string]string			"Internal server error"
//	@Router			/dashboard/activity [get]
func (h *DashboardHandler) getRecentActivity(c *gin.Context) {
	// Parse limit parameter
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	activity, err := h.getRecentActivityData(limit)
	if err != nil {
		logrus.WithError(err).Error("Error fetching recent activity")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": activity})
}

// GetUserInsights retrieves user insights and analytics
//
//	@Summary		Get user insights
//	@Description	Retrieve user insights including user distribution by status and role, new user statistics, and user activity data (Admin only)
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]UserInsights	"User insights data"
//	@Failure		401	{object}	map[string]string		"Unauthorized"
//	@Failure		403	{object}	map[string]string		"Forbidden - Admin access required"
//	@Failure		500	{object}	map[string]string		"Internal server error"
//	@Router			/dashboard/users [get]
func (h *DashboardHandler) getUserInsights(c *gin.Context) {
	// Check if user is admin
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
		return
	}

	insights, err := h.getUserInsightsData()
	if err != nil {
		logrus.WithError(err).Error("Error fetching user insights")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": insights})
}

// GetCollectionInsights retrieves collection metrics and analytics
//
//	@Summary		Get collection insights
//	@Description	Retrieve collection metrics including total collections, distribution by type, and activity data
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	map[string]CollectionMetrics	"Collection metrics data"
//	@Failure		401	{object}	map[string]string				"Unauthorized"
//	@Failure		500	{object}	map[string]string				"Internal server error"
//	@Router			/dashboard/collections [get]
func (h *DashboardHandler) getCollectionInsights(c *gin.Context) {
	metrics, err := h.getCollectionMetricsData()
	if err != nil {
		logrus.WithError(err).Error("Error fetching collection metrics")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": metrics})
}

// Helper functions to fetch data
func (h *DashboardHandler) getSystemStatsData() (SystemStats, error) {
	var stats SystemStats

	// Get total users
	err := h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	if err != nil {
		return stats, err
	}

	// Get active users (users with status 'active')
	err = h.db.QueryRow("SELECT COUNT(*) FROM users WHERE status = 'active'").Scan(&stats.ActiveUsers)
	if err != nil {
		return stats, err
	}

	// Get total roles
	err = h.db.QueryRow("SELECT COUNT(*) FROM roles").Scan(&stats.TotalRoles)
	if err != nil {
		return stats, err
	}

	// Get total collections
	err = h.db.QueryRow("SELECT COUNT(*) FROM collections").Scan(&stats.TotalCollections)
	if err != nil {
		return stats, err
	}

	// Get active sessions
	err = h.db.QueryRow("SELECT COUNT(*) FROM sessions WHERE expires > NOW()").Scan(&stats.TotalSessions)
	if err != nil {
		return stats, err
	}

	return stats, nil
}

func (h *DashboardHandler) getUserInsightsData() (UserInsights, error) {
	var insights UserInsights
	insights.UsersByStatus = make(map[string]int)
	insights.UsersByRole = make(map[string]int)

	// Get users by status
	rows, err := h.db.Query("SELECT status, COUNT(*) FROM users GROUP BY status")
	if err != nil {
		return insights, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return insights, err
		}
		insights.UsersByStatus[status] = count
	}

	// Get users by role
	rows, err = h.db.Query(`
		SELECT r.name, COUNT(u.id) 
		FROM roles r 
		LEFT JOIN users u ON r.id = u.role_id 
		GROUP BY r.name
	`)
	if err != nil {
		return insights, err
	}
	defer rows.Close()

	for rows.Next() {
		var roleName string
		var count int
		if err := rows.Scan(&roleName, &count); err != nil {
			return insights, err
		}
		insights.UsersByRole[roleName] = count
	}

	// Get new users this week
	err = h.db.QueryRow(`
		SELECT COUNT(*) FROM users 
		WHERE created_at >= DATE_TRUNC('week', NOW())
	`).Scan(&insights.NewUsersThisWeek)
	if err != nil {
		return insights, err
	}

	// Get new users this month
	err = h.db.QueryRow(`
		SELECT COUNT(*) FROM users 
		WHERE created_at >= DATE_TRUNC('month', NOW())
	`).Scan(&insights.NewUsersThisMonth)
	if err != nil {
		return insights, err
	}

	// Get recent registrations (last 10)
	rows, err = h.db.Query(`
		SELECT u.id, u.email, u.first_name, u.last_name, r.name as role, 
		       u.status, u.created_at, u.last_access
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		ORDER BY u.created_at DESC
		LIMIT 10
	`)
	if err != nil {
		return insights, err
	}
	defer rows.Close()

	for rows.Next() {
		var user UserSummary
		err := rows.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName,
			&user.Role, &user.Status, &user.CreatedAt, &user.LastAccess)
		if err != nil {
			return insights, err
		}
		insights.RecentRegistrations = append(insights.RecentRegistrations, user)
	}

	// Get most active users (by last access)
	rows, err = h.db.Query(`
		SELECT u.id, u.email, u.first_name, u.last_name, r.name as role, 
		       u.status, u.created_at, u.last_access
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.last_access IS NOT NULL
		ORDER BY u.last_access DESC
		LIMIT 10
	`)
	if err != nil {
		return insights, err
	}
	defer rows.Close()

	for rows.Next() {
		var user UserSummary
		err := rows.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName,
			&user.Role, &user.Status, &user.CreatedAt, &user.LastAccess)
		if err != nil {
			return insights, err
		}
		insights.MostActiveUsers = append(insights.MostActiveUsers, user)
	}

	return insights, nil
}

func (h *DashboardHandler) getCollectionMetricsData() (CollectionMetrics, error) {
	var metrics CollectionMetrics
	metrics.CollectionsByType = make(map[string]int)

	// Get total collections
	err := h.db.QueryRow("SELECT COUNT(*) FROM collections").Scan(&metrics.TotalCollections)
	if err != nil {
		return metrics, err
	}

	// Get collections by type (using group field)
	rows, err := h.db.Query(`
		SELECT COALESCE("group", 'Ungrouped') as group_type, COUNT(*) 
		FROM collections 
		GROUP BY "group"
	`)
	if err != nil {
		return metrics, err
	}
	defer rows.Close()

	for rows.Next() {
		var groupType string
		var count int
		if err := rows.Scan(&groupType, &count); err != nil {
			return metrics, err
		}
		metrics.CollectionsByType[groupType] = count
	}

	// Get recent collections
	rows, err = h.db.Query(`
		SELECT collection, icon, note, hidden, singleton, created_at
		FROM collections
		ORDER BY created_at DESC
		LIMIT 10
	`)
	if err != nil {
		return metrics, err
	}
	defer rows.Close()

	for rows.Next() {
		var collection CollectionSummary
		err := rows.Scan(&collection.Collection, &collection.Icon, &collection.Note,
			&collection.Hidden, &collection.Singleton, &collection.CreatedAt)
		if err != nil {
			return metrics, err
		}
		// Note: Item count would require dynamic table queries which is complex
		// For now, we'll set it to 0 as a placeholder
		collection.ItemCount = 0
		metrics.RecentCollections = append(metrics.RecentCollections, collection)
	}

	// Get most active collections (by activity log)
	rows, err = h.db.Query(`
		SELECT c.collection, c.icon, c.note, c.hidden, c.singleton, c.created_at,
		       COUNT(a.id) as activity_count
		FROM collections c
		LEFT JOIN activity a ON c.collection = a.collection
		WHERE a.timestamp >= NOW() - INTERVAL '30 days'
		GROUP BY c.collection, c.icon, c.note, c.hidden, c.singleton, c.created_at
		ORDER BY activity_count DESC
		LIMIT 10
	`)
	if err != nil {
		return metrics, err
	}
	defer rows.Close()

	for rows.Next() {
		var collection CollectionSummary
		var activityCount int
		err := rows.Scan(&collection.Collection, &collection.Icon, &collection.Note,
			&collection.Hidden, &collection.Singleton, &collection.CreatedAt, &activityCount)
		if err != nil {
			return metrics, err
		}
		collection.ItemCount = activityCount // Using activity count as a proxy
		metrics.MostActiveCollections = append(metrics.MostActiveCollections, collection)
	}

	return metrics, nil
}

func (h *DashboardHandler) getRecentActivityData(limit int) ([]ActivityItem, error) {
	var activities []ActivityItem

	rows, err := h.db.Query(`
		SELECT a.id, a.action, a.user_id, u.first_name, u.last_name, 
		       a.collection, a.item, a.comment, a.timestamp, a.ip
		FROM activity a
		LEFT JOIN users u ON a.user_id = u.id
		ORDER BY a.timestamp DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return activities, err
	}
	defer rows.Close()

	for rows.Next() {
		var activity ActivityItem
		var firstName, lastName sql.NullString
		err := rows.Scan(&activity.ID, &activity.Action, &activity.UserID,
			&firstName, &lastName, &activity.Collection, &activity.Item,
			&activity.Comment, &activity.Timestamp, &activity.IP)
		if err != nil {
			return activities, err
		}

		// Construct user name from first and last name
		if firstName.Valid && lastName.Valid {
			userName := firstName.String + " " + lastName.String
			activity.UserName = &userName
		}

		activities = append(activities, activity)
	}

	return activities, nil
}

func (h *DashboardHandler) getSystemHealthData() SystemHealth {
	health := SystemHealth{
		DatabaseConnected: true, // If we reach here, DB is connected
		Version:           "1.0.0",
		ServerUptime:      "N/A", // This would require storing server start time
	}

	// Check for last backup (this would depend on your backup strategy)
	// For now, we'll leave it as nil since there's no backup table in the schema

	return health
}
