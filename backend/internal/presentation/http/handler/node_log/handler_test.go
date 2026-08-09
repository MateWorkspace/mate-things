package presentationhttphandlernodelog

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnodelog "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node_log"
	"github.com/labstack/echo/v5"
)

func TestNodeLogGetListParsesFiltersDelegatesAndReturnsCountedData(t *testing.T) {
	loggedAt := time.Date(2026, time.July, 30, 14, 3, 12, 45_000_000, time.UTC)
	createdAt := loggedAt.Add(55 * time.Millisecond)
	query := &nodeLogQueryFake{
		readLogs: []domainmodels.NodeLog{{
			Id:           918273,
			NodeDeviceId: "AC276E5E030C",
			Level:        domainmodels.NodeLogLevelInfo,
			Tag:          "wifi_manager",
			Message:      "WiFi connected, IP: 192.168.1.42",
			LoggedAt:     loggedAt,
			CreatedAt:    createdAt,
		}},
		readTotal: 23,
	}
	handler := NewHandler(query)
	recorder, ctx := nodeLogContext(
		http.MethodGet,
		"/api/v1/node-logs?logged_at_start=2026-07-30T14%3A00%3A00Z&logged_at_end=2026-07-30T15%3A00%3A00.123Z&node_device_id=%20AC276E5E030C%20&level=INFO",
	)

	if err := handler.NodeLogGetList(ctx); err != nil {
		t.Fatalf("NodeLogGetList() error = %v", err)
	}

	if query.readCalls != 1 {
		t.Fatalf("ReadByFilter() calls = %d, want 1", query.readCalls)
	}
	assertReadFilter(t, query.readRequest, "2026-07-30T14:00:00Z", "2026-07-30T15:00:00.123Z", "AC276E5E030C", domainmodels.NodeLogLevelInfo)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var response nodeLogListResponse
	decodeNodeLogJSON(t, recorder, &response)
	if response.TotalItems != 23 {
		t.Errorf("total_items = %d, want 23", response.TotalItems)
	}
	if len(response.Data) != 1 {
		t.Fatalf("data length = %d, want 1", len(response.Data))
	}
	want := nodeLogResponse{
		Id:           918273,
		NodeDeviceId: "AC276E5E030C",
		Level:        "INFO",
		Tag:          "wifi_manager",
		Message:      "WiFi connected, IP: 192.168.1.42",
		LoggedAt:     loggedAt,
		CreatedAt:    createdAt,
	}
	if response.Data[0] != want {
		t.Errorf("data[0] = %#v, want %#v", response.Data[0], want)
	}
}

func TestNodeLogDeleteParsesFiltersDelegatesAndReturnsCount(t *testing.T) {
	query := &nodeLogQueryFake{deleteCount: 7}
	handler := NewHandler(query)
	recorder, ctx := nodeLogContext(
		http.MethodDelete,
		"/api/v1/node-logs?logged_at_start=2026-07-29T00%3A00%3A00Z&logged_at_end=2026-07-30T00%3A00%3A00Z&node_device_id=node-42&level=WARN",
	)

	if err := handler.NodeLogDelete(ctx); err != nil {
		t.Fatalf("NodeLogDelete() error = %v", err)
	}

	if query.deleteCalls != 1 {
		t.Fatalf("DeleteByFilter() calls = %d, want 1", query.deleteCalls)
	}
	assertDeleteFilter(t, query.deleteRequest, "2026-07-29T00:00:00Z", "2026-07-30T00:00:00Z", "node-42", domainmodels.NodeLogLevelWarn)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var response struct {
		Count int `json:"count"`
	}
	decodeNodeLogJSON(t, recorder, &response)
	if response.Count != 7 {
		t.Errorf("count = %d, want 7", response.Count)
	}
}

func TestNodeLogHandlersAcceptEverySupportedLevel(t *testing.T) {
	levels := []domainmodels.NodeLogLevel{
		domainmodels.NodeLogLevelNone,
		domainmodels.NodeLogLevelError,
		domainmodels.NodeLogLevelWarn,
		domainmodels.NodeLogLevelInfo,
		domainmodels.NodeLogLevelDebug,
	}

	for _, level := range levels {
		t.Run(string(level), func(t *testing.T) {
			query := &nodeLogQueryFake{}
			handler := NewHandler(query)
			recorder, ctx := nodeLogContext(http.MethodGet, "/api/v1/node-logs?level="+string(level))

			if err := handler.NodeLogGetList(ctx); err != nil {
				t.Fatalf("NodeLogGetList() error = %v", err)
			}

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
			}
			if query.readCalls != 1 {
				t.Fatalf("ReadByFilter() calls = %d, want 1", query.readCalls)
			}
			if query.readRequest.Level == nil || *query.readRequest.Level != level {
				t.Fatalf("level = %v, want %q", query.readRequest.Level, level)
			}
		})
	}
}

func TestNodeLogHandlersRejectInvalidTimeFiltersBeforeDelegation(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		target      string
		invoke      func(*handler, *echo.Context) error
		wantMessage string
	}{
		{
			name:        "GET invalid start",
			method:      http.MethodGet,
			target:      "/api/v1/node-logs?logged_at_start=yesterday",
			invoke:      (*handler).NodeLogGetList,
			wantMessage: "logged_at_start must be a valid RFC3339 timestamp",
		},
		{
			name:        "DELETE invalid end",
			method:      http.MethodDelete,
			target:      "/api/v1/node-logs?logged_at_end=2026-07-30",
			invoke:      (*handler).NodeLogDelete,
			wantMessage: "logged_at_end must be a valid RFC3339 timestamp",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := &nodeLogQueryFake{}
			handler := NewHandler(query)
			recorder, ctx := nodeLogContext(test.method, test.target)

			if err := test.invoke(handler, ctx); err != nil {
				t.Fatalf("handler error = %v", err)
			}

			response := assertInvalidFilterResponse(t, recorder)
			if response.Message != test.wantMessage {
				t.Errorf("message = %q, want %q", response.Message, test.wantMessage)
			}
			if query.readCalls != 0 || query.deleteCalls != 0 {
				t.Fatalf("use case calls = read %d, delete %d; want none", query.readCalls, query.deleteCalls)
			}
		})
	}
}

func TestNodeLogHandlersRejectUnsupportedLevelsBeforeDelegation(t *testing.T) {
	tests := []struct {
		name   string
		method string
		level  string
		invoke func(*handler, *echo.Context) error
	}{
		{
			name:   "GET unknown level",
			method: http.MethodGet,
			level:  "TRACE",
			invoke: (*handler).NodeLogGetList,
		},
		{
			name:   "DELETE lowercase level",
			method: http.MethodDelete,
			level:  "info",
			invoke: (*handler).NodeLogDelete,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			query := &nodeLogQueryFake{}
			handler := NewHandler(query)
			recorder, ctx := nodeLogContext(test.method, "/api/v1/node-logs?level="+test.level)

			if err := test.invoke(handler, ctx); err != nil {
				t.Fatalf("handler error = %v", err)
			}

			response := assertInvalidFilterResponse(t, recorder)
			if query.readCalls != 0 || query.deleteCalls != 0 {
				t.Fatalf("use case calls = read %d, delete %d; want none", query.readCalls, query.deleteCalls)
			}
			if response.Message != "level must be one of NONE, ERROR, WARN, INFO, DEBUG" {
				t.Errorf("message = %q, want supported-level validation message", response.Message)
			}
		})
	}
}

func TestNodeLogHandlersRenderUseCaseErrors(t *testing.T) {
	tests := []struct {
		name   string
		method string
		query  *nodeLogQueryFake
		invoke func(*handler, *echo.Context) error
	}{
		{
			name:   "GET read failure",
			method: http.MethodGet,
			query:  &nodeLogQueryFake{readErr: errors.New("database unavailable")},
			invoke: (*handler).NodeLogGetList,
		},
		{
			name:   "DELETE failure",
			method: http.MethodDelete,
			query:  &nodeLogQueryFake{deleteErr: errors.New("database unavailable")},
			invoke: (*handler).NodeLogDelete,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := NewHandler(test.query)
			recorder, ctx := nodeLogContext(test.method, "/api/v1/node-logs")

			if err := test.invoke(handler, ctx); err != nil {
				t.Fatalf("handler error = %v", err)
			}

			// The use case's raw error ("database unavailable") is a plain
			// error, not a *domainmodels.Error, so presentationhttputils.Error
			// falls through to the generic 500 mapping — it never echoes
			// internal error text or a per-call-site message to the client,
			// by design (see error.go's errorMappings).
			if recorder.Code != http.StatusInternalServerError {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
			}
			var response nodeLogErrorResponse
			decodeNodeLogJSON(t, recorder, &response)
			if response.Error != "Internal Server Error" {
				t.Errorf("error = %q, want %q", response.Error, "Internal Server Error")
			}
			wantMessage := "Something went wrong on our end. Please try again later."
			if response.Message != wantMessage {
				t.Errorf("message = %q, want %q", response.Message, wantMessage)
			}
		})
	}
}

type nodeLogQueryFake struct {
	readLogs      []domainmodels.NodeLog
	readTotal     int
	readErr       error
	readCalls     int
	readRequest   domainusecasesnodelog.ReadNodeLogByFilterRequest
	deleteCount   int
	deleteErr     error
	deleteCalls   int
	deleteRequest domainusecasesnodelog.DeleteNodeLogByFilterRequest
}

type nodeLogListResponse struct {
	Data       []nodeLogResponse `json:"data"`
	TotalItems int               `json:"total_items"`
}

type nodeLogResponse struct {
	Id           int64     `json:"id"`
	NodeDeviceId string    `json:"node_device_id"`
	Level        string    `json:"level"`
	Tag          string    `json:"tag"`
	Message      string    `json:"message"`
	LoggedAt     time.Time `json:"logged_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type nodeLogErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (f *nodeLogQueryFake) ReadByFilter(_ context.Context, request domainusecasesnodelog.ReadNodeLogByFilterRequest) ([]domainmodels.NodeLog, int, error) {
	f.readCalls++
	f.readRequest = request
	return f.readLogs, f.readTotal, f.readErr
}

func (f *nodeLogQueryFake) DeleteByFilter(_ context.Context, request domainusecasesnodelog.DeleteNodeLogByFilterRequest) (int, error) {
	f.deleteCalls++
	f.deleteRequest = request
	return f.deleteCount, f.deleteErr
}

func nodeLogContext(method, target string) (*httptest.ResponseRecorder, *echo.Context) {
	recorder := httptest.NewRecorder()
	ctx := echo.New().NewContext(httptest.NewRequest(method, target, nil), recorder)
	return recorder, ctx
}

func decodeNodeLogJSON(t *testing.T, recorder *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.NewDecoder(recorder.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v; body = %q", err, recorder.Body.String())
	}
}

func assertReadFilter(
	t *testing.T,
	got domainusecasesnodelog.ReadNodeLogByFilterRequest,
	start, end, nodeDeviceID string,
	level domainmodels.NodeLogLevel,
) {
	t.Helper()
	assertTimePointer(t, "LoggedAtStart", got.LoggedAtStart, start)
	assertTimePointer(t, "LoggedAtEnd", got.LoggedAtEnd, end)
	assertStringPointer(t, "NodeDeviceId", got.NodeDeviceId, nodeDeviceID)
	assertLevelPointer(t, got.Level, level)
}

func assertDeleteFilter(
	t *testing.T,
	got domainusecasesnodelog.DeleteNodeLogByFilterRequest,
	start, end, nodeDeviceID string,
	level domainmodels.NodeLogLevel,
) {
	t.Helper()
	assertTimePointer(t, "LoggedAtStart", got.LoggedAtStart, start)
	assertTimePointer(t, "LoggedAtEnd", got.LoggedAtEnd, end)
	assertStringPointer(t, "NodeDeviceId", got.NodeDeviceId, nodeDeviceID)
	assertLevelPointer(t, got.Level, level)
}

func assertTimePointer(t *testing.T, name string, got *time.Time, want string) {
	t.Helper()
	if got == nil || got.Format(time.RFC3339Nano) != want {
		t.Errorf("%s = %v, want %s", name, got, want)
	}
}

func assertStringPointer(t *testing.T, name string, got *string, want string) {
	t.Helper()
	if got == nil || *got != want {
		t.Errorf("%s = %v, want %q", name, got, want)
	}
}

func assertLevelPointer(t *testing.T, got *domainmodels.NodeLogLevel, want domainmodels.NodeLogLevel) {
	t.Helper()
	if got == nil || *got != want {
		t.Errorf("Level = %v, want %q", got, want)
	}
}

func assertInvalidFilterResponse(t *testing.T, recorder *httptest.ResponseRecorder) nodeLogErrorResponse {
	t.Helper()
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	var response nodeLogErrorResponse
	decodeNodeLogJSON(t, recorder, &response)
	if response.Error != "Invalid Format" {
		t.Errorf("error = %q, want %q", response.Error, "Invalid Format")
	}
	return response
}
