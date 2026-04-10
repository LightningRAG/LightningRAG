package channel

import (
	"slices"
	"testing"
)

func TestRegisteredKinds(t *testing.T) {
	kinds := RegisteredKinds()
	if len(kinds) == 0 {
		t.Fatal("RegisteredKinds: expected non-empty list")
	}
	if !slices.IsSorted(kinds) {
		t.Fatalf("RegisteredKinds: want sorted, got %v", kinds)
	}
	// mock 适配器始终注册，作为基线
	if !slices.Contains(kinds, "mock") {
		t.Fatalf("RegisteredKinds: missing mock, got %v", kinds)
	}
}
