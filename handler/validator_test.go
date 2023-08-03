package handler

import (
	"context"
	"strings"

	"github.com/SawitProRecruitment/UserService/repository"
	"github.com/golang/mock/gomock"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Validator", func() {
	var (
		validator *Validator
		ctrl      *gomock.Controller
		mockRepo  *repository.MockRepositoryInterface
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())

		mockRepo = repository.NewMockRepositoryInterface(ctrl)
		optsValidator := NewValidatorOptions{
			Repository: mockRepo,
		}

		validator = NewValidator(optsValidator)
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Describe("IsValidPhoneNumber", func() {
		ginkgo.Context("when the phone number is valid", func() {
			ginkgo.It("should return nil error", func() {
				mockRepo.EXPECT().IsPhoneNumberExists(context.Background(), "+6281234567890").Return(false, nil)
				err := validator.IsValidPhoneNumber("+6281234567890")
				gomega.Expect(err).To(gomega.BeNil())

			})
		})

		ginkgo.Context("when the phone number does not start with \"+62\"", func() {
			ginkgo.It("should return an error", func() {
				mockRepo.EXPECT().IsPhoneNumberExists(context.Background(), "081234567890").Return(false, nil)
				err := validator.IsValidPhoneNumber("081234567890")
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("Phone numbers must start with \"+62\""))
			})
		})

		ginkgo.Context("when the phone number has fewer than 10 digits", func() {
			ginkgo.It("should return an error", func() {
				mockRepo.EXPECT().IsPhoneNumberExists(context.Background(), "+6281234").Return(false, nil)
				err := validator.IsValidPhoneNumber("+6281234")
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("Phone numbers must start with \"+62\" and have 10 to 13 digits"))
			})
		})

		ginkgo.Context("when the phone number has more than 13 digits", func() {
			ginkgo.It("should return an error", func() {
				mockRepo.EXPECT().IsPhoneNumberExists(context.Background(), "+62812345678901234").Return(false, nil)
				err := validator.IsValidPhoneNumber("+62812345678901234")
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("Phone numbers must start with \"+62\" and have 10 to 13 digits"))
			})
		})

		ginkgo.Context("when the phone number already exists in the repository", func() {
			ginkgo.It("should return an error", func() {
				mockRepo.EXPECT().IsPhoneNumberExists(context.Background(), "+6281234567890").Return(true, nil)
				err := validator.IsValidPhoneNumber("+6281234567890")
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("Phone numbers already exists"))
			})
		})
	})

	ginkgo.Describe("IsValidFullName", func() {
		ginkgo.Context("when the full name is valid", func() {
			ginkgo.It("should return nil error", func() {
				err := validator.IsValidFullName("John Doe")
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when the full name has fewer than 3 characters", func() {
			ginkgo.It("should return an error", func() {
				err := validator.IsValidFullName("Jo")
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("Full name must be at minimum 3 characters"))
			})
		})

		ginkgo.Context("when the full name has more than 60 characters", func() {
			ginkgo.It("should return an error", func() {
				err := validator.IsValidFullName(strings.Repeat("a", 61))
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("Full name must be at minimum 3 characters and maximum 60 characters."))
			})
		})
	})

	ginkgo.Describe("IsValidPassword", func() {
		errMessage := "Passwords must have at least 6 characters, including 1 capital letter, 1 number, and 1 special character."
		ginkgo.Context("when the password is valid", func() {
			ginkgo.It("should return nil error", func() {
				err := validator.IsValidPassword("StrongP@ssw0rd")
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when the password is too short", func() {
			ginkgo.It("should return an error", func() {
				err := validator.IsValidPassword("Sh0rt")
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring(errMessage))
			})
		})

		ginkgo.Context("when the password does not contain an uppercase letter", func() {
			ginkgo.It("should return an error", func() {
				err := validator.IsValidPassword("weakpassword1@")
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring(errMessage))
			})
		})

		ginkgo.Context("when the password does not contain a digit", func() {
			ginkgo.It("should return an error", func() {
				err := validator.IsValidPassword("WeakPassword@")
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring(errMessage))
			})
		})

		ginkgo.Context("when the password does not contain a special character", func() {
			ginkgo.It("should return an error", func() {
				err := validator.IsValidPassword("WeakPassword1")
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring(errMessage))
			})
		})
	})
})
