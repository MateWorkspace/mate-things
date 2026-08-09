package applicationinfraredrecordsessionmanagement

import (
	"context"
	"encoding/json"
	"time"

	applicationinfraredcasegeneration "github.com/MateWorkspace/mate-things/backend/internal/application/infrared/case_generation"
	applicationshared "github.com/MateWorkspace/mate-things/backend/internal/application/shared"
	domaincontractsbroadcaster "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/broadcaster"
	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsnode "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/node"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesinfrared "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/infrared"
	"github.com/google/uuid"
)

// LlmClientFactory is the subset of infrastructurellm.ClientFactory this
// usecase needs — declared as a domain-layer interface so the usecase
// doesn't import the infrastructure package directly.
type LlmClientFactory interface {
	Current(ctx context.Context) (domaincontractsllm.Client, error)
}

const caseGenerationTimeout = 2 * time.Minute

type usecase struct {
	session       domaincontractsrepository.InfraredRecordSession
	device        domaincontractsrepository.InfraredDevice
	definition    domaincontractsrepository.InfraredStateDeviceDefinition
	state         domaincontractsrepository.InfraredState
	recordCase    domaincontractsrepository.InfraredStateDeviceRecordCase
	broadcaster   domaincontractsbroadcaster.InfraredRecordSession
	subscriptions domaincontractsnode.Subscriptions
	llmFactory    LlmClientFactory
	node          domaincontractsrepository.Node
	logger        domaincontractslogger.Leveled
}

func NewUsecaseImpl(
	session domaincontractsrepository.InfraredRecordSession,
	device domaincontractsrepository.InfraredDevice,
	definition domaincontractsrepository.InfraredStateDeviceDefinition,
	state domaincontractsrepository.InfraredState,
	recordCase domaincontractsrepository.InfraredStateDeviceRecordCase,
	broadcaster domaincontractsbroadcaster.InfraredRecordSession,
	subscriptions domaincontractsnode.Subscriptions,
	llmFactory LlmClientFactory,
	node domaincontractsrepository.Node,
	logger domaincontractslogger.Leveled,
) domainusecasesinfrared.RecordSessionManagement {
	return &usecase{
		session: session, device: device, definition: definition, state: state,
		recordCase: recordCase, broadcaster: broadcaster, subscriptions: subscriptions,
		llmFactory: llmFactory, node: node, logger: logger,
	}
}

func (u *usecase) Start(ctx context.Context, request domainusecasesinfrared.StartRecordSessionRequest) (uuid.UUID, error) {
	const tag = "infrared/record_session_management/Start"

	brand, err := applicationshared.RequiredInfraredBrand(request.Brand, "brand")
	if err != nil {
		return uuid.Nil, err
	}
	model, err := applicationshared.RequiredInfraredModel(request.Model, "model")
	if err != nil {
		return uuid.Nil, err
	}

	deviceId, err := u.device.Create(ctx, request.InfraredDeviceTypeId, brand, model)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to create infrared device", domainmodels.LoggerMeta{"err": err})
		return uuid.Nil, err
	}

	definitions := make([]domainmodels.InfraredStateDeviceDefinition, 0, len(request.Definitions))
	for _, d := range request.Definitions {
		definitions = append(definitions, domainmodels.InfraredStateDeviceDefinition{
			InfraredDeviceId: deviceId, InfraredStateId: d.InfraredStateId,
			Options: d.Options, Minimum: d.Minimum, Maximum: d.Maximum, Step: d.Step,
		})
	}
	if err := u.definition.CreateMany(ctx, definitions); err != nil {
		u.logger.Error(ctx, tag, "failed to create infrared state device definitions", domainmodels.LoggerMeta{"err": err})
		return uuid.Nil, err
	}

	sessionId, err := u.session.Create(ctx, request.NodeId, deviceId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to create infrared record session", domainmodels.LoggerMeta{"err": err})
		return uuid.Nil, err
	}

	u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateCasesGenerating)

	go u.runCaseGeneration(sessionId, deviceId, request.InfraredDeviceTypeId, brand, model)

	return sessionId, nil
}

func (u *usecase) runCaseGeneration(sessionId uuid.UUID, deviceId uuid.UUID, deviceTypeId uuid.UUID, brand string, model string) {
	const tag = "infrared/record_session_management/runCaseGeneration"

	ctx, cancel := context.WithTimeout(context.Background(), caseGenerationTimeout)
	defer cancel()

	states, err := u.state.ListByDeviceTypeId(ctx, deviceTypeId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list infrared states", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	definitions, err := u.definition.ListByDeviceId(ctx, deviceId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list infrared state device definitions", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	cases := applicationinfraredcasegeneration.Generate(states, definitions)

	client, err := u.llmFactory.Current(ctx)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to resolve llm client", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	scriptedCases, err := applicationinfraredcasegeneration.WriteScript(ctx, client, brand, model, states, cases)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to write case script", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	var firstCaseId uuid.UUID
	for i, c := range scriptedCases {
		states := make([]domainmodels.InfraredStateDeviceRecordState, 0, len(c.States))
		for stateId, value := range c.States {
			states = append(states, domainmodels.InfraredStateDeviceRecordState{InfraredStateId: stateId, StateValue: value})
		}
		caseId, err := u.recordCase.CreateWithStates(ctx, sessionId, c.Step, c.Description, states)
		if err != nil {
			u.logger.Error(ctx, tag, "failed to persist record case", domainmodels.LoggerMeta{"err": err, "session_id": sessionId, "step": c.Step})
			u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
			return
		}
		if i == 0 {
			firstCaseId = caseId
		}
	}

	if err := u.recordCase.UpdateStatusById(ctx, firstCaseId, domainmodels.InfraredRecordCaseStatusActive); err != nil {
		u.logger.Error(ctx, tag, "failed to activate first record case", domainmodels.LoggerMeta{"err": err, "session_id": sessionId, "case_id": firstCaseId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	if err := u.session.UpdateCurrentRecordCaseIdById(ctx, sessionId, &firstCaseId); err != nil {
		u.logger.Error(ctx, tag, "failed to set session's current record case", domainmodels.LoggerMeta{"err": err, "session_id": sessionId, "case_id": firstCaseId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	session, err := u.session.GetById(ctx, sessionId)
	if err != nil || session == nil {
		u.logger.Error(ctx, tag, "failed to look up session before subscribing to ir capture", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	node, err := u.node.ReadById(ctx, session.NodeId)
	if err != nil || node == nil {
		u.logger.Error(ctx, tag, "failed to look up node before subscribing to ir capture", domainmodels.LoggerMeta{"err": err, "session_id": sessionId, "node_id": session.NodeId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	if err := u.subscriptions.IrCapture(ctx, node.DeviceId); err != nil {
		u.logger.Error(ctx, tag, "failed to subscribe to ir capture topic", domainmodels.LoggerMeta{"err": err, "session_id": sessionId, "device_id": node.DeviceId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateRecording)
}

func (u *usecase) transition(ctx context.Context, tag string, sessionId uuid.UUID, state string) {
	// The passed-in ctx may already carry a deadline (e.g. the case-generation
	// goroutine's overall timeout) that has just expired — if that's exactly
	// why we're transitioning to FAILED, writing the terminal state on that
	// same ctx would silently fail too, wedging the session forever. Give the
	// actual writes a fresh, short-lived context; ctx is still fine for the
	// logging calls below, which do no network I/O.
	writeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	isCompleted := state == domainmodels.InfraredRecordingStateCompleted || state == domainmodels.InfraredRecordingStateFailed
	if err := u.session.UpdateRecordingStateById(writeCtx, sessionId, state, isCompleted); err != nil {
		u.logger.Error(ctx, tag, "failed to update recording state", domainmodels.LoggerMeta{"err": err, "session_id": sessionId, "state": state})
		return
	}
	if err := u.broadcaster.Send(writeCtx, domainmodels.InfraredRecordSessionEvent{SessionId: sessionId, RecordingState: state}); err != nil {
		u.logger.Warn(ctx, tag, "failed to broadcast recording state", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
	}
}

func (u *usecase) GetById(ctx context.Context, id uuid.UUID) (*domainmodels.InfraredRecordSession, error) {
	return u.session.GetById(ctx, id)
}

func (u *usecase) ListCases(ctx context.Context, sessionId uuid.UUID) ([]domainusecasesinfrared.CaseWithStatesAndRaw, error) {
	cases, err := u.recordCase.ListBySessionId(ctx, sessionId)
	if err != nil {
		return nil, err
	}
	result := make([]domainusecasesinfrared.CaseWithStatesAndRaw, 0, len(cases))
	for _, c := range cases {
		states, err := u.recordCase.ListStatesByCaseId(ctx, c.Id)
		if err != nil {
			return nil, err
		}
		raw, err := u.recordCase.ListRawByCaseId(ctx, c.Id)
		if err != nil {
			return nil, err
		}
		result = append(result, domainusecasesinfrared.CaseWithStatesAndRaw{Case: c, States: states, Raw: raw})
	}
	return result, nil
}

func (u *usecase) AcceptRaw(ctx context.Context, rawId uuid.UUID) error {
	return u.recordCase.UpdateRawStatusById(ctx, rawId, domainmodels.InfraredRecordRawStatusAccepted, nil)
}

func (u *usecase) DiscardRaw(ctx context.Context, rawId uuid.UUID, reason string) error {
	if reason == "" {
		return domainmodels.NewError("reason is required", domainmodels.ErrTypeValidation, nil)
	}
	return u.recordCase.UpdateRawStatusById(ctx, rawId, domainmodels.InfraredRecordRawStatusDiscarded, &reason)
}

func (u *usecase) RetryCase(ctx context.Context, caseId uuid.UUID) error {
	const tag = "infrared/record_session_management/RetryCase"

	raw, err := u.recordCase.ListRawByCaseId(ctx, caseId)
	if err != nil {
		return err
	}
	for _, r := range raw {
		if r.Status == domainmodels.InfraredRecordRawStatusAccepted {
			reason := "case marked for redo"
			if err := u.recordCase.UpdateRawStatusById(ctx, r.Id, domainmodels.InfraredRecordRawStatusDiscarded, &reason); err != nil {
				u.logger.Error(ctx, tag, "failed to discard accepted raw for retry", domainmodels.LoggerMeta{"err": err, "raw_id": r.Id})
				return err
			}
		}
	}
	if err := u.recordCase.UpdateStatusById(ctx, caseId, domainmodels.InfraredRecordCaseStatusActive); err != nil {
		u.logger.Error(ctx, tag, "failed to activate case for retry", domainmodels.LoggerMeta{"err": err, "case_id": caseId})
		return err
	}

	recordCase, err := u.recordCase.GetById(ctx, caseId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to look up case for retry", domainmodels.LoggerMeta{"err": err, "case_id": caseId})
		return err
	}
	if err := u.session.UpdateCurrentRecordCaseIdById(ctx, recordCase.InfraredRecordSessionId, &caseId); err != nil {
		u.logger.Error(ctx, tag, "failed to set session's current record case for retry", domainmodels.LoggerMeta{"err": err, "case_id": caseId})
		return err
	}
	return nil
}

func (u *usecase) SetCurrentCase(ctx context.Context, sessionId uuid.UUID, caseId uuid.UUID) error {
	const tag = "infrared/record_session_management/SetCurrentCase"

	recordCase, err := u.recordCase.GetById(ctx, caseId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to look up case", domainmodels.LoggerMeta{"err": err, "case_id": caseId})
		return err
	}
	if recordCase == nil || recordCase.InfraredRecordSessionId != sessionId {
		return domainmodels.NewError("case does not belong to session", domainmodels.ErrTypeValidation, nil)
	}

	if err := u.recordCase.UpdateStatusById(ctx, caseId, domainmodels.InfraredRecordCaseStatusActive); err != nil {
		u.logger.Error(ctx, tag, "failed to activate case", domainmodels.LoggerMeta{"err": err, "case_id": caseId})
		return err
	}
	if err := u.session.UpdateCurrentRecordCaseIdById(ctx, sessionId, &caseId); err != nil {
		u.logger.Error(ctx, tag, "failed to set session's current record case", domainmodels.LoggerMeta{"err": err, "session_id": sessionId, "case_id": caseId})
		return err
	}

	if err := u.broadcaster.Send(ctx, domainmodels.InfraredRecordSessionEvent{SessionId: sessionId, CurrentRecordCaseId: &caseId}); err != nil {
		u.logger.Warn(ctx, tag, "failed to broadcast current record case", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
	}

	return nil
}

func (u *usecase) CaptureIrRaw(ctx context.Context, request domainusecasesinfrared.CaptureIrRawRequest) error {
	const tag = "infrared/record_session_management/CaptureIrRaw"

	node, err := u.node.ReadByDeviceId(ctx, request.NodeDeviceId)
	if err != nil || node == nil {
		u.logger.Warn(ctx, tag, "ir capture from unknown node", domainmodels.LoggerMeta{"device_id": request.NodeDeviceId})
		return nil
	}

	session, err := u.session.GetActiveByNodeId(ctx, node.Id)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to look up active session for node", domainmodels.LoggerMeta{"err": err, "node_id": node.Id})
		return err
	}
	if session == nil || session.CurrentRecordCaseId == nil {
		u.logger.Warn(ctx, tag, "ir capture with no active recording case", domainmodels.LoggerMeta{"node_id": node.Id})
		return nil
	}

	rawBytes, err := json.Marshal(request.RawData)
	if err != nil {
		return domainmodels.NewError("failed to encode raw ir data", domainmodels.ErrTypeFailure, err)
	}
	if _, err := u.recordCase.CreateRaw(ctx, *session.CurrentRecordCaseId, rawBytes); err != nil {
		u.logger.Error(ctx, tag, "failed to persist raw capture", domainmodels.LoggerMeta{"err": err, "case_id": *session.CurrentRecordCaseId})
		return err
	}

	if err := u.broadcaster.Send(ctx, domainmodels.InfraredRecordSessionEvent{
		SessionId: session.Id, RecordingState: session.RecordingState, CurrentRecordCaseId: session.CurrentRecordCaseId,
	}); err != nil {
		u.logger.Warn(ctx, tag, "failed to broadcast raw capture", domainmodels.LoggerMeta{"err": err, "session_id": session.Id})
	}
	return nil
}
