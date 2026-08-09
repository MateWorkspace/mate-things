package presentationhttphandlerinfrared

import (
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	presentationhttputils "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/utils"
	"github.com/labstack/echo/v5"
)

const broadcastPermission = "infrared_record_session:get"

// InfraredRecordSessionBroadcastRegister godoc
//
// @Summary Infrared Record Session Broadcast Register
// @Description Upgrades to a websocket connection and streams infrared
// record session status events as they occur, as JSON matching
// InfraredRecordSessionEvent. The browser WebSocket API can't set an
// Authorization header, so the access token travels as a query parameter
// instead of the usual Bearer header.
// @Tags Infrared
// @Param token query string true "access token"
// @Param id path string true "session id"
// @Success 101
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Router /v1/infrared/record-sessions/{id}/broadcast [get]
func (h *handler) InfraredRecordSessionBroadcastRegister(c *echo.Context) error {
	req := c.Request()

	accessToken := req.URL.Query().Get("token")
	if accessToken == "" {
		return presentationhttputils.Error(c, domainmodels.NewError("authorization is required", domainmodels.ErrTypeUnauthorized, nil))
	}

	claims, err := h.token.ValidateAccess(accessToken)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	if !hasPermission(claims, broadcastPermission) {
		return presentationhttputils.Error(c, domainmodels.NewError("you do not have permission to perform this action", domainmodels.ErrTypeForbidden, nil))
	}

	sessionId, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	// Past this point the upgrader owns the response — any failure can no
	// longer be reported through presentationhttputils.Error.
	return h.broadcaster.Register(req.Context(), c.Response(), req, sessionId)
}

func hasPermission(claims *domainmodels.TokenClaimsAccess, required string) bool {
	for _, permission := range claims.Permissions {
		if permission == required {
			return true
		}
	}
	return false
}
