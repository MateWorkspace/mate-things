package infrastructurerepositoryinfraredrecordsession

import (
	"strings"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func TestQueryMarkChecksumClarificationUsedByIdSetsTimestamp(t *testing.T) {
	sqrDollar := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	p := &postgresImpl{}
	p.SqrD = &sqrDollar

	query, _, err := p.queryMarkChecksumClarificationUsedById(uuid.New())
	if err != nil {
		t.Fatalf("queryMarkChecksumClarificationUsedById() error = %v, want nil", err)
	}
	if !strings.Contains(query, "checksum_clarification_used_at") {
		t.Fatalf("query = %q, want it to set checksum_clarification_used_at", query)
	}
	if !strings.Contains(query, "infrared_record_session") {
		t.Fatalf("query = %q, want it to target infrared_record_session", query)
	}
}

func TestQueryReadByFilterJoinsDeviceAndDeviceType(t *testing.T) {
	sqrDollar := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	p := &postgresImpl{}
	p.SqrD = &sqrDollar

	_, _, query, _, err := p.queryReadByFilter(nil, nil, nil, nil, 1, 10)
	if err != nil {
		t.Fatalf("queryReadByFilter() error = %v, want nil", err)
	}
	if !strings.Contains(query, "infrared_device_type.name") {
		t.Fatalf("query = %q, want it to select the device type name", query)
	}
	if !strings.Contains(query, "deleted_at IS NULL") {
		t.Fatalf("query = %q, want it to exclude soft-deleted sessions", query)
	}
}

func TestQueryReadByFilterAppliesRecordingStateFilter(t *testing.T) {
	sqrDollar := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	p := &postgresImpl{}
	p.SqrD = &sqrDollar

	state := "RECORDING"
	_, _, query, args, err := p.queryReadByFilter(&state, nil, nil, nil, 1, 10)
	if err != nil {
		t.Fatalf("queryReadByFilter() error = %v, want nil", err)
	}
	if !strings.Contains(query, "recording_state") {
		t.Fatalf("query = %q, want a recording_state condition", query)
	}
	found := false
	for _, a := range args {
		if a == state {
			found = true
		}
	}
	if !found {
		t.Fatalf("args = %v, want %q among them", args, state)
	}
}
