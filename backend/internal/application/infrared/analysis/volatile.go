package applicationinfraredanalysis

import "fmt"

// DetectVolatileBits compares two bit sequences captured for the SAME case
// and returns the set of positions where they differ — these are volatile
// (rolling checksum, timestamp, toggle bit), not semantically meaningful.
func DetectVolatileBits(a, b []int) (map[int]struct{}, error) {
	if len(a) != len(b) {
		return nil, fmt.Errorf("bit sequence length mismatch: %d vs %d", len(a), len(b))
	}

	volatile := make(map[int]struct{})
	for i := range a {
		if a[i] != b[i] {
			volatile[i] = struct{}{}
		}
	}
	return volatile, nil
}

// UnionVolatileBits combines volatile-bit sets detected across multiple
// cases into one mask, since a bit volatile in any case must be excluded
// from semantic analysis everywhere.
func UnionVolatileBits(sets ...map[int]struct{}) map[int]struct{} {
	union := make(map[int]struct{})
	for _, set := range sets {
		for bit := range set {
			union[bit] = struct{}{}
		}
	}
	return union
}
