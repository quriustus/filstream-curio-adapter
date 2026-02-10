// Package adapter defines the core interfaces for the FilStream-Curio adapter.
package adapter

import (
	"context"
	"io"
)

// RetrieverAPI retrieves content from Curio storage by CID.
//
// Range semantics: [Start, End) — End is EXCLUSIVE (half-open).
// Full object of size N: Start=0, End=N.
// Start and End are both required for range reads.
type RetrieverAPI interface {
	// Get retrieves the full content for the given CID.
	Get(ctx context.Context, cid string) (io.ReadCloser, error)

	// GetRange retrieves a byte range [start, end) for the given CID.
	// Both start and end are required. End is exclusive.
	GetRange(ctx context.Context, cid string, start, end uint64) (io.ReadCloser, error)
}
