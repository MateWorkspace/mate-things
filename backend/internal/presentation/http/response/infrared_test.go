package presentationhttpresponse

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

func TestInfraredStateDeviceRecordRawComputesPulseCountAndDuration(t *testing.T) {
	durations := []int32{9000, 4500, 560, 560, 560, 1690}
	rawData, err := json.Marshal(durations)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v, want nil", err)
	}

	model := domainmodels.InfraredStateDeviceRecordRaw{
		Id:        uuid.New(),
		RawData:   rawData,
		Status:    "CAPTURED",
		CreatedAt: time.Now(),
	}

	got := InfraredStateDeviceRecordRaw(model)

	if got.PulseCount != len(durations) {
		t.Fatalf("PulseCount = %d, want %d", got.PulseCount, len(durations))
	}

	wantDuration := 0
	for _, d := range durations {
		wantDuration += int(d)
	}
	if got.DurationUs != wantDuration {
		t.Fatalf("DurationUs = %d, want %d", got.DurationUs, wantDuration)
	}
}

func TestInfraredStateDeviceRecordRawHandlesMalformedRawDataGracefully(t *testing.T) {
	model := domainmodels.InfraredStateDeviceRecordRaw{
		Id:        uuid.New(),
		RawData:   []byte("not json"),
		Status:    "CAPTURED",
		CreatedAt: time.Now(),
	}

	got := InfraredStateDeviceRecordRaw(model)

	if got.PulseCount != 0 || got.DurationUs != 0 {
		t.Fatalf("got PulseCount=%d DurationUs=%d, want 0/0 for malformed raw_data", got.PulseCount, got.DurationUs)
	}
}
