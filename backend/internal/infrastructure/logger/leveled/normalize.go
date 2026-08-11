package infrastructureloggerleveled

import domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"

// normalizeMeta replaces every error value in meta with its .Error() string.
// Structured log encoders (zerolog's Any, slog's Any) fall back to
// encoding/json for values they don't recognize, and json.Marshal on an
// error interface only sees the concrete type's exported fields - which is
// usually none (*errors.errorString, *fmt.wrapError, this codebase's own
// *domainmodels.Error), producing a useless "{}" in the log instead of the
// message. Converting to a string up front sidesteps that entirely.
func normalizeMeta(meta domainmodels.LoggerMeta) map[string]any {
	normalized := make(map[string]any, len(meta))
	for key, value := range meta {
		if err, ok := value.(error); ok && err != nil {
			normalized[key] = err.Error()
			continue
		}
		normalized[key] = value
	}
	return normalized
}
