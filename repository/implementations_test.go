package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/utils"
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

	ginkgo.Describe("Loging", func() {
		utils := utils.NewUtils()
		ginkgo.It("should return the user ID on successful login", func() {
			password := "password"
			salt := utils.GenerateRandomSalt()
			hashedPassword, _ := utils.HashingPassword(password, salt)
			mock.ExpectBegin()
			// Expect query for finding the user.
			mock.ExpectQuery("SELECT id FROM public.user").
				WithArgs("some_phone_number").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("expected_user_id"))

			// Expect query for finding the password and salt.
			mock.ExpectQuery("SELECT password, salt FROM public.password").
				WithArgs("expected_user_id").
				WillReturnRows(sqlmock.NewRows([]string{"password", "salt"}).AddRow(hashedPassword, salt))

			// Expect the update query for successful login.
			mock.ExpectExec("UPDATE public.login").
				WithArgs("expected_user_id").
				WillReturnResult(sqlmock.NewResult(1, 1))

			mock.ExpectCommit()

			// Perform the Login function.
			userID, err := repo.Login(context.Background(), generated.LoginRequest{
				PhoneNumber: "some_phone_number",
				Password:    password,
			})

			// Perform your assertions using gomega.Expect().
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(userID).To(gomega.Equal("expected_user_id"))

			// Ensure all expectations were met.
			gomega.Expect(mock.ExpectationsWereMet()).To(gomega.BeNil())
		})

		ginkgo.It("should return error on failed to commit", func() {
			password := "password"
			salt := utils.GenerateRandomSalt()
			hashedPassword, _ := utils.HashingPassword(password, salt)
			mock.ExpectBegin()
			// Expect query for finding the user.
			mock.ExpectQuery("SELECT id FROM public.user").
				WithArgs("some_phone_number").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("expected_user_id"))

			// Expect query for finding the password and salt.
			mock.ExpectQuery("SELECT password, salt FROM public.password").
				WithArgs("expected_user_id").
				WillReturnRows(sqlmock.NewRows([]string{"password", "salt"}).AddRow(hashedPassword, salt))

			// Expect the update query for successful login.
			mock.ExpectExec("UPDATE public.login").
				WithArgs("expected_user_id").
				WillReturnResult(sqlmock.NewResult(1, 1))

			mock.ExpectCommit().WillReturnError(errors.New("failed to commit"))

			// Perform the Login function.
			userID, err := repo.Login(context.Background(), generated.LoginRequest{
				PhoneNumber: "some_phone_number",
				Password:    password,
			})

			// Perform your assertions using gomega.Expect().
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(userID).To(gomega.Equal(""))

			// Ensure all expectations were met.
			gomega.Expect(mock.ExpectationsWereMet()).To(gomega.BeNil())
		})

		ginkgo.It("should return an error when user not found", func() {
			// Expect query for finding the user, but no rows returned (user not found).
			mock.ExpectBegin()

			mock.ExpectQuery("SELECT id FROM public.user").
				WithArgs("non_existent_phone_number").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(""))

			// Perform the Login function.
			userID, err := repo.Login(context.Background(), generated.LoginRequest{
				PhoneNumber: "non_existent_phone_number",
				Password:    "some_password",
			})

			// Perform your assertions using gomega.Expect().
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.Equal("User not found."))
			gomega.Expect(userID).To(gomega.BeEmpty())

			// Ensure all expectations were met.
			gomega.Expect(mock.ExpectationsWereMet()).To(gomega.BeNil())
		})

		ginkgo.It("should return an error on Begin failure", func() {
			mock.ExpectBegin().WillReturnError(errors.New("Begin failed"))

			// Perform the Login function.
			userID, err := repo.Login(context.Background(), generated.LoginRequest{
				PhoneNumber: "some_phone_number",
				Password:    "some_password",
			})

			// Perform your assertions using gomega.Expect().
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.Equal("Begin failed"))
			gomega.Expect(userID).To(gomega.BeEmpty())

			// Ensure all expectations were met.
			gomega.Expect(mock.ExpectationsWereMet()).To(gomega.BeNil())
		})

		ginkgo.It("should return an error on finding user ID failure", func() {
			mock.ExpectBegin()

			mock.ExpectQuery("SELECT id FROM public.user").
				WithArgs("some_phone_number").
				WillReturnError(errors.New("Finding user ID failed"))

			// Perform the Login function.
			userID, err := repo.Login(context.Background(), generated.LoginRequest{
				PhoneNumber: "some_phone_number",
				Password:    "some_password",
			})

			// Perform your assertions using gomega.Expect().
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.Equal("Finding user ID failed"))
			gomega.Expect(userID).To(gomega.BeEmpty())

			// Ensure all expectations were met.
			gomega.Expect(mock.ExpectationsWereMet()).To(gomega.BeNil())
		})

		ginkgo.It("should return an error on finding password and salt failure", func() {
			mock.ExpectBegin()

			mock.ExpectQuery("SELECT id FROM public.user").
				WithArgs("some_phone_number").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("expected_user_id"))

			mock.ExpectQuery("SELECT password, salt FROM public.password").
				WithArgs("expected_user_id").
				WillReturnError(errors.New("Finding password and salt failed"))

			// Perform the Login function.
			userID, err := repo.Login(context.Background(), generated.LoginRequest{
				PhoneNumber: "some_phone_number",
				Password:    "some_password",
			})

			// Perform your assertions using gomega.Expect().
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.Equal("Finding password and salt failed"))
			gomega.Expect(userID).To(gomega.BeEmpty())

			// Ensure all expectations were met.
			gomega.Expect(mock.ExpectationsWereMet()).To(gomega.BeNil())
		})

		ginkgo.It("should return an error on wrong passoword", func() {
			mock.ExpectBegin()

			mock.ExpectQuery("SELECT id FROM public.user").
				WithArgs("some_phone_number").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("expected_user_id"))

			mock.ExpectQuery("SELECT password, salt FROM public.password").
				WithArgs("expected_user_id").
				WillReturnRows(sqlmock.NewRows([]string{"password", "salt"}).AddRow("hashed_password", "some_salt"))

			// Perform the Login function.
			userID, err := repo.Login(context.Background(), generated.LoginRequest{
				PhoneNumber: "some_phone_number",
				Password:    "some_password",
			})

			// Perform your assertions using gomega.Expect().
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.Equal("Wrong password"))
			gomega.Expect(userID).To(gomega.BeEmpty())

			// Ensure all expectations were met.
			gomega.Expect(mock.ExpectationsWereMet()).To(gomega.BeNil())
		})

		ginkgo.It("should return an error on update query failure", func() {
			password := "password"
			salt := utils.GenerateRandomSalt()
			hashedPassword, _ := utils.HashingPassword(password, salt)
			mock.ExpectBegin()

			mock.ExpectQuery("SELECT id FROM public.user").
				WithArgs("some_phone_number").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("expected_user_id"))

			mock.ExpectQuery("SELECT password, salt FROM public.password").
				WithArgs("expected_user_id").
				WillReturnRows(sqlmock.NewRows([]string{"password", "salt"}).AddRow(hashedPassword, salt))

			mock.ExpectExec("UPDATE public.login").
				WithArgs("expected_user_id").
				WillReturnError(errors.New("Update query failed"))

			// Perform the Login function.
			userID, err := repo.Login(context.Background(), generated.LoginRequest{
				PhoneNumber: "some_phone_number",
				Password:    password,
			})

			// Perform your assertions using gomega.Expect().
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.Equal("Update query failed"))
			gomega.Expect(userID).To(gomega.BeEmpty())

			// Ensure all expectations were met.
			gomega.Expect(mock.ExpectationsWereMet()).To(gomega.BeNil())
		})
	})

	ginkgo.Context("GetUserProfile", func() {
		userID := "some_user_id"
		fullName := "some_full_name"
		phoneNumber := "123456789"

		ginkgo.It("get user profile success", func() {
			mock.ExpectQuery("SELECT full_name, phone_number FROM public.user WHERE id = \\$1").
				WithArgs(userID).
				WillReturnRows(sqlmock.NewRows([]string{"full_name", "phone_number"}).
					AddRow(fullName, phoneNumber))

			profile, err := repo.GetUserProfile(context.Background(), userID)
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(profile).To(gomega.Equal(
				generated.UserProfile{FullName: &fullName, PhoneNumber: &phoneNumber}))
		})

		ginkgo.It("get user profile error query", func() {
			mock.ExpectQuery("SELECT full_name, phone_number FROM public.user WHERE id = \\$1").
				WithArgs(userID).WillReturnError(errors.New("error"))

			profile, err := repo.GetUserProfile(context.Background(), userID)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(profile).To(gomega.Equal(generated.UserProfile{}))
		})
	})

	ginkgo.Context("UpdateUserProfile", func() {
		userID := "some_user_id"
		fullName := "some user"
		phoneNumber := "123456789"
		userProfile := generated.UserProfile{
			FullName:    &fullName,
			PhoneNumber: &phoneNumber,
		}
		ginkgo.It("should return the user profile when the update request is empty", func() {
			// Set up mock database query expectations for GetUserProfile
			rows := sqlmock.NewRows([]string{"full_name", "phone_number"}).
				AddRow(userProfile.FullName, userProfile.PhoneNumber)
			mock.ExpectQuery("SELECT full_name, phone_number FROM public.user WHERE id = \\$1").
				WithArgs(userID).
				WillReturnRows(rows)

			// Call the function with an empty update request
			result, err := repo.UpdateUserProfile(ctx, map[string]string{}, userID)
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(result).To(gomega.Equal(userProfile))
		})

		ginkgo.It("should update the user profile and return the updated profile", func() {
			// Set up mock database query expectations for UPDATE statement
			mock.ExpectExec("UPDATE public.user SET full_name = 'New Name', phone_number = '9876543210' WHERE id = \\$1").
				WithArgs(userID).
				WillReturnResult(sqlmock.NewResult(1, 1))

			// Set up mock database query expectations for GetUserProfile
			rows := sqlmock.NewRows([]string{"full_name", "phone_number"}).
				AddRow("New Name", "9876543210")
			mock.ExpectQuery("SELECT full_name, phone_number FROM public.user WHERE id = \\$1").
				WithArgs(userID).
				WillReturnRows(rows)

			// Call the function with an update request
			updateReq := map[string]string{
				"full_name":    "New Name",
				"phone_number": "9876543210",
			}

			result, err := repo.UpdateUserProfile(ctx, updateReq, userID)
			gomega.Expect(err).To(gomega.BeNil())
			gomega.Expect(*result.FullName).To(gomega.Equal(updateReq["full_name"]))
			gomega.Expect(*result.PhoneNumber).To(gomega.Equal(updateReq["phone_number"]))
		})

		ginkgo.It("should return an error when the UPDATE statement fails", func() {
			// Set up mock database query expectations for UPDATE statement error
			mock.ExpectExec("UPDATE public.user SET full_name = 'New Name', phone_number = '9876543210' WHERE id = \\$1").
				WithArgs(userID).
				WillReturnError(sql.ErrConnDone)

			// Call the function with an update request
			updateReq := map[string]string{
				"full_name":    "New Name",
				"phone_number": "9876543210",
			}
			result, err := repo.UpdateUserProfile(ctx, updateReq, userID)
			gomega.Expect(err).To(gomega.Not(gomega.BeNil()))
			gomega.Expect(result).To(gomega.Equal(generated.UserProfile{}))
		})
	})
})
