package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/labstack/echo/v4"
	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

type mockService struct {
	RegisterFunc func(ctx context.Context, regRequest *generated.RegistrationRequest) (string, []string)
}

func NewMockService() mockService {
	return mockService{RegisterFunc: func(ctx context.Context, regRequest *generated.RegistrationRequest) (string, []string) {
		fmt.Println("heheheh")
		return "", []string{}
	}}
}

func (m *mockService) Register(ctx context.Context, regRequest *generated.RegistrationRequest) (string, []string) {
	return m.RegisterFunc(ctx, regRequest)
}

var _ = ginkgo.Describe("endpoints", func() {
	var (
		server *Server
		svc    mockService
		repo   mockRepository
	)

	ginkgo.BeforeEach(func() {
		repo = NewMockRepository()
		svc = NewMockService()

		server = &Server{
			Repository: &repo,
			Service:    &svc,
		}
	})

	ginkgo.Describe("Register", func() {
		ginkgo.Context("when request body is valid", func() {
			ginkgo.It("should return 201 Created with user_id", func() {
				// Prepare the request and response recorder
				svc.RegisterFunc = func(ctx context.Context, regRequest *generated.RegistrationRequest) (string, []string) {
					return "some_user_id", []string{}
				}
				body := `{"username": "testuser", "password": "password123"}`
				req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()

				// Call the function
				err := server.Register(echo.New().NewContext(req, recorder))

				// Assertions
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusCreated))

				expectedResponse := `{"user_id": "some_user_id"}`
				gomega.Expect(recorder.Body.String()).To(gomega.MatchJSON(expectedResponse))
			})
		})

		ginkgo.Context("when request body is invalid", func() {
			ginkgo.It("should return 400 Bad Request with error message", func() {
				svc.RegisterFunc = func(ctx context.Context, regRequest *generated.RegistrationRequest) (string, []string) {
					return "", []string{"Bad Request"}
				}
				// Prepare the request and response recorder
				body := `{"invalid_field": "testuser", "password": "password123"}`
				req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				recorder := httptest.NewRecorder()
				ctx := echo.New().NewContext(req, recorder)

				// Call the function
				err := server.Register(ctx)

				// Assertions
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusBadRequest))

				expectedResponse := `{"message": ["Bad Request"]}`
				gomega.Expect(recorder.Body.String()).To(gomega.MatchJSON(expectedResponse))
			})
		})
	})
})
