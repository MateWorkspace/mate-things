package applicationinfraredanalysis

const frameGapThresholdMicros = 5000

// SegmentFrames splits a flat mark/space duration sequence into candidate
// frames, treating any SPACE (odd index — raw alternates mark, space,
// mark, space...) longer than frameGapThresholdMicros as an inter-frame
// gap rather than a data bit. Marks are never checked against this
// threshold: a long header mark (e.g. ~9000µs, common in NEC-style
// protocols) is not a gap and must not trigger a split. See this task's
// design note in the implementation plan for why this threshold and not a
// per-protocol one.
func SegmentFrames(raw []int32) [][]int32 {
	if len(raw) == 0 {
		return nil
	}

	var frames [][]int32
	start := 0
	for i := 1; i < len(raw); i += 2 {
		if raw[i] > frameGapThresholdMicros {
			frames = append(frames, raw[start:i])
			start = i + 1
		}
	}
	frames = append(frames, raw[start:])
	return frames
}

// DemodulateBits converts one frame's mark/space durations into a bit
// sequence, treating the first mark+space pair as a header (excluded from
// the result) and classifying every subsequent space against the midpoint
// of that frame's own minimum and maximum space duration.
func DemodulateBits(frame []int32) []int {
	if len(frame) < 4 {
		return nil
	}

	spaces := make([]int32, 0, len(frame)/2)
	for i := 3; i < len(frame); i += 2 {
		spaces = append(spaces, frame[i])
	}
	if len(spaces) == 0 {
		return nil
	}

	min, max := spaces[0], spaces[0]
	for _, s := range spaces {
		if s < min {
			min = s
		}
		if s > max {
			max = s
		}
	}
	midpoint := (min + max) / 2

	bits := make([]int, len(spaces))
	for i, s := range spaces {
		if s > midpoint {
			bits[i] = 1
		}
	}
	return bits
}
