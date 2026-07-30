package presentationhttphandlernodelog

import (
	"net/http"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnodelog "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node_log"
	presentationhttpresponse "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/mate-things/backend/internal/presentation/http/utils"
	"github.com/labstack/echo/v5"
)

type handler struct {
	queryUseCase domainusecasesnodelog.Query
}

func NewHandler(queryUseCase domainusecasesnodelog.Query) *handler {
	return &handler{queryUseCase: queryUseCase}
}

// NodeLogGetList godoc
//
// @Summary Node Log List
// @Tags Node Log
// @Produce json
// @Security BearerAuth
// @Success 200 {object} presentationhttpresponse.CountDataResponse[presentationhttpresponse.NodeLogResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/node-logs [get]
func (h *handler) NodeLogGetList(c *echo.Context) error {
	filter, err := h.filter(c)
	if err != nil {
		return presentationhttputils.Error(c, err, "One or more of the filters provided is invalid.")
	}

	logs, total, err := h.queryUseCase.ReadByFilter(c.Request().Context(), filter)
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to load node logs right now. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.CountDataResponse[presentationhttpresponse.NodeLogResponse]{
		Data:       presentationhttpresponse.NodeLogs(logs),
		TotalItems: total,
	})
}

// NodeLogDelete godoc
//
// @Summary Node Log Delete
// @Tags Node Log
// @Produce json
// @Security BearerAuth
// @Success 200 {object} presentationhttpresponse.CountResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse "Invalid Format"
// @Failure 401 {object} presentationhttpresponse.ErrorResponse "Unauthorized"
// @Failure 403 {object} presentationhttpresponse.ErrorResponse "Access Denied"
// @Failure 500 {object} presentationhttpresponse.ErrorResponse "Internal Server Error"
// @Router /v1/node-logs [delete]
func (h *handler) NodeLogDelete(c *echo.Context) error {
	filter, err := h.deleteFilter(c)
	if err != nil {
		return presentationhttputils.Error(c, err, "One or more of the filters provided is invalid.")
	}

	count, err := h.queryUseCase.DeleteByFilter(c.Request().Context(), filter)
	if err != nil {
		return presentationhttputils.Error(c, err, "Unable to delete node logs right now. Please try again.")
	}

	return c.JSON(http.StatusOK, presentationhttpresponse.CountResponse{Count: count})
}

func (h *handler) filter(c *echo.Context) (domainusecasesnodelog.ReadNodeLogByFilterRequest, error) {
	loggedAtStart, err := presentationhttputils.QueryTime(c, "logged_at_start")
	if err != nil {
		return domainusecasesnodelog.ReadNodeLogByFilterRequest{}, err
	}
	loggedAtEnd, err := presentationhttputils.QueryTime(c, "logged_at_end")
	if err != nil {
		return domainusecasesnodelog.ReadNodeLogByFilterRequest{}, err
	}
	level, err := nodeLogLevel(presentationhttputils.QueryString(c, "level"))
	if err != nil {
		return domainusecasesnodelog.ReadNodeLogByFilterRequest{}, err
	}

	return domainusecasesnodelog.ReadNodeLogByFilterRequest{
		LoggedAtStart: loggedAtStart,
		LoggedAtEnd:   loggedAtEnd,
		NodeDeviceId:  presentationhttputils.QueryString(c, "node_device_id"),
		Level:         level,
	}, nil
}

func (h *handler) deleteFilter(c *echo.Context) (domainusecasesnodelog.DeleteNodeLogByFilterRequest, error) {
	filter, err := h.filter(c)
	if err != nil {
		return domainusecasesnodelog.DeleteNodeLogByFilterRequest{}, err
	}

	return domainusecasesnodelog.DeleteNodeLogByFilterRequest{
		LoggedAtStart: filter.LoggedAtStart,
		LoggedAtEnd:   filter.LoggedAtEnd,
		NodeDeviceId:  filter.NodeDeviceId,
		Level:         filter.Level,
	}, nil
}

func nodeLogLevel(value *string) (*domainmodels.NodeLogLevel, error) {
	if value == nil {
		return nil, nil
	}

	level := domainmodels.NodeLogLevel(*value)
	switch level {
	case domainmodels.NodeLogLevelNone,
		domainmodels.NodeLogLevelError,
		domainmodels.NodeLogLevelWarn,
		domainmodels.NodeLogLevelInfo,
		domainmodels.NodeLogLevelDebug:
		return &level, nil
	default:
		return nil, domainmodels.NewError(
			"level must be one of NONE, ERROR, WARN, INFO, DEBUG",
			domainmodels.ErrTypeValidation,
			nil,
		)
	}
}
