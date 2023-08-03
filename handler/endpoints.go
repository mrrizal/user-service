package handler

import (
	"net/http"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/labstack/echo/v4"
)

func (s *Server) Register(ctx echo.Context) error {
	var regRequest generated.RegistrationRequest
	if err := ctx.Bind(&regRequest); err != nil {
		errResp := generated.RegistrationErrResponse{Message: &[]string{"Bad Request"}}
		return ctx.JSON(http.StatusBadRequest, errResp)
	}

	userID, errs := s.Service.Register(ctx.Request().Context(), &regRequest)
	if len(errs) > 0 {
		return ctx.JSON(http.StatusBadRequest, generated.RegistrationErrResponse{
			Message: &errs,
		})
	}

	return ctx.JSON(http.StatusCreated, map[string]string{"user_id": userID})
}

func (s *Server) Login(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, "")
}
