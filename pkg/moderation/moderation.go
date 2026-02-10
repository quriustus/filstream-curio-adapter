// Package moderation provides content moderation for FilStream's decentralized
// video network. Since content on Filecoin is immutable, moderation operates at
// the index/distribution layer — denying discovery and delivery rather than
// deleting underlying data.
package moderation

import (
	"time"
)

// FlagCategory classifies the type of content violation.
type FlagCategory string

const (
	CategoryCopyright FlagCategory = "copyright"
	CategoryIllegal   FlagCategory = "illegal"
	CategoryAbuse     FlagCategory = "abuse"
)

// ReviewAction represents the outcome of reviewing a content flag.
type ReviewAction string

const (
	ActionApprove ReviewAction = "approve" // content is fine, dismiss flag
	ActionDeny    ReviewAction = "deny"    // add to denylist
	ActionDismiss ReviewAction = "dismiss" // flag invalid, no action
)

// ContentFlag represents a report against a piece of content.
type ContentFlag struct {
	ID        string       `json:"id"`
	ContentID string       `json:"content_id"`
	FlaggedBy string       `json:"flagged_by"`
	Category  FlagCategory `json:"category"`
	Evidence  string       `json:"evidence"`
	Timestamp time.Time    `json:"timestamp"`
}

// DenyEntry is a record in the denylist.
type DenyEntry struct {
	ContentID string    `json:"content_id"`
	Reason    string    `json:"reason"`
	DeniedAt  time.Time `json:"denied_at"`
	DeniedBy  string    `json:"denied_by"`
}

// AuditRecord captures every moderation action for accountability.
type AuditRecord struct {
	ID         string       `json:"id"`
	FlagID     string       `json:"flag_id"`
	ContentID  string       `json:"content_id"`
	Action     ReviewAction `json:"action"`
	ActionBy   string       `json:"action_by"`
	Reason     string       `json:"reason"`
	Timestamp  time.Time    `json:"timestamp"`
}

// DMCANotice represents a DMCA takedown request per 17 U.S.C. § 512.
type DMCANotice struct {
	ID            string    `json:"id"`
	ContentID     string    `json:"content_id"`
	ClaimantName  string    `json:"claimant_name"`
	ClaimantEmail string    `json:"claimant_email"`
	WorkDesc      string    `json:"work_description"`    // description of copyrighted work
	InfringingURL string    `json:"infringing_url"`
	Statement     string    `json:"statement"`           // good-faith statement
	Signature     string    `json:"signature"`
	ReceivedAt    time.Time `json:"received_at"`
}

// DMCACounterNotice represents a counter-notification from the content uploader.
// Per DMCA, the service provider must wait 10 business days after receiving a
// counter-notice before restoring content (unless claimant files court action).
type DMCACounterNotice struct {
	ID             string    `json:"id"`
	NoticeID       string    `json:"notice_id"`        // references DMCANotice.ID
	ContentID      string    `json:"content_id"`
	ResponderName  string    `json:"responder_name"`
	ResponderEmail string    `json:"responder_email"`
	Statement      string    `json:"statement"`
	Signature      string    `json:"signature"`
	ReceivedAt     time.Time `json:"received_at"`
	// RestoreAfter is 10 business days from ReceivedAt (set by the system).
	RestoreAfter   time.Time `json:"restore_after"`
}

// EscalationConfig controls auto-escalation thresholds.
type EscalationConfig struct {
	// FlagThreshold: number of flags on a single content ID within Window
	// that triggers automatic escalation.
	FlagThreshold int           `json:"flag_threshold"`
	Window        time.Duration `json:"window"`
}

// DefaultEscalationConfig returns sensible defaults: 3 flags in 1 hour.
func DefaultEscalationConfig() EscalationConfig {
	return EscalationConfig{
		FlagThreshold: 3,
		Window:        time.Hour,
	}
}

// DMCARestorePeriod is the 10-business-day waiting period for counter-notices.
const DMCARestorePeriod = 10 * 24 * time.Hour // simplified to 10 calendar days

// DenyList manages a set of denied content IDs. Implementations must be
// safe for concurrent use.
type DenyList interface {
	Add(contentID, reason string) error
	Remove(contentID string) error
	IsDenied(contentID string) (bool, error)
	List() ([]DenyEntry, error)
}

// ModerationQueue handles the lifecycle of content flags.
type ModerationQueue interface {
	Submit(flag ContentFlag) error
	Review(flagID string, action ReviewAction, reviewedBy string) error
	Escalate(flagID string) error
	GetPending() ([]ContentFlag, error)
}

// SyncBroadcaster propagates denylist updates to seeder nodes.
type SyncBroadcaster interface {
	BroadcastDenylist(seederIDs []string) error
	BroadcastBloom(bloom *DenylistBloom) error
	SyncSeeder(seederID string) error
}

// AuditLog provides read access to the moderation audit trail.
type AuditLog interface {
	Append(record AuditRecord) error
	GetByContent(contentID string) ([]AuditRecord, error)
	GetByFlag(flagID string) ([]AuditRecord, error)
	GetAll() ([]AuditRecord, error)
}
