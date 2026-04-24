package loop

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ToolPermissionChecker is called before each tool invocation.
// A non-nil error denies the call; the error is surfaced to the LLM as the
// tool result, matching bashkit.Check's behavior. If the error is a
// *PermissionDeniedError, the loop will ask the user for approval before
// proceeding or surfacing the denial.
type ToolPermissionChecker func(ctx context.Context, toolName string, toolInput json.RawMessage) error

// PermissionDeniedError signals that an evaluator rejected a tool call.
// The loop treats this specially: it asks the user to override the decision
// rather than immediately returning an error to the model.
type PermissionDeniedError struct {
	Reason string
}

func (e *PermissionDeniedError) Error() string {
	if e.Reason != "" {
		return "tool call denied by permission check: " + e.Reason
	}
	return "tool call denied by permission check"
}

// permissionRequest is the JSON written to the evaluator command's stdin.
type permissionRequest struct {
	ToolName       string          `json:"tool_name"`
	ToolInput      json.RawMessage `json:"tool_input"`
	WorkingDir     string          `json:"working_dir,omitempty"`
	ConversationID string          `json:"conversation_id,omitempty"`
}

// permissionResponse is the JSON read from the evaluator command's stdout.
type permissionResponse struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason,omitempty"`
}

// NewExternalPermissionChecker returns a checker that invokes command via
// `bash -c`, writing a JSON permissionRequest to stdin and expecting a JSON
// permissionResponse on stdout. Any decision other than "allow" denies the
// tool call.
func NewExternalPermissionChecker(command string, getWorkingDir func() string, conversationID string) ToolPermissionChecker {
	return func(ctx context.Context, toolName string, toolInput json.RawMessage) error {
		wd := ""
		if getWorkingDir != nil {
			wd = getWorkingDir()
		}
		input := toolInput
		if len(input) == 0 {
			input = json.RawMessage("null")
		}
		reqBytes, err := json.Marshal(permissionRequest{
			ToolName:       toolName,
			ToolInput:      input,
			WorkingDir:     wd,
			ConversationID: conversationID,
		})
		if err != nil {
			return fmt.Errorf("permission check: marshal request: %w", err)
		}

		cmd := exec.CommandContext(ctx, "bash", "-c", command)
		cmd.Stdin = bytes.NewReader(reqBytes)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("permission check command failed: %w: %s", err, strings.TrimSpace(stderr.String()))
		}

		var resp permissionResponse
		if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &resp); err != nil {
			return fmt.Errorf("permission check: parse response %q: %w", strings.TrimSpace(stdout.String()), err)
		}

		if resp.Decision != "allow" {
			return &PermissionDeniedError{Reason: resp.Reason}
		}
		return nil
	}
}
