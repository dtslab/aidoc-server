package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/gin-gonic/gin"
	"github.com/stackvity/aidoc-server/internal/core/domain"
	"go.uber.org/zap"
)

// AuthMiddleware is a Gin middleware for authentication with Clerk (v2.0.9).
func AuthMiddleware(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Context for Clerk API calls.  Best practice to create a new context for each request.
		ctx := context.Background()

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "Authorization header is missing"})
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "Authorization header format is invalid"})
			return
		}

		sessionToken := tokenParts[1]

		claims, err := jwt.Verify(ctx, &jwt.VerifyParams{Token: sessionToken})
		if err != nil {
			log.Error("JWT verification failed", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Error: "Invalid token"})
			return
		}

		// Get user information from Clerk (optional, but often needed)
		userInfo, err := user.Get(ctx, claims.Subject) // Use claims.Subject for UserID. Updated
		if err != nil {
			log.Error("Failed to get user info from Clerk", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to retrieve user information"})
			return
		}

		// Set user ID, claims, and user info in the Gin context for use in handlers
		c.Set("userID", userInfo.ID)
		c.Set("claims", claims)
		c.Set("userInfo", userInfo) // Store the complete user object if needed

		c.Next() // Important: Call c.Next() to continue the request chain
	}
}

func RequirePermissions(requiredPermissions []string, log *zap.Logger) gin.HandlerFunc { // Add log parameter
	return func(c *gin.Context) {

		claims, ok := c.Get("claims").(*jwt.SessionClaims)
		if !ok {
			log.Error("claims not found or invalid type in context") // Log error if claims not found. Updated
			c.AbortWithStatusJSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to get session claims"})
			return
		}

		for _, perm := range requiredPermissions {
			if !claims.HasPermission(perm) {

				log.Warn("User does not have required permission", zap.String("permission", perm)) // Log the missing permission.  Updated
				c.AbortWithStatusJSON(http.StatusForbidden, domain.ErrorResponse{Error: "Insufficient permissions"})
				return
			}
		}

		c.Next()
	}
}
