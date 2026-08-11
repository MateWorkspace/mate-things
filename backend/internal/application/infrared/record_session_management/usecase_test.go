package applicationinfraredrecordsessionmanagement

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
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
	// checksumClarificationMarks counts MarkChecksumClarificationUsedById
	// calls, so tests can assert the once-per-session cap is actually
	// written (and, on the already-used fixture, never written again).
	checksumClarificationMarks int
	deleteErr                  error
	deletedId                  uuid.UUID
	deletedBy                  *uuid.UUID
}

func (f *fakeSessionRepository) Create(_ context.Context, nodeId uuid.UUID, deviceId uuid.UUID, _ *uuid.UUID) (uuid.UUID, error) {
	f.createdNodeId, f.createdDeviceId = nodeId, deviceId
	f.created = uuid.New()
	return f.created, nil
}
func (f *fakeSessionRepository) ReadById(_ context.Context, id uuid.UUID) (*domainmodels.InfraredRecordSession, error) {
	return f.getResult, nil
}
func (f *fakeSessionRepository) UpdateRecordingStateById(_ context.Context, _ uuid.UUID, state string, _ bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.statusUpdates = append(f.statusUpdates, state)
	return nil
}
func (f *fakeSessionRepository) DeleteById(_ context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	f.deletedId, f.deletedBy = id, deletedBy
	return f.deleteErr
}
func (f *fakeSessionRepository) MarkChecksumClarificationUsedById(_ context.Context, _ uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.checksumClarificationMarks++
	return nil
}

// ChecksumClarificationMarks returns a snapshot of how many times the
// once-per-session clarification marker was written — mirrors this file's
// existing StatusUpdates() convention for state a background goroutine
// might mutate concurrently with a test.
func (f *fakeSessionRepository) ChecksumClarificationMarks() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.checksumClarificationMarks
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
func (f *fakeSessionRepository) ReadActiveByNodeId(_ context.Context, _ uuid.UUID) (*domainmodels.InfraredRecordSession, error) {
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

func (f *fakeDeviceRepository) Create(_ context.Context, _ uuid.UUID, _ string, _ string, _ *uuid.UUID) (uuid.UUID, error) {
	f.created = uuid.New()
	return f.created, nil
}
func (f *fakeDeviceRepository) ReadById(_ context.Context, id uuid.UUID) (*domainmodels.InfraredDevice, error) {
	return &domainmodels.InfraredDevice{Id: id}, nil
}
func (f *fakeDeviceRepository) DeleteById(_ context.Context, _ uuid.UUID, _ *uuid.UUID) error {
	return nil
}

type fakeDefinitionRepository struct {
	domaincontractsrepository.InfraredStateDeviceDefinition
	createCalls          int
	listByDeviceIdResult []domainmodels.InfraredStateDeviceDefinition
}

func (f *fakeDefinitionRepository) CreateMany(_ context.Context, _ []domainmodels.InfraredStateDeviceDefinition, _ *uuid.UUID) error {
	f.createCalls++
	return nil
}
func (f *fakeDefinitionRepository) ReadListByDeviceId(_ context.Context, _ uuid.UUID) ([]domainmodels.InfraredStateDeviceDefinition, error) {
	return f.listByDeviceIdResult, nil
}
func (f *fakeDefinitionRepository) DeleteById(_ context.Context, _ uuid.UUID, _ *uuid.UUID) error {
	return nil
}

type fakeStateRepository struct {
	domaincontractsrepository.InfraredState
	listByDeviceTypeIdResult []domainmodels.InfraredState
}

func (f *fakeStateRepository) ReadListByDeviceTypeId(_ context.Context, _ uuid.UUID) ([]domainmodels.InfraredState, error) {
	return f.listByDeviceTypeIdResult, nil
}

type fakeEncoderRunner struct {
	err           error
	receivedState map[string]string
	raw           []int32            // if set, always returned regardless of state (repair-loop tests)
	byState       map[string][]int32 // if set, looked up by state["POWER"]+"|"+state["MODE"]
	bySource      map[string][]int32 // if set, looked up by the encoder source itself — lets a test give a repaired encoder different behavior from the broken one it replaced
	bySourceState map[string][]int32 // if set, looked up by source+"|"+state["POWER"]+"|"+state["MODE"] — for tests needing per-state behavior that also CHANGES across a repair round
}

func (f *fakeEncoderRunner) RunEncoder(source string, state map[string]string, _ time.Duration) ([]int32, error) {
	f.receivedState = state
	if f.err != nil {
		return nil, f.err
	}
	if f.bySourceState != nil {
		if raw, ok := f.bySourceState[source+"|"+state["POWER"]+"|"+state["MODE"]]; ok {
			return raw, nil
		}
	}
	if f.bySource != nil {
		if raw, ok := f.bySource[source]; ok {
			return raw, nil
		}
	}
	if f.byState != nil {
		if raw, ok := f.byState[state["POWER"]+"|"+state["MODE"]]; ok {
			return raw, nil
		}
	}
	if f.raw != nil {
		return f.raw, nil
	}
	return []int32{9000, 4500, 560, 560}, nil
}

type fakeCoderRepository struct {
	domaincontractsrepository.InfraredStateCoder
	created       domainmodels.InfraredStateCoder
	createCalled  bool
	getResult     *domainmodels.InfraredStateCoder
	getByIdResult *domainmodels.InfraredStateCoder
	activateCalls int
	deleteErr     error
	deletedId     uuid.UUID
	deletedBy     *uuid.UUID
}

func (f *fakeCoderRepository) Create(_ context.Context, coder domainmodels.InfraredStateCoder) (uuid.UUID, error) {
	f.created = coder
	f.createCalled = true
	return uuid.New(), nil
}
func (f *fakeCoderRepository) ReadBySessionId(_ context.Context, _ uuid.UUID) (*domainmodels.InfraredStateCoder, error) {
	return f.getResult, nil
}
func (f *fakeCoderRepository) ReadById(_ context.Context, _ uuid.UUID) (*domainmodels.InfraredStateCoder, error) {
	return f.getByIdResult, nil
}
func (f *fakeCoderRepository) Activate(_ context.Context, _ uuid.UUID, _ uuid.UUID) error {
	f.activateCalls++
	return nil
}
func (f *fakeCoderRepository) DeleteById(_ context.Context, id uuid.UUID, deletedBy *uuid.UUID) error {
	f.deletedId, f.deletedBy = id, deletedBy
	return f.deleteErr
}

type fakeTestCaseRepository struct {
	domaincontractsrepository.InfraredTestCase
	mu                           sync.Mutex
	createdTestCaseIds           []uuid.UUID
	getResult                    *domainmodels.InfraredTestCase
	listStatesByTestCaseIdResult []domainmodels.InfraredTestCaseState
	listByCoderIdResult          []domainmodels.InfraredTestCase
	deleteErr                    error
	deletedId                    uuid.UUID
	deleteStateErr               error
	deletedStateId               uuid.UUID
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
func (f *fakeTestCaseRepository) ReadById(_ context.Context, _ uuid.UUID) (*domainmodels.InfraredTestCase, error) {
	return f.getResult, nil
}
func (f *fakeTestCaseRepository) ReadListStatesByTestCaseId(_ context.Context, _ uuid.UUID) ([]domainmodels.InfraredTestCaseState, error) {
	return f.listStatesByTestCaseIdResult, nil
}
func (f *fakeTestCaseRepository) ReadListByCoderId(_ context.Context, _ uuid.UUID) ([]domainmodels.InfraredTestCase, error) {
	return f.listByCoderIdResult, nil
}
func (f *fakeTestCaseRepository) UpdateStatusById(_ context.Context, _ uuid.UUID, _ domainmodels.InfraredTestCaseStatus) error {
	return nil
}
func (f *fakeTestCaseRepository) DeleteById(_ context.Context, id uuid.UUID, _ *uuid.UUID) error {
	f.deletedId = id
	return f.deleteErr
}
func (f *fakeTestCaseRepository) DeleteStateById(_ context.Context, id uuid.UUID, _ *uuid.UUID) error {
	f.deletedStateId = id
	return f.deleteStateErr
}

type fakePublish struct {
	domaincontractsnode.Publish
	mu                 sync.Mutex
	transmittedRawData []int32
	transmitCalls      int
}

func (f *fakePublish) IrTransmit(_ context.Context, _ string, _ uuid.UUID, rawData []int32) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.transmittedRawData = rawData
	f.transmitCalls++
	return nil
}

type fakeCaseRepository struct {
	domaincontractsrepository.InfraredStateDeviceRecordCase
	mu              sync.Mutex
	createRawCaseId uuid.UUID
	createRawData   []byte
	createRawCalls  int
	createdCaseIds  []uuid.UUID
	createdSteps    []int32
	statusUpdateIds []uuid.UUID
	getResult       *domainmodels.InfraredStateDeviceRecordCase
	getErr          error

	acceptedRawCounts        map[uuid.UUID]int
	listBySessionIdResult    []domainmodels.InfraredStateDeviceRecordCase
	rawById                  map[uuid.UUID]*domainmodels.InfraredStateDeviceRecordRaw
	statusValuesByCase       map[uuid.UUID][]domainmodels.InfraredRecordCaseStatus
	listRawByCaseIdResult    []domainmodels.InfraredStateDeviceRecordRaw
	listStatesByCaseIdResult []domainmodels.InfraredStateDeviceRecordState
	listRawByCaseId          map[uuid.UUID][]domainmodels.InfraredStateDeviceRecordRaw
	listStatesByCaseId       map[uuid.UUID][]domainmodels.InfraredStateDeviceRecordState
	deleteErr                error
	deletedId                uuid.UUID
	deleteStateErr           error
	deletedStateId           uuid.UUID
	deleteRawErr             error
	deletedRawId             uuid.UUID
}

func (f *fakeCaseRepository) CreateRaw(_ context.Context, caseId uuid.UUID, rawData []byte) (uuid.UUID, error) {
	f.createRawCaseId, f.createRawData = caseId, rawData
	f.createRawCalls++
	return uuid.New(), nil
}
func (f *fakeCaseRepository) CreateWithStates(_ context.Context, _ uuid.UUID, step int32, _ string, _ []domainmodels.InfraredStateDeviceRecordState) (uuid.UUID, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id := uuid.New()
	f.createdCaseIds = append(f.createdCaseIds, id)
	f.createdSteps = append(f.createdSteps, step)
	return id, nil
}

// CreatedCaseIds returns a snapshot, safe to read while the background
// goroutine may still be appending to createdCaseIds concurrently.
func (f *fakeCaseRepository) CreatedCaseIds() []uuid.UUID {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]uuid.UUID(nil), f.createdCaseIds...)
}

// CreatedSteps returns a snapshot, safe to read while the background
// goroutine may still be appending to createdSteps concurrently — mirrors
// CreatedCaseIds()'s existing convention.
func (f *fakeCaseRepository) CreatedSteps() []int32 {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]int32(nil), f.createdSteps...)
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
func (f *fakeCaseRepository) ReadById(_ context.Context, _ uuid.UUID) (*domainmodels.InfraredStateDeviceRecordCase, error) {
	return f.getResult, f.getErr
}
func (f *fakeCaseRepository) UpdateRawStatusById(_ context.Context, _ uuid.UUID, _ domainmodels.InfraredRecordRawStatus, _ *string) error {
	return nil
}
func (f *fakeCaseRepository) DeleteById(_ context.Context, id uuid.UUID, _ *uuid.UUID) error {
	f.deletedId = id
	return f.deleteErr
}
func (f *fakeCaseRepository) DeleteStateById(_ context.Context, id uuid.UUID, _ *uuid.UUID) error {
	f.deletedStateId = id
	return f.deleteStateErr
}
func (f *fakeCaseRepository) DeleteRawById(_ context.Context, id uuid.UUID, _ *uuid.UUID) error {
	f.deletedRawId = id
	return f.deleteRawErr
}
func (f *fakeCaseRepository) ReadListRawByCaseId(_ context.Context, caseId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordRaw, error) {
	if f.listRawByCaseId != nil {
		return f.listRawByCaseId[caseId], nil
	}
	return f.listRawByCaseIdResult, nil
}
func (f *fakeCaseRepository) ReadListStatesByCaseId(_ context.Context, caseId uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordState, error) {
	if f.listStatesByCaseId != nil {
		return f.listStatesByCaseId[caseId], nil
	}
	return f.listStatesByCaseIdResult, nil
}
func (f *fakeCaseRepository) CountAcceptedRawByCaseId(_ context.Context, caseId uuid.UUID) (int, error) {
	return f.acceptedRawCounts[caseId], nil
}
func (f *fakeCaseRepository) ReadListBySessionId(_ context.Context, _ uuid.UUID) ([]domainmodels.InfraredStateDeviceRecordCase, error) {
	return f.listBySessionIdResult, nil
}
func (f *fakeCaseRepository) ReadRawById(_ context.Context, rawId uuid.UUID) (*domainmodels.InfraredStateDeviceRecordRaw, error) {
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
type fakeLlmClient struct {
	text        string
	err         error
	texts       []string // if set, successive GenerateText calls on THIS client return these in order, sticking on the last entry once exhausted
	callCount   int
	lastRequest domaincontractsllm.GenerateTextRequest
}

func (f *fakeLlmClient) GenerateText(_ context.Context, req domaincontractsllm.GenerateTextRequest) (domaincontractsllm.GenerateTextResult, error) {
	f.lastRequest = req
	if f.err != nil {
		return domaincontractsllm.GenerateTextResult{}, f.err
	}
	if len(f.texts) > 0 {
		i := f.callCount
		if i >= len(f.texts) {
			i = len(f.texts) - 1
		}
		f.callCount++
		return domaincontractsllm.GenerateTextResult{Text: f.texts[i]}, nil
	}
	return domaincontractsllm.GenerateTextResult{Text: f.text}, nil
}

type fakeLlmClientFactory struct {
	err            error
	responseText   string   // existing field — single-response tests keep using this
	responseTexts  []string // if set, Current() returns these in order, one per call
	callIndex      int
	clientSequence []string // when set, Current() returns ONE client whose successive GenerateText calls cycle through this list — for testing multi-call loops (e.g. the repair loop) on a single client
}

func (f *fakeLlmClientFactory) Current(_ context.Context) (domaincontractsllm.Client, error) {
	if f.err != nil {
		return nil, f.err
	}
	if len(f.clientSequence) > 0 {
		return &fakeLlmClient{texts: f.clientSequence}, nil
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

func (f *fakeLlmClientFactory) FromCredentials(_ domainmodels.LlmProvider, _ string, _ *string, _ string) (domaincontractsllm.Client, error) {
	return f.Current(context.Background())
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
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
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
		&fakeLlmClientFactory{err: errors.New("llm not configured for this test")}, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
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
		&fakeLlmClientFactory{}, nodeRepo, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
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
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
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
		nil, nodeRepo, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
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
		nil, &fakeNodeRepository{result: nil}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
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
		&fakeLlmClientFactory{}, nodeRepo, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
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
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
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
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
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
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
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
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
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
		&fakeLlmClientFactory{err: errors.New("llm not configured for this test")}, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
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
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
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
		llmFactory, &fakeNodeRepository{}, encoderRunner, coderRepo, testCaseRepo, &fakePublish{}, &noopLogger{},
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
		llmFactory, &fakeNodeRepository{}, encoderRunner, coderRepo, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
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

func TestRunAnalysisAndGenerationRepairsAnInitiallyFailingEncoder(t *testing.T) {
	sessionId := uuid.New()
	powerId := uuid.New()

	rawBytes, err := json.Marshal([]int32{9000, 4500, 560, 560})
	if err != nil {
		t.Fatalf("failed to marshal fixture raw data: %v", err)
	}
	baselineCaseId := uuid.New()

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
	coderRepo := &fakeCoderRepository{getResult: &domainmodels.InfraredStateCoder{Id: uuid.New(), SummaryReadme: "summary", DetailReadme: "detail"}}

	// clientSequence drives successive GenerateText calls on the ONE client
	// runAnalysisAndGeneration resolves: call 1 is WriteCoder's, call 2 is
	// the first repair round's.
	brokenSource := "function encode(state) { return [1]; }"
	repairedSource := "function encode(state) { return [9000, 4500, 560, 560]; }"
	llmFactory := &fakeLlmClientFactory{
		clientSequence: []string{
			`{"encoder_source": "` + brokenSource + `", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`,
			`{"encoder_source": "` + repairedSource + `", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s2", "detail_readme": "d2"}`,
		},
	}

	// The broken encoder emits a header-only signal (zero decoded bits), so
	// the single known baseline bit is wrong; the repaired one reproduces
	// the fixture exactly.
	encoderRunner := &fakeEncoderRunner{bySource: map[string][]int32{
		brokenSource:   {9000, 4500},
		repairedSource: {9000, 4500, 560, 560},
	}}
	testCaseRepo := &fakeTestCaseRepository{}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, definitionRepo,
		stateRepo, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, encoderRunner, coderRepo, testCaseRepo, &fakePublish{}, &noopLogger{},
	).(*usecase)

	impl.runAnalysisAndGeneration(sessionId)

	if !coderRepo.createCalled {
		t.Fatal("coder repository Create() was never called, want the repair loop to eventually succeed and persist")
	}
	if !strings.Contains(coderRepo.created.EncoderSource, "560, 560") {
		t.Fatalf("persisted EncoderSource = %q, want the REPAIRED (second) attempt, not the initial broken one", coderRepo.created.EncoderSource)
	}
}

func TestRunAnalysisAndGenerationPersistsDegradedBestAttemptAboveFatalFloor(t *testing.T) {
	sessionId := uuid.New()
	powerId := uuid.New()

	// header(9000,4500) then three data bits with spaces 560/1690/1690 ->
	// baseline bits [0, 1, 1]. Three owned bits is the smallest fixture that
	// can land strictly between the 60% fatal floor and a full pass.
	rawBytes, err := json.Marshal([]int32{9000, 4500, 560, 560, 560, 1690, 560, 1690})
	if err != nil {
		t.Fatalf("failed to marshal fixture raw data: %v", err)
	}
	baselineCaseId := uuid.New()

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
	coderRepo := &fakeCoderRepository{getResult: &domainmodels.InfraredStateCoder{Id: uuid.New(), SummaryReadme: "summary", DetailReadme: "detail"}}

	// Every attempt (initial + all three repair rounds) runs through this
	// same runner, which decodes to [0, 1, 0] against the real [0, 1, 1] ->
	// 2/3 = 66.7% owned accuracy: never a pass, but above the 60% floor, so
	// the best attempt must still be persisted.
	encoderRunner := &fakeEncoderRunner{raw: []int32{9000, 4500, 560, 560, 560, 1690, 560, 560}}

	llmFactory := &fakeLlmClientFactory{
		clientSequence: []string{
			`{"encoder_source": "function encode(state) { return [1]; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`,
			`{"encoder_source": "function encode(state) { return [2]; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`,
			`{"encoder_source": "function encode(state) { return [3]; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`,
			`{"encoder_source": "function encode(state) { return [4]; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`,
		},
	}

	testCaseRepo := &fakeTestCaseRepository{}
	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, definitionRepo,
		stateRepo, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, encoderRunner, coderRepo, testCaseRepo, &fakePublish{}, &noopLogger{},
	).(*usecase)

	impl.runAnalysisAndGeneration(sessionId)

	if !coderRepo.createCalled {
		t.Fatal("coder repository Create() was never called, want the best (66.7%, above the 60% floor) attempt persisted")
	}
	// Every repair round scores an identical 2/3, so none of them IMPROVES
	// on the initial attempt — the loop's "strictly greater OwnedCorrect"
	// rule must therefore discard all three and keep the very first coder
	// ("return [1]"). Asserting the persisted source (not merely that
	// something was persisted) is what proves non-improving rounds are
	// discarded rather than blindly overwriting best.
	if !strings.Contains(coderRepo.created.EncoderSource, "[1]") {
		t.Fatalf("persisted EncoderSource = %q, want the INITIAL attempt kept (no repair round beat its 2/3 score)", coderRepo.created.EncoderSource)
	}
}

func TestRunAnalysisAndGenerationFailsBelowFatalFloorAfterRepairExhausted(t *testing.T) {
	sessionId := uuid.New()
	powerId := uuid.New()

	rawBytes, err := json.Marshal([]int32{9000, 4500, 560, 560})
	if err != nil {
		t.Fatalf("failed to marshal fixture raw data: %v", err)
	}
	baselineCaseId := uuid.New()

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
	// RunEncoder always throws -> every case always fails -> 0% accuracy,
	// below the 60% floor even after 3 exhausted repair rounds.
	encoderRunner := &fakeEncoderRunner{err: errors.New("encoder threw")}
	llmFactory := &fakeLlmClientFactory{responseText: `{"encoder_source": "function encode(state) { return [9000, 4500]; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, definitionRepo,
		stateRepo, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, encoderRunner, coderRepo, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
	).(*usecase)

	impl.runAnalysisAndGeneration(sessionId)

	if coderRepo.createCalled {
		t.Fatal("coder repository Create() was called despite 0% accuracy after exhausted repair, want no persisted row")
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

func TestRunAnalysisAndGenerationRequestsChecksumClarificationOnChecksumOnlyGap(t *testing.T) {
	sessionId := uuid.New()
	powerId, modeId := uuid.New(), uuid.New()
	baselineCaseId, case1Id, case2Id := uuid.New(), uuid.New(), uuid.New()

	// 4-bit fixture: header(9000,4500) + 4 data bits. Baseline is all zero.
	// case1 targets POWER (bits 0 and 3 flip). case2 targets MODE (bits 1
	// and 3 flip). Bit 3 changes for BOTH cases -> Attribute() flags it as a
	// checksum bit via its intersection rule.
	marshalRaw := func(durations []int32) []byte {
		b, err := json.Marshal(durations)
		if err != nil {
			t.Fatalf("failed to marshal fixture raw data: %v", err)
		}
		return b
	}
	baselineRaw := marshalRaw([]int32{9000, 4500, 560, 560, 560, 560, 560, 560, 560, 560})
	case1Raw := marshalRaw([]int32{9000, 4500, 560, 1690, 560, 560, 560, 560, 560, 1690}) // decodes [1,0,0,1]
	case2Raw := marshalRaw([]int32{9000, 4500, 560, 560, 560, 1690, 560, 560, 560, 1690}) // decodes [0,1,0,1]

	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, InfraredDeviceId: uuid.New()}}
	stateRepo := &fakeStateRepository{listByDeviceTypeIdResult: []domainmodels.InfraredState{
		{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum},
		{Id: modeId, Name: "MODE", Type: domainmodels.InfraredStateTypeEnum},
	}}
	definitionRepo := &fakeDefinitionRepository{listByDeviceIdResult: []domainmodels.InfraredStateDeviceDefinition{
		{InfraredStateId: powerId, Options: []string{"OFF", "ON"}},   // baseline = Options[0] = "OFF"
		{InfraredStateId: modeId, Options: []string{"COOL", "HEAT"}}, // baseline = Options[0] = "COOL"
	}}
	caseRepo := &fakeCaseRepository{
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: baselineCaseId, InfraredRecordSessionId: sessionId, Step: 1, Status: domainmodels.InfraredRecordCaseStatusAccepted},
			{Id: case1Id, InfraredRecordSessionId: sessionId, Step: 2, Status: domainmodels.InfraredRecordCaseStatusAccepted},
			{Id: case2Id, InfraredRecordSessionId: sessionId, Step: 3, Status: domainmodels.InfraredRecordCaseStatusAccepted},
		},
		listRawByCaseId: map[uuid.UUID][]domainmodels.InfraredStateDeviceRecordRaw{
			baselineCaseId: {
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: baselineRaw, Status: domainmodels.InfraredRecordRawStatusAccepted},
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: baselineRaw, Status: domainmodels.InfraredRecordRawStatusAccepted},
			},
			case1Id: {
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: case1Id, RawData: case1Raw, Status: domainmodels.InfraredRecordRawStatusAccepted},
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: case1Id, RawData: case1Raw, Status: domainmodels.InfraredRecordRawStatusAccepted},
			},
			case2Id: {
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: case2Id, RawData: case2Raw, Status: domainmodels.InfraredRecordRawStatusAccepted},
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: case2Id, RawData: case2Raw, Status: domainmodels.InfraredRecordRawStatusAccepted},
			},
		},
		listStatesByCaseId: map[uuid.UUID][]domainmodels.InfraredStateDeviceRecordState{
			baselineCaseId: {
				{InfraredStateDeviceRecordCaseId: baselineCaseId, InfraredStateId: powerId, StateValue: "OFF"},
				{InfraredStateDeviceRecordCaseId: baselineCaseId, InfraredStateId: modeId, StateValue: "COOL"},
			},
			case1Id: {
				{InfraredStateDeviceRecordCaseId: case1Id, InfraredStateId: powerId, StateValue: "ON"},
				{InfraredStateDeviceRecordCaseId: case1Id, InfraredStateId: modeId, StateValue: "COOL"},
			},
			case2Id: {
				{InfraredStateDeviceRecordCaseId: case2Id, InfraredStateId: powerId, StateValue: "OFF"},
				{InfraredStateDeviceRecordCaseId: case2Id, InfraredStateId: modeId, StateValue: "HEAT"},
			},
		},
	}
	coderRepo := &fakeCoderRepository{}
	// The encoder reproduces every owned bit (0 and 1) correctly for every
	// known state but always emits 0 at the checksum bit (3) -> a pure
	// ChecksumOnlyGap.
	encoderRunner := &fakeEncoderRunner{byState: map[string][]int32{
		"OFF|COOL": {9000, 4500, 560, 560, 560, 560, 560, 560, 560, 560},  // [0,0,0,0] vs real [0,0,0,0] -> all correct
		"ON|COOL":  {9000, 4500, 560, 1690, 560, 560, 560, 560, 560, 560}, // [1,0,0,0] vs real [1,0,0,1] -> bit3 wrong only
		"OFF|HEAT": {9000, 4500, 560, 560, 560, 1690, 560, 560, 560, 560}, // [0,1,0,0] vs real [0,1,0,1] -> bit3 wrong only
	}}
	llmFactory := &fakeLlmClientFactory{clientSequence: []string{
		`{"encoder_source": "function encode(state) { return []; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`,
		`[{"description": "repeat baseline twice more", "states": {"POWER": "OFF", "MODE": "COOL"}}]`,
	}}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, definitionRepo,
		stateRepo, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, encoderRunner, coderRepo, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
	).(*usecase)

	impl.runAnalysisAndGeneration(sessionId)

	if coderRepo.createCalled {
		t.Fatal("coder repository Create() was called on a checksum-only gap, want the session sent back to RECORDING instead")
	}
	if len(caseRepo.CreatedCaseIds()) == 0 {
		t.Fatal("no new case was persisted, want WriteChecksumClarificationCases' plan to have been recorded via persistNewCasesAndReturnToRecording")
	}
	// The once-per-session cap is only real if the marker is actually
	// written — without this the sibling "already used" test below could
	// never be reached in production.
	if got := sessionRepo.ChecksumClarificationMarks(); got != 1 {
		t.Fatalf("MarkChecksumClarificationUsedById called %d times, want exactly 1 (the session just spent its one clarification attempt)", got)
	}
	found := false
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateRecording {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include RECORDING (checksum clarification sends the session back to record more)", sessionRepo.StatusUpdates())
	}
}

// The checksum-only-gap check runs AFTER the repair loop, not only on the
// initial attempt: here the initial attempt gets owned bits AND checksum
// bits wrong (so it enters the repair loop), and round 1 fixes every owned
// bit while leaving the checksum bit wrong. That post-repair state is a
// clean Passed(), so checking only before the loop would persist a coder
// with a known-broken checksum instead of gathering more data.
func TestRunAnalysisAndGenerationRequestsChecksumClarificationAfterRepairRound(t *testing.T) {
	sessionId := uuid.New()
	powerId, modeId := uuid.New(), uuid.New()
	baselineCaseId, case1Id, case2Id := uuid.New(), uuid.New(), uuid.New()

	marshalRaw := func(durations []int32) []byte {
		b, err := json.Marshal(durations)
		if err != nil {
			t.Fatalf("failed to marshal fixture raw data: %v", err)
		}
		return b
	}
	// Same 4-bit fixture as the sibling test: bit 3 flips for both cases, so
	// Attribute() flags it as the checksum bit; bits 0-2 are owned.
	baselineRaw := marshalRaw([]int32{9000, 4500, 560, 560, 560, 560, 560, 560, 560, 560})
	case1Raw := marshalRaw([]int32{9000, 4500, 560, 1690, 560, 560, 560, 560, 560, 1690}) // [1,0,0,1]
	case2Raw := marshalRaw([]int32{9000, 4500, 560, 560, 560, 1690, 560, 560, 560, 1690}) // [0,1,0,1]

	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, InfraredDeviceId: uuid.New()}}
	stateRepo := &fakeStateRepository{listByDeviceTypeIdResult: []domainmodels.InfraredState{
		{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum},
		{Id: modeId, Name: "MODE", Type: domainmodels.InfraredStateTypeEnum},
	}}
	definitionRepo := &fakeDefinitionRepository{listByDeviceIdResult: []domainmodels.InfraredStateDeviceDefinition{
		{InfraredStateId: powerId, Options: []string{"OFF", "ON"}},
		{InfraredStateId: modeId, Options: []string{"COOL", "HEAT"}},
	}}
	caseRepo := &fakeCaseRepository{
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: baselineCaseId, InfraredRecordSessionId: sessionId, Step: 1, Status: domainmodels.InfraredRecordCaseStatusAccepted},
			{Id: case1Id, InfraredRecordSessionId: sessionId, Step: 2, Status: domainmodels.InfraredRecordCaseStatusAccepted},
			{Id: case2Id, InfraredRecordSessionId: sessionId, Step: 3, Status: domainmodels.InfraredRecordCaseStatusAccepted},
		},
		listRawByCaseId: map[uuid.UUID][]domainmodels.InfraredStateDeviceRecordRaw{
			baselineCaseId: {
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: baselineRaw, Status: domainmodels.InfraredRecordRawStatusAccepted},
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: baselineRaw, Status: domainmodels.InfraredRecordRawStatusAccepted},
			},
			case1Id: {
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: case1Id, RawData: case1Raw, Status: domainmodels.InfraredRecordRawStatusAccepted},
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: case1Id, RawData: case1Raw, Status: domainmodels.InfraredRecordRawStatusAccepted},
			},
			case2Id: {
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: case2Id, RawData: case2Raw, Status: domainmodels.InfraredRecordRawStatusAccepted},
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: case2Id, RawData: case2Raw, Status: domainmodels.InfraredRecordRawStatusAccepted},
			},
		},
		listStatesByCaseId: map[uuid.UUID][]domainmodels.InfraredStateDeviceRecordState{
			baselineCaseId: {
				{InfraredStateDeviceRecordCaseId: baselineCaseId, InfraredStateId: powerId, StateValue: "OFF"},
				{InfraredStateDeviceRecordCaseId: baselineCaseId, InfraredStateId: modeId, StateValue: "COOL"},
			},
			case1Id: {
				{InfraredStateDeviceRecordCaseId: case1Id, InfraredStateId: powerId, StateValue: "ON"},
				{InfraredStateDeviceRecordCaseId: case1Id, InfraredStateId: modeId, StateValue: "COOL"},
			},
			case2Id: {
				{InfraredStateDeviceRecordCaseId: case2Id, InfraredStateId: powerId, StateValue: "OFF"},
				{InfraredStateDeviceRecordCaseId: case2Id, InfraredStateId: modeId, StateValue: "HEAT"},
			},
		},
	}
	coderRepo := &fakeCoderRepository{}

	brokenSource := "function encode(state) { return []; }"
	repairedSource := "function encode(state) { return [1]; }"
	// Broken: always emits all-zero bits, so owned bit 0 (case1) and owned
	// bit 1 (case2) are wrong on top of the checksum bit -> NOT a
	// checksum-only gap, the repair loop runs. Repaired: every owned bit
	// correct, checksum bit 3 still always 0 -> a checksum-only gap that only
	// exists after round 1.
	allZero := []int32{9000, 4500, 560, 560, 560, 560, 560, 560, 560, 560}
	encoderRunner := &fakeEncoderRunner{bySourceState: map[string][]int32{
		brokenSource + "|OFF|COOL": allZero,
		brokenSource + "|ON|COOL":  allZero,
		brokenSource + "|OFF|HEAT": allZero,

		repairedSource + "|OFF|COOL": allZero,
		repairedSource + "|ON|COOL":  {9000, 4500, 560, 1690, 560, 560, 560, 560, 560, 560}, // [1,0,0,0], real [1,0,0,1]
		repairedSource + "|OFF|HEAT": {9000, 4500, 560, 560, 560, 1690, 560, 560, 560, 560}, // [0,1,0,0], real [0,1,0,1]
	}}
	llmFactory := &fakeLlmClientFactory{clientSequence: []string{
		`{"encoder_source": "` + brokenSource + `", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`,
		`{"encoder_source": "` + repairedSource + `", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s2", "detail_readme": "d2"}`,
		`[{"description": "repeat baseline twice more", "states": {"POWER": "OFF", "MODE": "COOL"}}]`,
	}}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, definitionRepo,
		stateRepo, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, encoderRunner, coderRepo, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
	).(*usecase)

	impl.runAnalysisAndGeneration(sessionId)

	if coderRepo.createCalled {
		t.Fatal("coder repository Create() was called, want the post-repair checksum-only gap routed to clarification instead")
	}
	if len(caseRepo.CreatedCaseIds()) == 0 {
		t.Fatal("no new case was persisted, want the clarification plan recorded after the repair round")
	}
	if got := sessionRepo.ChecksumClarificationMarks(); got != 1 {
		t.Fatalf("MarkChecksumClarificationUsedById called %d times, want exactly 1", got)
	}
	found := false
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateRecording {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include RECORDING", sessionRepo.StatusUpdates())
	}
}

// Same checksum-only-gap fixture as the test above, but the session has
// ALREADY spent its one clarification attempt. The cap must hold: no second
// trip back to RECORDING, no second marker write, and the coder is persisted
// anyway — owned bits are 100% correct, so Passed() is true and an imperfect
// checksum is informational only, never a reason to block persistence.
func TestRunAnalysisAndGenerationSkipsChecksumClarificationWhenAlreadyUsed(t *testing.T) {
	sessionId := uuid.New()
	powerId, modeId := uuid.New(), uuid.New()
	baselineCaseId, case1Id, case2Id := uuid.New(), uuid.New(), uuid.New()

	marshalRaw := func(durations []int32) []byte {
		b, err := json.Marshal(durations)
		if err != nil {
			t.Fatalf("failed to marshal fixture raw data: %v", err)
		}
		return b
	}
	baselineRaw := marshalRaw([]int32{9000, 4500, 560, 560, 560, 560, 560, 560, 560, 560})
	case1Raw := marshalRaw([]int32{9000, 4500, 560, 1690, 560, 560, 560, 560, 560, 1690}) // decodes [1,0,0,1]
	case2Raw := marshalRaw([]int32{9000, 4500, 560, 560, 560, 1690, 560, 560, 560, 1690}) // decodes [0,1,0,1]

	alreadyUsed := time.Now().Add(-time.Hour)
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{
		Id: sessionId, InfraredDeviceId: uuid.New(), ChecksumClarificationUsedAt: &alreadyUsed,
	}}
	stateRepo := &fakeStateRepository{listByDeviceTypeIdResult: []domainmodels.InfraredState{
		{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum},
		{Id: modeId, Name: "MODE", Type: domainmodels.InfraredStateTypeEnum},
	}}
	definitionRepo := &fakeDefinitionRepository{listByDeviceIdResult: []domainmodels.InfraredStateDeviceDefinition{
		{InfraredStateId: powerId, Options: []string{"OFF", "ON"}},
		{InfraredStateId: modeId, Options: []string{"COOL", "HEAT"}},
	}}
	caseRepo := &fakeCaseRepository{
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: baselineCaseId, InfraredRecordSessionId: sessionId, Step: 1, Status: domainmodels.InfraredRecordCaseStatusAccepted},
			{Id: case1Id, InfraredRecordSessionId: sessionId, Step: 2, Status: domainmodels.InfraredRecordCaseStatusAccepted},
			{Id: case2Id, InfraredRecordSessionId: sessionId, Step: 3, Status: domainmodels.InfraredRecordCaseStatusAccepted},
		},
		listRawByCaseId: map[uuid.UUID][]domainmodels.InfraredStateDeviceRecordRaw{
			baselineCaseId: {
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: baselineRaw, Status: domainmodels.InfraredRecordRawStatusAccepted},
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: baselineRaw, Status: domainmodels.InfraredRecordRawStatusAccepted},
			},
			case1Id: {
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: case1Id, RawData: case1Raw, Status: domainmodels.InfraredRecordRawStatusAccepted},
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: case1Id, RawData: case1Raw, Status: domainmodels.InfraredRecordRawStatusAccepted},
			},
			case2Id: {
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: case2Id, RawData: case2Raw, Status: domainmodels.InfraredRecordRawStatusAccepted},
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: case2Id, RawData: case2Raw, Status: domainmodels.InfraredRecordRawStatusAccepted},
			},
		},
		listStatesByCaseId: map[uuid.UUID][]domainmodels.InfraredStateDeviceRecordState{
			baselineCaseId: {
				{InfraredStateDeviceRecordCaseId: baselineCaseId, InfraredStateId: powerId, StateValue: "OFF"},
				{InfraredStateDeviceRecordCaseId: baselineCaseId, InfraredStateId: modeId, StateValue: "COOL"},
			},
			case1Id: {
				{InfraredStateDeviceRecordCaseId: case1Id, InfraredStateId: powerId, StateValue: "ON"},
				{InfraredStateDeviceRecordCaseId: case1Id, InfraredStateId: modeId, StateValue: "COOL"},
			},
			case2Id: {
				{InfraredStateDeviceRecordCaseId: case2Id, InfraredStateId: powerId, StateValue: "OFF"},
				{InfraredStateDeviceRecordCaseId: case2Id, InfraredStateId: modeId, StateValue: "HEAT"},
			},
		},
	}
	coderRepo := &fakeCoderRepository{getResult: &domainmodels.InfraredStateCoder{Id: uuid.New(), SummaryReadme: "summary", DetailReadme: "detail"}}
	encoderRunner := &fakeEncoderRunner{byState: map[string][]int32{
		"OFF|COOL": {9000, 4500, 560, 560, 560, 560, 560, 560, 560, 560},  // [0,0,0,0] vs real [0,0,0,0] -> all correct
		"ON|COOL":  {9000, 4500, 560, 1690, 560, 560, 560, 560, 560, 560}, // [1,0,0,0] vs real [1,0,0,1] -> bit3 wrong only
		"OFF|HEAT": {9000, 4500, 560, 560, 560, 1690, 560, 560, 560, 560}, // [0,1,0,0] vs real [0,1,0,1] -> bit3 wrong only
	}}
	llmFactory := &fakeLlmClientFactory{responseTexts: []string{
		`{"encoder_source": "function encode(state) { return []; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`,
		`[{"description": "d1", "states": {"POWER": "OFF"}}]`,
	}}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, definitionRepo,
		stateRepo, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, encoderRunner, coderRepo, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
	).(*usecase)

	impl.runAnalysisAndGeneration(sessionId)

	if !coderRepo.createCalled {
		t.Fatal("coder repository Create() was never called, want the coder persisted anyway (owned bits 100% correct -> Passed(), checksum gap is informational only)")
	}
	if got := sessionRepo.ChecksumClarificationMarks(); got != 0 {
		t.Fatalf("MarkChecksumClarificationUsedById called %d times, want 0 (this session already spent its one attempt)", got)
	}
	if len(caseRepo.CreatedCaseIds()) != 0 {
		t.Fatalf("created %d new record cases, want 0 (the clarification path must not run a second time)", len(caseRepo.CreatedCaseIds()))
	}
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateRecording {
			t.Fatalf("status updates = %v, want NO second trip back to RECORDING once the clarification attempt is spent", sessionRepo.StatusUpdates())
		}
	}
}

// A recorded case that can't be attributed to exactly one changed state
// (here: it carries fewer states than the baseline, so no single differing
// state is identifiable) must be skipped, not fail the whole session — one
// unusable case is not a reason to throw away every other case's analysis.
func TestRunAnalysisAndGenerationSkipsNonOfatCaseInsteadOfFailing(t *testing.T) {
	sessionId := uuid.New()
	powerId, modeId := uuid.New(), uuid.New()
	baselineCaseId, oddCaseId := uuid.New(), uuid.New()

	marshalRaw := func(durations []int32) []byte {
		b, err := json.Marshal(durations)
		if err != nil {
			t.Fatalf("failed to marshal fixture raw data: %v", err)
		}
		return b
	}
	baselineRaw := marshalRaw([]int32{9000, 4500, 560, 560, 560, 560})
	oddRaw := marshalRaw([]int32{9000, 4500, 560, 1690, 560, 560})

	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, InfraredDeviceId: uuid.New()}}
	stateRepo := &fakeStateRepository{listByDeviceTypeIdResult: []domainmodels.InfraredState{
		{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum},
		{Id: modeId, Name: "MODE", Type: domainmodels.InfraredStateTypeEnum},
	}}
	definitionRepo := &fakeDefinitionRepository{listByDeviceIdResult: []domainmodels.InfraredStateDeviceDefinition{
		{InfraredStateId: powerId, Options: []string{"OFF", "ON"}},
		{InfraredStateId: modeId, Options: []string{"COOL", "HEAT"}},
	}}
	caseRepo := &fakeCaseRepository{
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: baselineCaseId, InfraredRecordSessionId: sessionId, Step: 1, Status: domainmodels.InfraredRecordCaseStatusAccepted},
			{Id: oddCaseId, InfraredRecordSessionId: sessionId, Step: 2, Status: domainmodels.InfraredRecordCaseStatusAccepted},
		},
		listRawByCaseId: map[uuid.UUID][]domainmodels.InfraredStateDeviceRecordRaw{
			baselineCaseId: {
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: baselineRaw, Status: domainmodels.InfraredRecordRawStatusAccepted},
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: baselineCaseId, RawData: baselineRaw, Status: domainmodels.InfraredRecordRawStatusAccepted},
			},
			oddCaseId: {
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: oddCaseId, RawData: oddRaw, Status: domainmodels.InfraredRecordRawStatusAccepted},
				{Id: uuid.New(), InfraredStateDeviceRecordCaseId: oddCaseId, RawData: oddRaw, Status: domainmodels.InfraredRecordRawStatusAccepted},
			},
		},
		listStatesByCaseId: map[uuid.UUID][]domainmodels.InfraredStateDeviceRecordState{
			baselineCaseId: {
				{InfraredStateDeviceRecordCaseId: baselineCaseId, InfraredStateId: powerId, StateValue: "OFF"},
				{InfraredStateDeviceRecordCaseId: baselineCaseId, InfraredStateId: modeId, StateValue: "COOL"},
			},
			// Only one state recorded, and it matches baseline: neither the
			// baseline case nor an attributable OFAT delta.
			oddCaseId: {
				{InfraredStateDeviceRecordCaseId: oddCaseId, InfraredStateId: powerId, StateValue: "OFF"},
			},
		},
	}
	coderRepo := &fakeCoderRepository{getResult: &domainmodels.InfraredStateCoder{Id: uuid.New(), SummaryReadme: "s", DetailReadme: "d"}}
	encoderRunner := &fakeEncoderRunner{raw: []int32{9000, 4500, 560, 560, 560, 560}}
	llmFactory := &fakeLlmClientFactory{responseTexts: []string{
		`{"encoder_source": "function encode(state) { return []; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`,
		`[{"description": "d1", "states": {"POWER": "OFF"}}]`,
	}}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, definitionRepo,
		stateRepo, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, encoderRunner, coderRepo, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
	).(*usecase)

	impl.runAnalysisAndGeneration(sessionId)

	if !coderRepo.createCalled {
		t.Fatal("coder repository Create() was never called, want the unattributable case skipped and the rest of the pipeline to proceed")
	}
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateFailed {
			t.Fatalf("status updates = %v, want NO FAILED (one unattributable case must not kill the session)", sessionRepo.StatusUpdates())
		}
	}
}

func TestBuildAnalysisPayloadAttributesCasesRecordedBeforeTheBaselineCase(t *testing.T) {
	sessionId := uuid.New()
	powerId, modeId := uuid.New(), uuid.New()
	powerCaseId, modeCaseId, baselineCaseId := uuid.New(), uuid.New(), uuid.New()

	marshalRaw := func(durations []int32) []byte {
		b, err := json.Marshal(durations)
		if err != nil {
			t.Fatalf("failed to marshal fixture raw data: %v", err)
		}
		return b
	}
	// 9000/4500 header, then four 560-pair bits: 560,560 = 0 and 560,1690 = 1.
	baselineRaw := marshalRaw([]int32{9000, 4500, 560, 560, 560, 560, 560, 560, 560, 560})
	powerRaw := marshalRaw([]int32{9000, 4500, 560, 1690, 560, 560, 560, 560, 560, 560})
	modeRaw := marshalRaw([]int32{9000, 4500, 560, 560, 560, 1690, 560, 560, 560, 560})

	states := []domainmodels.InfraredState{
		{Id: powerId, Name: "POWER", Type: domainmodels.InfraredStateTypeEnum},
		{Id: modeId, Name: "MODE", Type: domainmodels.InfraredStateTypeEnum},
	}
	definitions := []domainmodels.InfraredStateDeviceDefinition{
		{InfraredStateId: powerId, Options: []string{"OFF", "ON"}},
		{InfraredStateId: modeId, Options: []string{"COOL", "HEAT"}},
	}
	twoAcceptedRaws := func(caseId uuid.UUID, raw []byte) []domainmodels.InfraredStateDeviceRecordRaw {
		return []domainmodels.InfraredStateDeviceRecordRaw{
			{Id: uuid.New(), InfraredStateDeviceRecordCaseId: caseId, RawData: raw, Status: domainmodels.InfraredRecordRawStatusAccepted},
			{Id: uuid.New(), InfraredStateDeviceRecordCaseId: caseId, RawData: raw, Status: domainmodels.InfraredRecordRawStatusAccepted},
		}
	}

	// The LLM's press order put the baseline case LAST: every non-baseline
	// case is visited before the baseline is known.
	cases := []domainmodels.InfraredStateDeviceRecordCase{
		{Id: powerCaseId, InfraredRecordSessionId: sessionId, Step: 1, Status: domainmodels.InfraredRecordCaseStatusAccepted},
		{Id: modeCaseId, InfraredRecordSessionId: sessionId, Step: 2, Status: domainmodels.InfraredRecordCaseStatusAccepted},
		{Id: baselineCaseId, InfraredRecordSessionId: sessionId, Step: 3, Status: domainmodels.InfraredRecordCaseStatusAccepted},
	}
	caseRepo := &fakeCaseRepository{
		listBySessionIdResult: cases,
		listRawByCaseId: map[uuid.UUID][]domainmodels.InfraredStateDeviceRecordRaw{
			powerCaseId:    twoAcceptedRaws(powerCaseId, powerRaw),
			modeCaseId:     twoAcceptedRaws(modeCaseId, modeRaw),
			baselineCaseId: twoAcceptedRaws(baselineCaseId, baselineRaw),
		},
		listStatesByCaseId: map[uuid.UUID][]domainmodels.InfraredStateDeviceRecordState{
			powerCaseId: {
				{InfraredStateDeviceRecordCaseId: powerCaseId, InfraredStateId: powerId, StateValue: "ON"},
				{InfraredStateDeviceRecordCaseId: powerCaseId, InfraredStateId: modeId, StateValue: "COOL"},
			},
			modeCaseId: {
				{InfraredStateDeviceRecordCaseId: modeCaseId, InfraredStateId: powerId, StateValue: "OFF"},
				{InfraredStateDeviceRecordCaseId: modeCaseId, InfraredStateId: modeId, StateValue: "HEAT"},
			},
			baselineCaseId: {
				{InfraredStateDeviceRecordCaseId: baselineCaseId, InfraredStateId: powerId, StateValue: "OFF"},
				{InfraredStateDeviceRecordCaseId: baselineCaseId, InfraredStateId: modeId, StateValue: "COOL"},
			},
		},
	}

	impl := NewUsecaseImpl(
		&fakeSessionRepository{}, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
	).(*usecase)

	payload, baselineState, baselineBits, recordedCases, err := impl.buildAnalysisPayload(context.Background(), "test", cases, states, definitions)
	if err != nil {
		t.Fatalf("buildAnalysisPayload() error = %v, want nil", err)
	}
	if len(baselineBits) == 0 {
		t.Fatal("baselineBits is empty, want the baseline case resolved even though it is last in case order")
	}
	if baselineState[powerId.String()] != "OFF" || baselineState[modeId.String()] != "COOL" {
		t.Fatalf("baselineState = %v, want POWER=OFF and MODE=COOL", baselineState)
	}
	if len(recordedCases) != 2 {
		t.Fatalf("len(recordedCases) = %d, want 2 (both pre-baseline cases attributed, none silently skipped)", len(recordedCases))
	}
	if len(payload.States) != 2 {
		t.Fatalf("len(payload.States) = %d, want 2 attributed states, got payload %+v", len(payload.States), payload)
	}
	attributedBits := make(map[uuid.UUID][]int, len(payload.States))
	for _, s := range payload.States {
		attributedBits[s.StateId] = s.BitOffsets
	}
	if len(attributedBits[powerId]) == 0 || len(attributedBits[modeId]) == 0 {
		t.Fatalf("payload.States = %+v, want owned bits for both POWER and MODE", payload.States)
	}
}

func TestGetCoderBySessionIdDelegatesToRepository(t *testing.T) {
	sessionId := uuid.New()
	want := &domainmodels.InfraredStateCoder{Id: uuid.New(), InfraredRecordSessionId: sessionId}
	coderRepo := &fakeCoderRepository{getResult: want}
	usecase := NewUsecaseImpl(
		&fakeSessionRepository{}, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
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
		llmFactory, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, testCaseRepo, &fakePublish{}, &noopLogger{},
	).(*usecase)

	impl.runTestCaseGeneration(sessionId, coderId)

	if len(testCaseRepo.CreatedTestCaseIds()) != 2 {
		t.Fatalf("created test case ids = %v, want 2", testCaseRepo.CreatedTestCaseIds())
	}
	statusUpdates := sessionRepo.StatusUpdates()
	sawTestCasesGenerating, sawTesting := false, false
	for _, s := range statusUpdates {
		if s == domainmodels.InfraredRecordingStateTestCasesGenerating {
			sawTestCasesGenerating = true
		}
		if s == domainmodels.InfraredRecordingStateTesting {
			sawTesting = true
		}
	}
	if !sawTestCasesGenerating {
		t.Fatalf("status updates = %v, want to include TEST_CASES_GENERATING", statusUpdates)
	}
	if !sawTesting {
		t.Fatalf("status updates = %v, want to include TESTING", statusUpdates)
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
		&fakeLlmClientFactory{err: errors.New("llm down")}, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, testCaseRepo, &fakePublish{}, &noopLogger{},
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

func TestTransmitTestCaseRunsEncoderAndPublishesResult(t *testing.T) {
	testCaseId := uuid.New()
	coderId := uuid.New()
	powerId := uuid.New()
	nodeId := uuid.New()
	sessionId := uuid.New()
	deviceId := "AC276E5E030C"

	testCaseRepo := &fakeTestCaseRepository{
		getResult: &domainmodels.InfraredTestCase{Id: testCaseId, InfraredStateCoderId: coderId},
		listStatesByTestCaseIdResult: []domainmodels.InfraredTestCaseState{
			{InfraredTestCaseId: testCaseId, InfraredStateId: powerId, StateValue: "ON"},
		},
	}
	coderRepo := &fakeCoderRepository{getByIdResult: &domainmodels.InfraredStateCoder{Id: coderId, InfraredRecordSessionId: sessionId, EncoderSource: "function encode(state) { return [9000, 4500]; }"}}
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, NodeId: nodeId}}
	nodeRepo := &fakeNodeRepository{result: &domainmodels.Node{Id: nodeId, DeviceId: deviceId}}
	stateRepo := &fakeStateRepository{listByDeviceTypeIdResult: []domainmodels.InfraredState{{Id: powerId, Name: "POWER"}}}
	publish := &fakePublish{}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		stateRepo, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, nodeRepo, &fakeEncoderRunner{}, coderRepo, testCaseRepo, publish, &noopLogger{},
	)

	if err := impl.TransmitTestCase(context.Background(), testCaseId); err != nil {
		t.Fatalf("TransmitTestCase() error = %v, want nil", err)
	}
	if publish.transmitCalls != 1 {
		t.Fatalf("IrTransmit() calls = %d, want 1", publish.transmitCalls)
	}
}

func TestRecordTestCaseResultDoesNothingElseWhilePendingCasesRemain(t *testing.T) {
	coderId := uuid.New()
	firstCaseId, secondCaseId := uuid.New(), uuid.New()

	testCaseRepo := &fakeTestCaseRepository{
		getResult: &domainmodels.InfraredTestCase{Id: firstCaseId, InfraredStateCoderId: coderId},
		listByCoderIdResult: []domainmodels.InfraredTestCase{
			{Id: firstCaseId, InfraredStateCoderId: coderId, Status: domainmodels.InfraredTestCaseStatusPassed},
			{Id: secondCaseId, InfraredStateCoderId: coderId, Status: domainmodels.InfraredTestCaseStatusPending},
		},
	}
	coderRepo := &fakeCoderRepository{}
	sessionRepo := &fakeSessionRepository{}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, testCaseRepo, &fakePublish{}, &noopLogger{},
	)

	if err := impl.RecordTestCaseResult(context.Background(), firstCaseId, true); err != nil {
		t.Fatalf("RecordTestCaseResult() error = %v, want nil", err)
	}
	if coderRepo.activateCalls != 0 {
		t.Fatal("Activate() was called despite a pending test case remaining")
	}
	if len(sessionRepo.StatusUpdates()) != 0 {
		t.Fatalf("session status updates = %v, want none (session should not transition while cases are still pending)", sessionRepo.StatusUpdates())
	}
}

func TestRecordTestCaseResultCompletesSessionWhenAllPass(t *testing.T) {
	coderId := uuid.New()
	deviceId := uuid.New()
	sessionId := uuid.New()
	onlyCaseId := uuid.New()

	testCaseRepo := &fakeTestCaseRepository{
		getResult: &domainmodels.InfraredTestCase{Id: onlyCaseId, InfraredStateCoderId: coderId},
		listByCoderIdResult: []domainmodels.InfraredTestCase{
			{Id: onlyCaseId, InfraredStateCoderId: coderId, Status: domainmodels.InfraredTestCaseStatusPassed},
		},
	}
	fixtureCoder := &domainmodels.InfraredStateCoder{Id: coderId, InfraredDeviceId: deviceId, InfraredRecordSessionId: sessionId}
	coderRepo := &fakeCoderRepository{getByIdResult: fixtureCoder, getResult: fixtureCoder}
	sessionRepo := &fakeSessionRepository{}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, testCaseRepo, &fakePublish{}, &noopLogger{},
	)

	if err := impl.RecordTestCaseResult(context.Background(), onlyCaseId, true); err != nil {
		t.Fatalf("RecordTestCaseResult() error = %v, want nil", err)
	}
	if coderRepo.activateCalls != 1 {
		t.Fatalf("Activate() calls = %d, want 1", coderRepo.activateCalls)
	}
	found := false
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateCompleted {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include COMPLETED", sessionRepo.StatusUpdates())
	}
}

func TestRecordTestCaseResultTriggersRetryLoopWhenAnyFail(t *testing.T) {
	coderId := uuid.New()
	deviceId := uuid.New()
	sessionId := uuid.New()
	powerId := uuid.New()
	failedCaseId := uuid.New()

	testCaseRepo := &fakeTestCaseRepository{
		getResult: &domainmodels.InfraredTestCase{Id: failedCaseId, InfraredStateCoderId: coderId},
		listByCoderIdResult: []domainmodels.InfraredTestCase{
			{Id: failedCaseId, InfraredStateCoderId: coderId, Status: domainmodels.InfraredTestCaseStatusFailed},
		},
		listStatesByTestCaseIdResult: []domainmodels.InfraredTestCaseState{
			{InfraredTestCaseId: failedCaseId, InfraredStateId: powerId, StateValue: "OFF"},
		},
	}
	fixtureCoder := &domainmodels.InfraredStateCoder{Id: coderId, InfraredDeviceId: deviceId, InfraredRecordSessionId: sessionId}
	coderRepo := &fakeCoderRepository{getByIdResult: fixtureCoder, getResult: fixtureCoder}
	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, InfraredDeviceId: deviceId}}
	stateRepo := &fakeStateRepository{listByDeviceTypeIdResult: []domainmodels.InfraredState{{Id: powerId, Name: "POWER"}}}
	caseRepo := &fakeCaseRepository{
		listBySessionIdResult: []domainmodels.InfraredStateDeviceRecordCase{
			{Id: uuid.New(), InfraredRecordSessionId: sessionId, Step: 3, Status: domainmodels.InfraredRecordCaseStatusAccepted},
		},
	}
	llmFactory := &fakeLlmClientFactory{responseText: `[{"description": "Re-record POWER OFF.", "states": {"POWER": "OFF"}}]`}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		stateRepo, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, testCaseRepo, &fakePublish{}, &noopLogger{},
	)

	if err := impl.RecordTestCaseResult(context.Background(), failedCaseId, false); err != nil {
		t.Fatalf("RecordTestCaseResult() error = %v, want nil", err)
	}

	// runRetryCaseGeneration runs on its own goroutine — poll with a timeout,
	// the same pattern Plan B1's async-job tests already established.
	deadline := time.Now().Add(2 * time.Second)
	found := false
	for time.Now().Before(deadline) && !found {
		for _, s := range sessionRepo.StatusUpdates() {
			if s == domainmodels.InfraredRecordingStateRecording {
				found = true
			}
		}
		if !found {
			time.Sleep(5 * time.Millisecond)
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to eventually include RECORDING", sessionRepo.StatusUpdates())
	}
	if coderRepo.activateCalls != 0 {
		t.Fatal("Activate() was called despite a failed test case")
	}

	if steps := caseRepo.CreatedSteps(); len(steps) != 1 || steps[0] != 4 {
		t.Fatalf("CreatedSteps() = %v, want [4] (continuing from the existing Step: 3 fixture)", steps)
	}
	createdIds := caseRepo.CreatedCaseIds()
	if len(createdIds) != 1 {
		t.Fatalf("CreatedCaseIds() = %v, want exactly 1 new case", createdIds)
	}
	if got := sessionRepo.CurrentCaseId(); got == nil || *got != createdIds[0] {
		t.Fatalf("session cursor = %v, want it set to the new retry case %v", got, createdIds[0])
	}
}

func TestRecordTestCaseResultIsIdempotentAgainstDoubleSubmit(t *testing.T) {
	coderId := uuid.New()
	testCaseId := uuid.New()

	testCaseRepo := &fakeTestCaseRepository{
		getResult: &domainmodels.InfraredTestCase{Id: testCaseId, InfraredStateCoderId: coderId, Status: domainmodels.InfraredTestCaseStatusPassed},
	}
	coderRepo := &fakeCoderRepository{}
	sessionRepo := &fakeSessionRepository{}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, testCaseRepo, &fakePublish{}, &noopLogger{},
	)

	if err := impl.RecordTestCaseResult(context.Background(), testCaseId, true); err != nil {
		t.Fatalf("RecordTestCaseResult() error = %v, want nil", err)
	}
	if coderRepo.activateCalls != 0 {
		t.Fatal("Activate() was called on a double-submit of an already-terminal test case")
	}
	if len(sessionRepo.StatusUpdates()) != 0 {
		t.Fatalf("session status updates = %v, want none (already-recorded result should be a no-op)", sessionRepo.StatusUpdates())
	}
}

func TestRecordTestCaseResultIgnoresSupersededCoderRound(t *testing.T) {
	staleCoderId := uuid.New()
	latestCoderId := uuid.New()
	deviceId := uuid.New()
	sessionId := uuid.New()
	testCaseId := uuid.New()

	testCaseRepo := &fakeTestCaseRepository{
		getResult: &domainmodels.InfraredTestCase{Id: testCaseId, InfraredStateCoderId: staleCoderId},
		listByCoderIdResult: []domainmodels.InfraredTestCase{
			{Id: testCaseId, InfraredStateCoderId: staleCoderId, Status: domainmodels.InfraredTestCaseStatusPassed},
		},
	}
	coderRepo := &fakeCoderRepository{
		getByIdResult: &domainmodels.InfraredStateCoder{Id: staleCoderId, InfraredDeviceId: deviceId, InfraredRecordSessionId: sessionId},
		getResult:     &domainmodels.InfraredStateCoder{Id: latestCoderId, InfraredDeviceId: deviceId, InfraredRecordSessionId: sessionId},
	}
	sessionRepo := &fakeSessionRepository{}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		&fakeLlmClientFactory{}, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, testCaseRepo, &fakePublish{}, &noopLogger{},
	)

	if err := impl.RecordTestCaseResult(context.Background(), testCaseId, true); err != nil {
		t.Fatalf("RecordTestCaseResult() error = %v, want nil", err)
	}
	if coderRepo.activateCalls != 0 {
		t.Fatal("Activate() was called on a test case belonging to a superseded coder round")
	}
	if len(sessionRepo.StatusUpdates()) != 0 {
		t.Fatalf("session status updates = %v, want none (superseded-round result should be ignored)", sessionRepo.StatusUpdates())
	}
}

func TestRunTestCaseGenerationFailsSessionWhenNoPlansProposed(t *testing.T) {
	sessionId := uuid.New()
	coderId := uuid.New()

	sessionRepo := &fakeSessionRepository{getResult: &domainmodels.InfraredRecordSession{Id: sessionId, InfraredDeviceId: uuid.New()}}
	coderRepo := &fakeCoderRepository{getResult: &domainmodels.InfraredStateCoder{Id: coderId, SummaryReadme: "s", DetailReadme: "d"}}
	testCaseRepo := &fakeTestCaseRepository{}
	llmFactory := &fakeLlmClientFactory{responseText: "[]"}

	impl := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		llmFactory, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, testCaseRepo, &fakePublish{}, &noopLogger{},
	).(*usecase)

	impl.runTestCaseGeneration(sessionId, coderId)

	found := false
	for _, s := range sessionRepo.StatusUpdates() {
		if s == domainmodels.InfraredRecordingStateFailed {
			found = true
		}
	}
	if !found {
		t.Fatalf("status updates = %v, want to include FAILED", sessionRepo.StatusUpdates())
	}
	if len(testCaseRepo.CreatedTestCaseIds()) != 0 {
		t.Fatal("test cases were created despite the llm proposing zero plans")
	}
}

func TestDeleteRecordSessionByIdForwardsToRepository(t *testing.T) {
	id, deletedBy := uuid.New(), uuid.New()
	sessionRepo := &fakeSessionRepository{}
	usecase := NewUsecaseImpl(
		sessionRepo, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
	)

	if err := usecase.DeleteRecordSessionById(context.Background(), id, &deletedBy); err != nil {
		t.Fatalf("DeleteRecordSessionById() error = %v, want nil", err)
	}
	if sessionRepo.deletedId != id || sessionRepo.deletedBy != &deletedBy {
		t.Fatalf("session repository DeleteById() called with (%v, %v), want (%v, %v)", sessionRepo.deletedId, sessionRepo.deletedBy, id, &deletedBy)
	}

	sessionRepo.deleteErr = errors.New("not found")
	if err := usecase.DeleteRecordSessionById(context.Background(), id, &deletedBy); !errors.Is(err, sessionRepo.deleteErr) {
		t.Fatalf("DeleteRecordSessionById() error = %v, want the repository's error propagated", err)
	}
}

func TestDeleteRecordCaseByIdForwardsToRepository(t *testing.T) {
	id, deletedBy := uuid.New(), uuid.New()
	caseRepo := &fakeCaseRepository{}
	usecase := NewUsecaseImpl(
		&fakeSessionRepository{}, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
	)

	if err := usecase.DeleteRecordCaseById(context.Background(), id, &deletedBy); err != nil {
		t.Fatalf("DeleteRecordCaseById() error = %v, want nil", err)
	}
	if caseRepo.deletedId != id {
		t.Fatalf("case repository DeleteById() called with id %v, want %v", caseRepo.deletedId, id)
	}
}

func TestDeleteRecordStateByIdForwardsToRepository(t *testing.T) {
	id, deletedBy := uuid.New(), uuid.New()
	caseRepo := &fakeCaseRepository{}
	usecase := NewUsecaseImpl(
		&fakeSessionRepository{}, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
	)

	if err := usecase.DeleteRecordStateById(context.Background(), id, &deletedBy); err != nil {
		t.Fatalf("DeleteRecordStateById() error = %v, want nil", err)
	}
	if caseRepo.deletedStateId != id {
		t.Fatalf("case repository DeleteStateById() called with id %v, want %v", caseRepo.deletedStateId, id)
	}
}

func TestDeleteRecordRawByIdForwardsToRepository(t *testing.T) {
	id, deletedBy := uuid.New(), uuid.New()
	caseRepo := &fakeCaseRepository{}
	usecase := NewUsecaseImpl(
		&fakeSessionRepository{}, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, caseRepo, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
	)

	if err := usecase.DeleteRecordRawById(context.Background(), id, &deletedBy); err != nil {
		t.Fatalf("DeleteRecordRawById() error = %v, want nil", err)
	}
	if caseRepo.deletedRawId != id {
		t.Fatalf("case repository DeleteRawById() called with id %v, want %v", caseRepo.deletedRawId, id)
	}
}

func TestDeleteStateCoderByIdForwardsToRepository(t *testing.T) {
	id, deletedBy := uuid.New(), uuid.New()
	coderRepo := &fakeCoderRepository{}
	usecase := NewUsecaseImpl(
		&fakeSessionRepository{}, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, coderRepo, &fakeTestCaseRepository{}, &fakePublish{}, &noopLogger{},
	)

	if err := usecase.DeleteStateCoderById(context.Background(), id, &deletedBy); err != nil {
		t.Fatalf("DeleteStateCoderById() error = %v, want nil", err)
	}
	if coderRepo.deletedId != id {
		t.Fatalf("coder repository DeleteById() called with id %v, want %v", coderRepo.deletedId, id)
	}
}

func TestDeleteTestCaseByIdForwardsToRepository(t *testing.T) {
	id, deletedBy := uuid.New(), uuid.New()
	testCaseRepo := &fakeTestCaseRepository{}
	usecase := NewUsecaseImpl(
		&fakeSessionRepository{}, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, testCaseRepo, &fakePublish{}, &noopLogger{},
	)

	if err := usecase.DeleteTestCaseById(context.Background(), id, &deletedBy); err != nil {
		t.Fatalf("DeleteTestCaseById() error = %v, want nil", err)
	}
	if testCaseRepo.deletedId != id {
		t.Fatalf("test case repository DeleteById() called with id %v, want %v", testCaseRepo.deletedId, id)
	}
}

func TestDeleteTestCaseStateByIdForwardsToRepository(t *testing.T) {
	id, deletedBy := uuid.New(), uuid.New()
	testCaseRepo := &fakeTestCaseRepository{}
	usecase := NewUsecaseImpl(
		&fakeSessionRepository{}, &fakeDeviceRepository{}, &fakeDefinitionRepository{},
		&fakeStateRepository{}, &fakeCaseRepository{}, &fakeBroadcaster{}, &fakeSubscriptions{},
		nil, &fakeNodeRepository{}, &fakeEncoderRunner{}, &fakeCoderRepository{}, testCaseRepo, &fakePublish{}, &noopLogger{},
	)

	if err := usecase.DeleteTestCaseStateById(context.Background(), id, &deletedBy); err != nil {
		t.Fatalf("DeleteTestCaseStateById() error = %v, want nil", err)
	}
	if testCaseRepo.deletedStateId != id {
		t.Fatalf("test case repository DeleteStateById() called with id %v, want %v", testCaseRepo.deletedStateId, id)
	}
}
