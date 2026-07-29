package presentationhttprequest

type NodeClassPostRequest struct {
	Name        string  `json:"name" example:"Espresso Machine"`
	Description *string `json:"description" example:"Dual-boiler espresso machines with ESP32-controlled brew heads."`
}

type NodeClassPatchRequest struct {
	Name        *string `json:"name" example:"Espresso Machine"`
	Description *string `json:"description" example:"Dual-boiler espresso machines with ESP32-controlled brew heads."`
}

type NodePatchRequest struct {
	NodeClassId *string `json:"node_class_id" example:"3f1c9a2e-6d4b-4e7a-8c2f-1a9b3d5e7f01"`
	DeviceId    *string `json:"device_id" example:"ESP32-BARISTA-07"`
	Name        *string `json:"name" example:"Kitchen Espresso Machine"`
	FirmwareId  *string `json:"firmware_id" example:"9c4e2b7a-1f3d-4a6c-8b5e-2d7f9a1c3e08"`
	Description *string `json:"description" example:"The espresso machine behind the office kitchen counter."`
}

type NodeFirmwarePatchRequest struct {
	FirmwareId string `json:"firmware_id" example:"9c4e2b7a-1f3d-4a6c-8b5e-2d7f9a1c3e08"`
}

type FirmwarePatchRequest struct {
	NodeClassId *string `json:"node_class_id" example:"3f1c9a2e-6d4b-4e7a-8c2f-1a9b3d5e7f01"`
	Name        *string `json:"name" example:"espresso-fw"`
}

type OtaDispatchRequest struct {
	FirmwareId  string `json:"firmware_id" example:"9c4e2b7a-1f3d-4a6c-8b5e-2d7f9a1c3e08"`
	FirmwareUrl string `json:"firmware_url" example:"https://cdn.mate-things.internal/firmwares/espresso-fw/v3/firmware.bin"`
}

type SetNodeConfigValueRequest struct {
	Key   string `json:"key" example:"mqtt_host"`
	Value string `json:"value" example:"broker.example.com"`
}
