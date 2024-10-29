// internal/auth/auth.go
package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/stackvity/aidoc-server/internal/core/domain"
	"go.uber.org/zap"
)

// ... (ClerkClaims type definition)

// AuthClient interface.
type AuthClient interface {
	Authorize(ctx context.Context, userID string, patientID int) (bool, error) // Change the signature of Authorize to return an error
	VerifyToken(tokenString string) (*domain.ClerkClaims, error)               // Add VerifyToken if you need to verify tokens manually
	GetUser(ctx context.Context, userID string) (*user.User, error)
}

type authClient struct { // Lowercase name for the struct. Updated.
	log *zap.Logger
}

func NewAuthClient(ctx context.Context, log *zap.Logger) (AuthClient, error) {

	return &authClient{log: log}, nil
}

func (c *authClient) GetUser(ctx context.Context, userID string) (*user.User, error) {

	user, err := clerk.User.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info from Clerk: %w", err)
	}

	return user, nil
}

func (c *authClient) VerifyToken(tokenString string) (*domain.ClerkClaims, error) { // updated and corrected
	c.log.Info("Verifying token")

	claims, err := jwt.Verify(context.Background(), &jwt.VerifyParams{
		Token: tokenString,
	})
	if err != nil {

		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	if claims.Claims.Subject == "" {

		return nil, fmt.Errorf("invalid user id claim") // Or a custom error type
	}

	c.log.Info("Token verified successfully", zap.String("user_id", claims.Subject)) // Log successful verification
	return &domain.ClerkClaims{UserID: claims.Subject}, nil                          // Correct this
}

// Authorize checks if a user is authorized to access a patient's data.
func (c *authClient) Authorize(ctx context.Context, userID string, patientID int) (bool, error) { // updated and corrected

	c.log.Info("Authorize method called", zap.Int("patient_id", patientID), zap.String("user_id", userID))

	// 1. Patients can access their own data
	if fmt.Sprint(patientID) == userID {
		return true, nil
	}
	clerkUser, err := clerk.User.GetUser(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, roles := range clerkUser.PublicMetadata {

		if roles == "physician" {
			return true, nil // Physicians can access all patient data
		}

		if roles == "clerk" {
			return true, nil // Clerks can access all patient data (for this example)
		}
	}

	if strings.Contains(fmt.Sprint(clerkUser.PublicMetadata), "patient:read_all") {
		return true, nil // allow patient to access with this permission
	}

	c.log.Warn("Authorization failed", zap.Int("patient_id", patientID), zap.String("user_id", userID))

	return false, nil // Return false if none of the conditions are met
}
