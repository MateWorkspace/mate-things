package applicationinfraredanalysis

import (
	"reflect"
	"testing"
)

func TestSegmentFramesSplitsOnLongGap(t *testing.T) {
	// header(9000,4500) + 2 data bits (560,560=0) (560,1690=1) + a trailing
	// mark(560) whose SPACE is the long inter-frame gap(20000) + repeat
	// frame. raw strictly alternates mark(even index), space(odd index);
	// the gap must land on an odd index to be detected as a space, not a
	// mark — a long header mark (9000µs) is not a gap.
	raw := []int32{9000, 4500, 560, 560, 560, 1690, 560, 20000, 9000, 4500, 560, 560}
	frames := SegmentFrames(raw)
	if len(frames) != 2 {
		t.Fatalf("SegmentFrames() returned %d frames, want 2", len(frames))
	}
	if !reflect.DeepEqual(frames[0], []int32{9000, 4500, 560, 560, 560, 1690, 560}) {
		t.Errorf("frames[0] = %v, want the first 7 values (including the trailing mark)", frames[0])
	}
	if !reflect.DeepEqual(frames[1], []int32{9000, 4500, 560, 560}) {
		t.Errorf("frames[1] = %v, want the repeat frame", frames[1])
	}
}

func TestSegmentFramesSingleFrameWhenNoLongGap(t *testing.T) {
	raw := []int32{9000, 4500, 560, 560, 560, 1690}
	frames := SegmentFrames(raw)
	if len(frames) != 1 {
		t.Fatalf("SegmentFrames() returned %d frames, want 1", len(frames))
	}
}

func TestDemodulateBitsSkipsHeaderAndClassifiesBySpaceMidpoint(t *testing.T) {
	// header(9000,4500) then bits: short space=0, long space=1, short=0, long=1
	frame := []int32{9000, 4500, 560, 560, 560, 1690, 560, 560, 560, 1690}
	bits := DemodulateBits(frame)
	want := []int{0, 1, 0, 1}
	if !reflect.DeepEqual(bits, want) {
		t.Fatalf("DemodulateBits() = %v, want %v", bits, want)
	}
}

func TestDemodulateBitsIgnoresDanglingTrailingMark(t *testing.T) {
	// header + one bit + a trailing mark with no paired space
	frame := []int32{9000, 4500, 560, 560, 560}
	bits := DemodulateBits(frame)
	if !reflect.DeepEqual(bits, []int{0}) {
		t.Fatalf("DemodulateBits() = %v, want [0] (trailing unpaired mark ignored)", bits)
	}
}
