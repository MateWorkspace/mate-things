package presentationhttpresponse

import (
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesinfrared "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/infrared"
)

type InfraredRecordSessionResponse struct {
	Id                  string  `json:"id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	NodeId              string  `json:"node_id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	InfraredDeviceId    string  `json:"infrared_device_id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	RecordingState      string  `json:"recording_state" example:"RECORDING"`
	CurrentRecordCaseId *string `json:"current_record_case_id,omitempty" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	IsCompleted         bool    `json:"is_completed" example:"false"`
	AuditResponse
}

func InfraredRecordSession(model domainmodels.InfraredRecordSession) InfraredRecordSessionResponse {
	return InfraredRecordSessionResponse{
		Id:                  UUIDString(model.Id),
		NodeId:              UUIDString(model.NodeId),
		InfraredDeviceId:    UUIDString(model.InfraredDeviceId),
		RecordingState:      model.RecordingState,
		CurrentRecordCaseId: UUIDPtrString(model.CurrentRecordCaseId),
		IsCompleted:         model.IsCompleted,
		AuditResponse:       Audit(model.CreatedAt, model.UpdatedAt, model.DeletedAt, model.CreatedBy, model.UpdatedBy, model.DeletedBy),
	}
}

func InfraredRecordSessions(models []domainmodels.InfraredRecordSession) []InfraredRecordSessionResponse {
	responses := make([]InfraredRecordSessionResponse, len(models))
	for i, model := range models {
		responses[i] = InfraredRecordSession(model)
	}
	return responses
}

type InfraredStateDeviceRecordStateResponse struct {
	Id              string `json:"id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	InfraredStateId string `json:"infrared_state_id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	StateValue      string `json:"state_value" example:"ON"`
	AuditResponse
}

func InfraredStateDeviceRecordState(model domainmodels.InfraredStateDeviceRecordState) InfraredStateDeviceRecordStateResponse {
	return InfraredStateDeviceRecordStateResponse{
		Id:              UUIDString(model.Id),
		InfraredStateId: UUIDString(model.InfraredStateId),
		StateValue:      model.StateValue,
		AuditResponse:   Audit(model.CreatedAt, model.UpdatedAt, model.DeletedAt, model.CreatedBy, model.UpdatedBy, model.DeletedBy),
	}
}

func InfraredStateDeviceRecordStates(models []domainmodels.InfraredStateDeviceRecordState) []InfraredStateDeviceRecordStateResponse {
	responses := make([]InfraredStateDeviceRecordStateResponse, len(models))
	for i, model := range models {
		responses[i] = InfraredStateDeviceRecordState(model)
	}
	return responses
}

type InfraredStateDeviceRecordRawResponse struct {
	Id              string  `json:"id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	Status          string  `json:"status" example:"CAPTURED"`
	DiscardedReason *string `json:"discarded_reason,omitempty" example:"pressed the wrong button"`
	AuditResponse
}

func InfraredStateDeviceRecordRaw(model domainmodels.InfraredStateDeviceRecordRaw) InfraredStateDeviceRecordRawResponse {
	return InfraredStateDeviceRecordRawResponse{
		Id:              UUIDString(model.Id),
		Status:          string(model.Status),
		DiscardedReason: model.DiscardedReason,
		AuditResponse:   Audit(model.CreatedAt, model.UpdatedAt, model.DeletedAt, model.CreatedBy, model.UpdatedBy, model.DeletedBy),
	}
}

func InfraredStateDeviceRecordRaws(models []domainmodels.InfraredStateDeviceRecordRaw) []InfraredStateDeviceRecordRawResponse {
	responses := make([]InfraredStateDeviceRecordRawResponse, len(models))
	for i, model := range models {
		responses[i] = InfraredStateDeviceRecordRaw(model)
	}
	return responses
}

type InfraredStateDeviceRecordCaseResponse struct {
	Id          string                                   `json:"id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	Step        int32                                    `json:"step" example:"1"`
	Description string                                   `json:"description" example:"ON, mode=COOL, temp=16"`
	Status      string                                   `json:"status" example:"PENDING"`
	States      []InfraredStateDeviceRecordStateResponse `json:"states"`
	Raw         []InfraredStateDeviceRecordRawResponse   `json:"raw"`
	AuditResponse
}

func InfraredStateDeviceRecordCase(
	model domainmodels.InfraredStateDeviceRecordCase,
	states []domainmodels.InfraredStateDeviceRecordState,
	raw []domainmodels.InfraredStateDeviceRecordRaw,
) InfraredStateDeviceRecordCaseResponse {
	return InfraredStateDeviceRecordCaseResponse{
		Id:            UUIDString(model.Id),
		Step:          model.Step,
		Description:   model.Description,
		Status:        string(model.Status),
		States:        InfraredStateDeviceRecordStates(states),
		Raw:           InfraredStateDeviceRecordRaws(raw),
		AuditResponse: Audit(model.CreatedAt, model.UpdatedAt, model.DeletedAt, model.CreatedBy, model.UpdatedBy, model.DeletedBy),
	}
}

func InfraredStateDeviceRecordCases(cases []domainusecasesinfrared.CaseWithStatesAndRaw) []InfraredStateDeviceRecordCaseResponse {
	responses := make([]InfraredStateDeviceRecordCaseResponse, len(cases))
	for i, c := range cases {
		responses[i] = InfraredStateDeviceRecordCase(c.Case, c.States, c.Raw)
	}
	return responses
}

type InfraredStateCoderResponse struct {
	Id            string `json:"id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	EncoderSource string `json:"encoder_source"`
	DecoderSource string `json:"decoder_source"`
	SummaryReadme string `json:"summary_readme"`
	DetailReadme  string `json:"detail_readme"`
	Status        string `json:"status" example:"UNVERIFIED"`
	AuditResponse
}

func InfraredStateCoder(model domainmodels.InfraredStateCoder) InfraredStateCoderResponse {
	return InfraredStateCoderResponse{
		Id:            UUIDString(model.Id),
		EncoderSource: model.EncoderSource,
		DecoderSource: model.DecoderSource,
		SummaryReadme: model.SummaryReadme,
		DetailReadme:  model.DetailReadme,
		Status:        string(model.Status),
		AuditResponse: Audit(model.CreatedAt, model.UpdatedAt, model.DeletedAt, model.CreatedBy, model.UpdatedBy, model.DeletedBy),
	}
}

type InfraredTestCaseStateResponse struct {
	Id              string `json:"id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	InfraredStateId string `json:"infrared_state_id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	StateValue      string `json:"state_value" example:"ON"`
	AuditResponse
}

func InfraredTestCaseStates(models []domainmodels.InfraredTestCaseState) []InfraredTestCaseStateResponse {
	responses := make([]InfraredTestCaseStateResponse, len(models))
	for i, model := range models {
		responses[i] = InfraredTestCaseStateResponse{
			Id:              UUIDString(model.Id),
			InfraredStateId: UUIDString(model.InfraredStateId),
			StateValue:      model.StateValue,
			AuditResponse:   Audit(model.CreatedAt, model.UpdatedAt, model.DeletedAt, model.CreatedBy, model.UpdatedBy, model.DeletedBy),
		}
	}
	return responses
}

type InfraredTestCaseResponse struct {
	Id          string                          `json:"id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	Step        int32                           `json:"step" example:"1"`
	Description string                          `json:"description" example:"Confirm the unit powers on."`
	Status      string                          `json:"status" example:"PENDING"`
	States      []InfraredTestCaseStateResponse `json:"states"`
	AuditResponse
}

func InfraredTestCases(models []domainusecasesinfrared.TestCaseWithStates) []InfraredTestCaseResponse {
	responses := make([]InfraredTestCaseResponse, len(models))
	for i, m := range models {
		responses[i] = InfraredTestCaseResponse{
			Id:            UUIDString(m.TestCase.Id),
			Step:          m.TestCase.Step,
			Description:   m.TestCase.Description,
			Status:        string(m.TestCase.Status),
			States:        InfraredTestCaseStates(m.States),
			AuditResponse: Audit(m.TestCase.CreatedAt, m.TestCase.UpdatedAt, m.TestCase.DeletedAt, m.TestCase.CreatedBy, m.TestCase.UpdatedBy, m.TestCase.DeletedBy),
		}
	}
	return responses
}

type InfraredDeviceTypeResponse struct {
	Id   string `json:"id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	Name string `json:"name" example:"Air Conditioner"`
	AuditResponse
}

func InfraredDeviceType(model domainmodels.InfraredDeviceType) InfraredDeviceTypeResponse {
	return InfraredDeviceTypeResponse{
		Id:            UUIDString(model.Id),
		Name:          model.Name,
		AuditResponse: Audit(model.CreatedAt, model.UpdatedAt, model.DeletedAt, model.CreatedBy, model.UpdatedBy, model.DeletedBy),
	}
}

func InfraredDeviceTypes(models []domainmodels.InfraredDeviceType) []InfraredDeviceTypeResponse {
	responses := make([]InfraredDeviceTypeResponse, len(models))
	for i, model := range models {
		responses[i] = InfraredDeviceType(model)
	}
	return responses
}

type InfraredStateResponse struct {
	Id                   string `json:"id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	InfraredDeviceTypeId string `json:"infrared_device_type_id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	Name                 string `json:"name" example:"POWER"`
	Type                 string `json:"type" example:"ENUM"`
	AuditResponse
}

func InfraredState(model domainmodels.InfraredState) InfraredStateResponse {
	return InfraredStateResponse{
		Id:                   UUIDString(model.Id),
		InfraredDeviceTypeId: UUIDString(model.InfraredDeviceTypeId),
		Name:                 model.Name,
		Type:                 string(model.Type),
		AuditResponse:        Audit(model.CreatedAt, model.UpdatedAt, model.DeletedAt, model.CreatedBy, model.UpdatedBy, model.DeletedBy),
	}
}

func InfraredStates(models []domainmodels.InfraredState) []InfraredStateResponse {
	responses := make([]InfraredStateResponse, len(models))
	for i, model := range models {
		responses[i] = InfraredState(model)
	}
	return responses
}

type InfraredStateDeviceDefinitionResponse struct {
	Id               string   `json:"id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	InfraredDeviceId string   `json:"infrared_device_id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	InfraredStateId  string   `json:"infrared_state_id" example:"e1f4b7c0-2d5e-4f8a-9b3c-6e0f2a5d8c01"`
	Options          []string `json:"options,omitempty" example:"ON,OFF"`
	Minimum          *float64 `json:"minimum,omitempty" example:"16"`
	Maximum          *float64 `json:"maximum,omitempty" example:"30"`
	Step             *float64 `json:"step,omitempty" example:"1"`
	AuditResponse
}

func InfraredStateDeviceDefinition(model domainmodels.InfraredStateDeviceDefinition) InfraredStateDeviceDefinitionResponse {
	return InfraredStateDeviceDefinitionResponse{
		Id:               UUIDString(model.Id),
		InfraredDeviceId: UUIDString(model.InfraredDeviceId),
		InfraredStateId:  UUIDString(model.InfraredStateId),
		Options:          model.Options,
		Minimum:          model.Minimum,
		Maximum:          model.Maximum,
		Step:             model.Step,
		AuditResponse:    Audit(model.CreatedAt, model.UpdatedAt, model.DeletedAt, model.CreatedBy, model.UpdatedBy, model.DeletedBy),
	}
}

func InfraredStateDeviceDefinitions(models []domainmodels.InfraredStateDeviceDefinition) []InfraredStateDeviceDefinitionResponse {
	responses := make([]InfraredStateDeviceDefinitionResponse, len(models))
	for i, model := range models {
		responses[i] = InfraredStateDeviceDefinition(model)
	}
	return responses
}
