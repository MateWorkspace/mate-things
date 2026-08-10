package presentationhttphandlerinfrared

import (
	"net/http"

	domaincontractsbroadcaster "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/broadcaster"
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesinfrared "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/infrared"
	presentationhttprequest "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/utils"
	"github.com/labstack/echo/v5"
)

type handler struct {
	recordSessionUseCase domainusecasesinfrared.RecordSessionManagement
	referenceUseCase     domainusecasesinfrared.ReferenceManagement
	broadcaster          domaincontractsbroadcaster.InfraredRecordSession
	token                domaincontractsutility.Token
}

func NewHandler(
	recordSessionUseCase domainusecasesinfrared.RecordSessionManagement,
	referenceUseCase domainusecasesinfrared.ReferenceManagement,
	broadcaster domaincontractsbroadcaster.InfraredRecordSession,
	token domaincontractsutility.Token,
) *handler {
	return &handler{recordSessionUseCase: recordSessionUseCase, referenceUseCase: referenceUseCase, broadcaster: broadcaster, token: token}
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
		CreatedBy:            presentationhttputils.ActorId(c),
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

// RecordSessionCoderGetById godoc
//
// @Summary Infrared Record Session Coder Get By ID
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {object} presentationhttpresponse.InfraredStateCoderResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/record-sessions/{id}/coder [get]
func (h *handler) RecordSessionCoderGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	coder, err := h.recordSessionUseCase.GetCoderBySessionId(c.Request().Context(), id)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	if coder == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("coder"))
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.InfraredStateCoder(*coder))
}

// RecordSessionTestCasesGetList godoc
//
// @Summary Infrared Record Session Test Cases List
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200 {array} presentationhttpresponse.InfraredTestCaseResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/record-sessions/{id}/test-cases [get]
func (h *handler) RecordSessionTestCasesGetList(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	testCases, err := h.recordSessionUseCase.ListTestCases(c.Request().Context(), id)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.InfraredTestCases(testCases))
}

// TestCaseTransmitPost godoc
//
// @Summary Infrared Test Case Transmit
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/test-cases/{id}/transmit [post]
func (h *handler) TestCaseTransmitPost(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.recordSessionUseCase.TransmitTestCase(c.Request().Context(), id); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// TestCaseResultPost godoc
//
// @Summary Infrared Test Case Result
// @Tags Infrared
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Param request body presentationhttprequest.RecordTestCaseResultRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/test-cases/{id}/result [post]
func (h *handler) TestCaseResultPost(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	var req presentationhttprequest.RecordTestCaseResultRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	if err := h.recordSessionUseCase.RecordTestCaseResult(c.Request().Context(), id, req.Passed); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// RecordSessionDelete godoc
//
// @Summary Infrared Record Session Delete
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/record-sessions/{id} [delete]
func (h *handler) RecordSessionDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.recordSessionUseCase.DeleteRecordSessionById(c.Request().Context(), id, presentationhttputils.ActorId(c)); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// RecordCaseDelete godoc
//
// @Summary Infrared Record Case Delete
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param caseId path string true "case id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/record-cases/{caseId} [delete]
func (h *handler) RecordCaseDelete(c *echo.Context) error {
	caseId, err := presentationhttputils.RequiredUUID(c.Param("caseId"), "caseId")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.recordSessionUseCase.DeleteRecordCaseById(c.Request().Context(), caseId, presentationhttputils.ActorId(c)); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// RecordStateDelete godoc
//
// @Summary Infrared Record State Delete
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/record-states/{id} [delete]
func (h *handler) RecordStateDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.recordSessionUseCase.DeleteRecordStateById(c.Request().Context(), id, presentationhttputils.ActorId(c)); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// RecordRawDelete godoc
//
// @Summary Infrared Record Raw Delete
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param caseId path string true "case id"
// @Param rawId path string true "raw id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/record-cases/{caseId}/raw/{rawId} [delete]
func (h *handler) RecordRawDelete(c *echo.Context) error {
	rawId, err := presentationhttputils.RequiredUUID(c.Param("rawId"), "rawId")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.recordSessionUseCase.DeleteRecordRawById(c.Request().Context(), rawId, presentationhttputils.ActorId(c)); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// StateCoderDelete godoc
//
// @Summary Infrared State Coder Delete
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/state-coders/{id} [delete]
func (h *handler) StateCoderDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.recordSessionUseCase.DeleteStateCoderById(c.Request().Context(), id, presentationhttputils.ActorId(c)); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// TestCaseDelete godoc
//
// @Summary Infrared Test Case Delete
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/test-cases/{id} [delete]
func (h *handler) TestCaseDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.recordSessionUseCase.DeleteTestCaseById(c.Request().Context(), id, presentationhttputils.ActorId(c)); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// TestCaseStateDelete godoc
//
// @Summary Infrared Test Case State Delete
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Param stateId path string true "state id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/test-cases/{id}/states/{stateId} [delete]
func (h *handler) TestCaseStateDelete(c *echo.Context) error {
	stateId, err := presentationhttputils.RequiredUUID(c.Param("stateId"), "stateId")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.recordSessionUseCase.DeleteTestCaseStateById(c.Request().Context(), stateId, presentationhttputils.ActorId(c)); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// DeviceTypePost godoc
//
// @Summary Infrared Device Type Create
// @Tags Infrared
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body presentationhttprequest.CreateDeviceTypeRequest true "request"
// @Success 201 {object} presentationhttpresponse.IdResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/device-types [post]
func (h *handler) DeviceTypePost(c *echo.Context) error {
	var req presentationhttprequest.CreateDeviceTypeRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	id, err := h.referenceUseCase.CreateDeviceType(c.Request().Context(), req.Name, presentationhttputils.ActorId(c))
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusCreated, presentationhttpresponse.IdResponse{Id: id.String()})
}

// DeviceTypeGetList godoc
//
// @Summary Infrared Device Type List
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Success 200 {array} presentationhttpresponse.InfraredDeviceTypeResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/device-types [get]
func (h *handler) DeviceTypeGetList(c *echo.Context) error {
	deviceTypes, err := h.referenceUseCase.ListDeviceTypes(c.Request().Context())
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.InfraredDeviceTypes(deviceTypes))
}

// DeviceTypeDelete godoc
//
// @Summary Infrared Device Type Delete
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/device-types/{id} [delete]
func (h *handler) DeviceTypeDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.referenceUseCase.DeleteDeviceTypeById(c.Request().Context(), id, presentationhttputils.ActorId(c)); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// StatePost godoc
//
// @Summary Infrared State Create
// @Tags Infrared
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param deviceTypeId path string true "device type id"
// @Param request body presentationhttprequest.CreateStateRequest true "request"
// @Success 201 {object} presentationhttpresponse.IdResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/device-types/{deviceTypeId}/states [post]
func (h *handler) StatePost(c *echo.Context) error {
	deviceTypeId, err := presentationhttputils.RequiredUUID(c.Param("deviceTypeId"), "deviceTypeId")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	var req presentationhttprequest.CreateStateRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	id, err := h.referenceUseCase.CreateState(c.Request().Context(), deviceTypeId, req.Name, domainmodels.InfraredStateType(req.Type), presentationhttputils.ActorId(c))
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusCreated, presentationhttpresponse.IdResponse{Id: id.String()})
}

// StateGetList godoc
//
// @Summary Infrared State List
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param deviceTypeId path string true "device type id"
// @Success 200 {array} presentationhttpresponse.InfraredStateResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/device-types/{deviceTypeId}/states [get]
func (h *handler) StateGetList(c *echo.Context) error {
	deviceTypeId, err := presentationhttputils.RequiredUUID(c.Param("deviceTypeId"), "deviceTypeId")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	states, err := h.referenceUseCase.ListStatesByDeviceTypeId(c.Request().Context(), deviceTypeId)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.InfraredStates(states))
}

// StateDelete godoc
//
// @Summary Infrared State Delete
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/states/{id} [delete]
func (h *handler) StateDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.referenceUseCase.DeleteStateById(c.Request().Context(), id, presentationhttputils.ActorId(c)); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// DeviceDelete godoc
//
// @Summary Infrared Device Delete
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/devices/{id} [delete]
func (h *handler) DeviceDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.referenceUseCase.DeleteDeviceById(c.Request().Context(), id, presentationhttputils.ActorId(c)); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// DefinitionGetList godoc
//
// @Summary Infrared State Device Definition List
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "device id"
// @Success 200 {array} presentationhttpresponse.InfraredStateDeviceDefinitionResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/devices/{id}/definitions [get]
func (h *handler) DefinitionGetList(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	definitions, err := h.referenceUseCase.ListDefinitionsByDeviceId(c.Request().Context(), id)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.InfraredStateDeviceDefinitions(definitions))
}

// DefinitionDelete godoc
//
// @Summary Infrared State Device Definition Delete
// @Tags Infrared
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/infrared/state-device-definitions/{id} [delete]
func (h *handler) DefinitionDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.referenceUseCase.DeleteDefinitionById(c.Request().Context(), id, presentationhttputils.ActorId(c)); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
