package loop

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExternalPermissionChecker_Allow(t *testing.T) {
	check := NewExternalPermissionChecker(
		`cat >/dev/null; echo '{"decision":"allow"}'`,
		func() string { return "/tmp" },
		"conv-1",
	)
	if err := check(context.Background(), "bash", json.RawMessage(`{"command":"ls"}`)); err != nil {
		t.Fatalf("expected allow, got error: %v", err)
	}
}

func TestExternalPermissionChecker_Deny(t *testing.T) {
	check := NewExternalPermissionChecker(
		`cat >/dev/null; echo '{"decision":"deny","reason":"nope"}'`,
		nil,
		"",
	)
	err := check(context.Background(), "bash", json.RawMessage(`{"command":"rm -rf /"}`))
	if err == nil {
		t.Fatal("expected deny error, got nil")
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Fatalf("expected reason in error, got: %v", err)
	}
}

func TestExternalPermissionChecker_PassesRequest(t *testing.T) {
	// Evaluator writes stdin to a file so we can inspect the request JSON.
	dir := t.TempDir()
	capturePath := filepath.Join(dir, "input.json")
	check := NewExternalPermissionChecker(
		`cat >`+capturePath+`; echo '{"decision":"allow"}'`,
		func() string { return "/work" },
		"conv-42",
	)
	if err := check(context.Background(), "bash", json.RawMessage(`{"command":"ls"}`)); err != nil {
		t.Fatalf("expected allow, got error: %v", err)
	}
	got, err := os.ReadFile(capturePath)
	if err != nil {
		t.Fatalf("read captured input: %v", err)
	}
	var req struct {
		ToolName       string          `json:"tool_name"`
		ToolInput      json.RawMessage `json:"tool_input"`
		WorkingDir     string          `json:"working_dir"`
		ConversationID string          `json:"conversation_id"`
	}
	if err := json.Unmarshal(got, &req); err != nil {
		t.Fatalf("unmarshal request: %v\ngot: %s", err, got)
	}
	if req.ToolName != "bash" {
		t.Errorf("tool_name = %q, want bash", req.ToolName)
	}
	if req.WorkingDir != "/work" {
		t.Errorf("working_dir = %q, want /work", req.WorkingDir)
	}
	if req.ConversationID != "conv-42" {
		t.Errorf("conversation_id = %q, want conv-42", req.ConversationID)
	}
	if !strings.Contains(string(req.ToolInput), `"command":"ls"`) {
		t.Errorf("tool_input = %s, want command=ls", req.ToolInput)
	}
}

func TestExternalPermissionChecker_NonZeroExit(t *testing.T) {
	check := NewExternalPermissionChecker(`exit 1`, nil, "")
	err := check(context.Background(), "bash", json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error on non-zero exit")
	}
}

func TestExternalPermissionChecker_BadJSON(t *testing.T) {
	check := NewExternalPermissionChecker(`cat >/dev/null; echo 'not json'`, nil, "")
	err := check(context.Background(), "bash", json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error on bad JSON")
	}
}
