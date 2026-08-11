package applicationinfraredcodergeneration

import (
	"errors"
	"testing"
	"time"
)

type fakeJSEngine struct {
	raw []int32
	err error
}

func (f *fakeJSEngine) RunEncoder(_ string, _ map[string]string, _ time.Duration) ([]int32, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.raw, nil
}

func TestValidatePassesWhenEncoderMatchesEveryKnownCase(t *testing.T) {
	// [9000,4500] header + [560,560] one bit that decodes to bit 0 (space
	// 560 is not > the frame's own min/max midpoint of 560) — same fixture
	// shape as the analysis package's own tests.
	engine := &fakeJSEngine{raw: []int32{9000, 4500, 560, 560}}
	knownCases := []KnownCase{
		{Label: "baseline", State: map[string]string{"POWER": "OFF"}, Bits: []int{0}},
	}

	result := Validate(engine, "function encode(state){}", nil, knownCases, time.Second)

	if !result.Passed() {
		t.Fatalf("Validate() Passed() = false, want true; result = %+v", result)
	}
	if result.OwnedCorrect != 1 || result.OwnedTotal != 1 {
		t.Fatalf("owned = %d/%d, want 1/1", result.OwnedCorrect, result.OwnedTotal)
	}
}

func TestValidateFailsOnBitMismatch(t *testing.T) {
	// Two bits: real is [0,1] (spaces 560 then 1690 -> midpoint 1125 ->
	// bit0=560<1125=>0, bit1=1690>1125=>1); encoder always returns a raw
	// that decodes to [0,0] instead.
	engine := &fakeJSEngine{raw: []int32{9000, 4500, 560, 560, 560, 560}}
	knownCases := []KnownCase{
		{Label: "baseline", State: map[string]string{}, Bits: []int{0, 1}},
	}

	result := Validate(engine, "function encode(state){}", nil, knownCases, time.Second)

	if result.Passed() {
		t.Fatal("Validate() Passed() = true, want false (bit 1 mismatches)")
	}
	if result.OwnedCorrect != 1 || result.OwnedTotal != 2 {
		t.Fatalf("owned = %d/%d, want 1/2", result.OwnedCorrect, result.OwnedTotal)
	}
}

func TestValidateSeparatesChecksumBitsFromOwnedBits(t *testing.T) {
	// 2 bits: bit0 owned, bit1 is a checksum bit (checksumBits=[1]). Real
	// bits [0,1]; encoder returns [0,0] -> owned bit0 correct, checksum
	// bit1 wrong, but Passed() only cares about owned bits.
	engine := &fakeJSEngine{raw: []int32{9000, 4500, 560, 560, 560, 560}}
	knownCases := []KnownCase{
		{Label: "baseline", State: map[string]string{}, Bits: []int{0, 1}},
	}

	result := Validate(engine, "function encode(state){}", []int{1}, knownCases, time.Second)

	if !result.Passed() {
		t.Fatalf("Validate() Passed() = false, want true (checksum bits must not gate pass/fail); result = %+v", result)
	}
	if result.OwnedTotal != 1 || result.OwnedCorrect != 1 {
		t.Fatalf("owned = %d/%d, want 1/1 (bit 1 excluded, it's a checksum bit)", result.OwnedCorrect, result.OwnedTotal)
	}
	if result.ChecksumTotal != 1 || result.ChecksumOK != 0 {
		t.Fatalf("checksum = %d/%d, want 0/1", result.ChecksumOK, result.ChecksumTotal)
	}
}

func TestValidateRunErrorFailsThatCaseWithoutVacuousPass(t *testing.T) {
	// A case whose RunEncoder throws must never be counted as a trivial
	// 0/0 pass — Passed() must be false even though OwnedCorrect==OwnedTotal==0.
	engine := &fakeJSEngine{err: errors.New("encoder threw")}
	knownCases := []KnownCase{
		{Label: "baseline", State: map[string]string{}, Bits: []int{0}},
	}

	result := Validate(engine, "function encode(state){}", nil, knownCases, time.Second)

	if result.Passed() {
		t.Fatal("Validate() Passed() = true, want false (RunEncoder errored)")
	}
	if result.Cases[0].RunError == "" {
		t.Fatal("Cases[0].RunError is empty, want the RunEncoder error captured")
	}
	if result.OwnedAccuracy() != 0 {
		t.Fatalf("OwnedAccuracy() = %v, want 0 (no case ever produced comparable bits)", result.OwnedAccuracy())
	}
}

func TestValidateChecksumOnlyGapDetection(t *testing.T) {
	// Owned bits perfect, checksum bits wrong -> ChecksumOnlyGap() true.
	engine := &fakeJSEngine{raw: []int32{9000, 4500, 560, 560, 560, 560}}
	knownCases := []KnownCase{
		{Label: "baseline", State: map[string]string{}, Bits: []int{0, 1}},
	}
	result := Validate(engine, "function encode(state){}", []int{1}, knownCases, time.Second)
	if !result.ChecksumOnlyGap() {
		t.Fatal("ChecksumOnlyGap() = false, want true (owned bits clean, checksum bit wrong)")
	}

	// Owned bits also wrong -> ChecksumOnlyGap() false, even with a checksum mismatch too.
	engine2 := &fakeJSEngine{raw: []int32{9000, 4500, 1690, 1690, 560, 560}}
	result2 := Validate(engine2, "function encode(state){}", []int{1}, knownCases, time.Second)
	if result2.ChecksumOnlyGap() {
		t.Fatal("ChecksumOnlyGap() = true, want false (owned bit 0 is also wrong)")
	}
}
