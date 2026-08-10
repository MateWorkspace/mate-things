package applicationinfraredanalysis

import (
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"
)

func TestAttributeAssignsNonOverlappingBitsToEachState(t *testing.T) {
	powerId, modeId := uuid.New(), uuid.New()

	baseline := []int{0, 0, 0, 0} // POWER at bit0, MODE at bits1-2, bit3 unused
	caseBits := map[uuid.UUID][]int{
		powerId: {1, 0, 0, 0}, // POWER=OFF flips only bit0
		modeId:  {0, 1, 0, 0}, // MODE=HEAT flips only bit1
	}
	caseTargetState := map[uuid.UUID]uuid.UUID{powerId: powerId, modeId: modeId}
	caseTargetValue := map[uuid.UUID]string{powerId: "OFF", modeId: "HEAT"}

	payload, err := Attribute(baseline, caseBits, caseTargetState, caseTargetValue, map[int]struct{}{})
	if err != nil {
		t.Fatalf("Attribute() error = %v, want nil", err)
	}
	if len(payload.ChecksumBits) != 0 {
		t.Fatalf("ChecksumBits = %v, want empty (no bit changed for more than one state)", payload.ChecksumBits)
	}

	byState := make(map[uuid.UUID]StateAttribution, len(payload.States))
	for _, s := range payload.States {
		byState[s.StateId] = s
	}
	if !reflect.DeepEqual(byState[powerId].BitOffsets, []int{0}) {
		t.Errorf("POWER bit offsets = %v, want [0]", byState[powerId].BitOffsets)
	}
	if !reflect.DeepEqual(byState[modeId].BitOffsets, []int{1}) {
		t.Errorf("MODE bit offsets = %v, want [1]", byState[modeId].BitOffsets)
	}
	if !reflect.DeepEqual(byState[powerId].ValueBits["OFF"], []int{1}) {
		t.Errorf("POWER OFF value bits = %v, want [1]", byState[powerId].ValueBits["OFF"])
	}
}

func TestAttributeIdentifiesChecksumBitSharedAcrossStates(t *testing.T) {
	powerId, modeId := uuid.New(), uuid.New()

	baseline := []int{0, 0, 0, 0} // bit3 = checksum, flips whenever anything changes
	caseBits := map[uuid.UUID][]int{
		powerId: {1, 0, 0, 1}, // POWER=OFF: bit0 (owned) + bit3 (checksum)
		modeId:  {0, 1, 0, 1}, // MODE=HEAT: bit1 (owned) + bit3 (checksum)
	}
	caseTargetState := map[uuid.UUID]uuid.UUID{powerId: powerId, modeId: modeId}
	caseTargetValue := map[uuid.UUID]string{powerId: "OFF", modeId: "HEAT"}

	payload, err := Attribute(baseline, caseBits, caseTargetState, caseTargetValue, map[int]struct{}{})
	if err != nil {
		t.Fatalf("Attribute() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(payload.ChecksumBits, []int{3}) {
		t.Fatalf("ChecksumBits = %v, want [3]", payload.ChecksumBits)
	}

	for _, s := range payload.States {
		for _, offset := range s.BitOffsets {
			if offset == 3 {
				t.Fatalf("state %v owns bit 3, which is a checksum bit and must be excluded", s.StateId)
			}
		}
	}
}

func TestAttributeExcludesVolatileBitsFromEverything(t *testing.T) {
	powerId := uuid.New()
	baseline := []int{0, 0, 0}
	caseBits := map[uuid.UUID][]int{powerId: {1, 0, 1}} // bit2 volatile, would otherwise look checksum-like
	caseTargetState := map[uuid.UUID]uuid.UUID{powerId: powerId}
	caseTargetValue := map[uuid.UUID]string{powerId: "OFF"}
	volatile := map[int]struct{}{2: {}}

	payload, err := Attribute(baseline, caseBits, caseTargetState, caseTargetValue, volatile)
	if err != nil {
		t.Fatalf("Attribute() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(payload.VolatileBits, []int{2}) {
		t.Fatalf("VolatileBits = %v, want [2]", payload.VolatileBits)
	}
	for _, s := range payload.States {
		for _, offset := range s.BitOffsets {
			if offset == 2 {
				t.Fatalf("state %v owns volatile bit 2, must be excluded", s.StateId)
			}
		}
	}
	sort.Ints(payload.ChecksumBits) // no-op if empty, keeps this test order-independent
	if len(payload.ChecksumBits) != 0 {
		t.Fatalf("ChecksumBits = %v, want empty", payload.ChecksumBits)
	}
}
