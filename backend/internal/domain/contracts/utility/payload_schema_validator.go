package domaincontractsutility

import (
	"context"
	"encoding/json"
)

type PayloadSchemaValidator interface {
	Validate(ctx context.Context, definition json.RawMessage, payload json.RawMessage) error
}
