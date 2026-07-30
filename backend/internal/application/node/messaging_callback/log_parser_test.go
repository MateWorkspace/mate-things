package applicationnodemessagingcallback

import (
	"context"
	"errors"
	"testing"
	"time"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
	domainusecasesnode "github.com/MateWorkspace/mate-things/backend/internal/domain/usecases/node"
)

func TestParseLogLine_ParsesExactWireFormatForEverySupportedLevel(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		level   domainmodels.NodeLogLevel
		message string
	}{
		{name: "none", raw: "01/02/2026 03:04:05.006 [NONE] [system] idle", level: domainmodels.NodeLogLevelNone, message: "idle"},
		{name: "error", raw: "01/02/2026 03:04:05.006 [ERROR] [system] failed", level: domainmodels.NodeLogLevelError, message: "failed"},
		{name: "warn", raw: "01/02/2026 03:04:05.006 [WARN] [system] slow", level: domainmodels.NodeLogLevelWarn, message: "slow"},
		{name: "info", raw: "01/02/2026 03:04:05.006 [INFO] [system] ready", level: domainmodels.NodeLogLevelInfo, message: "ready"},
		{name: "debug", raw: "01/02/2026 03:04:05.006 [DEBUG] [system] detail", level: domainmodels.NodeLogLevelDebug, message: "detail"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, tag, message, loggedAt := parseLogLine(tt.raw)

			if level != tt.level {
				t.Fatalf("level = %q, want %q", level, tt.level)
			}
			if tag != "system" {
				t.Fatalf("tag = %q, want system", tag)
			}
			if message != tt.message {
				t.Fatalf("message = %q, want %q", message, tt.message)
			}
			wantLoggedAt := time.Date(2026, time.February, 1, 3, 4, 5, 6_000_000, time.UTC)
			if !loggedAt.Equal(wantLoggedAt) {
				t.Fatalf("loggedAt = %s, want %s", loggedAt, wantLoggedAt)
			}
			if loggedAt.Location() != time.UTC {
				t.Fatalf("loggedAt location = %s, want UTC", loggedAt.Location())
			}
		})
	}
}

func TestParseLogLine_PreservesEmptyTagAndMessageFromExactWireFormat(t *testing.T) {
	level, tag, message, loggedAt := parseLogLine("31/12/2025 23:59:58.007 [INFO] [] ")

	if level != domainmodels.NodeLogLevelInfo {
		t.Fatalf("level = %q, want INFO", level)
	}
	if tag != "" {
		t.Fatalf("tag = %q, want empty", tag)
	}
	if message != "" {
		t.Fatalf("message = %q, want empty", message)
	}
	wantLoggedAt := time.Date(2025, time.December, 31, 23, 59, 58, 7_000_000, time.UTC)
	if !loggedAt.Equal(wantLoggedAt) {
		t.Fatalf("loggedAt = %s, want %s", loggedAt, wantLoggedAt)
	}
}

func TestParseLogLine_FallsBackForMalformedOrSemanticallyInvalidLines(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{name: "nonmatching", raw: "not a device log line"},
		{name: "impossible date", raw: "31/02/2026 03:04:05.006 [INFO] [system] impossible"},
		{name: "out of range hour", raw: "01/02/2026 24:04:05.006 [INFO] [system] impossible"},
		{name: "out of range minute", raw: "01/02/2026 03:60:05.006 [INFO] [system] impossible"},
		{name: "out of range second", raw: "01/02/2026 03:04:60.006 [INFO] [system] impossible"},
		{name: "unsupported level", raw: "01/02/2026 03:04:05.006 [VERBOSE] [system] unsupported"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().UTC()
			level, tag, message, loggedAt := parseLogLine(tt.raw)
			after := time.Now().UTC()

			if level != domainmodels.NodeLogLevelNone {
				t.Fatalf("level = %q, want NONE", level)
			}
			if tag != "" {
				t.Fatalf("tag = %q, want empty", tag)
			}
			if message != tt.raw {
				t.Fatalf("message = %q, want raw %q", message, tt.raw)
			}
			if loggedAt.Location() != time.UTC {
				t.Fatalf("loggedAt location = %s, want UTC", loggedAt.Location())
			}
			if loggedAt.Before(before) || loggedAt.After(after) {
				t.Fatalf("loggedAt = %s, want between %s and %s", loggedAt, before, after)
			}
		})
	}
}

func TestUsecaseLog_PersistsParsedValuesForDevice(t *testing.T) {
	repository := &recordingNodeLog{}
	usecase := &usecase{nodeLog: repository, logger: noOpLogger{}}

	err := usecase.Log(context.Background(), domainusecasesnode.NodeLogMessageRequest{
		DeviceId: "device-42",
		Payload:  []byte("02/03/2026 04:05:06.007 [WARN] [sensor] battery low"),
	})
	if err != nil {
		t.Fatalf("Log() error = %v, want nil", err)
	}
	if repository.createCalls != 1 {
		t.Fatalf("Create calls = %d, want 1", repository.createCalls)
	}
	if repository.nodeDeviceID != "device-42" || repository.level != domainmodels.NodeLogLevelWarn || repository.tag != "sensor" || repository.message != "battery low" {
		t.Fatalf("persisted values = (%q, %q, %q, %q), want (device-42, WARN, sensor, battery low)", repository.nodeDeviceID, repository.level, repository.tag, repository.message)
	}
	wantLoggedAt := time.Date(2026, time.March, 2, 4, 5, 6, 7_000_000, time.UTC)
	if !repository.loggedAt.Equal(wantLoggedAt) {
		t.Fatalf("persisted loggedAt = %s, want %s", repository.loggedAt, wantLoggedAt)
	}
}

func TestUsecaseLog_PropagatesRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	repository := &recordingNodeLog{createErr: wantErr}
	usecase := &usecase{nodeLog: repository, logger: noOpLogger{}}

	err := usecase.Log(context.Background(), domainusecasesnode.NodeLogMessageRequest{
		DeviceId: "device-42",
		Payload:  []byte("02/03/2026 04:05:06.007 [INFO] [sensor] reporting"),
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Log() error = %v, want %v", err, wantErr)
	}
	if repository.createCalls != 1 {
		t.Fatalf("Create calls = %d, want 1", repository.createCalls)
	}
}

type recordingNodeLog struct {
	createCalls  int
	nodeDeviceID string
	level        domainmodels.NodeLogLevel
	tag          string
	message      string
	loggedAt     time.Time
	createErr    error
}

func (r *recordingNodeLog) Create(_ context.Context, nodeDeviceID string, level domainmodels.NodeLogLevel, tag, message string, loggedAt time.Time) (int64, error) {
	r.createCalls++
	r.nodeDeviceID = nodeDeviceID
	r.level = level
	r.tag = tag
	r.message = message
	r.loggedAt = loggedAt
	return 1, r.createErr
}

func (r *recordingNodeLog) ReadByFilter(context.Context, *time.Time, *time.Time, *string, *domainmodels.NodeLogLevel) ([]domainmodels.NodeLog, int, error) {
	return nil, 0, nil
}

func (r *recordingNodeLog) DeleteByFilter(context.Context, *time.Time, *time.Time, *string, *domainmodels.NodeLogLevel) (int, error) {
	return 0, nil
}

type noOpLogger struct{}

func (noOpLogger) Error(context.Context, string, string, domainmodels.LoggerMeta) {}
func (noOpLogger) Warn(context.Context, string, string, domainmodels.LoggerMeta)  {}
func (noOpLogger) Info(context.Context, string, string, domainmodels.LoggerMeta)  {}
func (noOpLogger) Debug(context.Context, string, string, domainmodels.LoggerMeta) {}
