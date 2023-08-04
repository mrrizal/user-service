package utils

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type Utils interface {
	HashingPassword(password, salt string) (string, error)
	GenerateRandomSalt() string
	GenerateJWTToken(claims jwt.MapClaims) (string, error)
	ExtractJWTToken(ctx echo.Context) (string, error)
}

type utils struct{}

func NewUtils() *utils {
	return &utils{}
}

func (u *utils) HashingPassword(password, salt string) (string, error) {
	passwordWithSalt := []byte(password + salt)
	hashedPassword, err := bcrypt.GenerateFromPassword(passwordWithSalt, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func (u *utils) GenerateRandomSalt() string {
	const saltLength = 16
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())

	salt := make([]byte, saltLength)
	for i := range salt {
		salt[i] = charset[rand.Intn(len(charset))]
	}

	return string(salt)
}

func (u *utils) GenerateJWTToken(claims jwt.MapClaims) (string, error) {
	var privateKeyPath = os.Getenv("PRIVATE_KEY")
	privateKeyBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read private key file: %v", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %v", err)
	}

	// Set the expiration time
	expirationTime := time.Now().Add(24 * time.Hour)

	claims["exp"] = expirationTime.Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (u *utils) ExtractJWTToken(ctx echo.Context) (string, error) {
	authHeader := ctx.Request().Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("Authorization header is missing")
	}

	tokenString := ""
	fmt.Sscanf(authHeader, "Bearer %s", &tokenString)

	if tokenString == "" {
		return "", fmt.Errorf("JWT token is missing")
	}

	return tokenString, nil
}
