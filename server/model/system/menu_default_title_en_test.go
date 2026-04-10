package system

import "testing"

func TestDefaultMenuTitleEnglish_KnownRoutes(t *testing.T) {
	if got := DefaultMenuTitleEnglish("dashboard"); got != "Dashboard" {
		t.Fatalf("dashboard: got %q", got)
	}
	if got := DefaultMenuTitleEnglish("plugin-email"); got != "Email plugin" {
		t.Fatalf("plugin-email: got %q", got)
	}
	if got := DefaultMenuTitleEnglish("https://lightningrag.com"); got != "Official website" {
		t.Fatalf("external: got %q", got)
	}
	if got := DefaultMenuTitleEnglish("nonexistent_route_xyz"); got != "" {
		t.Fatalf("unknown: want empty, got %q", got)
	}
}
