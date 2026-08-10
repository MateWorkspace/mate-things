package applicationinfraredrecordsessionmanagement

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	domaincontractsbroadcaster "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/broadcaster"
	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domaincontractslogger "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/logger"
	domaincontractsnode "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/node"
	domaincontractsrepository "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/repository"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesinfrared "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/infrared"
	"github.com/google/uuid"
)

type fakeSessionRepository struct {
	domaincontractsrepository.InfraredRecordSession
	createdNodeId        uuid.UUID
	createdDeviceId      uuid.UUID
	created              uuid.UUID
	mu                   sync.Mutex
	statusUpdates        []string
	getResult            *domainmodels.InfraredRecordSession
	currentCaseSessionId uuid.UUID
	currentCaseId        *uuid.UUID
}

func (f *fakeSessionRepository) Create(_ context.Context, nodeId uuid.UUID, deviceId uuid.UUID) (uuid.UUID, error) {
	f.createdNodeId, f.createdDeviceId = nodeId, deviceId
	f.created = uuid.New()
	return f.created, nil
}
func (f *fakeSessionRepository) GetById(_ context.Context, id uuid.UUID) (*domainmodels.InfraredRecordSession, error) {
	return f.getResult, nil
}
func (f *fakeSessionRepository) UpdateRecordingStateById(_ context.Context, _ uuid.UUID, state string, _ bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.statusUpdates = append(f.statusUpdates, state)
	return nil
}

// StatusUpdates returns a snapshot, safe to read while the background
// goroutine may still be appending to statusUpdates concurrently.
func (f *fakeSessionRepository) StatusUpdates() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.statusUpdates...)
}
func (f *fakeSessionRepository) UpdateCurrentRecordCaseIdById(_ context.Context, sessionId uuid.UUID, currentRecordCaseId *uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.currentCaseSessionId = sessionId
	f.currentCaseId = currentRecordCaseId
	return nil
}

// CurrentCaseId returns a snapshot, safe to read while the background
// goroutine may still be calling UpdateCurrentRecordCaseIdById concurrently.
func (f *fakeSessionRepository) CurrentCaseId() *uuid.UUID {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.currentCaseId
}
func (f *fakeSessionRepository) GetActiveByNodeId(_ context.Context, _ uuid.UUID) (*domainmodels.InfraredRecordSession, error) {
	return f.getResult, nil
}

type fakeNodeRepository struct {
	domaincontractsrepository.Node
	result *domainmodels.Node
}

func (f *fakeNodeRepository) ReadByDeviceId(_ context.Context, _ string) (*domainmodels.Node, error) {
	return f.result, nil
}

// ReadById is needed because runCaseGeneration looks up the node by
// session.NodeId (not by device id) before subscribing to ir capture.
func (f *fakeNodeRepository) ReadById(_ context.Context, _ uuid.UUID) (*domainmodels.Node, error) {
	return f.result, nil
}

type fakeDeviceRepository struct {
	domaincontractsrepository.InfraredDevice
	created uuid.UUID
}

func (f *fakeDeviceRepository) Create(_ context.Context, _ uuid.UUID, _ string, _ string) (uuid.UUID, error) {
	f.created = uuid.New()
	return f.created, nil
}
func (f *fakeDeviceRepository) GetById(_ context.Context, id uuid.UUID) (*domainmodels.InfraredDevice, error) {
	return &domainmodels.InfraredDevice{Id: id}, nil
}

type fakeDefinitionRepository struct {
	domaincontractsrepository.InfraredStateDeviceDefinition
	createCalls          int
	listByDeviceIdResult []domainmodels.InfraredStateDeviceDefinition
}

func (f *fakeDefinitionRepository) CreateMany(_ context.Context, _ []domainmodels.InfraredStateDeviceDefinition) error {
	f.createCalls++
	return nil
}
func (f *fakeDefinitionRepository) ListByDeviceId(_ context.Context, _ uuid.UUID) ([]domainmodels.InfraredStateDeviceDefinition, error) {
	return f.listByDeviceIdResult, nil
}

type fakeStateRepository struct {
	domaincontractsrepository.InfraredState
	listByDeviceTypeIdResult []domainmodels.InfraredState
}

func (f *fakeStateRepository) ListByDeviceTypeId(_ context.Context, _ uuid.UUID) ([]domainmodels.InfraredState, error) {
	return f.listByDeviceTypeIdResult, nil
}

type fakeEncoderRunner struct {
	err           error
	receivedState map[string]string
}

func (f *fakeEncoderRunner) RunEncoder(_ string, state map[string]string, _ time.Duration) ([]int32, error) {
	f.receivedState = state
	if f.err != nil {
		return nil, f.err
	}
	return []int32{9000, 4500}, nil
}

type fakeCoderRepository struct {
	domaincontractsrepository.InfraredStateCoder
	created      domainmodels.InfraredStateCoder
	createCalled bool
	getResult    *domainmodels.InfraredStateCoder
}

func (f *fakeCoderRepository) Create(_ context.Context, coder domainmodels.InfraredStateCoder) (uuid.UUID, error) {
	f.created = coder
	f.createCalled = true
	return uuid.New(), nil
}
func (f *fakeCoderRepository) GetBySessionId(_ context.Context, _ uuid.UUID) (*domainmodels.InfraredStateCoder, error) {
	return f.getResult, nil
}

type fakeTestCaseRepository struct {
	domaincontractsrepository.InfraredTestCase
	mu                 sync.Mutex
	createdTestCaseIds []uuid.UUID
}

func (f *fakeTestCaseRepository) CreateWithStates(_ context.Context, _ uuid.UUID, _ int32, _ string, _ []domainmodels.InfraredTestCaseState) (uuid.UUID, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id := uuid.New()
	f.createdTestCaseIds = append(f.createdTestCaseIds, id)
	return id, nil
}

// CreatedTestCaseIds returns a snapshot, safe to read while the background
// goroutine may still be appending to createdTestCaseIds concurrently.
func (f *fakeTestCaseRepository) CreatedTestCaseIds() []uuid.UUID {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]uuid.UUID(nil), f.createdTestCaseIds...)
}

type fakeCaseRepository struct {
	domaincontractsrepository.InfraredStateDeviceRecordCase
	mu              sync.Mutex
	createRawCaseId uuid.UUID
	createRawData   []byte
	createRawCalls  int
	createdCaseIds  []uuid.UUID
	statusUpdateIds []uuid.UUID
	getResult       *domainmodels.InfraredStateDeviceRecordCase
	getErr          error

	acceptedRawCounts        map[uuid.UUID]int
	listBySessionIdResult    []domainmodels.InfraredStateDeviceRecordCase
	rawById                  map[uuid.UUID]*domainmodels.InfraredStateDeviceRecordRaw
	statusValuesByCase       map[uuid.UUID][]domainmodels.InfraredRecordCaseStatus
	listRawByCaseIdResult    []domainmodels.InfraredStateDeviceRecordRaw
	listStatesByCaseIdResult []domainmodels.InfraredStateDeviceRecordState
}

func (f *fakeCaseRepository) CreateRaw(_ context.Context, caseId uuid.UUID, rawData []byte) (uuid.UUID, error) {
	f.createRawCaseId, f.createRawData = caseId, rawData
	f.createRawCalls++
	return uuid.New(), nil
}
func (f *fakeCaseRepository) CreateWithStates(_ context.Context, _ uuid.UUID, _ int32, _ string, _ []domainmodels.InfraredStateDeviceRecordState) (uuid.UUID, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id := uuid.New()
	f.createdCaseIds = append(f.createdCaseIds, id)
	return id, nil
}

// CreatedCaseIds returns a snapshot, safe to read while the background
// goroutine may still be appending to createdCaseIds concurrently.
func (f *fakeCaseRepository) CreatedCaseIds() []uuid.UUID {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]uuid.UUID(nil), f.createdCaseIds...)
}
func (f *fakeCaseRepository) UpdateStatusById(_ context.Context, id uuid.UUID, status domainmodels.InfraredRecordCaseStatus) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.statusUpdateIds = append(f.statusUpdateIds, id)
	if f.statusValuesByCase == nil {
		f.statusValuesByCase = make(map[uuid.UUID][]domainmodels.InfraredRecordCaseStatus)
	}
	f.statusValuesByCase[id] = append(f.statusValuesByCase[id], status)
	return nil
}
func (f *fakeCaseRepository) GetById(_ context.Context, _ uuid.UUID) (*domainmodels.InfraredStateDeviceRecordCase, error) {
	return f.getResult, f.getErr
}
func (f *fakeCaseRepository) UpdateRawStatusById(_ context.Context, _ uuid.UUID, _ domainmodels.InfraredRecordRawStatus, _ *string) error {
	return nil
}
func (f *fakeCaseRepository) ListRawByCaseId(_ context.Context, _ uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordRaw, error) {
	return f.listRawByCaseIdResult, nil
}
func (f *fakeCaseRepository) ListStatesByCaseId(_ context.Context, _ uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordState, error) {
	return f.listStatesByCaseIdResult, nil
}
func (f *fakeCaseRepository) CountAcceptedRawByCaseId(_ context.Context, caseId uuid.UUID) (int, error) {
	return f.acceptedRawCounts[caseId], nil
}
func (f *fakeCaseRepository) ListBySessionId(_ context.Context, _ uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordCase, error) {
	return f.listBySessionIdResult, nil
}
func (f *fakeCaseRepository) GetRawById(_ context.Context, rawId uuid.UUID) (*domainmodels.InfraredStateDeviceRecordRaw, error) {
	return f.rawById[rawId], nil
}

// StatusUpdatesForCase returns a snapshot of every status this fake's
// UpdateStatusById has been called with for the given case, in call order —
// mirrors the file's existing StatusUpdates()/CurrentCaseId() convention
// for state a background goroutine might mutate concurrently with a test.
func (f *fakeCaseRepository) StatusUpdatesForCase(caseId uuid.UUID) []domainmodels.InfraredRecordCaseStatus {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]domainmodels.InfraredRecordCaseStatus(nil), f.statusValuesByCase[caseId]...)
}

// fakeLlmClient/fakeLlmClientFactory stand in for the merged prior plan's
// llm.Client/ClientFactory so Start's background goroutine can complete
// (or fail deterministically) without ever making a real provider call.
// The default response text is shaped for WriteScript's case-generation
// callers; runAnalysisAndGeneration's WriteCoder expects a differently
// shaped JSON object, so fakeLlmClientFactory.responseText lets a test
// override it.
type fakeLlmClient struct{ text string }

func (f *fakeLlmClient) GenerateText(_ context.Context, _ domaincontractsllm.GenerateTextRequest) (domaincontractsllm.GenerateTextResult, error) {
	return domaincontractsllm.GenerateTextResult{Text: f.text}, nil
}

type fakeLlmClientFactory struct {
	err           error
	responseText  string   // existing field — single-response tests keep using this
	responseTexts []string // if set, Current() returns these in order, one per call
	callIndex     int
}

func (f *fakeLlmClientFactory) Current(_ context.Context) (domaincontractsllm.Client, error) {
	if f.err != nil {
		return nil, f.err
	}
	if len(f.responseTexts) > 0 {
		i := f.callIndex
		if i >= len(f.responseTexts) {
			i = len(f.responseTexts) - 1
		}
		f.callIndex++
		return &fakeLlmClient{text: f.responseTexts[i]}, nil
	}
	text := f.responseText
	if text == "" {
		text = `[{"case_index": 0, "description": "baseline", "order": 1}]`
	}
	return &fakeLlmClient{text: text}, nil
}

type fakeBroadcaster struct {
	domaincontractsbroadcaster.InfraredRecordSession
	sentEvents []domainmodels.InfraredRecordSessionEvent
}

func (f *fakeBroadcaster) Send(_ context.Context, event domainmodels.InfraredRecordSessionEvent) error {
	f.sentEvents = append(f.sentEvents, event)
	return nil
}

type fakeSubscriptions struct {
	domaincontractsnode.Subscriptions
	mu                  sync.Mutex
	subscribedDeviceIds []string
}

func (f *fakeSubscriptions) IrCapture(_ context.Context, nodeDeviceId string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.subscribedDeviceIds = append(f.subscribedDeviceIds, nodeDeviceId)
	return nil
}

// SubscribedDeviceIds returns a snapshot, safe to read while the background
// goroutine may still be appending to subscribedDeviceIds concurrently.
func (f *fakeSubscriptions) SubscribedDeviceIds() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.subscribedDeviceIds...)
}

// noopLogger is a true no-op, not the embed-and-panic fake convention used
// elsewhere in this file: the implementation logs from inside the
// background goroutine on every failure path (e.g. an intentionally
// failing llmFactory in the tests below), so a logger that panics when
// actually called would crash the test binary asynchronously.
type noopLogger struct{ domaincontractslogger.Leveled }

func (f *noopLogger) Error(_ context.Context, _ string, _ string, _ domainmodels.LoggerMeta) {}
func (f *noopLogger) Warn(_ context.Context, _ string, _ string, _ domainmodels.LoggerMeta)  {}
func (f *noopLogger) Info(_ context.Context, _ string, _ string, _ domainmodels.LoggerMeta)  {}
func (f *noopLogger) Debug(_ context.Context, _ string, _ string, _ domainmodels.LoggerMeta) {}

func TestStartRejectsEmptyBrand(t *testing.T) {
	usecase := NewUsecaseImpl(
		&fakeSessionRepository{}, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &noopLogger{},
	)

	_, err := usecase.Start(context.Background(), domainusecasesinfrared.StartRecordSessionRequest{
		Brand: "", Model: "PAC-09HDN",
	})
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("Start() error = %v, want validation error", err)
	}
}

func TestStartCreatesDeviceDefinitionsAndSessionThenKicksOffCaseGeneration(t *testing.T) {
	sessionRepo := &fakeSessionRepository{}
	deviceRepo := &fakeDeviceRepository{}
	definitionRepo := &fakeDefinitionRepository{}
	// llmFactory must NOT be nil here: Start launches a background goroutine
	// that reaches u.llmFactory.Current(ctx) shortly after this call returns,
	// and calling a method on a nil interface panics — inside a goroutine,
	// that panic crashes the whole test binary, not just this test. Giving
	// it a factory that fails fast (rather than nil) makes the async path
	// end at a harmless FAILED transition instead.
	usecase := NewUsecaseImpl(
		sessionRepo, deviceRepo, definitionRepo,
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{err: errors.New("llm not configured for this test")}, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &noopLogger{},
	)

	nodeId := uuid.New()
	sessionId, err := usecase.Start(context.Background(), domainusecasesinfrared.StartRecordSessionRequest{
		NodeId: nodeId, Brand: "Polytron", Model: "PAC-09HDN",
	})
	if err != nil {
		t.Fatalf("Start() error = %v, want nil", err)
	}
	if sessionId != sessionRepo.created {
		t.Fatalf("Start() returned %v, want the repository's created id %v", sessionId, sessionRepo.created)
	}
	if sessionRepo.createdNodeId != nodeId {
		t.Fatalf("session created with node id %v, want %v", sessionRepo.createdNodeId, nodeId)
	}
	if definitionRepo.createCalls != 1 {
		t.Fatalf("definition repository CreateMany() calls = %d, want 1", definitionRepo.createCalls)
	}

	// The case-generation goroutine runs asynchronously; give it a moment,
	// then confirm the session transitioned into CASES_GENERATING at least.
	// (Task 13's real implementation must make this observable synchronously
	// for the *first* transition — see the "Step 3" note on making the DRAFT
	// -> CASES_GENERATING handoff itself synchronous up to the point the
	// goroutine is launched.)
	statusUpdates := sessionRepo.StatusUpdates()
	found := false
	for _, s := range statusUpdates {
		if s == domainmodels.InfraredRecordingStateCasesGenerating {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include CASES_GENERATING", statusUpdates)
	}
}

func TestStartEventuallySubscribesToIrCaptureAndTransitionsToRecording(t *testing.T) {
	nodeId := uuid.New()
	deviceId := "AC276E5E030C"
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{NodeId: nodeId}}
	caseRepo := &fakeCaseRepository{}
	subscriptions := &fakeSubscriptions{}
	nodeRepo := &fakeNodeRepository{result: &domainmodels.Node{Id: nodeId, DeviceId: deviceId}}

	usecase := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, subscriptions,
		&fakeLlmClientFactory{}, nodeRepo, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &noopLogger{},
	)

	_, err := usecase.Start(context.Background(), domainusecasesinfrared.StartRecordSessionRequest{
		NodeId: nodeId, Brand: "Polytron", Model: "PAC-09HDN",
	})
	if err != nil {
		t.Fatalf("Start() error = %v, want nil", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && len(subscriptions.SubscribedDeviceIds()) == 0 {
		time.Sleep(5 * time.Millisecond)
	}

	subscribedDeviceIds := subscriptions.SubscribedDeviceIds()
	if len(subscribedDeviceIds) != 1 || subscribedDeviceIds[0] != deviceId {
		t.Fatalf("subscribedDeviceIds = %v, want [%q]", subscribedDeviceIds, deviceId)
	}

	statusUpdates := sessionRepo.StatusUpdates()
	found := false
	for _, s := range statusUpdates {
		if s == domainmodels.InfraredRecordingStateRecording {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include RECORDING", statusUpdates)
	}
}

func TestDiscardRawRequiresNonEmptyReason(t *testing.T) {
	usecase := NewUsecaseImpl(
		&fakeSessionRepository{}, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &noopLogger{},
	)

	err := usecase.DiscardRaw(context.Background(), uuid.New(), "")
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("DiscardRaw() error = %v, want validation error", err)
	}
}

func TestCaptureIrRawPersistsRawAgainstCurrentCase(t *testing.T) {
	caseId := uuid.New()
	sessionId := uuid.New()
	nodeId := uuid.New()

	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{
		Id: sessionId, RecordingState: domainmodels.InfraredRecordingStateRecording, CurrentRecordCaseId: &caseId,
	}}
	caseRepo := &fakeCaseRepository{}
	nodeRepo := &fakeNodeRepository{result: &domainmodels.Node{Id: nodeId}}
	broadcaster := &fakeBroadcaster{}

	usecase := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, broadcaster, &fakeSubscriptions{},
		nil, nodeRepo, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &noopLogger{},
	)

	err := usecase.CaptureIrRaw(context.Background(), domainusecasesinfrared.CaptureIrRawRequest{
		NodeDeviceId: "AC276E5E030C", RawData: []int32{9000, 4500, 560, 560},
	})
	if err != nil {
		t.Fatalf("CaptureIrRaw() error = %v, want nil", err)
	}
	if caseRepo.createRawCalls != 1 {
		t.Fatalf("CreateRaw() calls = %d, want 1", caseRepo.createRawCalls)
	}
	if caseRepo.createRawCaseId != caseId {
		t.Fatalf("CreateRaw() case id = %v, want the session's CurrentRecordCaseId %v", caseRepo.createRawCaseId, caseId)
	}
	if len(broadcaster.sentEvents) != 1 || broadcaster.sentEvents[0].SessionId != sessionId {
		t.Fatalf("broadcaster sent events = %+v, want one event for session %v", broadcaster.sentEvents, sessionId)
	}
}

func TestCaptureIrRawDoesNothingWhenNodeUnknown(t *testing.T) {
	caseRepo := &fakeCaseRepository{}
	usecase := NewUsecaseImpl(
		&fakeSessionRepository{}, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{result: nil}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &noopLogger{},
	)

	err := usecase.CaptureIrRaw(context.Background(), domainusecasesinfrared.CaptureIrRawRequest{
		NodeDeviceId: "unknown-device", RawData: []int32{9000},
	})
	if err != nil {
		t.Fatalf("CaptureIrRaw() error = %v, want nil (unknown node is a no-op, not a failure)", err)
	}
	if caseRepo.createRawCalls != 0 {
		t.Fatalf("CreateRaw() calls = %d, want 0", caseRepo.createRawCalls)
	}
}

func TestRunCaseGenerationSetsFirstCaseAsSessionCurrentCase(t *testing.T) {
	nodeId := uuid.New()
	deviceId := "AC276E5E030C"
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{NodeId: nodeId}}
	caseRepo := &fakeCaseRepository{}
	subscriptions := &fakeSubscriptions{}
	nodeRepo := &fakeNodeRepository{result: &domainmodels.Node{Id: nodeId, DeviceId: deviceId}}

	usecase := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, subscriptions,
		&fakeLlmClientFactory{}, nodeRepo, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &noopLogger{},
	)

	_, err := usecase.Start(context.Background(), domainusecasesinfrared.StartRecordSessionRequest{
		NodeId: nodeId, Brand: "Polytron", Model: "PAC-09HDN",
	})
	if err != nil {
		t.Fatalf("Start() error = %v, want nil", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && sessionRepo.CurrentCaseId() == nil {
		time.Sleep(5 * time.Millisecond)
	}

	createdCaseIds := caseRepo.CreatedCaseIds()
	if len(createdCaseIds) == 0 {
		t.Fatalf("no cases were created")
	}
	currentCaseId := sessionRepo.CurrentCaseId()
	if currentCaseId == nil || *currentCaseId != createdCaseIds[0] {
		t.Fatalf("session current case id = %v, want the first created case %v", currentCaseId, createdCaseIds[0])
	}
}

func TestSetCurrentCaseRejectsCaseFromDifferentSession(t *testing.T) {
	sessionId := uuid.New()
	otherSessionId := uuid.New()
	caseId := uuid.New()
	caseRepo := &fakeCaseRepository{getResult: &domainmodels.InfraredStateDeviceRecordCase{
		Id: caseId, InfraredRecordSessionId: otherSessionId,
	}}
	sessionRepo := &fakeSessionRepository{}

	usecase := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &noopLogger{},
	)

	err := usecase.SetCurrentCase(context.Background(), sessionId, caseId)
	if !errors.Is(err, domainmodels.ErrTypeValidation) {
		t.Fatalf("SetCurrentCase() error = %v, want validation error", err)
	}
	if sessionRepo.CurrentCaseId() != nil {
		t.Fatalf("session current case id = %v, want nil (case belongs to a different session)", sessionRepo.CurrentCaseId())
	}
}

func TestSetCurrentCaseSetsSessionCurrentCase(t *testing.T) {
	sessionId := uuid.New()
	caseId := uuid.New()
	caseRepo := &fakeCaseRepository{getResult: &domainmodels.InfraredStateDeviceRecordCase{
		Id: caseId, InfraredRecordSessionId: sessionId,
	}}
	sessionRepo := &fakeSessionRepository{}
	broadcaster := &fakeBroadcaster{}

	usecase := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, broadcaster, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &noopLogger{},
	)

	if err := usecase.SetCurrentCase(context.Background(), sessionId, caseId); err != nil {
		t.Fatalf("SetCurrentCase() error = %v, want nil", err)
	}

	currentCaseId := sessionRepo.CurrentCaseId()
	if currentCaseId == nil || *currentCaseId != caseId {
		t.Fatalf("session current case id = %v, want %v", currentCaseId, caseId)
	}
	if len(caseRepo.statusUpdateIds) != 1 || caseRepo.statusUpdateIds[0] != caseId {
		t.Fatalf("case status updates = %v, want [%v]", caseRepo.statusUpdateIds, caseId)
	}
	if len(broadcaster.sentEvents) != 1 || broadcaster.sentEvents[0].CurrentRecordCaseId == nil || *broadcaster.sentEvents[0].CurrentRecordCaseId != caseId {
		t.Fatalf("broadcaster sent events = %+v, want one event with current case id %v", broadcaster.sentEvents, caseId)
	}
}

func TestRetryCaseSetsCurrentCase(t *testing.T) {
	sessionId := uuid.New()
	caseId := uuid.New()
	caseRepo := &fakeCaseRepository{getResult: &domainmodels.InfraredStateDeviceRecordCase{
		Id: caseId, InfraredRecordSessionId: sessionId,
	}}
	sessionRepo := &fakeSessionRepository{}

	usecase := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &noopLogger{},
	)

	if err := usecase.RetryCase(context.Background(), caseId); err != nil {
		t.Fatalf("RetryCase() error = %v, want nil", err)
	}

	currentCaseId := sessionRepo.CurrentCaseId()
	if currentCaseId == nil || *currentCaseId != caseId {
		t.Fatalf("session current case id = %v, want %v", currentCaseId, caseId)
	}
	if sessionRepo.currentCaseSessionId != sessionId {
		t.Fatalf("session id used for current case update = %v, want %v", sessionRepo.currentCaseSessionId, sessionId)
	}
}

func TestAcceptRawAdvancesCursorToNextPendingCase(t *testing.T) {
	sessionId := uuid.New()
	firstCaseId, secondCaseId := uuid.New(), uuid.New()
	rawId := uuid.New()

	sessionRepo := &fakeSessionRepository{}
	caseRepo := &fakeCaseRepository{
		getResult:         &domainmodels.InfraredStateDeviceRecordCase{Id: firstCaseId, InfraredRecordSessionId: sessionId},
		acceptedRawCounts: map[uuid.UUID]int{firstCaseId: 2},
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: firstCaseId, InfraredRecordSessionId: sessionId, Step: 1, Status: domainmodels.InfraredRecordCaseStatusActive},
			{Id: secondCaseId, InfraredRecordSessionId: sessionId, Step: 2, Status: domainmodels.InfraredRecordCaseStatusPending},
		},
		rawById: map[uuid.UUID]*domainmodels.InfraredStateDeviceRecordRaw{
			rawId: {Id: rawId, InfraredStateDeviceRecordCaseId: firstCaseId},
		},
	}
	usecase := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &noopLogger{},
	)

	if err := usecase.AcceptRaw(context.Background(), rawId); err != nil {
		t.Fatalf("AcceptRaw() error = %v, want nil", err)
	}

	firstUpdates := caseRepo.StatusUpdatesForCase(firstCaseId)
	if len(firstUpdates) == 0 || firstUpdates[len(firstUpdates)-1] != domainmodels.InfraredRecordCaseStatusAccepted {
		t.Fatalf("first case status updates = %v, want to end with ACCEPTED", firstUpdates)
	}
	secondUpdates := caseRepo.StatusUpdatesForCase(secondCaseId)
	if len(secondUpdates) == 0 || secondUpdates[len(secondUpdates)-1] != domainmodels.InfraredRecordCaseStatusActive {
		t.Fatalf("second case status updates = %v, want to end with ACTIVE", secondUpdates)
	}
	if got := sessionRepo.CurrentCaseId(); got == nil || *got != secondCaseId {
		t.Fatalf("session current case = %v, want %v", got, secondCaseId)
	}
}

func TestAcceptRawTriggersAnalyzingWhenNoPendingCaseRemains(t *testing.T) {
	sessionId := uuid.New()
	onlyCaseId := uuid.New()
	rawId := uuid.New()

	sessionRepo := &fakeSessionRepository{}
	caseRepo := &fakeCaseRepository{
		getResult:         &domainmodels.InfraredStateDeviceRecordCase{Id: onlyCaseId, InfraredRecordSessionId: sessionId},
		acceptedRawCounts: map[uuid.UUID]int{onlyCaseId: 2},
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: onlyCaseId, InfraredRecordSessionId: sessionId, Step: 1, Status: domainmodels.InfraredRecordCaseStatusActive},
		},
		rawById: map[uuid.UUID]*domainmodels.InfraredStateDeviceRecordRaw{
			rawId: {Id: rawId, InfraredStateDeviceRecordCaseId: onlyCaseId},
		},
	}
	// runAnalysisAndGeneration is still Task 9's no-op stub at this point in
	// the plan (Task 10 fills it in), so llmFactory is never actually
	// called by the goroutine AcceptRaw launches — passed anyway for
	// forward-compatibility with Task 10's real body.
	usecase := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{err: errors.New("llm not configured for this test")}, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &noopLogger{},
	)

	if err := usecase.AcceptRaw(context.Background(), rawId); err != nil {
		t.Fatalf("AcceptRaw() error = %v, want nil", err)
	}

	found := false
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateAnalyzing {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include ANALYZING", sessionRepo.StatusUpdates())
	}
}

func TestAcceptRawDoesNotAdvanceCursorBeforeSecondRawAccepted(t *testing.T) {
	sessionId := uuid.New()
	caseId := uuid.New()
	rawId := uuid.New()

	sessionRepo := &fakeSessionRepository{}
	caseRepo := &fakeCaseRepository{
		getResult:         &domainmodels.InfraredStateDeviceRecordCase{Id: caseId, InfraredRecordSessionId: sessionId},
		acceptedRawCounts: map[uuid.UUID]int{caseId: 1},
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: caseId, InfraredRecordSessionId: sessionId, Step: 1, Status: domainmodels.InfraredRecordCaseStatusActive},
		},
		rawById: map[uuid.UUID]*domainmodels.InfraredStateDeviceRecordRaw{
			rawId: {Id: rawId, InfraredStateDeviceRecordCaseId: caseId},
		},
	}
	usecase := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &noopLogger{},
	)

	if err := usecase.AcceptRaw(context.Background(), rawId); err != nil {
		t.Fatalf("AcceptRaw() error = %v, want nil", err)
	}
	if updates := caseRepo.StatusUpdatesForCase(caseId); len(updates) != 0 {
		t.Fatalf("case status updates = %v, want none (only 1 of 2 required raws accepted)", updates)
	}
	if got := sessionRepo.CurrentCaseId(); got != nil {
		t.Fatalf("session current case = %v, want nil (no change before the case's second raw is accepted)", got)
	}
}

func TestRunAnalysisAndGenerationPersistsCoderAndAdvancesToTestCaseGeneration(t *testing.T) {
	sessionId := uuid.New()
	powerId := uuid.New()

	// A single-bit "header + one bit" capture (matches the fixtures already
	// established in the analysis package's own tests): header(9000,4500)
	// then one data bit with a short space (=> bit 0). Two identical
	// accepted raws so DetectVolatileBits finds no volatile bits.
	rawBytes, err := json.Marshal([]int32{9000, 4500, 560, 560})
	if err != nil {
		t.Fatalf("failed to marshal fixture raw data: %v", err)
	}
	baselineCaseId := uuid.New()

	// runAnalysisAndGeneration looks the session up by id before anything
	// else; fakeSessionRepository{}'s default getResult is nil, which the
	// implementation treats as "session not found" and fails fast on — give
	// it a real session so the test actually exercises the analysis/coder
	// pipeline instead of bailing out on the very first lookup.
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, InfraredDeviceId: uuid.New()}}
	stateRepo := &fakeStateRepository{listByDeviceTypeIdResult: []domainmodels.InfraredState{{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}}
	definitionRepo := &fakeDefinitionRepository{listByDeviceIdResult: []domainmodels.InfraredStateDeviceDefinition{
		{InfraredStateId: powerId, Options: []string{"ON", "OFF"}},
	}}
	caseRepo := &fakeCaseRepository{
		// The case's Step is deliberately NOT 1 — the baseline must be
		// identified by its recorded state values matching
		// applicationinfraredcasegeneration.Baseline, never by Step, since
		// an earlier LLM-driven re-ordering step can reassign Step away
		// from 1 for the true baseline case.
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: baselineCaseId, InfraredRecordSessionId: sessionId, Step: 3, Status: domainmodels.InfraredRecordCaseStatusAccepted},
		},
		listRawByCaseIdResult: []domainmodels.InfraredStateDeviceRecordRaw{
			{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: rawBytes, Status: domainmodels.InfraredRecordRawStatusAccepted},
			{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: rawBytes, Status: domainmodels.InfraredRecordRawStatusAccepted},
		},
		listStatesByCaseIdResult: []domainmodels.InfraredStateDeviceRecordState{
			{InfraredStateDeviceRecordCaseId: baselineCaseId, InfraredStateId: powerId, StateValue: "ON"},
		},
	}
	// GetBySessionId backs runTestCaseGeneration's own coder lookup (launched
	// as a goroutine right after runAnalysisAndGeneration persists this same
	// coder) — must be non-nil or that job fails fast before ever reaching
	// WriteTestCases.
	coderRepo := &fakeCoderRepository{getResult: &domainmodels.InfraredStateCoder{Id: uuid.New(), SummaryReadme: "summary", DetailReadme: "detail"}}
	// WriteCoder expects a JSON object shaped like coderResponse, then
	// WriteTestCases (runTestCaseGeneration, called next) expects a JSON
	// array — fakeLlmClientFactory.responseTexts feeds them in that order.
	llmFactory := &fakeLlmClientFactory{responseTexts: []string{
		`{"encoder_source": "function encode(state) { return [9000, 4500]; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "summary", "detail_readme": "detail"}`,
		`[{"description": "d1", "states": {"POWER": "ON"}}]`,
	}}
	encoderRunner := &fakeEncoderRunner{}
	testCaseRepo := &fakeTestCaseRepository{}

	// NewUsecaseImpl returns the domainusecasesinfrared.RecordSessionManagement
	// interface, but runAnalysisAndGeneration is unexported — assert back to
	// the concrete *usecase type to call it directly. Name this "impl", not
	// "usecase": a local variable named "usecase" would shadow the package's
	// own "usecase" struct type, making a later `.(*usecase)` assertion a
	// compile error (the identifier would resolve to the variable, not the
	// type, in that scope).
	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, definitionRepo,
		stateRepo, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, encoderRunner, coderRepo, testCaseRepo, &noopLogger{},
	).(*usecase)

	impl.runAnalysisAndGeneration(sessionId)

	if !coderRepo.createCalled {
		t.Fatal("coder repository Create() was never called, want the analysis pipeline to produce and persist a coder")
	}
	if coderRepo.created.InfraredRecordSessionId != sessionId {
		t.Fatalf("created coder session id = %v, want %v", coderRepo.created.InfraredRecordSessionId, sessionId)
	}
	if coderRepo.created.EncoderSource == "" {
		t.Fatal("created coder has empty EncoderSource")
	}

	// The encoder smoke test must receive a state-NAME-keyed map (e.g.
	// "POWER"), not the UUID-keyed map buildAnalysisPayload builds
	// internally — this was a real bug caught during implementation.
	wantState := map[string]string{"POWER": "ON"}
	if !reflect.DeepEqual(encoderRunner.receivedState, wantState) {
		t.Fatalf("RunEncoder() received state = %v, want %v (name-keyed, not uuid-keyed)", encoderRunner.receivedState, wantState)
	}

	// runAnalysisAndGeneration now launches runTestCaseGeneration as a
	// goroutine right after persisting the coder — give it a moment to reach
	// its own terminal transition before asserting on it.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && len(testCaseRepo.CreatedTestCaseIds()) == 0 {
		time.Sleep(5 * time.Millisecond)
	}

	found := false
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateFunctionGenerating {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include FUNCTION_GENERATING", sessionRepo.StatusUpdates())
	}

	found = false
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateTesting {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include TESTING (runTestCaseGeneration closes the FUNCTION_GENERATING -> TEST_CASES_GENERATING boundary)", sessionRepo.StatusUpdates())
	}
}

func TestRunAnalysisAndGenerationFailsSessionWhenEncoderThrows(t *testing.T) {
	sessionId := uuid.New()
	powerId := uuid.New()

	rawBytes, err := json.Marshal([]int32{9000, 4500, 560, 560})
	if err != nil {
		t.Fatalf("failed to marshal fixture raw data: %v", err)
	}
	baselineCaseId := uuid.New()

	// See the sibling test above for why getResult must be non-nil: without
	// it this test would still end in FAILED, but for the wrong reason (the
	// session lookup itself failing rather than the encoder smoke test).
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, InfraredDeviceId: uuid.New()}}
	stateRepo := &fakeStateRepository{listByDeviceTypeIdResult: []domainmodels.InfraredState{{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}}
	definitionRepo := &fakeDefinitionRepository{listByDeviceIdResult: []domainmodels.InfraredStateDeviceDefinition{
		{InfraredStateId: powerId, Options: []string{"ON", "OFF"}},
	}}
	caseRepo := &fakeCaseRepository{
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: baselineCaseId, InfraredRecordSessionId: sessionId, Step: 1, Status: domainmodels.InfraredRecordCaseStatusAccepted},
		},
		listRawByCaseIdResult: []domainmodels.InfraredStateDeviceRecordRaw{
			{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: rawBytes, Status: domainmodels.InfraredRecordRawStatusAccepted},
			{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: rawBytes, Status: domainmodels.InfraredRecordRawStatusAccepted},
		},
		listStatesByCaseIdResult: []domainmodels.InfraredStateDeviceRecordState{
			{InfraredStateDeviceRecordCaseId: baselineCaseId, InfraredStateId: powerId, StateValue: "ON"},
		},
	}
	coderRepo := &fakeCoderRepository{}
	// Must be a coder-shaped response so WriteCoder actually succeeds and
	// the pipeline reaches the encoder smoke test — a bare
	// &fakeLlmClientFactory{} defaults to the case-generation JSON array
	// shape, which fails to unmarshal here and returns before RunEncoder is
	// ever called, making the "encoder throws" scenario this test names
	// never actually happen.
	llmFactory := &fakeLlmClientFactory{responseText: `{"encoder_source": "function encode(state) { return [9000, 4500]; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "summary", "detail_readme": "detail"}`}
	encoderRunner := &fakeEncoderRunner{err: errors.New("encoder threw")}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, definitionRepo,
		stateRepo, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, encoderRunner, coderRepo, &fakeTestCaseRepository{}, &noopLogger{},
	).(*usecase)

	impl.runAnalysisAndGeneration(sessionId)

	if coderRepo.createCalled {
		t.Fatal("coder repository Create() was called despite the encoder smoke test failing, want no persisted row")
	}
	found := false
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateFailed {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include FAILED", sessionRepo.StatusUpdates())
	}
}

func TestGetCoderBySessionIdDelegatesToRepository(t *testing.T) {
	sessionId := uuid.New()
	want := &domainmodels.InfraredStateCoder{Id: uuid.New(), InfraredRecordSessionId: sessionId}
	coderRepo := &fakeCoderRepository{getResult: want}
	usecase := NewUsecaseImpl(
		&fakeSessionRepository{}, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, &fakeTestCaseRepository{}, &noopLogger{},
	)

	got, err := usecase.GetCoderBySessionId(context.Background(), sessionId)
	if err != nil || got != want {
		t.Fatalf("GetCoderBySessionId() = %v, %v, want %v, nil", got, err, want)
	}
}

func TestRunTestCaseGenerationPersistsOneTestCasePerPlanAndTransitionsToTesting(t *testing.T) {
	sessionId := uuid.New()
	coderId := uuid.New()
	powerId := uuid.New()

	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, InfraredDeviceId: uuid.New()}}
	stateRepo := &fakeStateRepository{listByDeviceTypeIdResult: []domainmodels.InfraredState{{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum}}}
	coderRepo := &fakeCoderRepository{getResult: &domainmodels.InfraredStateCoder{Id: coderId, SummaryReadme: "s", DetailReadme: "d"}}
	testCaseRepo := &fakeTestCaseRepository{}
	llmFactory := &fakeLlmClientFactory{responseText: `[{"description": "d1", "states": {"POWER": "ON"}}, {"description": "d2", "states": {"POWER": "OFF"}}]`}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		stateRepo, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, testCaseRepo, &noopLogger{},
	).(*usecase)

	impl.runTestCaseGeneration(sessionId, coderId)

	if len(testCaseRepo.CreatedTestCaseIds()) != 2 {
		t.Fatalf("created test case ids = %v, want 2", testCaseRepo.CreatedTestCaseIds())
	}
	found := false
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateTesting {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include TESTING", sessionRepo.StatusUpdates())
	}
}

func TestRunTestCaseGenerationFailsSessionWhenLlmErrors(t *testing.T) {
	sessionId := uuid.New()
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, InfraredDeviceId: uuid.New()}}
	coderRepo := &fakeCoderRepository{getResult: &domainmodels.InfraredStateCoder{Id: uuid.New()}}
	testCaseRepo := &fakeTestCaseRepository{}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{err: errors.New("llm down")}, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, testCaseRepo, &noopLogger{},
	).(*usecase)

	impl.runTestCaseGeneration(sessionId, uuid.New())

	if len(testCaseRepo.CreatedTestCaseIds()) != 0 {
		t.Fatal("test cases were created despite the llm call failing")
	}
	found := false
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateFailed {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include FAILED", sessionRepo.StatusUpdates())
	}
}
