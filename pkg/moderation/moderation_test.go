package moderation

import (
	"testing"
	"time"
)

func TestDenyList_AddRemoveIsDenied(t *testing.T) {
	dl := NewMockDenyList()

	denied, _ := dl.IsDenied("vid-123")
	if denied {
		t.Fatal("expected not denied initially")
	}

	if err := dl.Add("vid-123", "copyright"); err != nil {
		t.Fatal(err)
	}

	denied, _ = dl.IsDenied("vid-123")
	if !denied {
		t.Fatal("expected denied after Add")
	}

	if err := dl.Remove("vid-123"); err != nil {
		t.Fatal(err)
	}

	denied, _ = dl.IsDenied("vid-123")
	if denied {
		t.Fatal("expected not denied after Remove")
	}
}

func TestDenyList_RemoveNonexistent(t *testing.T) {
	dl := NewMockDenyList()
	if err := dl.Remove("nonexistent"); err == nil {
		t.Fatal("expected error removing nonexistent entry")
	}
}

func TestDenyList_List(t *testing.T) {
	dl := NewMockDenyList()
	_ = dl.Add("vid-1", "abuse")
	_ = dl.Add("vid-2", "illegal")

	entries, _ := dl.List()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestModerationQueue_SubmitAndGetPending(t *testing.T) {
	dl := NewMockDenyList()
	al := NewMockAuditLog()
	q := NewMockModerationQueue(dl, al, DefaultEscalationConfig())

	flag := ContentFlag{
		ID:        "f1",
		ContentID: "vid-456",
		FlaggedBy: "user-1",
		Category:  CategoryCopyright,
		Evidence:  "matches known work X",
	}
	if err := q.Submit(flag); err != nil {
		t.Fatal(err)
	}

	pending, _ := q.GetPending()
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending, got %d", len(pending))
	}
	if pending[0].ContentID != "vid-456" {
		t.Fatalf("wrong content ID: %s", pending[0].ContentID)
	}
}

func TestModerationQueue_ReviewDeny(t *testing.T) {
	dl := NewMockDenyList()
	al := NewMockAuditLog()
	q := NewMockModerationQueue(dl, al, DefaultEscalationConfig())

	_ = q.Submit(ContentFlag{ID: "f1", ContentID: "vid-789", Category: CategoryIllegal})

	if err := q.Review("f1", ActionDeny, "admin-1"); err != nil {
		t.Fatal(err)
	}

	// Should be in denylist now
	denied, _ := dl.IsDenied("vid-789")
	if !denied {
		t.Fatal("expected content denied after review")
	}

	// Should have audit record
	records, _ := al.GetByFlag("f1")
	if len(records) != 1 {
		t.Fatalf("expected 1 audit record, got %d", len(records))
	}
	if records[0].ActionBy != "admin-1" {
		t.Fatalf("wrong reviewer: %s", records[0].ActionBy)
	}

	// No longer pending
	pending, _ := q.GetPending()
	if len(pending) != 0 {
		t.Fatalf("expected 0 pending after review, got %d", len(pending))
	}
}

func TestModerationQueue_ReviewApprove(t *testing.T) {
	dl := NewMockDenyList()
	al := NewMockAuditLog()
	q := NewMockModerationQueue(dl, al, DefaultEscalationConfig())

	_ = q.Submit(ContentFlag{ID: "f1", ContentID: "vid-ok", Category: CategoryAbuse})
	_ = q.Review("f1", ActionApprove, "admin-2")

	denied, _ := dl.IsDenied("vid-ok")
	if denied {
		t.Fatal("approved content should not be denied")
	}
}

func TestModerationQueue_Escalate(t *testing.T) {
	dl := NewMockDenyList()
	al := NewMockAuditLog()
	q := NewMockModerationQueue(dl, al, DefaultEscalationConfig())

	_ = q.Submit(ContentFlag{ID: "f1", ContentID: "vid-esc"})
	_ = q.Escalate("f1")

	if !q.IsEscalated("f1") {
		t.Fatal("expected flag to be escalated")
	}
}

func TestModerationQueue_AutoEscalation(t *testing.T) {
	dl := NewMockDenyList()
	al := NewMockAuditLog()
	cfg := EscalationConfig{FlagThreshold: 3, Window: time.Hour}
	q := NewMockModerationQueue(dl, al, cfg)

	// Submit 3 flags for same content within window
	for i := 0; i < 3; i++ {
		_ = q.Submit(ContentFlag{
			ContentID: "vid-hot",
			FlaggedBy: "user-" + string(rune('a'+i)),
			Category:  CategoryAbuse,
		})
	}

	// The third flag should trigger auto-escalation
	if !q.IsEscalated("flag-3") {
		t.Fatal("expected auto-escalation after 3 flags")
	}
}

func TestSyncBroadcaster(t *testing.T) {
	b := NewMockSyncBroadcaster()

	_ = b.BroadcastDenylist([]string{"seeder-1", "seeder-2"})
	_ = b.SyncSeeder("seeder-3")

	if len(b.Broadcasts) != 1 {
		t.Fatalf("expected 1 broadcast, got %d", len(b.Broadcasts))
	}
	if len(b.Broadcasts[0]) != 2 {
		t.Fatalf("expected 2 seeders in broadcast, got %d", len(b.Broadcasts[0]))
	}
	if len(b.SyncedPeers) != 1 || b.SyncedPeers[0] != "seeder-3" {
		t.Fatal("expected seeder-3 synced")
	}
}

func TestAuditLog(t *testing.T) {
	al := NewMockAuditLog()

	_ = al.Append(AuditRecord{ID: "a1", FlagID: "f1", ContentID: "vid-1", Action: ActionDeny, ActionBy: "admin"})
	_ = al.Append(AuditRecord{ID: "a2", FlagID: "f2", ContentID: "vid-1", Action: ActionApprove, ActionBy: "admin"})
	_ = al.Append(AuditRecord{ID: "a3", FlagID: "f3", ContentID: "vid-2", Action: ActionDeny, ActionBy: "admin"})

	byContent, _ := al.GetByContent("vid-1")
	if len(byContent) != 2 {
		t.Fatalf("expected 2 records for vid-1, got %d", len(byContent))
	}

	all, _ := al.GetAll()
	if len(all) != 3 {
		t.Fatalf("expected 3 total records, got %d", len(all))
	}
}

func TestDMCACounterNotice_RestoreAfter(t *testing.T) {
	notice := DMCACounterNotice{
		ReceivedAt:   time.Now(),
		RestoreAfter: time.Now().Add(DMCARestorePeriod),
	}
	if notice.RestoreAfter.Before(notice.ReceivedAt.Add(9 * 24 * time.Hour)) {
		t.Fatal("restore period should be at least 10 days")
	}
}

func TestFlagCategories(t *testing.T) {
	cats := []FlagCategory{CategoryCopyright, CategoryIllegal, CategoryAbuse}
	for _, c := range cats {
		if c == "" {
			t.Fatal("category should not be empty")
		}
	}
}
