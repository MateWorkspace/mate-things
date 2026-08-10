package infrastructurejsengine

import (
	"strings"
	"testing"
	"time"
)

func TestRunEncoderReturnsArrayFromValidSource(t *testing.T) {
	source := `function encode(state) { return state.POWER === "ON" ? [9000, 4500, 560, 560] : [9000, 4500, 560, 1690]; }`
	result, err := RunEncoder(source, map[string]string{"POWER": "ON"}, time.Second)
	if err != nil {
		t.Fatalf("RunEncoder() error = %v, want nil", err)
	}
	want := []int32{9000, 4500, 560, 560}
	if len(result) != len(want) {
		t.Fatalf("RunEncoder() = %v, want %v", result, want)
	}
	for i := range want {
		if result[i] != want[i] {
			t.Fatalf("RunEncoder() = %v, want %v", result, want)
		}
	}
}

func TestRunEncoderPropagatesThrownError(t *testing.T) {
	source := `function encode(state) { throw new Error("unsupported state"); }`
	_, err := RunEncoder(source, map[string]string{}, time.Second)
	if err == nil {
		t.Fatal("RunEncoder() error = nil, want the thrown error propagated")
	}
	if !strings.Contains(err.Error(), "unsupported state") {
		t.Fatalf("RunEncoder() error = %v, want it to mention the thrown message", err)
	}
}

func TestRunEncoderErrorsOnNonArrayReturn(t *testing.T) {
	source := `function encode(state) { return "not an array"; }`
	_, err := RunEncoder(source, map[string]string{}, time.Second)
	if err == nil {
		t.Fatal("RunEncoder() error = nil, want an error for a non-array return value")
	}
}

func TestRunEncoderTimesOutOnInfiniteLoop(t *testing.T) {
	source := `function encode(state) { while (true) {} }`
	start := time.Now()
	_, err := RunEncoder(source, map[string]string{}, 100*time.Millisecond)
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("RunEncoder() error = nil, want a timeout error")
	}
	if elapsed > 2*time.Second {
		t.Fatalf("RunEncoder() took %v, want it to return promptly after the timeout", elapsed)
	}
}
