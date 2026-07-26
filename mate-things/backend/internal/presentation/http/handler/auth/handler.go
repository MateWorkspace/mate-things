package presentationhttphandlerauth

import (
	"net/http"

	domainusecasesauth "github.com/ABA-Developer/nusapala-things/backend/internal/domain/usecases/auth"
	presentationhttprequest "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/http/request"
	presentationhttpresponse "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/http/response"
	presentationhttputils "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/http/utils"
	"github.com/labstack/echo/v5"
)

type handler struct {
	sessionUseCase domainusecasesauth.Session
}

func NewHandler(sessionUseCase domainusecasesauth.Session) *handler {
	return &handler{sessionUseCase: sessionUseCase}
}

// AuthLoginPost godoc
//
// @Summary Auth Login
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body presentationhttprequest.AuthLoginRequest true "request"
// @Success 201
// @Router /v1/auth/login [post]
func (h *handler) AuthLoginPost(c *echo.Context) error {
	var req presentationhttprequest.AuthLoginRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	username, err := presentationhttputils.RequiredString(req.Username, "username")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	password, err := presentationhttputils.RequiredString(req.Password, "password")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	result, err := h.sessionUseCase.Login(c.Request().Context(), domainusecasesauth.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.LoginResponse{
		User:         presentationhttpresponse.User(result.User),
		Role:         presentationhttpresponse.Role(result.Role),
		Permissions:  presentationhttpresponse.Permissions(result.Permissions),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})
}

// AuthRefreshPost godoc
//
// @Summary Auth Refresh
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body presentationhttprequest.AuthRefreshRequest true "request"
// @Success 201
// @Router /v1/auth/refresh [post]
func (h *handler) AuthRefreshPost(c *echo.Context) error {
	var req presentationhttprequest.AuthRefreshRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	refreshToken, err := presentationhttputils.RequiredString(req.RefreshToken, "refresh_token")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	result, err := h.sessionUseCase.Refresh(c.Request().Context(), domainusecasesauth.RefreshRequest{
		RefreshToken: refreshToken,
	})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.LoginResponse{
		User:         presentationhttpresponse.User(result.User),
		Role:         presentationhttpresponse.Role(result.Role),
		Permissions:  presentationhttpresponse.Permissions(result.Permissions),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})
}
