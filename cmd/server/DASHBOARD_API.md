# Dashboard API Documentation

The Dashboard API provides comprehensive insights and analytics for system administrators to monitor and understand their GoRectus instance.

## Authentication

All dashboard endpoints require administrator-level access. Users with the "Administrator" role can access these endpoints with a valid JWT token.

## Endpoints

### `GET /api/v1/dashboard`

Returns a complete dashboard overview with all available metrics.

**Response Structure:**

```json
{
  "data": {
    "system_stats": {
      "total_users": 15,
      "active_users": 12,
      "total_roles": 3,
      "total_collections": 8,
      "total_sessions": 5
    },
    "user_insights": {
      "users_by_status": {
        "active": 12,
        "inactive": 2,
        "invited": 1
      },
      "users_by_role": {
        "Administrator": 2,
        "Editor": 5,
        "User": 8
      },
      "new_users_this_week": 3,
      "new_users_this_month": 7,
      "recent_registrations": [
        /* ... */
      ],
      "most_active_users": [
        /* ... */
      ]
    },
    "collection_metrics": {
      "total_collections": 8,
      "collections_by_type": {
        "Content": 5,
        "System": 2,
        "Ungrouped": 1
      },
      "recent_collections": [
        /* ... */
      ],
      "most_active_collections": [
        /* ... */
      ]
    },
    "recent_activity": [
      /* ... */
    ],
    "system_health": {
      "database_connected": true,
      "server_uptime": "N/A",
      "version": "1.0.0",
      "last_backup": null
    }
  }
}
```

### `GET /api/v1/dashboard/stats`

Returns basic system statistics.

**Response Structure:**

```json
{
  "data": {
    "total_users": 15,
    "active_users": 12,
    "total_roles": 3,
    "total_collections": 8,
    "total_sessions": 5
  }
}
```

### `GET /api/v1/dashboard/activity`

Returns recent system activity.

**Query Parameters:**

- `limit` (optional): Number of activity items to return (default: 20, max: 100)

**Response Structure:**

```json
{
  "data": [
    {
      "id": "uuid",
      "action": "create",
      "user_id": "user-uuid",
      "user_name": "John Doe",
      "collection": "posts",
      "item": "post-uuid",
      "comment": null,
      "timestamp": "2025-08-05T10:30:00Z",
      "ip": "127.0.0.1"
    }
  ]
}
```

### `GET /api/v1/dashboard/users`

Returns detailed user analytics and insights.

**Response Structure:**

```json
{
  "data": {
    "users_by_status": {
      "active": 12,
      "inactive": 2,
      "invited": 1
    },
    "users_by_role": {
      "Administrator": 2,
      "Editor": 5,
      "User": 8
    },
    "new_users_this_week": 3,
    "new_users_this_month": 7,
    "recent_registrations": [
      {
        "id": "user-uuid",
        "email": "john@example.com",
        "first_name": "John",
        "last_name": "Doe",
        "role": "User",
        "status": "active",
        "created_at": "2025-08-01T14:20:00Z",
        "last_access": "2025-08-05T09:15:00Z"
      }
    ],
    "most_active_users": [
      {
        "id": "user-uuid",
        "email": "jane@example.com",
        "first_name": "Jane",
        "last_name": "Smith",
        "role": "Editor",
        "status": "active",
        "created_at": "2025-07-15T10:00:00Z",
        "last_access": "2025-08-05T11:45:00Z"
      }
    ]
  }
}
```

### `GET /api/v1/dashboard/collections`

Returns collection metrics and insights.

**Response Structure:**

```json
{
  "data": {
    "total_collections": 8,
    "collections_by_type": {
      "Content": 5,
      "System": 2,
      "Ungrouped": 1
    },
    "recent_collections": [
      {
        "collection": "posts",
        "icon": "article",
        "note": "Blog posts and articles",
        "hidden": false,
        "singleton": false,
        "item_count": 0,
        "created_at": "2025-08-01T12:00:00Z"
      }
    ],
    "most_active_collections": [
      {
        "collection": "posts",
        "icon": "article",
        "note": "Blog posts and articles",
        "hidden": false,
        "singleton": false,
        "item_count": 15,
        "created_at": "2025-08-01T12:00:00Z"
      }
    ]
  }
}
```

## Error Responses

All endpoints can return the following error responses:

- **401 Unauthorized**: Missing or invalid authentication token
- **403 Forbidden**: User does not have administrator privileges
- **500 Internal Server Error**: Database or server error

Example error response:

```json
{
  "error": "Admin access required"
}
```

## Data Insights

### System Statistics

- **Total Users**: Count of all registered users
- **Active Users**: Count of users with "active" status
- **Total Roles**: Count of all defined roles
- **Total Collections**: Count of all collections
- **Total Sessions**: Count of active sessions (not expired)

### User Insights

- **Users by Status**: Distribution of users across different statuses
- **Users by Role**: Distribution of users across different roles
- **New Users This Week/Month**: Registration trends
- **Recent Registrations**: Latest user registrations (last 10)
- **Most Active Users**: Users sorted by last access time (last 10)

### Collection Metrics

- **Collections by Type**: Distribution based on the "group" field
- **Recent Collections**: Latest created collections (last 10)
- **Most Active Collections**: Collections with most activity in the last 30 days

### Activity Tracking

- Tracks all system actions (create, read, update, delete)
- Includes user information and timestamps
- Provides IP address for security monitoring

### System Health

- Database connectivity status
- Server version information
- Basic system health indicators

## Usage Examples

### Get Complete Dashboard

```bash
curl -H "Authorization: Bearer <token>" \
     https://api.example.com/api/v1/dashboard
```

### Get Recent Activity (Last 10 Items)

```bash
curl -H "Authorization: Bearer <token>" \
     "https://api.example.com/api/v1/dashboard/activity?limit=10"
```

### Monitor User Growth

```bash
curl -H "Authorization: Bearer <token>" \
     https://api.example.com/api/v1/dashboard/users
```
