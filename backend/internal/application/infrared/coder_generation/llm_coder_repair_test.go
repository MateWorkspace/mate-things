package applicationinfraredcodergeneration

import (
	"context"
	"strings"
	"testing"
)

func TestRepairCoderIncludesPriorSourceAndMismatchFeedback(t *testing.T) {
	responseBody := `{"encoder_source": "function encode(state) { return [9000, 4500]; }", "decoder_source": "function decode(raw) { return {}; }", "summary_readme": "s", "detail_readme": "d"}`
	client := &fakeLlmClient{responseText: responseBody}

	validation := ValidationResult{
		Cases: []CaseValidation{
			{Label: "baseline", RunError: "encoder threw: TEMP out of range"},
			{Label: "MODE=HEAT", OwnedCorrect: 10, OwnedTotal: 12},
		},
	}

	coder, err := RepairCoder(context.Background(), client, "Polytron", "PAC-09HDN", "function encode(old){}", "function decode(old){}", validation)
	if err != nil {
		t.Fatalf("RepairCoder() error = %v, want nil", err)
	}
	if coder.EncoderSource == "" {
		t.Fatal("RepairCoder() returned empty EncoderSource")
	}

	prompt := client.lastRequest.Prompt
	if !strings.Contains(prompt, "function encode(old){}") {
		t.Fatalf("prompt = %q, want it to include the prior encoder source", prompt)
	}
	if !strings.Contains(prompt, "TEMP out of range") {
		t.Fatalf("prompt = %q, want it to include the RunError feedback", prompt)
	}
	if !strings.Contains(prompt, "MODE=HEAT") {
		t.Fatalf("prompt = %q, want it to include the mismatching case's label", prompt)
	}
	if client.lastRequest.ResponseSchema == nil {
		t.Fatal("GenerateText() request had no ResponseSchema — structured output was not requested")
	}
}

func TestRepairCoderPropagatesLlmError(t *testing.T) {
	client := &fakeLlmClient{err: errPlaceholder}
	_, err := RepairCoder(context.Background(), client, "Polytron", "PAC-09HDN", "", "", ValidationResult{})
	if err == nil {
		t.Fatal("RepairCoder() error = nil, want propagated error")
	}
}
