package applicationinfraredrecordsessionmanagement

import (
	"context"
	"encoding/json"
	"time"

	applicationinfraredanalysis "github.com/MateWorkspace/mate-things/backend/internal/application/infrared/analysis"
	applicationinfraredcasegeneration "github.com/MateWorkspace/mate-things/backend/internal/application/infrared/case_generation"
	applicationinfraredcodergeneration "github.com/MateWorkspace/mate-things/backend/internal/application/infrared/coder_generation"
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

// EncoderRunner is the subset of infrastructurejsengine's exported surface
// this usecase needs — declared locally so the usecase doesn't import the
// infrastructure package directly, same pattern as LlmClientFactory.
type EncoderRunner interface {
	RunEncoder(source string, state map[string]string, timeout time.Duration) ([]int32, error)
}

const caseGenerationTimeout = 2 * time.Minute
const analysisTimeout = 3 * time.Minute
const encoderSmokeTestTimeout = 2 * time.Second
const testCaseGenerationTimeout = 2 * time.Minute

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
	encoderRunner EncoderRunner
	coder         domaincontractsrepository.InfraredStateCoder
	testCase      domaincontractsrepository.InfraredTestCase
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
	encoderRunner EncoderRunner,
	coder domaincontractsrepository.InfraredStateCoder,
	testCase domaincontractsrepository.InfraredTestCase,
	logger domaincontractslogger.Leveled,
) domainusecasesinfrared.RecordSessionManagement {
	return &usecase{
		session: session, device: device, definition: definition, state: state,
		recordCase: recordCase, broadcaster: broadcaster, subscriptions: subscriptions,
		llmFactory: llmFactory, node: node, encoderRunner: encoderRunner, coder: coder, testCase: testCase, logger: logger,
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
	const tag = "infrared/record_session_management/AcceptRaw"

	if err := u.recordCase.UpdateRawStatusById(ctx, rawId, domainmodels.InfraredRecordRawStatusAccepted, nil); err != nil {
		return err
	}

	raw, err := u.recordCase.GetRawById(ctx, rawId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to look up accepted raw", domainmodels.LoggerMeta{"err": err, "raw_id": rawId})
		return err
	}

	acceptedCount, err := u.recordCase.CountAcceptedRawByCaseId(ctx, raw.InfraredStateDeviceRecordCaseId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to count accepted raws", domainmodels.LoggerMeta{"err": err, "case_id": raw.InfraredStateDeviceRecordCaseId})
		return err
	}
	if acceptedCount < 2 {
		return nil
	}

	if err := u.recordCase.UpdateStatusById(ctx, raw.InfraredStateDeviceRecordCaseId, domainmodels.InfraredRecordCaseStatusAccepted); err != nil {
		u.logger.Error(ctx, tag, "failed to mark case accepted", domainmodels.LoggerMeta{"err": err, "case_id": raw.InfraredStateDeviceRecordCaseId})
		return err
	}

	// InfraredStateDeviceRecordRaw has no session id of its own (only a case
	// id) — look the case up to get the session id from a field that's
	// genuinely backed by a column, rather than adding a synthetic
	// join-only field to the raw model.
	acceptedCase, err := u.recordCase.GetById(ctx, raw.InfraredStateDeviceRecordCaseId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to look up accepted case", domainmodels.LoggerMeta{"err": err, "case_id": raw.InfraredStateDeviceRecordCaseId})
		return err
	}
	if acceptedCase == nil {
		err := domainmodels.NewError("accepted case not found", domainmodels.ErrTypeNotFound, nil)
		u.logger.Error(ctx, tag, "accepted case not found", domainmodels.LoggerMeta{"case_id": raw.InfraredStateDeviceRecordCaseId})
		return err
	}
	sessionId := acceptedCase.InfraredRecordSessionId

	cases, err := u.recordCase.ListBySessionId(ctx, sessionId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list cases for cursor advancement", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		return err
	}

	var next *domainmodels.InfraredStateDeviceRecordCase
	for i := range cases {
		if cases[i].Status != domainmodels.InfraredRecordCaseStatusPending {
			continue
		}
		if next == nil || cases[i].Step < next.Step {
			next = &cases[i]
		}
	}

	if next != nil {
		if err := u.recordCase.UpdateStatusById(ctx, next.Id, domainmodels.InfraredRecordCaseStatusActive); err != nil {
			u.logger.Error(ctx, tag, "failed to activate next case", domainmodels.LoggerMeta{"err": err, "case_id": next.Id})
			return err
		}
		if err := u.session.UpdateCurrentRecordCaseIdById(ctx, sessionId, &next.Id); err != nil {
			u.logger.Error(ctx, tag, "failed to advance session cursor", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
			return err
		}
		u.broadcastBestEffort(ctx, tag, sessionId, domainmodels.InfraredRecordingStateRecording, &next.Id)
		return nil
	}

	if err := u.session.UpdateCurrentRecordCaseIdById(ctx, sessionId, nil); err != nil {
		u.logger.Error(ctx, tag, "failed to clear session cursor before analysis", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		return err
	}
	u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateAnalyzing)

	go u.runAnalysisAndGeneration(sessionId)

	return nil
}

func (u *usecase) broadcastBestEffort(ctx context.Context, tag string, sessionId uuid.UUID, state string, currentCaseId *uuid.UUID) {
	if err := u.broadcaster.Send(ctx, domainmodels.InfraredRecordSessionEvent{SessionId: sessionId, RecordingState: state, CurrentRecordCaseId: currentCaseId}); err != nil {
		u.logger.Warn(ctx, tag, "failed to broadcast", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
	}
}

// runAnalysisAndGeneration is this plan's second background-goroutine job,
// structurally identical to runCaseGeneration above: it runs the
// deterministic bit-analysis pipeline over the session's accepted cases,
// asks the LLM to write an encoder/decoder pair, smoke-tests the encoder
// via goja, and persists the result.
func (u *usecase) runAnalysisAndGeneration(sessionId uuid.UUID) {
	const tag = "infrared/record_session_management/runAnalysisAndGeneration"

	ctx, cancel := context.WithTimeout(context.Background(), analysisTimeout)
	defer cancel()

	session, err := u.session.GetById(ctx, sessionId)
	if err != nil || session == nil {
		u.logger.Error(ctx, tag, "failed to look up session", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	device, err := u.device.GetById(ctx, session.InfraredDeviceId)
	if err != nil || device == nil {
		u.logger.Error(ctx, tag, "failed to look up device", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	states, err := u.state.ListByDeviceTypeId(ctx, device.InfraredDeviceTypeId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list states", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	definitions, err := u.definition.ListByDeviceId(ctx, session.InfraredDeviceId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list definitions", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	cases, err := u.recordCase.ListBySessionId(ctx, sessionId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list cases", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	payload, baselineState, err := u.buildAnalysisPayload(ctx, tag, cases, states, definitions)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to analyze recorded cases", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFunctionGenerating)

	client, err := u.llmFactory.Current(ctx)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to resolve llm client", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	written, err := applicationinfraredcodergeneration.WriteCoder(ctx, client, device.Brand, device.Model, states, definitions, payload)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to write coder", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	// RunEncoder (and the prompt WriteCoder just built) both expect a
	// state-NAME-keyed map (e.g. state.POWER); buildAnalysisPayload's
	// baselineState is UUID-keyed to match its own internal precision needs,
	// so it must be converted before crossing this boundary.
	if _, err := u.encoderRunner.RunEncoder(written.EncoderSource, stateIdToName(states, baselineState), encoderSmokeTestTimeout); err != nil {
		u.logger.Error(ctx, tag, "generated encoder failed its smoke test", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	coderId, err := u.coder.Create(ctx, domainmodels.InfraredStateCoder{
		InfraredDeviceId:        session.InfraredDeviceId,
		InfraredRecordSessionId: sessionId,
		EncoderSource:           written.EncoderSource,
		DecoderSource:           written.DecoderSource,
		SummaryReadme:           written.SummaryReadme,
		DetailReadme:            written.DetailReadme,
	})
	if err != nil {
		u.logger.Error(ctx, tag, "failed to persist coder", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	go u.runTestCaseGeneration(sessionId, coderId)
}

// runTestCaseGeneration is this plan's third background-goroutine job,
// launched right after runAnalysisAndGeneration persists a coder — it asks
// the LLM to propose a minimal test plan for that coder and persists one
// InfraredTestCase (with its target states) per proposed plan.
func (u *usecase) runTestCaseGeneration(sessionId uuid.UUID, coderId uuid.UUID) {
	const tag = "infrared/record_session_management/runTestCaseGeneration"

	ctx, cancel := context.WithTimeout(context.Background(), testCaseGenerationTimeout)
	defer cancel()

	coder, err := u.coder.GetBySessionId(ctx, sessionId)
	if err != nil || coder == nil {
		u.logger.Error(ctx, tag, "failed to look up coder", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	session, err := u.session.GetById(ctx, sessionId)
	if err != nil || session == nil {
		u.logger.Error(ctx, tag, "failed to look up session", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	device, err := u.device.GetById(ctx, session.InfraredDeviceId)
	if err != nil || device == nil {
		u.logger.Error(ctx, tag, "failed to look up device", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	states, err := u.state.ListByDeviceTypeId(ctx, device.InfraredDeviceTypeId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list states", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	definitions, err := u.definition.ListByDeviceId(ctx, session.InfraredDeviceId)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to list definitions", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	client, err := u.llmFactory.Current(ctx)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to resolve llm client", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}
	plans, err := applicationinfraredcodergeneration.WriteTestCases(ctx, client, device.Brand, device.Model, *coder, states, definitions)
	if err != nil {
		u.logger.Error(ctx, tag, "failed to write test cases", domainmodels.LoggerMeta{"err": err, "session_id": sessionId})
		u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
		return
	}

	for i, plan := range plans {
		testCaseStates := make([]domainmodels.InfraredTestCaseState, 0, len(plan.States))
		for stateId, value := range plan.States {
			testCaseStates = append(testCaseStates, domainmodels.InfraredTestCaseState{InfraredStateId: stateId, StateValue: value})
		}
		if _, err := u.testCase.CreateWithStates(ctx, coderId, int32(i+1), plan.Description, testCaseStates); err != nil {
			u.logger.Error(ctx, tag, "failed to persist test case", domainmodels.LoggerMeta{"err": err, "session_id": sessionId, "step": i + 1})
			u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateFailed)
			return
		}
	}

	u.transition(ctx, tag, sessionId, domainmodels.InfraredRecordingStateTesting)
}

func (u *usecase) ListTestCases(ctx context.Context, sessionId uuid.UUID) ([]domainusecasesinfrared.TestCaseWithStates, error) {
	coder, err := u.coder.GetBySessionId(ctx, sessionId)
	if err != nil {
		return nil, err
	}
	if coder == nil {
		return nil, nil
	}
	testCases, err := u.testCase.ListByCoderId(ctx, coder.Id)
	if err != nil {
		return nil, err
	}
	result := make([]domainusecasesinfrared.TestCaseWithStates, 0, len(testCases))
	for _, tc := range testCases {
		states, err := u.testCase.ListStatesByTestCaseId(ctx, tc.Id)
		if err != nil {
			return nil, err
		}
		result = append(result, domainusecasesinfrared.TestCaseWithStates{TestCase: tc, States: states})
	}
	return result, nil
}

// stateIdToName converts a state-id-keyed value map (as produced internally
// by buildAnalysisPayload) into a state-name-keyed map, since RunEncoder and
// WriteCoder's prompt both address states by name (e.g. state.POWER), not
// by their internal UUID.
func stateIdToName(states []domainmodels.InfraredState, byId map[string]string) map[string]string {
	nameById := make(map[string]string, len(states))
	for _, s := range states {
		nameById[s.Id.String()] = s.Name
	}
	byName := make(map[string]string, len(byId))
	for id, value := range byId {
		if name, ok := nameById[id]; ok {
			byName[name] = value
		}
	}
	return byName
}

// buildAnalysisPayload turns every case's accepted raws into the pure
// analysis package's inputs and runs the full frame -> demodulate ->
// volatile -> attribute pipeline. Returns the payload plus the baseline
// case's target state (state id -> value) for the encoder smoke test.
//
// The baseline case is identified by matching its recorded state values
// against applicationinfraredcasegeneration.Baseline(states, definitions),
// NOT by Step == 1: a later re-ordering step (WriteScript, Plan B1) can
// reassign every case's Step based on the LLM's chosen press order, so the
// case that was originally cases[0] can end up at any Step value. Comparing
// against the deterministically-known baseline values is the only reliable
// way to find it from persisted rows alone.
func (u *usecase) buildAnalysisPayload(
	ctx context.Context,
	tag string,
	cases []domainmodels.InfraredStateDeviceRecordCase,
	states []domainmodels.InfraredState,
	definitions []domainmodels.InfraredStateDeviceDefinition,
) (applicationinfraredanalysis.AnalysisPayload, map[string]string, error) {
	expectedBaseline := applicationinfraredcasegeneration.Baseline(states, definitions)

	var baselineBits []int
	baselineState := make(map[string]string)
	caseBits := make(map[uuid.UUID][]int)
	caseTargetState := make(map[uuid.UUID]uuid.UUID)
	caseTargetValue := make(map[uuid.UUID]string)
	var volatileSets []map[int]struct{}

	for _, c := range cases {
		raws, err := u.recordCase.ListRawByCaseId(ctx, c.Id)
		if err != nil {
			return applicationinfraredanalysis.AnalysisPayload{}, nil, err
		}
		var accepted [][]int32
		for _, r := range raws {
			if r.Status == domainmodels.InfraredRecordRawStatusAccepted {
				var durations []int32
				if err := json.Unmarshal(r.RawData, &durations); err != nil {
					return applicationinfraredanalysis.AnalysisPayload{}, nil, err
				}
				accepted = append(accepted, durations)
			}
		}
		if len(accepted) < 2 {
			continue
		}

		frameBitsPerRaw := make([][]int, 0, len(accepted))
		for _, raw := range accepted {
			frames := applicationinfraredanalysis.SegmentFrames(raw)
			if len(frames) == 0 {
				continue
			}
			frameBitsPerRaw = append(frameBitsPerRaw, applicationinfraredanalysis.DemodulateBits(frames[0]))
		}
		if len(frameBitsPerRaw) < 2 {
			continue
		}

		volatile, err := applicationinfraredanalysis.DetectVolatileBits(frameBitsPerRaw[0], frameBitsPerRaw[1])
		if err != nil {
			return applicationinfraredanalysis.AnalysisPayload{}, nil, err
		}
		volatileSets = append(volatileSets, volatile)

		caseStates, err := u.recordCase.ListStatesByCaseId(ctx, c.Id)
		if err != nil {
			return applicationinfraredanalysis.AnalysisPayload{}, nil, err
		}

		if isBaselineCase(expectedBaseline, caseStates) {
			baselineBits = frameBitsPerRaw[0]
			for _, s := range caseStates {
				baselineState[s.InfraredStateId.String()] = s.StateValue
			}
			continue
		}

		caseBits[c.Id] = frameBitsPerRaw[0]
		// OFAT construction (Plan B1, Task 8): a non-baseline case differs
		// from baseline in exactly one state — find it by comparing against
		// the baseline's own recorded state values.
		for _, s := range caseStates {
			if baselineValue, ok := baselineState[s.InfraredStateId.String()]; ok && baselineValue != s.StateValue {
				caseTargetState[c.Id] = s.InfraredStateId
				caseTargetValue[c.Id] = s.StateValue
				break
			}
		}
	}

	if baselineBits == nil {
		return applicationinfraredanalysis.AnalysisPayload{}, nil, domainmodels.NewError(
			"no baseline case with two accepted raws was found for analysis", domainmodels.ErrTypeFailure, nil,
		)
	}

	volatile := applicationinfraredanalysis.UnionVolatileBits(volatileSets...)
	payload, err := applicationinfraredanalysis.Attribute(baselineBits, caseBits, caseTargetState, caseTargetValue, volatile)
	if err != nil {
		return applicationinfraredanalysis.AnalysisPayload{}, nil, err
	}
	return payload, baselineState, nil
}

// isBaselineCase reports whether a case's recorded state values match the
// deterministically-known baseline exactly (same states, same values) —
// the only reliable way to identify the baseline case once Step may have
// been reassigned by a later re-ordering step.
func isBaselineCase(expectedBaseline map[uuid.UUID]string, caseStates []domainmodels.InfraredStateDeviceRecordState) bool {
	if len(caseStates) != len(expectedBaseline) {
		return false
	}
	for _, s := range caseStates {
		if expectedBaseline[s.InfraredStateId] != s.StateValue {
			return false
		}
	}
	return true
}

func (u *usecase) GetCoderBySessionId(ctx context.Context, sessionId uuid.UUID) (*domainmodels.InfraredStateCoder, error) {
	return u.coder.GetBySessionId(ctx, sessionId)
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
