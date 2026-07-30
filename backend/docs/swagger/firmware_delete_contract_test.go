package swagger

import (
	"encoding/json"
	"os"
	"testing"
)

func TestFirmwareDeleteDocumentsRequiredJSONConfirmation(t *testing.T) {
	data, err := os.ReadFile("swagger.json")
	if err != nil {
		t.Fatalf("ReadFile(swagger.json) error = %v", err)
	}

	var document struct {
		Definitions map[string]struct {
			Required []string `json:"required"`
		} `json:"definitions"`
		Paths map[string]map[string]struct {
			Consumes []string `json:"consumes"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatalf("Unmarshal(swagger.json) error = %v", err)
	}

	deleteOperation := document.Paths["/v1/firmwares/{id}"]["delete"]
	if len(deleteOperation.Consumes) != 1 || deleteOperation.Consumes[0] != "application/json" {
		t.Fatalf("DELETE consumes = %#v, want [application/json]", deleteOperation.Consumes)
	}

	const definition = "github_com_MateWorkspace_mate-things_backend_internal_presentation_http_request.FirmwareDeleteRequest"
	required := document.Definitions[definition].Required
	if len(required) != 1 || required[0] != "expected_name" {
		t.Fatalf("%s required = %#v, want [expected_name]", definition, required)
	}
}
