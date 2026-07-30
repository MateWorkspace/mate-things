package presentationhttproute

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestRouteNodeGatesOtaDispatchOnlyWithOtaPermission(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "node id", path: "/api/v1/nodes/3f1c9a2e-6d4b-4e7a-8c2f-1a9b3d5e7f01/ota"},
		{name: "device id", path: "/api/v1/nodes/by-device/AC276E5E030C/ota"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			routeNode(e.Group("/api/v1"), otaRouteHandlerFake{}, nodeLogPermissionMiddleware)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, test.path, nil)

			e.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d; body = %q", recorder.Code, http.StatusNoContent, recorder.Body.String())
			}
			if got := recorder.Header().Get("X-Test-Required-Permissions"); got != "ota:dispatch" {
				t.Fatalf("required permissions = %q, want %q", got, "ota:dispatch")
			}
		})
	}
}

type otaRouteHandlerFake struct {
	NodeHandler
}

func (otaRouteHandlerFake) OtaDispatchByNodeIdPost(c *echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

func (otaRouteHandlerFake) OtaDispatchByNodeDeviceIdPost(c *echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}
