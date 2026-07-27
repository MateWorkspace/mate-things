package presentationhttprequest

type NodeClassPostRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type NodeClassPatchRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type NodePatchRequest struct {
	NodeClassId *string `json:"node_class_id"`
	DeviceId    *string `json:"device_id"`
	Name        *string `json:"name"`
	FirmwareId  *string `json:"firmware_id"`
	Description *string `json:"description"`
}

type NodeFirmwarePatchRequest struct {
	FirmwareId string `json:"firmware_id"`
}

type FirmwarePatchRequest struct {
	NodeClassId *string `json:"node_class_id"`
	Name        *string `json:"name"`
}

type OtaDispatchRequest struct {
	FirmwareId  string `json:"firmware_id"`
	FirmwareUrl string `json:"firmware_url"`
}
