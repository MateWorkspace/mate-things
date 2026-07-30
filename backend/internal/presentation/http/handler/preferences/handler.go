package presentationhttphandlerpreferences

import (
	"net/http"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasespreferences "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/preferences"
	presentationhttprequest "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/utils"
	"github.com/labstack/echo/v5"
)

// This handler only ever returns 204/error, but @Failure annotations below
// need the type in scope for swag to resolve it.
var _ = presentationhttpresponse.ErrorResponse{}

type handler struct {
	updateUseCase domainusecasespreferences.Update
}

func NewHandler(updateUseCase domainusecasespreferences.Update) *handler {
	return &handler{updateUseCase: updateUseCase}
}

// PreferencesPatch godoc
//
// @Summary Preferences
// @Tags Preferences
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param resource path string true "resource"
// @Param id path string true "id"
// @Param request body presentationhttprequest.PreferencesPatchRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/preferences/{resource}/{id} [patch]
func (h *handler) PreferencesPatch(c *echo.Context) error {
	resource, err := presentationhttputils.RequiredString(c.Param("resource"), "resource")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	var req presentationhttprequest.PreferencesPatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	preferences, err := presentationhttputils.RequiredRawJSON(req.Preferences, "preferences")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	updatedBy := presentationhttputils.ActorId(c)
	switch resource {
	case "action":
		err = h.updateUseCase.Action(c.Request().Context(), domainusecasespreferences.UpdateActionPreferencesRequest{Id: id, Preferences: preferences, UpdatedBy: updatedBy})
	case "firmware":
		err = h.updateUseCase.Firmware(c.Request().Context(), domainusecasespreferences.UpdateFirmwarePreferencesRequest{Id: id, Preferences: preferences, UpdatedBy: updatedBy})
	case "node":
		err = h.updateUseCase.Node(c.Request().Context(), domainusecasespreferences.UpdateNodePreferencesRequest{Id: id, Preferences: preferences, UpdatedBy: updatedBy})
	case "node_class":
		err = h.updateUseCase.NodeClass(c.Request().Context(), domainusecasespreferences.UpdateNodeClassPreferencesRequest{Id: id, Preferences: preferences, UpdatedBy: updatedBy})
	case "payload_schema":
		err = h.updateUseCase.PayloadSchema(c.Request().Context(), domainusecasespreferences.UpdatePayloadSchemaPreferencesRequest{Id: id, Preferences: preferences, UpdatedBy: updatedBy})
	case "permission":
		err = h.updateUseCase.Permission(c.Request().Context(), domainusecasespreferences.UpdatePermissionPreferencesRequest{Id: id, Preferences: preferences, UpdatedBy: updatedBy})
	case "role":
		err = h.updateUseCase.Role(c.Request().Context(), domainusecasespreferences.UpdateRolePreferencesRequest{Id: id, Preferences: preferences, UpdatedBy: updatedBy})
	case "user":
		err = h.updateUseCase.User(c.Request().Context(), domainusecasespreferences.UpdateUserPreferencesRequest{Id: id, Preferences: preferences, UpdatedBy: updatedBy})
	default:
		err = domainmodels.NewError("resource is not supported", domainmodels.ErrTypeValidation, nil)
	}
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
