package server

import "testing"

func TestNormalizeBasePath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "empty is root", input: "", want: "/"},
		{name: "slash is root", input: "/", want: "/"},
		{name: "normal", input: "/shelley", want: "/shelley"},
		{name: "trim trailing slash", input: "/shelley/", want: "/shelley"},
		{name: "must start with slash", input: "shelley", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeBasePath(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("normalizeBasePath(%q) expected error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeBasePath(%q) error = %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("normalizeBasePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
