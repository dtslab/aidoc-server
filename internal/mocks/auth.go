package mocks

import (
	"context"

	"github.com/stackvity/aidoc-server/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

// AuthorizeMock mocks the auth.Authorize function
type AuthorizeMock struct {
	mock.Mock
}

// Authorize mocks the Authorize function.  It takes a context and patientID.
func (m *AuthorizeMock) Authorize(ctx context.Context, patientID int) bool {
	args := m.Called(ctx, patientID)
	return args.Bool(0)
}

// MockClerkClient mocks the auth.ClerkClient interface
type MockClerkClient struct {
	mock.Mock
}

// VerifyToken mocks the VerifyToken method
func (m *MockClerkClient) VerifyToken(tokenString string) (*domain.ClerkClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ClerkClaims), args.Error(1)
}

// GetUserRolesAndPermissions mocks the GetUserRolesAndPermissions method
func (m *MockClerkClient) GetUserRolesAndPermissions(userID string) ([]string, []string, error) {
	args := m.Called(userID)

	// Type assert the return values to avoid potential panics in tests.
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
