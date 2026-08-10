package applicationinfraredanalysis

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

type StateAttribution struct {
	StateId    uuid.UUID
	BitOffsets []int
	ValueBits  map[string][]int
}

type AnalysisPayload struct {
	FrameBitLength int
	BaselineBits   []int
	ChecksumBits   []int
	VolatileBits   []int
	States         []StateAttribution
}

// Attribute discovers checksum bits and per-state bit ownership from a
// baseline case's bits plus every non-baseline case's bits, given which
// single state each non-baseline case targets (by this plan's OFAT
// construction, exactly one). See this task's design note in the
// implementation plan for the intersection-based checksum-discovery
// rationale.
func Attribute(
	baselineBits []int,
	caseBits map[uuid.UUID][]int,
	caseTargetState map[uuid.UUID]uuid.UUID,
	caseTargetValue map[uuid.UUID]string,
	volatile map[int]struct{},
) (AnalysisPayload, error) {
	perStateDelta := make(map[uuid.UUID]map[int]struct{})
	for caseId, bits := range caseBits {
		if len(bits) != len(baselineBits) {
			return AnalysisPayload{}, fmt.Errorf("case %s bit length %d does not match baseline length %d", caseId, len(bits), len(baselineBits))
		}
		stateId, ok := caseTargetState[caseId]
		if !ok {
			return AnalysisPayload{}, fmt.Errorf("case %s has no target state", caseId)
		}
		if _, ok := perStateDelta[stateId]; !ok {
			perStateDelta[stateId] = make(map[int]struct{})
		}
		for i := range bits {
			if _, isVolatile := volatile[i]; isVolatile {
				continue
			}
			if bits[i] != baselineBits[i] {
				perStateDelta[stateId][i] = struct{}{}
			}
		}
	}

	checksumBits := make(map[int]struct{})
	stateIds := make([]uuid.UUID, 0, len(perStateDelta))
	for stateId := range perStateDelta {
		stateIds = append(stateIds, stateId)
	}
	for i := 0; i < len(stateIds); i++ {
		for j := i + 1; j < len(stateIds); j++ {
			for bit := range perStateDelta[stateIds[i]] {
				if _, sharedWithOther := perStateDelta[stateIds[j]][bit]; sharedWithOther {
					checksumBits[bit] = struct{}{}
				}
			}
		}
	}

	states := make([]StateAttribution, 0, len(perStateDelta))
	for stateId, delta := range perStateDelta {
		offsets := make([]int, 0, len(delta))
		for bit := range delta {
			if _, isChecksum := checksumBits[bit]; !isChecksum {
				offsets = append(offsets, bit)
			}
		}
		sort.Ints(offsets)

		valueBits := make(map[string][]int)
		for caseId, target := range caseTargetState {
			if target != stateId {
				continue
			}
			bits := caseBits[caseId]
			pattern := make([]int, len(offsets))
			for k, offset := range offsets {
				pattern[k] = bits[offset]
			}
			valueBits[caseTargetValue[caseId]] = pattern
		}

		states = append(states, StateAttribution{StateId: stateId, BitOffsets: offsets, ValueBits: valueBits})
	}
	sort.Slice(states, func(i, j int) bool { return states[i].StateId.String() < states[j].StateId.String() })

	checksumList := make([]int, 0, len(checksumBits))
	for bit := range checksumBits {
		checksumList = append(checksumList, bit)
	}
	sort.Ints(checksumList)

	volatileList := make([]int, 0, len(volatile))
	for bit := range volatile {
		volatileList = append(volatileList, bit)
	}
	sort.Ints(volatileList)

	return AnalysisPayload{
		FrameBitLength: len(baselineBits),
		BaselineBits:   baselineBits,
		ChecksumBits:   checksumList,
		VolatileBits:   volatileList,
		States:         states,
	}, nil
}
