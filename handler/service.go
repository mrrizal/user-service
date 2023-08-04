package handler

import (
	"context"
	"fmt"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/repository"
	"github.com/SawitProRecruitment/UserService/utils"
	"github.com/golang-jwt/jwt/v5"
)

type Service interface {
	Register(ctx context.Context, regRequest *generated.RegistrationRequest) (string, []string)
	Login(ctx context.Context, loginRequest *generated.LoginRequest) (string, error)
	GetUserProfile(ctx context.Context, token string) (generated.UserProfile, error)
	UpdateUserProfile(ctx context.Context,
		updateUserProfileRequest generated.UpdateUserProfileRequest, token string) (generated.UserProfile, error)
}

type service struct {
	Repository repository.RepositoryInterface
	Validator  Validator
	Utils      utils.Utils
}

type NewServiceOptions struct {
	Repository repository.RepositoryInterface
	Validator  Validator
	Utils      utils.Utils
}

func NewService(opts NewServiceOptions) *service {
	return &service{opts.Repository, opts.Validator, opts.Utils}
}

func (s *service) Register(ctx context.Context, regRequest *generated.RegistrationRequest) (string, []string) {
	errs := []string{}
	if err := s.Validator.IsValidPhoneNumber(regRequest.PhoneNumber); err != nil {
		errs = append(errs, fmt.Sprintf("phone_number: %s", err.Error()))
	}

	if err := s.Validator.IsValidFullName(regRequest.FullName); err != nil {
		errs = append(errs, fmt.Sprintf("full_name: %s", err.Error()))
	}

	if err := s.Validator.IsValidPassword(regRequest.Password); err != nil {
		errs = append(errs, fmt.Sprintf("password: %s", err.Error()))
	}

	if len(errs) > 0 {
		return "", errs
	}

	salt := s.Utils.GenerateRandomSalt()
	temp, err := s.Utils.HashingPassword(regRequest.Password, salt)
	if err != nil {
		errs = append(errs, err.Error())
		return "", errs
	}

	regRequest.Password = temp
	userID, err := s.Repository.Register(ctx, *regRequest, salt)
	if err != nil {
		errs = append(errs, err.Error())
	}
	return userID, errs
}

func (s *service) Login(ctx context.Context, loginRequest *generated.LoginRequest) (string, error) {
	data := make(map[string]interface{})
	userID, err := s.Repository.Login(ctx, *loginRequest)
	if err != nil {
		return "", err
	}

	data["user_id"] = userID
	data["phone_number"] = loginRequest.PhoneNumber
	jwtToken, err := s.Utils.GenerateJWTToken(data)
	if err != nil {
		return "", err
	}
	return jwtToken, nil
}

func (s *service) GetUserProfile(ctx context.Context, token string) (generated.UserProfile, error) {
	jwtToken, err := s.Validator.ValidateJWTToken(token)
	if err != nil {
		return generated.UserProfile{}, err
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return generated.UserProfile{}, err
	}

	userID := claims["user_id"].(string)
	return s.Repository.GetUserProfile(ctx, userID)
}

func (s *service) UpdateUserProfile(ctx context.Context, updateUserProfileRequest generated.UpdateUserProfileRequest,
	token string) (generated.UserProfile, error) {
	jwtToken, err := s.Validator.ValidateJWTToken(token)
	if err != nil {
		return generated.UserProfile{}, err
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return generated.UserProfile{}, err
	}

	userID := claims["user_id"].(string)

	updateUserProfileRequestMap := make(map[string]string)
	if updateUserProfileRequest.FullName != nil {
		if err := s.Validator.IsValidFullName(*updateUserProfileRequest.FullName); err == nil {
			updateUserProfileRequestMap["full_name"] = *updateUserProfileRequest.FullName
		} else {
			return generated.UserProfile{}, err
		}
	}

	if updateUserProfileRequest.PhoneNumber != nil {
		if err := s.Validator.IsValidPhoneNumber(*updateUserProfileRequest.PhoneNumber); err == nil {
			updateUserProfileRequestMap["phone_number"] = *updateUserProfileRequest.PhoneNumber
		} else {
			return generated.UserProfile{}, err
		}
	}

	return s.Repository.UpdateUserProfile(ctx, updateUserProfileRequestMap, userID)
}
