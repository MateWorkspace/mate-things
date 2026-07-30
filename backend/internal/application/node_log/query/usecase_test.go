package applicationnodelogquery

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnodelog "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node_log"
)

func TestUsecaseReadByFilter_ForwardsFiltersAndReturnsRepositoryResults(t *testing.T) {
	loggedAtStart := time.Date(2026, time.July, 1, 10, 0, 0, 0, time.UTC)
	loggedAtEnd := time.Date(2026, time.July, 2, 10, 0, 0, 0, time.UTC)
	nodeDeviceID := "node-42"
	level := domainmodels.NodeLogLevelWarn
	wantNodeLogs := []domainmodels.NodeLog{{Id: 17, NodeDeviceId: nodeDeviceID, Level: level, Tag: "sensor", Message: "battery low"}}
	repository := &recordingNodeLog{readNodeLogs: wantNodeLogs, readTotal: 1}
	usecase := NewUsecaseImpl(repository, &recordingLogger{})

	nodeLogs, total, err := usecase.ReadByFilter(context.Background(), domainusecasesnodelog.ReadNodeLogByFilterRequest{
		LoggedAtStart: &loggedAtStart,
		LoggedAtEnd:   &loggedAtEnd,
		NodeDeviceId:  &nodeDeviceID,
		Level:         &level,
	})

	if err != nil {
		t.Fatalf("ReadByFilter() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(nodeLogs, wantNodeLogs) {
		t.Fatalf("ReadByFilter() node logs = %#v, want %#v", nodeLogs, wantNodeLogs)
	}
	if total != 1 {
		t.Fatalf("ReadByFilter() total = %d, want 1", total)
	}
	if repository.loggedAtStart != &loggedAtStart || repository.loggedAtEnd != &loggedAtEnd || repository.nodeDeviceID != &nodeDeviceID || repository.level != &level {
		t.Fatalf("ReadByFilter() forwarded filters = (%p, %p, %p, %p), want (%p, %p, %p, %p)", repository.loggedAtStart, repository.loggedAtEnd, repository.nodeDeviceID, repository.level, &loggedAtStart, &loggedAtEnd, &nodeDeviceID, &level)
	}
}

func TestUsecaseReadByFilter_PropagatesRepositoryErrorAndLogsFilterMetadata(t *testing.T) {
	loggedAtStart := time.Date(2026, time.July, 1, 10, 0, 0, 0, time.UTC)
	loggedAtEnd := time.Date(2026, time.July, 2, 10, 0, 0, 0, time.UTC)
	nodeDeviceID := "node-42"
	level := domainmodels.NodeLogLevelError
	wantErr := errors.New("database unavailable")
	logger := &recordingLogger{}
	usecase := NewUsecaseImpl(&recordingNodeLog{readErr: wantErr}, logger)

	nodeLogs, total, err := usecase.ReadByFilter(context.Background(), domainusecasesnodelog.ReadNodeLogByFilterRequest{
		LoggedAtStart: &loggedAtStart,
		LoggedAtEnd:   &loggedAtEnd,
		NodeDeviceId:  &nodeDeviceID,
		Level:         &level,
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("ReadByFilter() error = %v, want %v", err, wantErr)
	}
	if nodeLogs != nil || total != 0 {
		t.Fatalf("ReadByFilter() results = (%#v, %d), want (nil, 0)", nodeLogs, total)
	}
	wantMeta := domainmodels.LoggerMeta{"err": wantErr, "node_device_id": &nodeDeviceID, "level": &level}
	if logger.tag != "node_log/query/ReadByFilter" || logger.message != "failed to read node logs" || !reflect.DeepEqual(logger.meta, wantMeta) {
		t.Fatalf("Error() = (%q, %q, %#v), want (%q, %q, %#v)", logger.tag, logger.message, logger.meta, "node_log/query/ReadByFilter", "failed to read node logs", wantMeta)
	}
}

func TestUsecaseDeleteByFilter_ForwardsFiltersAndReturnsRepositoryTotal(t *testing.T) {
	loggedAtStart := time.Date(2026, time.July, 1, 10, 0, 0, 0, time.UTC)
	loggedAtEnd := time.Date(2026, time.July, 2, 10, 0, 0, 0, time.UTC)
	nodeDeviceID := "node-42"
	level := domainmodels.NodeLogLevelInfo
	repository := &recordingNodeLog{deleteTotal: 3}
	usecase := NewUsecaseImpl(repository, &recordingLogger{})

	total, err := usecase.DeleteByFilter(context.Background(), domainusecasesnodelog.DeleteNodeLogByFilterRequest{
		LoggedAtStart: &loggedAtStart,
		LoggedAtEnd:   &loggedAtEnd,
		NodeDeviceId:  &nodeDeviceID,
		Level:         &level,
	})

	if err != nil {
		t.Fatalf("DeleteByFilter() error = %v, want nil", err)
	}
	if total != 3 {
		t.Fatalf("DeleteByFilter() total = %d, want 3", total)
	}
	if repository.loggedAtStart != &loggedAtStart || repository.loggedAtEnd != &loggedAtEnd || repository.nodeDeviceID != &nodeDeviceID || repository.level != &level {
		t.Fatalf("DeleteByFilter() forwarded filters = (%p, %p, %p, %p), want (%p, %p, %p, %p)", repository.loggedAtStart, repository.loggedAtEnd, repository.nodeDeviceID, repository.level, &loggedAtStart, &loggedAtEnd, &nodeDeviceID, &level)
	}
}

func TestUsecaseDeleteByFilter_PropagatesRepositoryErrorAndLogsFilterMetadata(t *testing.T) {
	loggedAtStart := time.Date(2026, time.July, 1, 10, 0, 0, 0, time.UTC)
	loggedAtEnd := time.Date(2026, time.July, 2, 10, 0, 0, 0, time.UTC)
	nodeDeviceID := "node-42"
	level := domainmodels.NodeLogLevelDebug
	wantErr := errors.New("database unavailable")
	logger := &recordingLogger{}
	usecase := NewUsecaseImpl(&recordingNodeLog{deleteErr: wantErr}, logger)

	total, err := usecase.DeleteByFilter(context.Background(), domainusecasesnodelog.DeleteNodeLogByFilterRequest{
		LoggedAtStart: &loggedAtStart,
		LoggedAtEnd:   &loggedAtEnd,
		NodeDeviceId:  &nodeDeviceID,
		Level:         &level,
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("DeleteByFilter() error = %v, want %v", err, wantErr)
	}
	if total != 0 {
		t.Fatalf("DeleteByFilter() total = %d, want 0", total)
	}
	wantMeta := domainmodels.LoggerMeta{"err": wantErr, "node_device_id": &nodeDeviceID, "level": &level}
	if logger.tag != "node_log/query/DeleteByFilter" || logger.message != "failed to delete node logs" || !reflect.DeepEqual(logger.meta, wantMeta) {
		t.Fatalf("Error() = (%q, %q, %#v), want (%q, %q, %#v)", logger.tag, logger.message, logger.meta, "node_log/query/DeleteByFilter", "failed to delete node logs", wantMeta)
	}
}

type recordingNodeLog struct {
	loggedAtStart *time.Time
	loggedAtEnd   *time.Time
	nodeDeviceID  *string
	level         *domainmodels.NodeLogLevel
	readNodeLogs  []domainmodels.NodeLog
	readTotal     int
	readErr       error
	deleteTotal   int
	deleteErr     error
}

func (r *recordingNodeLog) Create(context.Context, string, domainmodels.NodeLogLevel, string, string, time.Time) (int64, error) {
	return 0, nil
}

func (r *recordingNodeLog) ReadByFilter(_ context.Context, loggedAtStart, loggedAtEnd *time.Time, nodeDeviceID *string, level *domainmodels.NodeLogLevel) ([]domainmodels.NodeLog, int, error) {
	r.loggedAtStart = loggedAtStart
	r.loggedAtEnd = loggedAtEnd
	r.nodeDeviceID = nodeDeviceID
	r.level = level
	return r.readNodeLogs, r.readTotal, r.readErr
}

func (r *recordingNodeLog) DeleteByFilter(_ context.Context, loggedAtStart, loggedAtEnd *time.Time, nodeDeviceID *string, level *domainmodels.NodeLogLevel) (int, error) {
	r.loggedAtStart = loggedAtStart
	r.loggedAtEnd = loggedAtEnd
	r.nodeDeviceID = nodeDeviceID
	r.level = level
	return r.deleteTotal, r.deleteErr
}

type recordingLogger struct {
	tag     string
	message string
	meta    domainmodels.LoggerMeta
}

func (l *recordingLogger) Error(_ context.Context, tag, message string, meta domainmodels.LoggerMeta) {
	l.tag = tag
	l.message = message
	l.meta = meta
}

func (*recordingLogger) Warn(context.Context, string, string, domainmodels.LoggerMeta)  {}
func (*recordingLogger) Info(context.Context, string, string, domainmodels.LoggerMeta)  {}
func (*recordingLogger) Debug(context.Context, string, string, domainmodels.LoggerMeta) {}
