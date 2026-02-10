package adapter

import (
	"context"
	"time"
)

// HealthStatus represents the health of a storage node.
type HealthStatus struct {
	NodeID    string
	Healthy   bool
	Latency   time.Duration
	GeoLabel  string
	CheckedAt time.Time
	Message   string
}

// HealthChecker checks the health of Curio storage nodes.
type HealthChecker interface {
	// CheckHealth returns the current health status of the given node.
	// If the node's proof TTL has expired, re-verification is triggered.
	CheckHealth(ctx context.Context, nodeID string) (HealthStatus, error)
}
