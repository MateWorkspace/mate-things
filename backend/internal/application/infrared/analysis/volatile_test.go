package applicationinfraredanalysis

import (
	"reflect"
	"testing"
)

func TestDetectVolatileBitsFindsDifferingPositions(t *testing.T) {
	a := []int{0, 1, 0, 1, 0}
	b := []int{0, 1, 1, 1, 1}
	got, err := DetectVolatileBits(a, b)
	if err != nil {
		t.Fatalf("DetectVolatileBits() error = %v, want nil", err)
	}
	want := map[int]struct{}{2: {}, 4: {}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DetectVolatileBits() = %v, want %v", got, want)
	}
}

func TestDetectVolatileBitsReturnsEmptySetWhenIdentical(t *testing.T) {
	a := []int{1, 0, 1}
	got, err := DetectVolatileBits(a, append([]int{}, a...))
	if err != nil {
		t.Fatalf("DetectVolatileBits() error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Fatalf("DetectVolatileBits() = %v, want empty", got)
	}
}

func TestDetectVolatileBitsErrorsOnLengthMismatch(t *testing.T) {
	_, err := DetectVolatileBits([]int{0, 1}, []int{0, 1, 1})
	if err == nil {
		t.Fatal("DetectVolatileBits() error = nil, want a length-mismatch error")
	}
}

func TestUnionVolatileBitsCombinesMultipleSets(t *testing.T) {
	got := UnionVolatileBits(
		map[int]struct{}{1: {}},
		map[int]struct{}{2: {}, 3: {}},
		map[int]struct{}{1: {}},
	)
	want := map[int]struct{}{1: {}, 2: {}, 3: {}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("UnionVolatileBits() = %v, want %v", got, want)
	}
}
