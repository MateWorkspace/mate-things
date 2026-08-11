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
