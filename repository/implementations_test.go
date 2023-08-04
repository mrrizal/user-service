package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/SawitProRecruitment/UserService/generated"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestRepository(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Repository Suite")
}

var _ = ginkgo.Describe("Repository", func() {
	var (
		db   *sql.DB
		repo *Repository
		mock sqlmock.Sqlmock
		ctx  context.Context
	)

	ginkgo.BeforeEach(func() {
		var err error
		db, mock, err = sqlmock.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		repo = &Repository{Db: db}

		ctx = context.Background()
	})

	ginkgo.Describe("IsPhoneNumberExists", func() {
		ginkgo.Context("when phone number exists", func() {
			ginkgo.It("should return true", func() {
				phoneNumber := "1234567890"
				rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery("^SELECT count\\(phone_number\\) FROM public.user WHERE phone_number = \\$1$").
					WithArgs(phoneNumber).
					WillReturnRows(rows)

				exists, err := repo.IsPhoneNumberExists(ctx, phoneNumber)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				gomega.Expect(exists).To(gomega.BeTrue())
			})
		})

		ginkgo.Context("when phone number does not exist", func() {
			ginkgo.It("should return false", func() {
				phoneNumber := "9876543210"
				rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery("^SELECT count\\(phone_number\\) FROM public.user WHERE phone_number = \\$1$").
					WithArgs(phoneNumber).
					WillReturnRows(rows)

				exists, err := repo.IsPhoneNumberExists(ctx, phoneNumber)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				gomega.Expect(exists).To(gomega.BeFalse())
			})
		})

		ginkgo.Context("when an error occurs", func() {
			ginkgo.It("should return the error", func() {
				phoneNumber := "invalid_phone_number"
				mock.ExpectQuery("^SELECT count\\(phone_number\\) FROM public.user WHERE phone_number = \\$1$").
					WithArgs(phoneNumber).
					WillReturnError(sql.ErrConnDone)

				exists, err := repo.IsPhoneNumberExists(ctx, phoneNumber)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(exists).To(gomega.BeFalse())
			})
		})
	})

	ginkgo.Describe("Register", func() {
		regRequest := generated.RegistrationRequest{
			FullName:    "John Doe",
			PhoneNumber: "1234567890",
			Password:    "securePassword",
		}
		salt := "randomSalt"
		userID := "generatedUserID"

		ginkgo.It("should register a new user", func() {

			// Expect the first query for inserting user data and returning the user ID
			rows := sqlmock.NewRows([]string{"id"}).AddRow(userID)
			mock.ExpectBegin()
			mock.ExpectQuery("^INSERT INTO public.user \\(full_name, phone_number\\) VALUES \\(\\$1, \\$2\\) RETURNING id$").
				WithArgs(regRequest.FullName, regRequest.PhoneNumber).
				WillReturnRows(rows)

			mock.ExpectExec("^INSERT INTO public.login \\(user_id, success_login\\) VALUES \\(\\$1, \\$2\\)").
				WithArgs(userID, 0).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// Expect the second query for inserting password data
			mock.ExpectExec("^INSERT INTO public.password \\(user_id, password, salt\\) VALUES \\(\\$1, \\$2, \\$3\\)$").
				WithArgs(userID, regRequest.Password, salt).
				WillReturnResult(sqlmock.NewResult(1, 1))

			mock.ExpectCommit()

			createdUserID, err := repo.Register(ctx, regRequest, salt)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(createdUserID).To(gomega.Equal(userID))
		})

		ginkgo.It("should handle transaction rollback on error", func() {
			mock.ExpectBegin()
			mock.ExpectQuery("^INSERT INTO public.user \\(full_name, phone_number\\) VALUES \\(\\$1, \\$2\\) RETURNING id$").
				WithArgs(regRequest.FullName, regRequest.PhoneNumber).
				WillReturnError(sql.ErrConnDone)

			mock.ExpectRollback()

			_, err := repo.Register(ctx, regRequest, salt)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})

		ginkgo.Context("when Begin returns an error", func() {
			ginkgo.It("should return the error", func() {
				mock.ExpectBegin().WillReturnError(errors.New("begin error"))

				_, err := repo.Register(ctx, regRequest, salt)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("begin error"))
			})
		})

		ginkgo.Context("when Exec returns an error", func() {
			ginkgo.It("should return the error and rollback the transaction", func() {
				mock.ExpectBegin()

				mock.ExpectQuery("^INSERT INTO public.user \\(full_name, phone_number\\) VALUES \\(\\$1, \\$2\\) RETURNING id$").
					WithArgs(regRequest.FullName, regRequest.PhoneNumber).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("generatedUserID"))

				mock.ExpectExec("^INSERT INTO public.login \\(user_id, success_login\\) VALUES \\(\\$1, \\$2\\)").
					WithArgs(userID, 0).
					WillReturnError(errors.New("exec error"))

				mock.ExpectExec("^INSERT INTO public.password \\(user_id, password, salt\\) VALUES \\(\\$1, \\$2, \\$3\\)$").
					WithArgs("generatedUserID", regRequest.Password, salt).
					WillReturnError(errors.New("exec error"))

				mock.ExpectRollback()

				_, err := repo.Register(ctx, regRequest, salt)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("exec error"))
			})

			ginkgo.It("should return the error and rollback the transaction", func() {
				mock.ExpectBegin()

				mock.ExpectQuery("^INSERT INTO public.user \\(full_name, phone_number\\) VALUES \\(\\$1, \\$2\\) RETURNING id$").
					WithArgs(regRequest.FullName, regRequest.PhoneNumber).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("generatedUserID"))

				mock.ExpectExec("^INSERT INTO public.login \\(user_id, success_login\\) VALUES \\(\\$1, \\$2\\)").
					WithArgs(userID, 0).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec("^INSERT INTO public.password \\(user_id, password, salt\\) VALUES \\(\\$1, \\$2, \\$3\\)$").
					WithArgs("generatedUserID", regRequest.Password, salt).
					WillReturnError(errors.New("exec error"))

				mock.ExpectRollback()

				_, err := repo.Register(ctx, regRequest, salt)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("exec error"))
			})
		})

		ginkgo.Context("when Commit returns an error", func() {
			ginkgo.It("should return the error and rollback the transaction", func() {
				mock.ExpectBegin()

				mock.ExpectQuery("^INSERT INTO public.user \\(full_name, phone_number\\) VALUES \\(\\$1, \\$2\\) RETURNING id$").
					WithArgs(regRequest.FullName, regRequest.PhoneNumber).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("generatedUserID"))

				mock.ExpectExec("^INSERT INTO public.login \\(user_id, success_login\\) VALUES \\(\\$1, \\$2\\)").
					WithArgs(userID, 0).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec("^INSERT INTO public.password \\(user_id, password, salt\\) VALUES \\(\\$1, \\$2, \\$3\\)$").
					WithArgs("generatedUserID", regRequest.Password, salt).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit().WillReturnError(errors.New("commit error"))

				mock.ExpectRollback()

				_, err := repo.Register(ctx, regRequest, salt)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("commit error"))
			})
		})
	})
})
