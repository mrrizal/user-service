package handler

import (
	"context"
	"errors"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

type mockRepository struct {
	isPhoneNumberExistsFunc func(context.Context, string) (bool, error)
	registerFunc            func(ctx context.Context, regRequest generated.RegistrationRequest, salt string) (string, error)
	loginFunc               func(ctx context.Context, loginRequest generated.LoginRequest) (string, error)
	getProfileFunc          func(ctx context.Context, userID string) (generated.UserProfile, error)
	updateProfileFunc       func(ctx context.Context, updateUserProfileRequest map[string]string,
		userID string) (generated.UserProfile, error)
}

func NewMockRepository() mockRepository {
	return mockRepository{
		isPhoneNumberExistsFunc: func(ctx context.Context, s string) (bool, error) {
			return false, nil
		},
		registerFunc: func(ctx context.Context, regRequest generated.RegistrationRequest, salt string) (string, error) {
			return "mockedUserID", nil
		},
		getProfileFunc: func(ctx context.Context, userID string) (generated.UserProfile, error) {
			return generated.UserProfile{}, nil
		},
		updateProfileFunc: func(ctx context.Context, updateUserProfileRequest map[string]string, userID string) (generated.UserProfile, error) {
			return generated.UserProfile{}, nil
		},
	}
}

func (m *mockRepository) IsPhoneNumberExists(ctx context.Context, phoneNumber string) (bool, error) {
	return m.isPhoneNumberExistsFunc(ctx, phoneNumber)
}

func (m *mockRepository) Register(ctx context.Context, regRequest generated.RegistrationRequest, salt string) (string, error) {
	return m.registerFunc(ctx, regRequest, salt)
}

func (m *mockRepository) Login(ctx context.Context, loginRequest generated.LoginRequest) (string, error) {
	return m.loginFunc(ctx, loginRequest)
}

func (m *mockRepository) GetUserProfile(ctx context.Context, userID string) (generated.UserProfile, error) {
	return m.getProfileFunc(ctx, userID)
}

func (m *mockRepository) UpdateUserProfile(ctx context.Context,
	updateUserProfileRequest map[string]string, userID string) (generated.UserProfile, error) {
	return m.updateProfileFunc(ctx, updateUserProfileRequest, userID)
}

type mockUtils struct {
	hashingPasswordFunc    func(password, salt string) (string, error)
	generateRandomSaltFunc func() string
	generateJWTTokenFunc   func(claims jwt.MapClaims) (string, error)
	extractJWTTokenFunc    func(ctx echo.Context) (string, error)
}

func (m *mockUtils) HashingPassword(password, salt string) (string, error) {
	return m.hashingPasswordFunc(password, salt)
}

func (m *mockUtils) GenerateRandomSalt() string {
	return m.generateRandomSaltFunc()
}

func (m *mockUtils) GenerateJWTToken(claims jwt.MapClaims) (string, error) {
	return m.generateJWTTokenFunc(claims)
}

func (m *mockUtils) ExtractJWTToken(ctx echo.Context) (string, error) {
	return m.extractJWTTokenFunc(ctx)
}

func NewMockUtils() mockUtils {
	return mockUtils{
		hashingPasswordFunc: func(password, salt string) (string, error) {
			return "", nil
		},
		generateRandomSaltFunc: func() string {
			return ""
		},
		generateJWTTokenFunc: func(claims jwt.MapClaims) (string, error) {
			return "", nil
		},
		extractJWTTokenFunc: func(ctx echo.Context) (string, error) {
			return "", nil
		},
	}
}

type MockValidator struct {
	MockIsValidPhoneNumber func(phoneNumber string) error
	MockIsValidFullName    func(fullName string) error
	MockIsValidPassword    func(password string) error
	MockValidateJWTToken   func(tokenString string) (*jwt.Token, error)
}

func (m *MockValidator) IsValidPhoneNumber(phoneNumber string) error {
	if m.MockIsValidPhoneNumber != nil {
		return m.MockIsValidPhoneNumber(phoneNumber)
	}
	// Replace this with your desired mock behavior
	return nil
}

func (m *MockValidator) IsValidFullName(fullName string) error {
	if m.MockIsValidFullName != nil {
		return m.MockIsValidFullName(fullName)
	}
	// Replace this with your desired mock behavior
	return nil
}

func (m *MockValidator) IsValidPassword(password string) error {
	if m.MockIsValidPassword != nil {
		return m.MockIsValidPassword(password)
	}
	// Replace this with your desired mock behavior
	return nil
}

func (m *MockValidator) ValidateJWTToken(tokenString string) (*jwt.Token, error) {
	if m.MockValidateJWTToken != nil {
		return m.MockValidateJWTToken(tokenString)
	}
	// Replace this with your desired mock behavior
	return nil, nil
}

func NewMockValidator() MockValidator {
	return MockValidator{}
}

var _ = ginkgo.Describe("Service", func() {
	var (
		service   *service
		ctx       context.Context
		regReq    *generated.RegistrationRequest
		repo      mockRepository
		utils     mockUtils
		validator Validator
	)

	ginkgo.BeforeEach(func() {
		utils = NewMockUtils()
		repo = NewMockRepository()
		validatorOpts := NewValidatorOptions{
			Repository: &repo,
		}
		validator = NewValidator(validatorOpts)

		serviceOpts := NewServiceOptions{
			Repository: &repo,
			Validator:  validator,
			Utils:      &utils,
		}
		service = NewService(serviceOpts)

		ctx = context.Background()
		regReq = &generated.RegistrationRequest{
			PhoneNumber: "+621234567890",
			FullName:    "John Doe",
			Password:    "P@ssw0rd",
		}
	})

	ginkgo.Context("Register", func() {
		ginkgo.It("should register a user", func() {
			userID, errs := service.Register(ctx, regReq)
			gomega.Expect(userID).To(gomega.Equal("mockedUserID"))
			gomega.Expect(errs).To(gomega.BeEmpty())
		})

		ginkgo.It("should return validation errors for invalid phone number", func() {
			regReq.PhoneNumber = ""
			userID, errs := service.Register(ctx, regReq)

			gomega.Expect(userID).To(gomega.BeEmpty())
			gomega.Expect(errs).To(gomega.HaveLen(1))
			gomega.Expect(errs[0]).To(gomega.ContainSubstring("Phone number"))
		})

		ginkgo.It("should return validation errors for invalid full name", func() {
			regReq.FullName = ""
			userID, errs := service.Register(ctx, regReq)

			gomega.Expect(userID).To(gomega.BeEmpty())
			gomega.Expect(errs).To(gomega.HaveLen(1))
			gomega.Expect(errs[0]).To(gomega.ContainSubstring("Full name"))
		})

		ginkgo.It("should return validation errors for invalid password", func() {
			regReq.Password = "short"
			userID, errs := service.Register(ctx, regReq)

			gomega.Expect(userID).To(gomega.BeEmpty())
			gomega.Expect(errs).To(gomega.HaveLen(1))
			gomega.Expect(errs[0]).To(gomega.ContainSubstring("Password"))
		})
	})

	ginkgo.Context("Login", func() {
		ginkgo.It("error when repository return error", func() {
			repo.loginFunc = func(ctx context.Context, loginRequest generated.LoginRequest) (string, error) {
				return "", errors.New("error")
			}
			token, err := service.Login(context.Background(), &generated.LoginRequest{})
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(token).To(gomega.Equal(""))
		})

		ginkgo.It("success login", func() {
			repo.loginFunc = func(ctx context.Context, loginRequest generated.LoginRequest) (string, error) {
				return "some_user_id", nil
			}
			utils.generateJWTTokenFunc = func(claims jwt.MapClaims) (string, error) {
				return "token", nil
			}
			token, err := service.Login(context.Background(), &generated.LoginRequest{})
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(token).NotTo(gomega.Equal(""))
		})

		ginkgo.It("raise error when generate token failed", func() {
			repo.loginFunc = func(ctx context.Context, loginRequest generated.LoginRequest) (string, error) {
				return "some_user_id", nil
			}
			utils.generateJWTTokenFunc = func(claims jwt.MapClaims) (string, error) {
				return "", errors.New("error")
			}
			token, err := service.Login(context.Background(), &generated.LoginRequest{})
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(token).To(gomega.Equal(""))
		})
	})

	ginkgo.Context("GetProfile", func() {
		var (
			fullName    = "test user"
			phoneNumber = "123456789"
		)

		ginkgo.BeforeEach(func() {
			mockRepo := NewMockRepository()
			mockRepo.getProfileFunc = func(ctx context.Context, userID string) (generated.UserProfile, error) {
				return generated.UserProfile{FullName: &fullName, PhoneNumber: &phoneNumber}, nil
			}

			mockValidator := NewMockValidator()
			mockValidator.MockValidateJWTToken = func(tokenString string) (*jwt.Token, error) {
				return &jwt.Token{}, nil
			}
			serviceOpts := NewServiceOptions{
				Repository: &mockRepo,
				Validator:  &mockValidator,
				Utils:      &utils,
			}
			service = NewService(serviceOpts)

		})

		ginkgo.It("should return the user profile when the JWT token is valid", func() {
			token := "test"
			userProfile, err := service.GetUserProfile(ctx, token)

			// Assertions
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(userProfile).ToNot(gomega.BeNil())
			// gomega.Expect(userProfile.FullName).To(gomega.Equal(fullName))
			// gomega.Expect(userProfile.PhoneNumber).To(gomega.Equal(phoneNumber))
		})
	})
})
