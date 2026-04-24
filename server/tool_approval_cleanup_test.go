package server

import (
	"context"
	"encoding/json"
	"testing"

	"shelley.exe.dev/db"
)

func TestCancelOrphanedToolApprovals(t *testing.T) {
	svr, database, _ := newTestServer(t)
	ctx := context.Background()

	conv, err := database.CreateConversation(ctx, nil, true, nil, nil, db.ConversationOptions{})
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	mkApproval := func(status string) string {
		msg, err := database.CreateMessage(ctx, db.CreateMessageParams{
			ConversationID: conv.ConversationID,
			Type:           db.MessageTypeSystem,
			UserData: map[string]any{
				"kind":        "tool_approval_request",
				"approval_id": "id-" + status,
				"tool_name":   "bash",
				"status":      status,
			},
			ExcludedFromContext: true,
		})
		if err != nil {
			t.Fatalf("create message: %v", err)
		}
		return msg.MessageID
	}

	pendingID := mkApproval("pending")
	approvedID := mkApproval("approved")
	deniedID := mkApproval("denied")

	// Unrelated system message (e.g. distill_status) should be untouched.
	unrelated, err := database.CreateMessage(ctx, db.CreateMessageParams{
		ConversationID:      conv.ConversationID,
		Type:                db.MessageTypeSystem,
		UserData:            map[string]any{"distill_status": "in_progress"},
		ExcludedFromContext: true,
	})
	if err != nil {
		t.Fatalf("create unrelated: %v", err)
	}

	if err := svr.CancelOrphanedToolApprovals(ctx); err != nil {
		t.Fatalf("CancelOrphanedToolApprovals: %v", err)
	}

	readStatus := func(id string) string {
		msg, err := database.GetMessageByID(ctx, id)
		if err != nil {
			t.Fatalf("get %s: %v", id, err)
		}
		if msg.UserData == nil {
			return ""
		}
		var data map[string]any
		if err := json.Unmarshal([]byte(*msg.UserData), &data); err != nil {
			t.Fatalf("unmarshal %s: %v", id, err)
		}
		s, _ := data["status"].(string)
		return s
	}

	if got := readStatus(pendingID); got != "cancelled" {
		t.Errorf("pending approval status = %q, want cancelled", got)
	}
	if got := readStatus(approvedID); got != "approved" {
		t.Errorf("approved approval status = %q, want approved (unchanged)", got)
	}
	if got := readStatus(deniedID); got != "denied" {
		t.Errorf("denied approval status = %q, want denied (unchanged)", got)
	}

	unrelatedMsg, err := database.GetMessageByID(ctx, unrelated.MessageID)
	if err != nil {
		t.Fatalf("get unrelated: %v", err)
	}
	var unrelatedData map[string]any
	if err := json.Unmarshal([]byte(*unrelatedMsg.UserData), &unrelatedData); err != nil {
		t.Fatalf("unmarshal unrelated: %v", err)
	}
	if unrelatedData["distill_status"] != "in_progress" {
		t.Errorf("unrelated message disturbed: %v", unrelatedData)
	}
	if _, hasStatus := unrelatedData["status"]; hasStatus {
		t.Errorf("unrelated message gained a status field: %v", unrelatedData)
	}
}
