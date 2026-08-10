package presentationhttprequest

type StartRecordSessionRequest struct {
	NodeId               string                         `json:"node_id" example:"..."`
	InfraredDeviceTypeId string                         `json:"infrared_device_type_id" example:"..."`
	Brand                string                         `json:"brand" example:"Polytron"`
	Model                string                         `json:"model" example:"PAC-09HDN"`
	Definitions          []StartRecordSessionDefinition `json:"definitions"`
}

type StartRecordSessionDefinition struct {
	InfraredStateId string   `json:"infrared_state_id" example:"..."`
	Options         []string `json:"options,omitempty" example:"ON,OFF"`
	Minimum         *float64 `json:"minimum,omitempty" example:"16"`
	Maximum         *float64 `json:"maximum,omitempty" example:"30"`
	Step            *float64 `json:"step,omitempty" example:"1"`
}

type DiscardRawRequest struct {
	Reason string `json:"reason" example:"pressed the wrong button"`
}

type RecordTestCaseResultRequest struct {
	Passed bool `json:"passed" example:"true"`
}

type CreateDeviceTypeRequest struct {
	Name string `json:"name" example:"Air Conditioner"`
}

type CreateStateRequest struct {
	Name string `json:"name" example:"POWER"`
	Type string `json:"type" example:"ENUM"`
}
