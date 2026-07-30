package presentationhttproute

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestRouteNodeLogRegistersConsumerFacingContract(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		wantHandler    string
		wantPermission string
	}{
		{
			name:           "GET list",
			method:         http.MethodGet,
			wantHandler:    "get",
			wantPermission: "node_log:get",
		},
		{
			name:           "DELETE logs",
			method:         http.MethodDelete,
			wantHandler:    "delete",
			wantPermission: "node_log:remove",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			routeNodeLog(e.Group("/api/v1"), nodeLogRouteHandlerFake{}, nodeLogPermissionMiddleware)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(test.method, "/api/v1/node-logs", nil)

			e.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d; body = %q", recorder.Code, http.StatusOK, recorder.Body.String())
			}
			if got := recorder.Header().Get("X-Test-Required-Permissions"); got != test.wantPermission {
				t.Errorf("required permissions = %q, want %q", got, test.wantPermission)
			}
			var response struct {
				Handler string `json:"handler"`
			}
			if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
				t.Fatalf("decode response: %v; body = %q", err, recorder.Body.String())
			}
			if response.Handler != test.wantHandler {
				t.Errorf("handler = %q, want %q", response.Handler, test.wantHandler)
			}
		})
	}
}

type nodeLogRouteHandlerFake struct{}

func (nodeLogRouteHandlerFake) NodeLogGetList(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"handler": "get"})
}

func (nodeLogRouteHandlerFake) NodeLogDelete(c *echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"handler": "delete"})
}

func nodeLogPermissionMiddleware(requiredPermissions ...string) echo.MiddlewareFunc {
	permissions := strings.Join(requiredPermissions, ",")
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Response().Header().Set("X-Test-Required-Permissions", permissions)
			return next(c)
		}
	}
}
