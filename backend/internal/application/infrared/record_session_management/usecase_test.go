package applicationinfraredrecordsessionmanagement

import (
	"context"
	"errors"
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
	createdNodeId   uuid.UUID
	createdDeviceId uuid.UUID
	created         uuid.UUID
	mu              sync.Mutex
	statusUpdates   []string
	getResult       *domainmodels.InfraredRecordSession
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
func (f *fakeSessionRepository) UpdateCurrentRecordCaseIdById(_ context.Context, _ uuid.UUID, _ *uuid.UUID) error {
	return nil
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
	createCalls int
}

func (f *fakeDefinitionRepository) CreateMany(_ context.Context, _ []domainmodels.InfraredStateDeviceDefinition) error {
	f.createCalls++
	return nil
}
func (f *fakeDefinitionRepository) ListByDeviceId(_ context.Context, _ uuid.UUID) ([]domainmodels.InfraredStateDeviceDefinition, error) {
	return nil, nil
}

type fakeStateRepository struct {
	domaincontractsrepository.InfraredState
}

func (f *fakeStateRepository) ListByDeviceTypeId(_ context.Context, _ uuid.UUID) ([]domainmodels.InfraredState, error) {
	return nil, nil
}

type fakeCaseRepository struct {
	domaincontractsrepository.InfraredStateDeviceRecordCase
	createRawCaseId uuid.UUID
	createRawData   []byte
	createRawCalls  int
}

func (f *fakeCaseRepository) CreateRaw(_ context.Context, caseId uuid.UUID, rawData []byte) (uuid.UUID, error) {
	f.createRawCaseId, f.createRawData = caseId, rawData
	f.createRawCalls++
	return uuid.New(), nil
}
func (f *fakeCaseRepository) CreateWithStates(_ context.Context, _ uuid.UUID, _ int32, _ string, _ []domainmodels.InfraredStateDeviceRecordState) (uuid.UUID, error) {
	return uuid.New(), nil
}

// fakeLlmClient/fakeLlmClientFactory stand in for the merged prior plan's
// llm.Client/ClientFactory so Start's background goroutine can complete
// (or fail deterministically) without ever making a real provider call.
type fakeLlmClient struct{}

func (f *fakeLlmClient) GenerateText(_ context.Context, _ domaincontractsllm.GenerateTextRequest) (domaincontractsllm.GenerateTextResult, error) {
	return domaincontractsllm.GenerateTextResult{Text: `[{"case_index": 0, "description": "baseline", "order": 1}]`}, nil
}

type fakeLlmClientFactory struct{ err error }

func (f *fakeLlmClientFactory) Current(_ context.Context) (domaincontractsllm.Client, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &fakeLlmClient{}, nil
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
		nil, &fakeNodeRepository{}, &noopLogger{},
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
		&fakeLlmClientFactory{err: errors.New("llm not configured for this test")}, &fakeNodeRepository{}, &noopLogger{},
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
		&fakeLlmClientFactory{}, nodeRepo, &noopLogger{},
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
		nil, &fakeNodeRepository{}, &noopLogger{},
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
		nil, nodeRepo, &noopLogger{},
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
		nil, &fakeNodeRepository{result: nil}, &noopLogger{},
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
