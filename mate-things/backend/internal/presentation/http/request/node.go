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
	NodeClassId  *string `json:"node_class_id"`
	DeviceId     *string `json:"device_id"`
	Name         *string `json:"name"`
	FirmwareName *string `json:"firmware_name"`
	Description  *string `json:"description"`
}

type NodeFirmwarePatchRequest struct {
	FirmwareName string `json:"firmware_name"`
}

type FirmwarePatchRequest struct {
	NodeClassId *string `json:"node_class_id"`
	Name        *string `json:"name"`
}

type OtaDispatchRequest struct {
	FirmwareName string `json:"firmware_name"`
	FirmwareUrl  string `json:"firmware_url"`
}
