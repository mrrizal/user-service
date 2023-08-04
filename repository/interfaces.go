// This file contains the interfaces for the repository layer.
// The repository layer is responsible for interacting with the database.
// For testing purpose we will generate mock implementations of these
// interfaces using mockgen. See the Makefile for more information.
package repository

import (
	"context"

	"github.com/SawitProRecruitment/UserService/generated"
)

type RepositoryInterface interface {
	IsPhoneNumberExists(ctx context.Context, phoneNumber string) (bool, error)
	Register(ctx context.Context, regRequest generated.RegistrationRequest, salt string) (string, error)
	Login(ctx context.Context, loginRequest generated.LoginRequest) (string, error)
	GetUserProfile(ctx context.Context, userID string) (generated.UserProfile, error)
	UpdateUserProfile(ctx context.Context,
		updateUserProfileRequest map[string]string, userID string) (generated.UserProfile, error)
}
