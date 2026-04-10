package llm

import (
	"strings"
	"testing"
)

func TestNewStreamLineScannerLongLine(t *testing.T) {
	var sb strings.Builder
	sb.WriteString(`data: {"k":"`)
	sb.WriteString(strings.Repeat("z", 200_000))
	sb.WriteString(`"}`)
	s := newStreamLineScanner(strings.NewReader(sb.String() + "\n"))
	if !s.Scan() {
		t.Fatalf("Scan: %v", s.Err())
	}
	if len(s.Text()) < 150_000 {
		t.Fatalf("expected long line, got len %d", len(s.Text()))
	}
}
