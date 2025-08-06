# GoRectus API Documentation

This document provides information about accessing the OpenAPI/Swagger documentation for the GoRectus API.

## Swagger/OpenAPI Documentation

The GoRectus API now includes comprehensive OpenAPI 3.0 documentation with Swagger UI integration.

### Accessing the Documentation

Once the server is running, you can access the interactive Swagger documentation at:

```
http://localhost:8080/swagger/index.html
```

### What's Included

The API documentation includes:

- **Complete API Reference**: All endpoints with detailed descriptions
- **Request/Response Models**: Comprehensive data models with examples
- **Authentication**: JWT Bearer token authentication documentation
- **Interactive Testing**: Try out API endpoints directly from the documentation
- **Response Examples**: Sample responses for all endpoints

### API Endpoints Coverage

The documentation covers the following endpoint groups:

1. **Authentication** (`/api/v1/auth`)

   - Login (`POST /auth/login`)
   - Logout (`POST /auth/logout`)
   - Refresh Token (`POST /auth/refresh`)
   - Get Current User (`GET /auth/me`)

2. **Collections** (`/api/v1/collections`)

   - Get all collections (`GET /collections`)
   - Create collection (`POST /collections`)
   - Get collection by name (`GET /collections/{collection}`)
   - Update collection (`PATCH /collections/{collection}`)
   - Delete collection (`DELETE /collections/{collection}`)

3. **Items** (`/api/v1/items`)

   - Get items from collection (`GET /items/{collection}`)
   - Create item (`POST /items/{collection}`)
   - Get item by ID (`GET /items/{collection}/{id}`)
   - Update item (`PATCH /items/{collection}/{id}`)
   - Delete item (`DELETE /items/{collection}/{id}`)

4. **Health Check** (`/api/v1/health`)
   - Health status (`GET /health`)

### Authentication in Swagger UI

To use authenticated endpoints in the Swagger UI:

1. First, use the `/auth/login` endpoint to obtain a JWT token
2. Click the "Authorize" button at the top of the Swagger UI
3. Enter `Bearer <your-jwt-token>` in the authorization field
4. Click "Authorize" to apply the token to all subsequent requests

### Development

#### Regenerating Documentation

If you make changes to the API endpoints or add new Swagger annotations, regenerate the documentation:

```bash
# Install swag CLI tool (if not already installed)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate updated documentation
swag init -g cmd/server/main.go -o docs
```

#### Adding New Endpoint Documentation

To add documentation for new endpoints:

1. Add Swagger annotations above your handler functions:

```go
// YourHandler does something
//
//	@Summary		Brief description
//	@Description	Detailed description
//	@Tags			tag-name
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth  // If authentication required
//	@Param			param-name	path/query/body	type	required	"Description"
//	@Success		200			{object}		ResponseModel	"Success description"
//	@Failure		400			{object}		ErrorResponse	"Error description"
//	@Router			/endpoint [method]
func (h *Handler) YourHandler(c *gin.Context) {
    // handler implementation
}
```

2. Regenerate the documentation using the swag command above

### API Models

The documentation includes the following data models:

- `LoginRequest` / `LoginResponse`: Authentication request/response
- `UserModel`: User information structure
- `CollectionModel`: Collection data structure
- `ItemModel`: Item data structure with flexible content
- `ErrorResponse`: Standard error response format
- `SuccessMessage`: Standard success response format

### File Structure

```
docs/
├── docs.go         # Generated Go documentation
├── swagger.json    # OpenAPI 3.0 JSON specification
└── swagger.yaml    # OpenAPI 3.0 YAML specification
```

The generated files are automatically imported by the main application and served at the `/swagger/*` endpoint.
