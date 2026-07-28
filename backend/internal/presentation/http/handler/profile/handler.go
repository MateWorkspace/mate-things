package presentationhttphandlerprofile

import (
	"net/http"

	domainusecasesprofile "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/profile"
	presentationhttprequest "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/utils"
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
// @Success 200 {object} presentationhttpresponse.UserResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/profile [get]
func (h *handler) ProfileGet(c *echo.Context) error {
	userId, err := presentationhttputils.RequiredActorId(c)
	if err != nil {
		return presentationhttputils.Error(c, err, "Please sign in to continue.")
	}

	user, err := h.meUseCase.GetProfile(c.Request().Context(), domainusecasesprofile.GetProfileRequest{UserId: userId})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to load your profile right now. Please try again.")
	}
	if user == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("user"), "Unable to load your profile right now. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.User(*user))
}

// ProfilePermissionsGet godoc
//
// @Summary Profile Permissions Get
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Success 200 {array} presentationhttpresponse.PermissionResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/profile/permissions [get]
func (h *handler) ProfilePermissionsGet(c *echo.Context) error {
	userId, err := presentationhttputils.RequiredActorId(c)
	if err != nil {
		return presentationhttputils.Error(c, err, "Please sign in to continue.")
	}

	permissions, err := h.meUseCase.GetPermissions(c.Request().Context(), domainusecasesprofile.GetProfilePermissionsRequest{UserId: userId})
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to load your permissions right now. Please try again.")
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
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 409 {object} presentationhttpresponse.ErrorResponse "Already Exists"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/profile [patch]
func (h *handler) ProfilePatch(c *echo.Context) error {
	userId, err := presentationhttputils.RequiredActorId(c)
	if err != nil {
		return presentationhttputils.Error(c, err, "Please sign in to continue.")
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
		return presentationhttputils.Error(c, err, "Unable to update your profile. Please check your input and try again.")
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
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/profile/password [patch]
func (h *handler) ProfilePasswordPatch(c *echo.Context) error {
	userId, err := presentationhttputils.RequiredActorId(c)
	if err != nil {
		return presentationhttputils.Error(c, err, "Please sign in to continue.")
	}

	var req presentationhttprequest.ProfilePasswordPatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	currentPassword, err := presentationhttputils.RequiredString(req.CurrentPassword, "current_password")
	if err != nil {
		return presentationhttputils.Error(c, err, "Please enter your current password.")
	}

	if err := h.securityUseCase.ChangePassword(c.Request().Context(), domainusecasesprofile.ChangePasswordRequest{
		UserId:          userId,
		CurrentPassword: currentPassword,
		NewPassword:     req.NewPassword,
		UpdatedBy:       &userId,
	}); err != nil {
		return presentationhttputils.Error(c, err, "Unable to change your password. Please check your current password and try again.")
	}

	return c.NoContent(http.StatusNoContent)
}
