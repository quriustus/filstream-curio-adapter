package moderation

import (
	"fmt"
	"sync"
	"time"
)

// MockDenyList is an in-memory DenyList for testing.
type MockDenyList struct {
	mu      sync.RWMutex
	entries map[string]DenyEntry
}

func NewMockDenyList() *MockDenyList {
	return &MockDenyList{entries: make(map[string]DenyEntry)}
}

func (m *MockDenyList) Add(contentID, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[contentID] = DenyEntry{
		ContentID: contentID,
		Reason:    reason,
		DeniedAt:  time.Now(),
		DeniedBy:  "system",
	}
	return nil
}

func (m *MockDenyList) Remove(contentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.entries[contentID]; !ok {
		return fmt.Errorf("content %s not in denylist", contentID)
	}
	delete(m.entries, contentID)
	return nil
}

func (m *MockDenyList) IsDenied(contentID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.entries[contentID]
	return ok, nil
}

func (m *MockDenyList) List() ([]DenyEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]DenyEntry, 0, len(m.entries))
	for _, e := range m.entries {
		out = append(out, e)
	}
	return out, nil
}

// MockModerationQueue is an in-memory ModerationQueue for testing.
type MockModerationQueue struct {
	mu        sync.Mutex
	flags     map[string]ContentFlag
	escalated map[string]bool
	reviewed  map[string]ReviewAction
	denyList  DenyList
	auditLog  AuditLog
	escConfig EscalationConfig
	// track flags per content for auto-escalation
	contentFlags map[string][]time.Time
}

func NewMockModerationQueue(dl DenyList, al AuditLog, cfg EscalationConfig) *MockModerationQueue {
	return &MockModerationQueue{
		flags:        make(map[string]ContentFlag),
		escalated:    make(map[string]bool),
		reviewed:     make(map[string]ReviewAction),
		denyList:     dl,
		auditLog:     al,
		escConfig:    cfg,
		contentFlags: make(map[string][]time.Time),
	}
}

func (m *MockModerationQueue) Submit(flag ContentFlag) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if flag.ID == "" {
		flag.ID = fmt.Sprintf("flag-%d", len(m.flags)+1)
	}
	if flag.Timestamp.IsZero() {
		flag.Timestamp = time.Now()
	}
	m.flags[flag.ID] = flag

	// Auto-escalation check
	now := time.Now()
	cutoff := now.Add(-m.escConfig.Window)
	times := m.contentFlags[flag.ContentID]
	// prune old
	var recent []time.Time
	for _, t := range times {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	recent = append(recent, now)
	m.contentFlags[flag.ContentID] = recent

	if len(recent) >= m.escConfig.FlagThreshold {
		m.escalated[flag.ID] = true
	}
	return nil
}

func (m *MockModerationQueue) Review(flagID string, action ReviewAction, reviewedBy string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	flag, ok := m.flags[flagID]
	if !ok {
		return fmt.Errorf("flag %s not found", flagID)
	}
	m.reviewed[flagID] = action

	if action == ActionDeny && m.denyList != nil {
		_ = m.denyList.Add(flag.ContentID, string(flag.Category))
	}

	if m.auditLog != nil {
		_ = m.auditLog.Append(AuditRecord{
			ID:        fmt.Sprintf("audit-%s", flagID),
			FlagID:    flagID,
			ContentID: flag.ContentID,
			Action:    action,
			ActionBy:  reviewedBy,
			Reason:    string(flag.Category),
			Timestamp: time.Now(),
		})
	}
	return nil
}

func (m *MockModerationQueue) Escalate(flagID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.flags[flagID]; !ok {
		return fmt.Errorf("flag %s not found", flagID)
	}
	m.escalated[flagID] = true
	return nil
}

func (m *MockModerationQueue) IsEscalated(flagID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.escalated[flagID]
}

func (m *MockModerationQueue) GetPending() ([]ContentFlag, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []ContentFlag
	for id, f := range m.flags {
		if _, done := m.reviewed[id]; !done {
			out = append(out, f)
		}
	}
	return out, nil
}

// MockSyncBroadcaster records broadcast calls for testing.
type MockSyncBroadcaster struct {
	mu          sync.Mutex
	Broadcasts  [][]string
	SyncedPeers []string
}

func NewMockSyncBroadcaster() *MockSyncBroadcaster {
	return &MockSyncBroadcaster{}
}

func (m *MockSyncBroadcaster) BroadcastDenylist(seederIDs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Broadcasts = append(m.Broadcasts, seederIDs)
	return nil
}

func (m *MockSyncBroadcaster) SyncSeeder(seederID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SyncedPeers = append(m.SyncedPeers, seederID)
	return nil
}

// MockAuditLog is an in-memory AuditLog for testing.
type MockAuditLog struct {
	mu      sync.Mutex
	records []AuditRecord
}

func NewMockAuditLog() *MockAuditLog {
	return &MockAuditLog{}
}

func (m *MockAuditLog) Append(record AuditRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records = append(m.records, record)
	return nil
}

func (m *MockAuditLog) GetByContent(contentID string) ([]AuditRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []AuditRecord
	for _, r := range m.records {
		if r.ContentID == contentID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (m *MockAuditLog) GetByFlag(flagID string) ([]AuditRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []AuditRecord
	for _, r := range m.records {
		if r.FlagID == flagID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (m *MockAuditLog) GetAll() ([]AuditRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]AuditRecord, len(m.records))
	copy(out, m.records)
	return out, nil
}
