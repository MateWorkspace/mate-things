package presentationhttprequest

import "encoding/json"

type PreferencesPatchRequest struct {
	Preferences json.RawMessage `json:"preferences" swaggertype:"object"`
}
