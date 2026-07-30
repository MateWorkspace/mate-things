package swagger

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
)

func TestFleetMutationSchemasDocumentProtectedIdentityAndPresenceContracts(t *testing.T) {
	data, err := os.ReadFile("swagger.json")
	if err != nil {
		t.Fatalf("ReadFile(swagger.json) error = %v", err)
	}

	var document struct {
		Definitions map[string]struct {
			Properties map[string]json.RawMessage `json:"properties"`
			Required   []string                   `json:"required"`
		} `json:"definitions"`
		Paths map[string]map[string]struct {
			Responses map[string]json.RawMessage `json:"responses"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("Unmarshal(swagger.json) error = %v", err)
	}

	const prefix = "github_com_MateWorkspace_mate-things_backend_internal_presentation_http_request."
	ota := document.Definitions[prefix+"OtaDispatchRequest"]
	if got := sortedKeys(ota.Properties); !equalStrings(got, []string{"firmware_id"}) {
		t.Fatalf("OtaDispatchRequest properties = %#v, want [firmware_id]", got)
	}
	if !equalStrings(ota.Required, []string{"firmware_id"}) {
		t.Fatalf("OtaDispatchRequest required = %#v, want [firmware_id]", ota.Required)
	}

	nodePatch := document.Definitions[prefix+"NodePatchRequest"]
	if _, exists := nodePatch.Properties["device_id"]; exists {
		t.Fatal("NodePatchRequest documents mutable device_id, want immutable identity")
	}

	config := document.Definitions[prefix+"SetNodeConfigValueRequest"]
	if !equalStrings(config.Required, []string{"value"}) {
		t.Fatalf("SetNodeConfigValueRequest required = %#v, want [value]", config.Required)
	}

	for _, path := range []string{
		"/v1/nodes/{id}/ota",
		"/v1/nodes/by-device/{device_id}/ota",
	} {
		if _, exists := document.Paths[path]["post"].Responses["204"]; !exists {
			t.Fatalf("%s POST responses omit 204", path)
		}
	}
}

func sortedKeys(values map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func equalStrings(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
