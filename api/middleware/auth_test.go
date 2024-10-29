package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stackvity/aidoc-server/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockClerkClient for testing
type MockClerkClient struct {
	mock.Mock
}

// VerifyToken mocks Clerk's token verification
func (m *MockClerkClient) VerifyToken(tokenString string) (*domain.ClerkClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ClerkClaims), args.Error(1)
}

// GetUserRolesAndPermissions mocks retrieving user roles and permissions. Updated.
func (m *MockClerkClient) GetUserRolesAndPermissions(userID string) ([]string, []string, error) {
	args := m.Called(userID)

	// Handle nil return values from mock to prevent test panics
	var roles []string
	if args.Get(0) != nil {
		roles = args.Get(0).([]string)
	}
	var permissions []string
	if args.Get(1) != nil {
		permissions = args.Get(1).([]string)
	}

	return roles, permissions, args.Error(2)
}

func (m *MockClerkClient) Authorize(ctx context.Context, patientID int) bool { // Add mock for Authorize method
	args := m.Called(ctx, patientID)
	return args.Bool(0)
}

func TestAuthMiddleware(t *testing.T) {
	log := zap.NewNop()

	t.Run("valid_token", func(t *testing.T) {
		mockClerk := new(MockClerkClient)
		mockClerk.On("VerifyToken", "valid-token").Return(&domain.ClerkClaims{UserID: "user-123"}, nil)
		mockClerk.On("GetUserRolesAndPermissions", "user-123").Return([]string{"patient"}, []string{"patient:read"}, nil) // Mock roles and permissions

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
		c.Request.Header.Set("Authorization", "Bearer valid-token")

		middleware := AuthMiddleware(log, mockClerk)
		middleware(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "user-123", c.GetString("userID"))              // Assert userID set in context
		assert.Equal(t, []string{"patient"}, c.GetStringSlice("roles")) // Assert roles
		assert.Equal(t, []string{"patient:read"}, c.GetStringSlice("permissions"))

		mockClerk.AssertExpectations(t)
	})

	t.Run("invalid_token_verification", func(t *testing.T) {
		mockClerk := new(MockClerkClient)
		mockClerk.On("VerifyToken", "invalid-token").Return(nil, errors.New("invalid token"))
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
		c.Request.Header.Set("Authorization", "Bearer invalid-token")

		middleware := AuthMiddleware(log, mockClerk)
		middleware(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "Invalid token", errResp.Error)
		mockClerk.AssertExpectations(t)
	})

	t.Run("missing_auth_header", func(t *testing.T) {
		mockClerk := new(MockClerkClient)
		// No mock setup needed as the middleware should fail before reaching Clerk
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
		// Explicitly DO NOT set the Authorization header

		middleware := AuthMiddleware(log, mockClerk)
		middleware(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)               // Expected result
		var errResp domain.ErrorResponse                               // Check for expected response
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)                   // Unmarshal JSON response.
		assert.Equal(t, "Authorization header missing", errResp.Error) // Assert the error message. Updated.
		mockClerk.AssertNotCalled(t, "VerifyToken")                    // Ensure no calls to Clerk API

	})

	t.Run("invalid_auth_header_format", func(t *testing.T) {
		mockClerk := new(MockClerkClient)
		// No mock setup needed, should fail before hitting Clerk client.
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
		c.Request.Header.Set("Authorization", "InvalidFormat")

		middleware := AuthMiddleware(log, mockClerk)
		middleware(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)                      // Expected status.
		var errResp domain.ErrorResponse                                      // Unmarshal the JSON response
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)                          // Updated: Unmarshal JSON response
		assert.Equal(t, "Invalid authorization header format", errResp.Error) // Expected message. Updated
		mockClerk.AssertNotCalled(t, "VerifyToken")                           // Assert not call
	})

	t.Run("clerk_server_error_get_user", func(t *testing.T) {
		mockClerk := new(MockClerkClient)
		mockClerk.On("VerifyToken", "valid-token").Return(&domain.ClerkClaims{UserID: "user-123"}, nil)
		mockClerk.On("GetUserRolesAndPermissions", "user-123").
			Return(nil, nil, errors.New("clerk server error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
		c.Request.Header.Set("Authorization", "Bearer valid-token")

		middleware := AuthMiddleware(log, mockClerk)
		middleware(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "Failed to retrieve user information", errResp.Error)

		mockClerk.AssertExpectations(t)

	})

	t.Run("no_user_id_claim", func(t *testing.T) { // Add test cases for missing user ID claim
		mockClerk := new(MockClerkClient)
		mockClerk.On("VerifyToken", "no-user-id-token").
			Return(&domain.ClerkClaims{}, nil) // Token verification succeeds, but claims are empty.

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/", nil)
		c.Request.Header.Set("Authorization", "Bearer no-user-id-token")

		middleware := AuthMiddleware(log, mockClerk)
		middleware(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect unauthorized error

		// Add assertions to check the response body and logged error messages
		var errResp domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp) // Updated: Unmarshal the response
		if err != nil {
			t.Fatal("Failed to unmarshal error response:", err) // Handle unmarshaling error
		}
		assert.Equal(t, "Invalid token", errResp.Error) // Check if the correct error message is returned. Updated

	})

}

// ... other relevant test functions

// TestRequirePermissions tests the RequirePermissions middleware.
func TestRequirePermissions(t *testing.T) {
	log := zap.NewNop()

	t.Run("has_permissions", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("permissions", []string{"patient:read", "patient:write"}) // Set permissions in context
		middleware := RequirePermissions([]string{"patient:read"}, log) // Middleware with one required permission
		middleware(c)
		assert.Equal(t, http.StatusOK, w.Code) // Expect 200 OK
	})

	t.Run("missing_permissions", func(t *testing.T) { // Correct test name. Updated.
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("permissions", []string{"patient:write"})                                   // User does not have "patient:read" permission
		middleware := RequirePermissions([]string{"patient:read", "patient:update"}, log) // Middleware requiring patient:read and patient:update permissions
		middleware(c)

		assert.Equal(t, http.StatusForbidden, w.Code) // Expect Forbidden (403)

		var errResp domain.ErrorResponse
		_ = json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.Equal(t, "Insufficient permissions", errResp.Error) // Assert error message

	})

	t.Run("no_permissions_in_context", func(t *testing.T) { // Correct test name. Updated
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		// Permissions are *not* set in the context
		middleware := RequirePermissions([]string{"patient:read"}, log) // Middleware with "patient:read" permission
		middleware(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code) // Internal Server Error (500)

		var errResp domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err) // Ensure unmarshaling the error response doesn't fail
		assert.Equal(t, domain.ErrorResponse{Error: "Failed to retrieve user permissions"}, errResp)
	})

	t.Run("invalid_permission_type", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("permissions", 123) // Set an invalid type for permissions
		middleware := RequirePermissions([]string{"patient:read"}, log)

		middleware(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errResp domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errResp)
		assert.NoError(t, err)
		assert.Equal(t, domain.ErrorResponse{Error: "Failed to retrieve user permissions"}, errResp)
	})
	// Add test cases for Clerk server errors if applicable
}
