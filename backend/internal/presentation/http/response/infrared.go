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
}

func InfraredRecordSession(model domainmodels.InfraredRecordSession) InfraredRecordSessionResponse {
	return InfraredRecordSessionResponse{
		Id:                  UUIDString(model.Id),
		NodeId:              UUIDString(model.NodeId),
		InfraredDeviceId:    UUIDString(model.InfraredDeviceId),
		RecordingState:      model.RecordingState,
		CurrentRecordCaseId: UUIDPtrString(model.CurrentRecordCaseId),
		IsCompleted:         model.IsCompleted,
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
}

func InfraredStateDeviceRecordState(model domainmodels.InfraredStateDeviceRecordState) InfraredStateDeviceRecordStateResponse {
	return InfraredStateDeviceRecordStateResponse{
		Id:              UUIDString(model.Id),
		InfraredStateId: UUIDString(model.InfraredStateId),
		StateValue:      model.StateValue,
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
}

func InfraredStateDeviceRecordRaw(model domainmodels.InfraredStateDeviceRecordRaw) InfraredStateDeviceRecordRawResponse {
	return InfraredStateDeviceRecordRawResponse{
		Id:              UUIDString(model.Id),
		Status:          string(model.Status),
		DiscardedReason: model.DiscardedReason,
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
}

func InfraredStateDeviceRecordCase(
	model domainmodels.InfraredStateDeviceRecordCase,
	states []domainmodels.InfraredStateDeviceRecordState,
	raw []domainmodels.InfraredStateDeviceRecordRaw,
) InfraredStateDeviceRecordCaseResponse {
	return InfraredStateDeviceRecordCaseResponse{
		Id:          UUIDString(model.Id),
		Step:        model.Step,
		Description: model.Description,
		Status:      string(model.Status),
		States:      InfraredStateDeviceRecordStates(states),
		Raw:         InfraredStateDeviceRecordRaws(raw),
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
}

func InfraredStateCoder(model domainmodels.InfraredStateCoder) InfraredStateCoderResponse {
	return InfraredStateCoderResponse{
		Id:            UUIDString(model.Id),
		EncoderSource: model.EncoderSource,
		DecoderSource: model.DecoderSource,
		SummaryReadme: model.SummaryReadme,
		DetailReadme:  model.DetailReadme,
		Status:        string(model.Status),
	}
}
