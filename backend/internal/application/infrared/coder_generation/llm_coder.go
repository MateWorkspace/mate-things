package applicationinfraredcodergeneration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	applicationinfraredanalysis "github.com/MateWorkspace/mate-things/backend/internal/application/infrared/analysis"
	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

const responseSchema = `{
	"type": "object",
	"properties": {
		"encoder_source": {"type": "string"},
		"decoder_source": {"type": "string"},
		"summary_readme": {"type": "string"},
		"detail_readme": {"type": "string"}
	},
	"required": ["encoder_source", "decoder_source", "summary_readme", "detail_readme"],
	"additionalProperties": false
}`

type Coder struct {
	EncoderSource string
	DecoderSource string
	SummaryReadme string
	DetailReadme  string
}

type coderResponse struct {
	EncoderSource string `json:"encoder_source"`
	DecoderSource string `json:"decoder_source"`
	SummaryReadme string `json:"summary_readme"`
	DetailReadme  string `json:"detail_readme"`
}

// WriteCoder asks the LLM to write a JavaScript encoder/decoder pair plus
// two README fields, given the deterministic bit-analysis payload as
// context. The LLM's job is pattern recognition and code generation over
// data Go has already extracted — it never re-derives which bits belong to
// which state; that's payload.States, computed in Task 6.
func WriteCoder(
	ctx context.Context,
	client domaincontractsllm.Client,
	deviceBrand string,
	deviceModel string,
	states []domainmodels.InfraredState,
	definitions []domainmodels.InfraredStateDeviceDefinition,
	payload applicationinfraredanalysis.AnalysisPayload,
	baselineValues map[string]string,
) (Coder, error) {
	stateNameById := make(map[string]string, len(states))
	for _, state := range states {
		stateNameById[state.Id.String()] = state.Name
	}

	var promptBuilder strings.Builder
	fmt.Fprintf(&promptBuilder, "Device: %s %s\n\n", deviceBrand, deviceModel)
	fmt.Fprintf(&promptBuilder, "Frame is %d bits. Baseline bits: %v. Checksum bit offsets (do not treat as state data): %v. Volatile bit offsets (rolling/timestamp, ignore): %v.\n\n",
		payload.FrameBitLength, payload.BaselineBits, payload.ChecksumBits, payload.VolatileBits)
	if len(baselineValues) > 0 {
		var baselineParts []string
		for _, s := range states {
			if value, ok := baselineValues[s.Name]; ok {
				baselineParts = append(baselineParts, fmt.Sprintf("%s=%s", s.Name, value))
			}
		}
		if len(baselineParts) > 0 {
			fmt.Fprintf(&promptBuilder, "The baseline (most common) state is: %s. Every state's baseline value MUST be a valid, correctly-encoded input to your encoder — do not write validation logic that only accepts the specific non-baseline values mentioned below.\n\n", strings.Join(baselineParts, ", "))
		}
	}
	promptBuilder.WriteString("Per-state bit ownership and observed value patterns:\n")
	for _, s := range payload.States {
		fmt.Fprintf(&promptBuilder, "- %s: bit offsets %v\n", stateNameById[s.StateId.String()], s.BitOffsets)
		for value, bits := range s.ValueBits {
			fmt.Fprintf(&promptBuilder, "  value %q -> bits %v\n", value, bits)
		}
	}
	promptBuilder.WriteString("\nWrite a JavaScript encoder function `function encode(state)` taking an object keyed by state name (e.g. state.POWER, state.MODE) with string values, returning an array of mark/space microsecond durations for the full IR frame including header and any checksum computation your analysis of the bit layout implies. Also write a best-effort `function decode(raw)` inverse (not verified by this pipeline). Then write a short protocol summary and a longer detailed explanation of the encoding (frame structure, checksum algorithm if any, per-state bit meaning). Respond as JSON matching the given schema.")

	result, err := client.GenerateText(ctx, domaincontractsllm.GenerateTextRequest{
		System:          "You are an expert at reverse-engineering infrared remote control protocols and writing correct JavaScript encoders/decoders from bit-level analysis data.",
		Prompt:          promptBuilder.String(),
		MaxOutputTokens: 8192,
		ResponseSchema:  []byte(responseSchema),
	})
	if err != nil {
		return Coder{}, domainmodels.NewError("failed to generate encoder/decoder", domainmodels.ErrTypeFailure, err)
	}

	var response coderResponse
	if err := json.Unmarshal([]byte(result.Text), &response); err != nil {
		return Coder{}, domainmodels.NewError("llm returned malformed coder response", domainmodels.ErrTypeFailure, err)
	}

	return Coder{
		EncoderSource: response.EncoderSource,
		DecoderSource: response.DecoderSource,
		SummaryReadme: response.SummaryReadme,
		DetailReadme:  response.DetailReadme,
	}, nil
}
