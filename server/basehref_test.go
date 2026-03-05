package server

import (
	"strings"
	"testing"
)

func TestBaseHrefForPath(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: "/"},
		{name: "root", in: "/", want: "/"},
		{name: "normal", in: "/shelley", want: "/shelley/"},
		{name: "trailing slash", in: "/shelley/", want: "/shelley/"},
		{name: "trim spaces", in: "  /shelley  ", want: "/shelley/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := baseHrefForPath(tt.in); got != tt.want {
				t.Fatalf("baseHrefForPath(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestInjectBaseHrefReplacesPlaceholder(t *testing.T) {
	input := `<html><head><base href="/" /><link rel="stylesheet" href="styles.css"></head><body></body></html>`
	got := injectBaseHref(input, "/__arca/shelley")
	if strings.Contains(got, `<base href="/" />`) {
		t.Fatalf("expected placeholder to be replaced: %q", got)
	}
	if !strings.Contains(got, `<base href="/__arca/shelley/" />`) {
		t.Fatalf("expected rewritten base href in %q", got)
	}
}

func TestInjectBaseHrefAddsWhenMissing(t *testing.T) {
	input := `<html><head><title>x</title></head><body></body></html>`
	got := injectBaseHref(input, "/__arca/shelley")
	if !strings.Contains(got, `<base href="/__arca/shelley/" />`) {
		t.Fatalf("expected injected base href in %q", got)
	}
}
