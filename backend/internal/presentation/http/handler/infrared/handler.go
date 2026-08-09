package presentationhttphandlerinfrared

import (
	"net/http"

	domaincontractsbroadcaster "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/broadcaster"
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainusecasesinfrared "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/infrared"
	presentationhttprequest "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/utils"
	"github.com/labstack/echo/v5"
)

type handler struct {
	recordSessionUseCase domainusecasesinfrared.RecordSessionManagement
	broadcaster          domaincontractsbroadcaster.InfraredRecordSession
	token                domaincontractsutility.Token
}

func NewHandler(
	recordSessionUseCase domainusecasesinfrared.RecordSessionManagement,
	broadcaster domaincontractsbroadcaster.InfraredRecordSession,
	token domaincontractsutility.Token,
) *handler {
	return &handler{recordSessionUseCase: recordSessionUseCase, broadcaster: broadcaster, token: token}
}

// RecordSessionPost godoc
//
// @Summary Infrared Record Session
// @Tags Infrared
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body presentationhttprequest.StartRecordSessionRequest true "request"
// @Success 201 {object} presentationhttpresponse.IdResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/record-sessions [post]
func (h *handler) RecordSessionPost(c *echo.Context) error {
	var req presentationhttprequest.StartRecordSessionRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	nodeId, err := presentationhttputils.RequiredUUID(req.NodeId, "node_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	infraredDeviceTypeId, err := presentationhttputils.RequiredUUID(req.InfraredDeviceTypeId, "infrared_device_type_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	definitions := make([]domainusecasesinfrared.StartRecordSessionDefinition, len(req.Definitions))
	for i, definition := range req.Definitions {
		infraredStateId, err := presentationhttputils.RequiredUUID(definition.InfraredStateId, "infrared_state_id")
		if err != nil {
			return presentationhttputils.Error(c, err)
		}
		definitions[i] = domainusecasesinfrared.StartRecordSessionDefinition{
			InfraredStateId: infraredStateId,
			Options:         definition.Options,
			Minimum:         definition.Minimum,
			Maximum:         definition.Maximum,
			Step:            definition.Step,
		}
	}

	id, err := h.recordSessionUseCase.Start(c.Request().Context(), domainusecasesinfrared.StartRecordSessionRequest{
		NodeId:               nodeId,
		InfraredDeviceTypeId: infraredDeviceTypeId,
		Brand:                req.Brand,
		Model:                req.Model,
		Definitions:          definitions,
	})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusCreated, presentationhttpresponse.IdResponse{Id: id.String()})
}

// RecordSessionGetById godoc
//
// @Summary Infrared Record Session Get By ID
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} presentationhttpresponse.InfraredRecordSessionResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/record-sessions/{id} [get]
func (h *handler) RecordSessionGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	session, err := h.recordSessionUseCase.GetById(c.Request().Context(), id)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	if session == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("record session"))
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.InfraredRecordSession(*session))
}

// RecordSessionCasesGetList godoc
//
// @Summary Infrared Record Session Cases List
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {array} presentationhttpresponse.InfraredStateDeviceRecordCaseResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/record-sessions/{id}/cases [get]
func (h *handler) RecordSessionCasesGetList(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	cases, err := h.recordSessionUseCase.ListCases(c.Request().Context(), id)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.InfraredStateDeviceRecordCases(cases))
}

// RecordCaseRawAccept godoc
//
// @Summary Infrared Record Case Raw Accept
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param caseId path string true "case id"
// @Param rawId path string true "raw id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/record-cases/{caseId}/raw/{rawId}/accept [post]
func (h *handler) RecordCaseRawAccept(c *echo.Context) error {
	rawId, err := presentationhttputils.RequiredUUID(c.Param("rawId"), "rawId")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.recordSessionUseCase.AcceptRaw(c.Request().Context(), rawId); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// RecordCaseRawDiscard godoc
//
// @Summary Infrared Record Case Raw Discard
// @Tags Infrared
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param caseId path string true "case id"
// @Param rawId path string true "raw id"
// @Param request body presentationhttprequest.DiscardRawRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/record-cases/{caseId}/raw/{rawId}/discard [post]
func (h *handler) RecordCaseRawDiscard(c *echo.Context) error {
	rawId, err := presentationhttputils.RequiredUUID(c.Param("rawId"), "rawId")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	var req presentationhttprequest.DiscardRawRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	if err := h.recordSessionUseCase.DiscardRaw(c.Request().Context(), rawId, req.Reason); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// RecordCaseRetry godoc
//
// @Summary Infrared Record Case Retry
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param caseId path string true "case id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/record-cases/{caseId}/retry [post]
func (h *handler) RecordCaseRetry(c *echo.Context) error {
	caseId, err := presentationhttputils.RequiredUUID(c.Param("caseId"), "caseId")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.recordSessionUseCase.RetryCase(c.Request().Context(), caseId); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// RecordSessionCaseCurrentPost godoc
//
// @Summary Infrared Record Session Set Current Case
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "session id"
// @Param caseId path string true "case id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/record-sessions/{id}/cases/{caseId}/current [post]
func (h *handler) RecordSessionCaseCurrentPost(c *echo.Context) error {
	sessionId, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	caseId, err := presentationhttputils.RequiredUUID(c.Param("caseId"), "caseId")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.recordSessionUseCase.SetCurrentCase(c.Request().Context(), sessionId, caseId); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
