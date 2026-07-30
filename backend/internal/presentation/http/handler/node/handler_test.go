package presentationhttphandlernode

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

func TestFirmwareBinaryPutForwardsExplicitEmptyConfigSchema(t *testing.T) {
	firmwareID := uuid.New()
	firmware := &recordingFirmwareManagement{}
	handler := NewHandler(nil, nil, firmware, nil, nil, nil)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("file", "firmware.bin")
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := file.Write([]byte{0xE9, 0x01}); err != nil {
		t.Fatalf("file Write() error = %v", err)
	}
	if err := writer.WriteField("config_schema", "[]"); err != nil {
		t.Fatalf("WriteField() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("multipart Close() error = %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/firmwares/"+firmwareID.String()+"/binary",
		&body,
	)
	request.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	ctx := echo.New().NewContext(request, recorder)
	ctx.SetPathValues(echo.PathValues{{Name: "id", Value: firmwareID.String()}})

	if err := handler.FirmwareBinaryPut(ctx); err != nil {
		t.Fatalf("FirmwareBinaryPut() error = %v", err)
	}

	if firmware.replaceCalls != 1 {
		t.Fatalf("ReplaceBinaryById() calls = %d, want 1", firmware.replaceCalls)
	}
	if firmware.replaceRequest.Id != firmwareID {
		t.Fatalf("ReplaceBinaryById() ID = %s, want %s", firmware.replaceRequest.Id, firmwareID)
	}
	if firmware.replaceRequest.ConfigSchema == nil {
		t.Fatal("ReplaceBinaryById() config schema = nil, want explicit empty slice")
	}
	if len(firmware.replaceRequest.ConfigSchema) != 0 {
		t.Fatalf(
			"ReplaceBinaryById() config schema = %#v, want empty",
			firmware.replaceRequest.ConfigSchema,
		)
	}
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestFirmwareBinaryPutPreservesNilConfigSchemaWhenFieldIsOmitted(t *testing.T) {
	firmwareID := uuid.New()
	firmware := &recordingFirmwareManagement{}
	handler := NewHandler(nil, nil, firmware, nil, nil, nil)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("file", "legacy-firmware.bin")
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := file.Write([]byte{0xE9, 0x01}); err != nil {
		t.Fatalf("file Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("multipart Close() error = %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/firmwares/"+firmwareID.String()+"/binary",
		&body,
	)
	request.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	recorder := httptest.NewRecorder()
	ctx := echo.New().NewContext(request, recorder)
	ctx.SetPathValues(echo.PathValues{{Name: "id", Value: firmwareID.String()}})

	if err := handler.FirmwareBinaryPut(ctx); err != nil {
		t.Fatalf("FirmwareBinaryPut() error = %v", err)
	}

	if firmware.replaceCalls != 1 {
		t.Fatalf("ReplaceBinaryById() calls = %d, want 1", firmware.replaceCalls)
	}
	if firmware.replaceRequest.ConfigSchema != nil {
		t.Fatalf(
			"ReplaceBinaryById() config schema = %#v, want nil",
			firmware.replaceRequest.ConfigSchema,
		)
	}
}

func TestFirmwareDeleteForwardsExpectedNameForAuthoritativeConfirmation(t *testing.T) {
	firmwareID := uuid.New()
	firmware := &recordingFirmwareManagement{}
	handler := NewHandler(nil, nil, firmware, nil, nil, nil)
	request := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/firmwares/"+firmwareID.String(),
		bytes.NewBufferString(`{"expected_name":"authoritative-firmware"}`),
	)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	ctx := echo.New().NewContext(request, recorder)
	ctx.SetPathValues(echo.PathValues{{Name: "id", Value: firmwareID.String()}})

	if err := handler.FirmwareDelete(ctx); err != nil {
		t.Fatalf("FirmwareDelete() error = %v", err)
	}

	if firmware.deleteCalls != 1 {
		t.Fatalf("DeleteById() calls = %d, want 1", firmware.deleteCalls)
	}
	if firmware.deleteRequest.Id != firmwareID {
		t.Fatalf("DeleteById() ID = %s, want %s", firmware.deleteRequest.Id, firmwareID)
	}
	if firmware.deleteRequest.ExpectedName != "authoritative-firmware" {
		t.Fatalf(
			"DeleteById() expected name = %q, want %q",
			firmware.deleteRequest.ExpectedName,
			"authoritative-firmware",
		)
	}
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestFirmwareDeleteRejectsMissingOrEmptyExpectedName(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "missing field", body: `{}`},
		{name: "empty field", body: `{"expected_name":"   "}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			firmwareID := uuid.New()
			firmware := &recordingFirmwareManagement{}
			handler := NewHandler(nil, nil, firmware, nil, nil, nil)
			request := httptest.NewRequest(
				http.MethodDelete,
				"/api/v1/firmwares/"+firmwareID.String(),
				bytes.NewBufferString(test.body),
			)
			request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			recorder := httptest.NewRecorder()
			ctx := echo.New().NewContext(request, recorder)
			ctx.SetPathValues(echo.PathValues{{Name: "id", Value: firmwareID.String()}})

			if err := handler.FirmwareDelete(ctx); err != nil {
				t.Fatalf("FirmwareDelete() error = %v", err)
			}

			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
			}
			if firmware.deleteCalls != 0 {
				t.Fatalf("DeleteById() calls = %d, want 0", firmware.deleteCalls)
			}
		})
	}
}

func TestOtaDispatchByNodeIdForwardsFirmwareIdOnly(t *testing.T) {
	nodeID := uuid.New()
	firmwareID := uuid.New()
	ota := &recordingOta{}
	handler := NewHandler(nil, nil, nil, ota, nil, nil)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/nodes/"+nodeID.String()+"/ota",
		bytes.NewBufferString(`{"firmware_id":"`+firmwareID.String()+`"}`),
	)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	ctx := echo.New().NewContext(request, recorder)
	ctx.SetPathValues(echo.PathValues{{Name: "id", Value: nodeID.String()}})

	if err := handler.OtaDispatchByNodeIdPost(ctx); err != nil {
		t.Fatalf("OtaDispatchByNodeIdPost() error = %v", err)
	}

	if ota.byNodeIDCalls != 1 {
		t.Fatalf("DispatchByNodeId() calls = %d, want 1", ota.byNodeIDCalls)
	}
	if ota.byNodeIDRequest.NodeId != nodeID || ota.byNodeIDRequest.FirmwareId != firmwareID {
		t.Fatalf(
			"DispatchByNodeId() request = %#v, want node ID %s and firmware ID %s",
			ota.byNodeIDRequest,
			nodeID,
			firmwareID,
		)
	}
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestNodePatchIgnoresDeviceIdentityMutation(t *testing.T) {
	nodeID := uuid.New()
	device := &recordingDeviceManagement{}
	handler := NewHandler(nil, device, nil, nil, nil, nil)
	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/nodes/"+nodeID.String(),
		bytes.NewBufferString(`{"name":"node_AC276E5E030C","device_id":"FFFFFFFFFFFF"}`),
	)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	ctx := echo.New().NewContext(request, recorder)
	ctx.SetPathValues(echo.PathValues{{Name: "id", Value: nodeID.String()}})

	if err := handler.NodePatch(ctx); err != nil {
		t.Fatalf("NodePatch() error = %v", err)
	}

	if device.updateCalls != 1 {
		t.Fatalf("UpdateById() calls = %d, want 1", device.updateCalls)
	}
	if device.updateRequest.Name == nil || *device.updateRequest.Name != "node_AC276E5E030C" {
		t.Fatalf("UpdateById() name = %v, want persisted node name", device.updateRequest.Name)
	}
}

func TestNodeConfigPutPreservesPresentStringValueVerbatim(t *testing.T) {
	tests := []struct {
		name  string
		value string
		body  string
	}{
		{name: "empty", value: "", body: `{"key":"mqtt_host","value":""}`},
		{name: "whitespace", value: "  ", body: `{"key":"mqtt_host","value":"  "}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			nodeID := uuid.New()
			config := &recordingConfigValue{}
			handler := NewHandler(nil, nil, nil, nil, nil, config)
			request := httptest.NewRequest(
				http.MethodPut,
				"/api/v1/nodes/"+nodeID.String()+"/config",
				bytes.NewBufferString(test.body),
			)
			request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			recorder := httptest.NewRecorder()
			ctx := echo.New().NewContext(request, recorder)
			ctx.SetPathValues(echo.PathValues{{Name: "id", Value: nodeID.String()}})

			if err := handler.NodeConfigPut(ctx); err != nil {
				t.Fatalf("NodeConfigPut() error = %v", err)
			}

			if config.setCalls != 1 {
				t.Fatalf("SetByNodeId() calls = %d, want 1", config.setCalls)
			}
			if config.setRequest.Value != test.value {
				t.Fatalf("SetByNodeId() value = %q, want %q", config.setRequest.Value, test.value)
			}
		})
	}
}

func TestNodeConfigPutRejectsMissingValueField(t *testing.T) {
	nodeID := uuid.New()
	config := &recordingConfigValue{}
	handler := NewHandler(nil, nil, nil, nil, nil, config)
	request := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/nodes/"+nodeID.String()+"/config",
		bytes.NewBufferString(`{"key":"mqtt_host"}`),
	)
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	recorder := httptest.NewRecorder()
	ctx := echo.New().NewContext(request, recorder)
	ctx.SetPathValues(echo.PathValues{{Name: "id", Value: nodeID.String()}})

	if err := handler.NodeConfigPut(ctx); err != nil {
		t.Fatalf("NodeConfigPut() error = %v", err)
	}

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if config.setCalls != 0 {
		t.Fatalf("SetByNodeId() calls = %d, want 0", config.setCalls)
	}
}

type recordingFirmwareManagement struct {
	domainusecasesnode.FirmwareManagement
	deleteCalls    int
	deleteRequest  domainusecasesnode.DeleteFirmwareRequest
	replaceCalls   int
	replaceRequest domainusecasesnode.ReplaceFirmwareBinaryByIdRequest
}

type recordingConfigValue struct {
	domainusecasesnode.ConfigValue
	setCalls   int
	setRequest domainusecasesnode.SetConfigValueRequest
}

func (r *recordingConfigValue) SetByNodeId(
	_ context.Context,
	request domainusecasesnode.SetConfigValueRequest,
) error {
	r.setCalls++
	r.setRequest = request
	return nil
}

type recordingDeviceManagement struct {
	domainusecasesnode.DeviceManagement
	updateCalls   int
	updateRequest domainusecasesnode.UpdateNodeRequest
}

func (r *recordingDeviceManagement) UpdateById(
	_ context.Context,
	request domainusecasesnode.UpdateNodeRequest,
) error {
	r.updateCalls++
	r.updateRequest = request
	return nil
}

type recordingOta struct {
	domainusecasesnode.Ota
	byNodeIDCalls   int
	byNodeIDRequest domainusecasesnode.DispatchOtaByNodeIdRequest
}

func (r *recordingOta) DispatchByNodeId(
	_ context.Context,
	request domainusecasesnode.DispatchOtaByNodeIdRequest,
) error {
	r.byNodeIDCalls++
	r.byNodeIDRequest = request
	return nil
}

func (r *recordingFirmwareManagement) ReplaceBinaryById(
	_ context.Context,
	request domainusecasesnode.ReplaceFirmwareBinaryByIdRequest,
) (domainusecasesnode.FirmwareBinaryStatResult, error) {
	r.replaceCalls++
	r.replaceRequest = request
	return domainusecasesnode.FirmwareBinaryStatResult{}, nil
}

func (r *recordingFirmwareManagement) DeleteById(
	_ context.Context,
	request domainusecasesnode.DeleteFirmwareRequest,
) error {
	r.deleteCalls++
	r.deleteRequest = request
	return nil
}
