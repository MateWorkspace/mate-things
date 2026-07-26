package presentationhttphandlerprofile

import (
	"net/http"

	domainusecasesprofile "github.com/ABA-Developer/nusapala-things/backend/internal/domain/usecases/profile"
	presentationhttprequest "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/http/request"
	presentationhttpresponse "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/http/response"
	presentationhttputils "github.com/ABA-Developer/nusapala-things/backend/internal/presentation/http/utils"
	"github.com/labstack/echo/v5"
)

type handler struct {
	meUseCase       domainusecasesprofile.Me
	accountUseCase  domainusecasesprofile.Account
	securityUseCase domainusecasesprofile.Security
}

func NewHandler(
	meUseCase domainusecasesprofile.Me,
	accountUseCase domainusecasesprofile.Account,
	securityUseCase domainusecasesprofile.Security,
) *handler {
	return &handler{
		meUseCase:       meUseCase,
		accountUseCase:  accountUseCase,
		securityUseCase: securityUseCase,
	}
}

// ProfileGet godoc
//
// @Summary Profile Get
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Success 200
// @Router /v1/profile [get]
func (h *handler) ProfileGet(c *echo.Context) error {
	userId, err := presentationhttputils.RequiredActorId(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	user, err := h.meUseCase.GetProfile(c.Request().Context(), domainusecasesprofile.GetProfileRequest{UserId: userId})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	if user == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("user"))
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.User(*user))
}

// ProfilePermissionsGet godoc
//
// @Summary Profile Permissions Get
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Success 200
// @Router /v1/profile/permissions [get]
func (h *handler) ProfilePermissionsGet(c *echo.Context) error {
	userId, err := presentationhttputils.RequiredActorId(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	permissions, err := h.meUseCase.GetPermissions(c.Request().Context(), domainusecasesprofile.GetProfilePermissionsRequest{UserId: userId})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.Permissions(permissions))
}

// ProfilePatch godoc
//
// @Summary Profile
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body presentationhttprequest.ProfilePatchRequest true "request"
// @Success 204
// @Router /v1/profile [patch]
func (h *handler) ProfilePatch(c *echo.Context) error {
	userId, err := presentationhttputils.RequiredActorId(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	var req presentationhttprequest.ProfilePatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	if err := h.accountUseCase.UpdateProfile(c.Request().Context(), domainusecasesprofile.UpdateProfileRequest{
		UserId:    userId,
		Name:      req.Name,
		Bio:       req.Bio,
		Username:  req.Username,
		UpdatedBy: &userId,
	}); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// ProfilePasswordPatch godoc
//
// @Summary Profile Password
// @Tags Profile
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body presentationhttprequest.ProfilePasswordPatchRequest true "request"
// @Success 204
// @Router /v1/profile/password [patch]
func (h *handler) ProfilePasswordPatch(c *echo.Context) error {
	userId, err := presentationhttputils.RequiredActorId(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	var req presentationhttprequest.ProfilePasswordPatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	currentPassword, err := presentationhttputils.RequiredString(req.CurrentPassword, "current_password")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	newPassword, err := presentationhttputils.RequiredString(req.NewPassword, "new_password")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.securityUseCase.ChangePassword(c.Request().Context(), domainusecasesprofile.ChangePasswordRequest{
		UserId:          userId,
		CurrentPassword: currentPassword,
		NewPassword:     newPassword,
		UpdatedBy:       &userId,
	}); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
