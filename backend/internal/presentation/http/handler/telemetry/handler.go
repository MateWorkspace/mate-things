package presentationhttphandlertelemetry

import (
	"net/http"

	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasestelemetry "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/telemetry"
	presentationhttpresponse "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/utils"
	"github.com/labstack/echo/v5"
)

const broadcastPermission = "telemetry_record:get"

type handler struct {
	queryUseCase     domainusecasestelemetry.Query
	broadcastUseCase domainusecasestelemetry.Broadcast
	token            domaincontractsutility.Token
}

func NewHandler(
	queryUseCase domainusecasestelemetry.Query,
	broadcastUseCase domainusecasestelemetry.Broadcast,
	token domaincontractsutility.Token,
) *handler {
	return &handler{
		queryUseCase:     queryUseCase,
		broadcastUseCase: broadcastUseCase,
		token:            token,
	}
}

// TelemetryRecordGetList godoc
//
// @Summary Telemetry Record List
// @Tags Telemetry
// @Produce json
// @Security BearerAuth
// @Param recorded_at_start query string false "RFC3339 timestamp"
// @Param recorded_at_end query string false "RFC3339 timestamp"
// @Param node_device_id query string false "node device id"
// @Param metric_name query string false "metric name"
// @Param payload_schema_name query string false "payload schema name"
// @Param payload_schema_version query int false "payload schema version"
// @Success 200 {object} presentationhttpresponse.CountDataResponse[presentationhttpresponse.TelemetryRecordResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/telemetry [get]
func (h *handler) TelemetryRecordGetList(c *echo.Context) error {
	filter, err := h.filter(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	records, total, err := h.queryUseCase.ReadByFilter(c.Request().Context(), filter)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.CountDataResponse[presentationhttpresponse.TelemetryRecordResponse]{
		Data:       presentationhttpresponse.TelemetryRecords(records),
		TotalItems: total,
	})
}

// TelemetryRecordGetLatest godoc
//
// @Summary Telemetry Record Get Latest
// @Tags Telemetry
// @Produce json
// @Security BearerAuth
// @Param node_device_id query string false "node device id"
// @Param metric_name query string false "metric name"
// @Success 200 {object} presentationhttpresponse.TelemetryRecordResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/telemetry/latest [get]
func (h *handler) TelemetryRecordGetLatest(c *echo.Context) error {
	request := domainusecasestelemetry.ReadLatestTelemetryRequest{
		NodeDeviceId: presentationhttputils.QueryString(c, "node_device_id"),
		MetricName:   presentationhttputils.QueryString(c, "metric_name"),
	}

	telemetryRecord, err := h.queryUseCase.ReadLatest(c.Request().Context(), request)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	if telemetryRecord == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("telemetry record"))
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.TelemetryRecord(*telemetryRecord))
}

// TelemetryRecordDelete godoc
//
// @Summary Telemetry Record Delete
// @Tags Telemetry
// @Produce json
// @Security BearerAuth
// @Param recorded_at_start query string false "RFC3339 timestamp"
// @Param recorded_at_end query string false "RFC3339 timestamp"
// @Param node_device_id query string false "node device id"
// @Param metric_name query string false "metric name"
// @Param payload_schema_name query string false "payload schema name"
// @Param payload_schema_version query int false "payload schema version"
// @Success 200 {object} presentationhttpresponse.CountResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/telemetry [delete]
func (h *handler) TelemetryRecordDelete(c *echo.Context) error {
	filter, err := h.deleteFilter(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	count, err := h.queryUseCase.DeleteByFilter(c.Request().Context(), filter)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.CountResponse{Count: count})
}

// TelemetryBroadcastRegister godoc
//
// @Summary Telemetry Broadcast Register
// @Description Upgrades to a websocket connection and streams every telemetry
// record as it is ingested, optionally filtered to node_device_id/metric_name,
// as JSON matching TelemetryRecordResponse. The browser WebSocket API can't
// set an Authorization header, so the access token travels as a query
// parameter instead of the usual Bearer header.
// @Tags Telemetry
// @Param token query string true "access token"
// @Param node_device_id query string false "node device id"
// @Param metric_name query string false "metric name"
// @Success 101
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Router /v1/telemetry/broadcast [get]
func (h *handler) TelemetryBroadcastRegister(c *echo.Context) error {
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

	nodeDeviceId := presentationhttputils.QueryString(c, "node_device_id")
	metricName := presentationhttputils.QueryString(c, "metric_name")

	// Past this point the upgrader owns the response — any failure can no
	// longer be reported through presentationhttputils.Error.
	return h.broadcastUseCase.Register(req.Context(), c.Response(), req, claims.UserId, nodeDeviceId, metricName)
}

// TelemetryBroadcastSessionList godoc
//
// @Summary Telemetry Broadcast Session List
// @Tags Telemetry
// @Produce json
// @Security BearerAuth
// @Success 200 {object} presentationhttpresponse.CountDataResponse[presentationhttpresponse.BroadcastSessionTelemetryResponse]
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/telemetry/broadcast/sessions [get]
func (h *handler) TelemetryBroadcastSessionList(c *echo.Context) error {
	sessions, err := h.broadcastUseCase.SessionList(c.Request().Context())
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.CountDataResponse[presentationhttpresponse.BroadcastSessionTelemetryResponse]{
		Data:       presentationhttpresponse.BroadcastSessionTelemetries(sessions),
		TotalItems: len(sessions),
	})
}

func hasPermission(claims *domainmodels.TokenClaimsAccess, required string) bool {
	for _, permission := range claims.Permissions {
		if permission == required {
			return true
		}
	}
	return false
}

func (h *handler) filter(c *echo.Context) (domainusecasestelemetry.ReadTelemetryByFilterRequest, error) {
	recordedAtStart, err := presentationhttputils.QueryTime(c, "recorded_at_start")
	if err != nil {
		return domainusecasestelemetry.ReadTelemetryByFilterRequest{}, err
	}
	recordedAtEnd, err := presentationhttputils.QueryTime(c, "recorded_at_end")
	if err != nil {
		return domainusecasestelemetry.ReadTelemetryByFilterRequest{}, err
	}
	payloadSchemaVersion, err := presentationhttputils.QueryInt32(c, "payload_schema_version")
	if err != nil {
		return domainusecasestelemetry.ReadTelemetryByFilterRequest{}, err
	}

	return domainusecasestelemetry.ReadTelemetryByFilterRequest{
		RecordedAtStart:      recordedAtStart,
		RecordedAtEnd:        recordedAtEnd,
		NodeDeviceId:         presentationhttputils.QueryString(c, "node_device_id"),
		MetricName:           presentationhttputils.QueryString(c, "metric_name"),
		PayloadSchemaName:    presentationhttputils.QueryString(c, "payload_schema_name"),
		PayloadSchemaVersion: payloadSchemaVersion,
	}, nil
}

func (h *handler) deleteFilter(c *echo.Context) (domainusecasestelemetry.DeleteTelemetryByFilterRequest, error) {
	filter, err := h.filter(c)
	if err != nil {
		return domainusecasestelemetry.DeleteTelemetryByFilterRequest{}, err
	}

	return domainusecasestelemetry.DeleteTelemetryByFilterRequest{
		RecordedAtStart:      filter.RecordedAtStart,
		RecordedAtEnd:        filter.RecordedAtEnd,
		NodeDeviceId:         filter.NodeDeviceId,
		MetricName:           filter.MetricName,
		PayloadSchemaName:    filter.PayloadSchemaName,
		PayloadSchemaVersion: filter.PayloadSchemaVersion,
	}, nil
}
