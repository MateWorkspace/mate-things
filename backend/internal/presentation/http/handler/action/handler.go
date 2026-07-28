package presentationhttphandleraction

import (
	"net/http"
	"time"

	domainusecasesaction "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/action"
	presentationhttprequest "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/utils"
	"github.com/labstack/echo/v5"
)

type handler struct {
	definitionUseCase domainusecasesaction.Definition
	executionUseCase  domainusecasesaction.Execution
	historyUseCase    domainusecasesaction.History
}

func NewHandler(
	definitionUseCase domainusecasesaction.Definition,
	executionUseCase domainusecasesaction.Execution,
	historyUseCase domainusecasesaction.History,
) *handler {
	return &handler{
		definitionUseCase: definitionUseCase,
		executionUseCase:  executionUseCase,
		historyUseCase:    historyUseCase,
	}
}

// ActionPost godoc
//
// @Summary Action
// @Tags Actions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body presentationhttprequest.ActionPostRequest true "request"
// @Success 201
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 409 {object} presentationhttpresponse.ErrorResponse "Already Exists"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/actions [post]
func (h *handler) ActionPost(c *echo.Context) error {
	var req presentationhttprequest.ActionPostRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}

	nodeClassId, err := presentationhttputils.RequiredUUID(req.NodeClassId, "node_class_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	id, err := h.definitionUseCase.Create(c.Request().Context(), domainusecasesaction.CreateActionRequest{
		NodeClassId:          nodeClassId,
		Name:                 req.Name,
		Description:          req.Description,
		PayloadSchemaName:    req.PayloadSchemaName,
		PayloadSchemaVersion: req.PayloadSchemaVersion,
		CreatedBy:            presentationhttputils.ActorId(c),
	})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusCreated, presentationhttpresponse.IdResponse{Id: id.String()})
}

// ActionGetList godoc
//
// @Summary Action List
// @Tags Actions
// @Produce json
// @Security BearerAuth
// @Success 200
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/actions [get]
func (h *handler) ActionGetList(c *echo.Context) error {
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	nodeClassId, err := presentationhttputils.QueryUUID(c, "node_class_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	payloadSchemaVersion, err := presentationhttputils.QueryInt32(c, "payload_schema_version")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	actions, total, err := h.definitionUseCase.ReadByPagination(c.Request().Context(), domainusecasesaction.ReadActionsByPaginationRequest{
		Page:                 page.Page,
		Limit:                page.Limit,
		Search:               page.Search,
		NodeClassId:          nodeClassId,
		PayloadSchemaName:    presentationhttputils.QueryString(c, "payload_schema_name"),
		PayloadSchemaVersion: payloadSchemaVersion,
	})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.ActionResponse]{
		Data: presentationhttpresponse.Actions(actions),
		Page: presentationhttputils.PageResponse(page, total),
	})
}

// ActionGetByName godoc
//
// @Summary Action Get By Name
// @Tags Actions
// @Produce json
// @Security BearerAuth
// @Param name path string true "name"
// @Success 200
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/actions/by-name/{name} [get]
func (h *handler) ActionGetByName(c *echo.Context) error {
	name, err := presentationhttputils.RequiredString(c.Param("name"), "name")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	action, err := h.definitionUseCase.ReadByName(c.Request().Context(), domainusecasesaction.ReadActionByNameRequest{Name: name})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	if action == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("action"))
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.Action(*action))
}

// ActionGetById godoc
//
// @Summary Action Get By ID
// @Tags Actions
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 200
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/actions/{id} [get]
func (h *handler) ActionGetById(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	action, err := h.definitionUseCase.ReadById(c.Request().Context(), domainusecasesaction.ReadActionByIdRequest{Id: id})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	if action == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("action"))
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.Action(*action))
}

// ActionPatch godoc
//
// @Summary Action
// @Tags Actions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Param request body presentationhttprequest.ActionPatchRequest true "request"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 409 {object} presentationhttpresponse.ErrorResponse "Already Exists"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/actions/{id} [patch]
func (h *handler) ActionPatch(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	var req presentationhttprequest.ActionPatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	nodeClassId, err := presentationhttputils.OptionalUUID(req.NodeClassId, "node_class_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.definitionUseCase.UpdateById(c.Request().Context(), domainusecasesaction.UpdateActionRequest{
		Id:                   id,
		NodeClassId:          nodeClassId,
		Name:                 req.Name,
		Description:          req.Description,
		PayloadSchemaName:    req.PayloadSchemaName,
		PayloadSchemaVersion: req.PayloadSchemaVersion,
		UpdatedBy:            presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// ActionDelete godoc
//
// @Summary Action Delete
// @Tags Actions
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Success 204
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/actions/{id} [delete]
func (h *handler) ActionDelete(c *echo.Context) error {
	id, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	if err := h.definitionUseCase.DeleteById(c.Request().Context(), domainusecasesaction.DeleteActionRequest{
		Id:        id,
		DeletedBy: presentationhttputils.ActorId(c),
	}); err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

// ActionDispatchPost godoc
//
// @Summary Action Dispatch
// @Tags Actions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "id"
// @Param request body presentationhttprequest.ActionDispatchRequest true "request"
// @Success 201
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 404 {object} presentationhttpresponse.ErrorResponse "Not Found"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Failure 504 {object} presentationhttpresponse.ErrorResponse "Request Timeout"
// @Router /v1/actions/{id}/dispatch [post]
func (h *handler) ActionDispatchPost(c *echo.Context) error {
	actionId, err := presentationhttputils.RequiredUUID(c.Param("id"), "id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	var req presentationhttprequest.ActionDispatchRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		return err
	}
	nodeId, err := presentationhttputils.RequiredUUID(req.NodeId, "node_id")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	payload, err := presentationhttputils.RequiredRawJSON(req.Payload, "payload")
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	executedAt := time.Now().UTC()
	if req.ExecutedAt != nil {
		executedAt = req.ExecutedAt.UTC()
	}

	actionLog, err := h.executionUseCase.Dispatch(c.Request().Context(), domainusecasesaction.DispatchActionRequest{
		ActionId:   actionId,
		NodeId:     nodeId,
		Payload:    payload,
		ExecutedAt: executedAt,
		ActorId:    presentationhttputils.ActorId(c),
	})
	if err != nil {
		return presentationhttputils.Error(c, err)
	}
	if actionLog == nil {
		return presentationhttputils.Error(c, presentationhttputils.MissingResponse("action log"))
	}

	return c.JSON(http.StatusCreated, presentationhttpresponse.ActionLog(*actionLog))
}

// ActionLogGetList godoc
//
// @Summary Action Log List
// @Tags Action Logs
// @Produce json
// @Security BearerAuth
// @Success 200
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/action-logs [get]
func (h *handler) ActionLogGetList(c *echo.Context) error {
	filter, err := h.actionLogFilter(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	actionLogs, total, err := h.historyUseCase.ReadByFilter(c.Request().Context(), filter)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.CountDataResponse[presentationhttpresponse.ActionLogResponse]{
		Data:       presentationhttpresponse.ActionLogs(actionLogs),
		TotalItems: total,
	})
}

// ActionLogDelete godoc
//
// @Summary Action Log Delete
// @Tags Action Logs
// @Produce json
// @Security BearerAuth
// @Success 200
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/action-logs [delete]
func (h *handler) ActionLogDelete(c *echo.Context) error {
	filter, err := h.actionLogDeleteFilter(c)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	count, err := h.historyUseCase.DeleteByFilter(c.Request().Context(), filter)
	if err != nil {
		return presentationhttputils.Error(c, err)
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.CountResponse{Count: count})
}

func (h *handler) actionLogFilter(c *echo.Context) (domainusecasesaction.ReadActionLogsByFilterRequest, error) {
	executedAtStart, err := presentationhttputils.QueryTime(c, "executed_at_start")
	if err != nil {
		return domainusecasesaction.ReadActionLogsByFilterRequest{}, err
	}
	executedAtEnd, err := presentationhttputils.QueryTime(c, "executed_at_end")
	if err != nil {
		return domainusecasesaction.ReadActionLogsByFilterRequest{}, err
	}
	actionId, err := presentationhttputils.QueryUUID(c, "action_id")
	if err != nil {
		return domainusecasesaction.ReadActionLogsByFilterRequest{}, err
	}
	nodeId, err := presentationhttputils.QueryUUID(c, "node_id")
	if err != nil {
		return domainusecasesaction.ReadActionLogsByFilterRequest{}, err
	}

	return domainusecasesaction.ReadActionLogsByFilterRequest{
		ExecutedAtStart: executedAtStart,
		ExecutedAtEnd:   executedAtEnd,
		ActionId:        actionId,
		NodeId:          nodeId,
	}, nil
}

func (h *handler) actionLogDeleteFilter(c *echo.Context) (domainusecasesaction.DeleteActionLogsByFilterRequest, error) {
	filter, err := h.actionLogFilter(c)
	if err != nil {
		return domainusecasesaction.DeleteActionLogsByFilterRequest{}, err
	}

	return domainusecasesaction.DeleteActionLogsByFilterRequest{
		ExecutedAtStart: filter.ExecutedAtStart,
		ExecutedAtEnd:   filter.ExecutedAtEnd,
		ActionId:        filter.ActionId,
		NodeId:          filter.NodeId,
	}, nil
}
