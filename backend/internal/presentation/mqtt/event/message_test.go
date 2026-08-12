package presentationmqttevent

import "testing"

func TestExtractTopicPartsHandlesNestedSuffix(t *testing.T) {
	got, ok := extractTopicParts("/pub/ABCDEF012345/ir/rx")
	if !ok {
		t.Fatalf("expected ok=true, got false")
	}
	want := topicParts{Direction: "pub", DeviceId: "ABCDEF012345", Suffix: "ir/rx"}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestExtractTopicPartsHandlesFlatSuffix(t *testing.T) {
	got, ok := extractTopicParts("/pub/ABCDEF012345/status")
	if !ok {
		t.Fatalf("expected ok=true, got false")
	}
	want := topicParts{Direction: "pub", DeviceId: "ABCDEF012345", Suffix: "status"}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestExtractTopicPartsHandlesGlobalSuffix(t *testing.T) {
	got, ok := extractTopicParts("/pub/registration")
	if !ok {
		t.Fatalf("expected ok=true, got false")
	}
	want := topicParts{Direction: "pub", Suffix: "registration", IsGlobal: true}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestExtractTopicPartsRejectsEmptyNestedSegment(t *testing.T) {
	if _, ok := extractTopicParts("/pub/ABCDEF012345/ir//rx"); ok {
		t.Fatalf("expected ok=false for empty inner segment")
	}
}

func TestExtractTopicPartsRejectsTooFewSegments(t *testing.T) {
	if _, ok := extractTopicParts("/pub"); ok {
		t.Fatalf("expected ok=false for a single segment")
	}
}
