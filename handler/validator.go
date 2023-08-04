package handler

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/SawitProRecruitment/UserService/repository"
	"github.com/golang-jwt/jwt/v5"
)

type Validator struct {
	Repository repository.RepositoryInterface
}

type NewValidatorOptions struct {
	Repository repository.RepositoryInterface
}

func NewValidator(opts NewValidatorOptions) *Validator {
	return &Validator{opts.Repository}
}

func (v *Validator) IsValidPhoneNumber(phoneNumber string) error {
	message := `Phone numbers must start with "+62" and have 10 to 13 digits.`
	isValid := true
	if len(phoneNumber)-3 < 10 || len(phoneNumber)-3 > 13 {
		isValid = false
	}

	if !strings.HasPrefix(phoneNumber, "+62") {
		isValid = false
	}

	isExists, _ := v.Repository.IsPhoneNumberExists(context.Background(), phoneNumber)
	if isExists {
		message = "Phone numbers already exists."
		isValid = false
	}

	if !isValid {
		return errors.New(message)
	}
	return nil
}

func (v *Validator) IsValidFullName(fullName string) error {
	message := "Full name must be at minimum 3 characters and maximum 60 characters."
	if len(fullName) < 3 || len(fullName) > 60 {
		return errors.New(message)
	}
	return nil
}

func (v *Validator) IsValidPassword(password string) error {
	upperCaseRegex := `[A-Z]`
	digitRegex := `\d`
	specialCharRegex := `[^A-Za-z0-9]`
	lengthRegex := `.{6,}`

	message := "Passwords must have at least 6 characters, including 1 capital letter, 1 number, and 1 special character."

	if len(password) < 6 {
		return errors.New(message)
	}

	for _, pattern := range []string{upperCaseRegex, digitRegex, specialCharRegex, lengthRegex} {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}

		if !regex.MatchString(password) {
			return errors.New(message)
		}
	}

	return nil
}

func (v *Validator) ValidateJWTToken(tokenString string) (*jwt.Token, error) {
	var publicKeyPath = os.Getenv("PUBLIC_KEY")
	publicKeyBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %v", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}
