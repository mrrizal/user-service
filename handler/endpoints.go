package handler

import (
	"net/http"
	"strings"

	"github.com/SawitProRecruitment/UserService/generated"
	"github.com/SawitProRecruitment/UserService/utils"
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
	var loginRequest generated.LoginRequest
	if err := ctx.Bind(&loginRequest); err != nil {
		errResp := generated.ErrorResponse{Message: "Bad Request"}
		return ctx.JSON(http.StatusBadRequest, errResp)
	}

	token, err := s.Service.Login(ctx.Request().Context(), &loginRequest)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: err.Error()})
	}
	expireIn := "24 hours"
	return ctx.JSON(http.StatusOK, generated.LoginResponse{Token: &token, ExpireIn: &expireIn})
}

func (s *Server) GetProfile(ctx echo.Context) error {
	token, err := utils.ExtractJWTToken(ctx)
	if err != nil {
		return ctx.JSON(http.StatusForbidden, generated.ErrorResponse{Message: err.Error()})
	}

	userProfile, err := s.Service.GetUserProfile(ctx.Request().Context(), token)
	if err != nil {
		return ctx.JSON(http.StatusForbidden, generated.ErrorResponse{Message: err.Error()})
	}

	return ctx.JSON(http.StatusOK, userProfile)
}

func (s *Server) UpdateProfile(ctx echo.Context) error {
	token, err := utils.ExtractJWTToken(ctx)
	if err != nil {
		return ctx.JSON(http.StatusForbidden, generated.ErrorResponse{Message: err.Error()})
	}

	var updateUserProfileRequest generated.UpdateUserProfileRequest
	if err := ctx.Bind(&updateUserProfileRequest); err != nil {
		errResp := generated.RegistrationErrResponse{Message: &[]string{"Bad Request"}}
		return ctx.JSON(http.StatusBadRequest, errResp)
	}

	userProfile, err := s.Service.UpdateUserProfile(ctx.Request().Context(), updateUserProfileRequest, token)
	if err != nil {
		status := http.StatusBadRequest
		if strings.Contains(err.Error(), "Phone number already exists") {
			status = http.StatusConflict
		}
		return ctx.JSON(status, generated.ErrorResponse{Message: err.Error()})
	}
	return ctx.JSON(http.StatusOK, userProfile)
}
