package presentationhttphandlertelemetry

import (
	"net/http"

	domainusecasestelemetry "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/telemetry"
	presentationhttpresponse "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/utils"
	"github.com/labstack/echo/v5"
)

type handler struct {
	queryUseCase domainusecasestelemetry.Query
}

func NewHandler(queryUseCase domainusecasestelemetry.Query) *handler {
	return &handler{queryUseCase: queryUseCase}
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
// @Router /v1/telemetry-records [get]
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
// @Router /v1/telemetry-records [delete]
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
