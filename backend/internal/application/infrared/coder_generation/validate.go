package applicationinfraredcodergeneration

import (
	"time"

	applicationinfraredanalysis "github.com/MateWorkspace/mate-things/backend/internal/application/infrared/analysis"
	domaincontractsutility "github.com/MateWorkspace/mate-things/backend/internal/domain/contracts/utility"
)

// KnownCase is one state the pipeline already has real, recorded bits for
// — the baseline, plus every already-recorded non-baseline OFAT case.
type KnownCase struct {
	Label string
	State map[string]string
	Bits  []int
}

type CaseValidation struct {
	Label         string
	OwnedCorrect  int
	OwnedTotal    int
	ChecksumOK    int
	ChecksumTotal int
	RunError      string
}

type ValidationResult struct {
	Cases         []CaseValidation
	OwnedCorrect  int
	OwnedTotal    int
	ChecksumOK    int
	ChecksumTotal int
}

// Passed reports whether every known case ran without error and matched
// the real recorded bits exactly on every owned bit position. A
// ValidationResult with no cases is never "passed" — buildAnalysisPayload
// always requires at least a baseline case to exist before WriteCoder is
// ever called, so an empty Cases slice here is a caller bug, not a valid
// pass.
func (r ValidationResult) Passed() bool {
	if len(r.Cases) == 0 {
		return false
	}
	for _, c := range r.Cases {
		if c.RunError != "" || c.OwnedCorrect != c.OwnedTotal {
			return false
		}
	}
	return true
}

// ChecksumOnlyGap reports whether every owned bit is correct across every
// case (no RunError, no owned mismatch) but at least one checksum bit is
// wrong — the signal used to route to the checksum-clarification path
// instead of the repair loop, since a repair prompt can't fix a data gap.
func (r ValidationResult) ChecksumOnlyGap() bool {
	if len(r.Cases) == 0 {
		return false
	}
	for _, c := range r.Cases {
		if c.RunError != "" || c.OwnedCorrect != c.OwnedTotal {
			return false
		}
	}
	return r.ChecksumTotal > 0 && r.ChecksumOK < r.ChecksumTotal
}

// OwnedAccuracy returns 0 when OwnedTotal is 0 (e.g. every case errored)
// rather than dividing by zero.
func (r ValidationResult) OwnedAccuracy() float64 {
	if r.OwnedTotal == 0 {
		return 0
	}
	return float64(r.OwnedCorrect) / float64(r.OwnedTotal)
}

// Validate runs encoderSource against every knownCases entry via the real
// JSEngine, decodes its output with the same DemodulateBits production
// uses on real captures, and compares it bit-for-bit against that case's
// real recorded bits — owned-bit positions scored separately from
// checksumBits positions.
func Validate(
	runner domaincontractsutility.JSEngine,
	encoderSource string,
	checksumBits []int,
	knownCases []KnownCase,
	timeout time.Duration,
) ValidationResult {
	checksumSet := make(map[int]struct{}, len(checksumBits))
	for _, b := range checksumBits {
		checksumSet[b] = struct{}{}
	}

	var result ValidationResult
	for _, kc := range knownCases {
		cv := CaseValidation{Label: kc.Label}

		raw, err := runner.RunEncoder(encoderSource, kc.State, timeout)
		if err != nil {
			cv.RunError = err.Error()
			result.Cases = append(result.Cases, cv)
			continue
		}

		decoded := applicationinfraredanalysis.DemodulateBits(raw)
		for i, want := range kc.Bits {
			got := -1
			if i < len(decoded) {
				got = decoded[i]
			}
			if _, isChecksum := checksumSet[i]; isChecksum {
				cv.ChecksumTotal++
				if got == want {
					cv.ChecksumOK++
				}
			} else {
				cv.OwnedTotal++
				if got == want {
					cv.OwnedCorrect++
				}
			}
		}

		result.Cases = append(result.Cases, cv)
		result.OwnedCorrect += cv.OwnedCorrect
		result.OwnedTotal += cv.OwnedTotal
		result.ChecksumOK += cv.ChecksumOK
		result.ChecksumTotal += cv.ChecksumTotal
	}
	return result
}
