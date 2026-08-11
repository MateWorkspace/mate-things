package applicationinfraredcodergeneration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	domaincontractsllm "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/llm"
	domainmodels "github.com/MateWorkspace/mate-things/backend/internal/domain/models"
)

// RepairCoder asks the LLM to fix a previously-generated encoder/decoder
// pair, given the concrete per-case validation results (RunEncoder errors
// and/or bit-level mismatches against real recorded signals) from
// Validate(). Reuses WriteCoder's response schema and return type — a
// repair response is shaped identically to an initial one.
func RepairCoder(
	ctx context.Context,
	client domaincontractsllm.Client,
	deviceBrand string,
	deviceModel string,
	priorEncoderSource string,
	priorDecoderSource string,
	validation ValidationResult,
) (Coder, error) {
	var prompt strings.Builder
	fmt.Fprintf(&prompt, "Device: %s %s\n\n", deviceBrand, deviceModel)
	prompt.WriteString("Your previous encoder:\n\n```javascript\n")
	prompt.WriteString(priorEncoderSource)
	prompt.WriteString("\n```\n\nYour previous decoder:\n\n```javascript\n")
	prompt.WriteString(priorDecoderSource)
	prompt.WriteString("\n```\n\n")
	prompt.WriteString(buildRepairFeedback(validation))
	prompt.WriteString("\nFix the encoder (and decoder if relevant) so every case above is correct. Respond with the complete corrected JSON matching the same schema as before (encoder_source, decoder_source, summary_readme, detail_readme) — do not send a partial diff.")

	result, err := client.GenerateText(ctx, domaincontractsllm.GenerateTextRequest{
		System:          "You are an expert at reverse-engineering infrared remote control protocols and writing correct JavaScript encoders/decoders from bit-level analysis data and test feedback.",
		Prompt:          prompt.String(),
		MaxOutputTokens: 8192,
		ResponseSchema:  []byte(responseSchema),
	})
	if err != nil {
		return Coder{}, domainmodels.NewError("failed to generate repaired encoder/decoder", domainmodels.ErrTypeFailure, err)
	}

	var response coderResponse
	if err := json.Unmarshal([]byte(result.Text), &response); err != nil {
		return Coder{}, domainmodels.NewError("llm returned malformed repair response", domainmodels.ErrTypeFailure, err)
	}

	return Coder{
		EncoderSource: response.EncoderSource,
		DecoderSource: response.DecoderSource,
		SummaryReadme: response.SummaryReadme,
		DetailReadme:  response.DetailReadme,
	}, nil
}

// buildRepairFeedback turns a ValidationResult into the concrete,
// per-case feedback text the repair prompt sends back to the LLM. For each
// case, it reports: RunError verbatim if the encoder threw, "correct" if
// all owned bits matched, or an aggregate count of mismatched owned bits.
func buildRepairFeedback(validation ValidationResult) string {
	var b strings.Builder
	b.WriteString("Your previous encoder was tested against the real recorded signal for several states of this device and had problems. Details per test case:\n\n")
	for _, c := range validation.Cases {
		if c.RunError != "" {
			fmt.Fprintf(&b, "- Case %q: the encoder THREW A RUNTIME ERROR: %q. This must not happen for any valid state value within the documented ranges.\n", c.Label, c.RunError)
			continue
		}
		if c.OwnedCorrect == c.OwnedTotal {
			fmt.Fprintf(&b, "- Case %q: correct.\n", c.Label)
			continue
		}
		fmt.Fprintf(&b, "- Case %q: %d/%d owned bits wrong.\n", c.Label, c.OwnedTotal-c.OwnedCorrect, c.OwnedTotal)
	}
	return b.String()
}
