package handler

import (
	"context"

	"github.com/SawitProRecruitment/UserService/generated"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

type mockRepository struct {
	isPhoneNumberExistsFunc func(context.Context, string) (bool, error)
	registerFunc            func(ctx context.Context, regRequest generated.RegistrationRequest, salt string) (string, error)
}

func NewMockRepository() mockRepository {
	return mockRepository{
		isPhoneNumberExistsFunc: func(ctx context.Context, s string) (bool, error) {
			return false, nil
		},
		registerFunc: func(ctx context.Context, regRequest generated.RegistrationRequest, salt string) (string, error) {
			return "mockedUserID", nil
		},
	}
}

func (m *mockRepository) IsPhoneNumberExists(ctx context.Context, phoneNumber string) (bool, error) {
	return m.isPhoneNumberExistsFunc(ctx, phoneNumber)
}

func (m *mockRepository) Register(ctx context.Context, regRequest generated.RegistrationRequest, salt string) (string, error) {
	return m.registerFunc(ctx, regRequest, salt)
}

var _ = ginkgo.Describe("Service", func() {
	var (
		service *Service
		ctx     context.Context
		regReq  *generated.RegistrationRequest
	)

	ginkgo.BeforeEach(func() {
		repo := NewMockRepository()
		validatorOpts := NewValidatorOptions{
			Repository: &repo,
		}
		validator := NewValidator(validatorOpts)

		serviceOpts := NewServiceOptions{
			Repository: &repo,
			Validator:  *validator,
		}
		service = NewService(serviceOpts)

		ctx = context.Background()
		regReq = &generated.RegistrationRequest{
			PhoneNumber: "+621234567890",
			FullName:    "John Doe",
			Password:    "P@ssw0rd",
		}
	})

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
