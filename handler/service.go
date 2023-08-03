package handler

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/repository"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Repository repository.RepositoryInterface
	Validator  Validator
}

type NewServiceOptions struct {
	Repository repository.RepositoryInterface
	Validator  Validator
}

func NewService(opts NewServiceOptions) *Service {
	return &Service{opts.Repository, opts.Validator}
}

func (s *Service) Register(ctx context.Context, regRequest *generated.RegistrationRequest) (string, []string) {
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

	salt := generateRandomSalt()
	temp, err := hashingPassword(regRequest.Password, salt)
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

func hashingPassword(password, salt string) (string, error) {
	passwordWithSalt := []byte(password + salt)
	hashedPassword, err := bcrypt.GenerateFromPassword(passwordWithSalt, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func generateRandomSalt() string {
	const saltLength = 16
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())

	salt := make([]byte, saltLength)
	for i := range salt {
		salt[i] = charset[rand.Intn(len(charset))]
	}

	return string(salt)
}
