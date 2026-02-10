// Package policy implements the node scoring and selection framework.
package policy

import (
	"sync"
	"time"
)

// Config holds configurable weights for the scoring engine.
type Config struct {
	// LatencyWeight is the weight for P95 latency in the score (0-1).
	LatencyWeight float64

	// GeoBoost is the additive bonus for nodes matching the preferred geo label.
	GeoBoost float64

	// MinSamples is the minimum number of latency samples before scoring applies.
	// Nodes below this threshold get a grace period (neutral score).
	MinSamples int

	// ProofGraceMisses is how many consecutive missed proofs are tolerated
	// before a scoring penalty. Default: 2.
	ProofGraceMisses int

	// ProofTTL overrides the default proof TTL for scoring decisions.
	ProofTTL time.Duration

	// HalfOpenProbeInterval is how often to send a probe to a degraded node.
	HalfOpenProbeInterval time.Duration
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		LatencyWeight:         0.7,
		GeoBoost:              0.1,
		MinSamples:            10,
		ProofGraceMisses:      2,
		ProofTTL:              24 * time.Hour,
		HalfOpenProbeInterval: 5 * time.Minute,
	}
}

// NodeScore represents the computed score for a storage node.
type NodeScore struct {
	NodeID         string
	Score          float64
	P95Latency     time.Duration
	SampleCount    int
	MissedProofs   int
	GeoLabel       string
	HalfOpen       bool
	LastProofCheck time.Time
}

// Engine is the scoring and selection engine.
type Engine struct {
	mu     sync.RWMutex
	config Config
	nodes  map[string]*nodeState
}

type nodeState struct {
	latencies    []time.Duration // sliding window
	missedProofs int
	geoLabel     string
	lastProof    time.Time
	halfOpen     bool
}

// NewEngine creates a new scoring engine with the given config.
func NewEngine(cfg Config) *Engine {
	return &Engine{
		config: cfg,
		nodes:  make(map[string]*nodeState),
	}
}

// RecordLatency adds a latency sample for the given node.
func (e *Engine) RecordLatency(nodeID string, d time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	ns := e.getOrCreate(nodeID)
	ns.latencies = append(ns.latencies, d)

	// Keep sliding window at 100 samples max.
	if len(ns.latencies) > 100 {
		ns.latencies = ns.latencies[len(ns.latencies)-100:]
	}
}

// RecordProofResult records a proof verification result for the given node.
func (e *Engine) RecordProofResult(nodeID string, passed bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	ns := e.getOrCreate(nodeID)
	ns.lastProof = time.Now()
	if passed {
		ns.missedProofs = 0
		ns.halfOpen = false
	} else {
		ns.missedProofs++
		if ns.missedProofs > e.config.ProofGraceMisses {
			ns.halfOpen = true
		}
	}
}

// SetGeoLabel sets the geographic label for a node.
func (e *Engine) SetGeoLabel(nodeID, label string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.getOrCreate(nodeID).geoLabel = label
}

// Score computes the current score for a node given a preferred geo label.
func (e *Engine) Score(nodeID, preferredGeo string) NodeScore {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ns, ok := e.nodes[nodeID]
	if !ok {
		return NodeScore{NodeID: nodeID}
	}

	score := NodeScore{
		NodeID:         nodeID,
		SampleCount:    len(ns.latencies),
		MissedProofs:   ns.missedProofs,
		GeoLabel:       ns.geoLabel,
		HalfOpen:       ns.halfOpen,
		LastProofCheck: ns.lastProof,
	}

	// Grace period: not enough samples yet.
	if len(ns.latencies) < e.config.MinSamples {
		score.Score = 0.5 // neutral
		return score
	}

	score.P95Latency = p95(ns.latencies)

	// Base latency score: lower is better. Normalize to 0-1 (cap at 10s).
	latencyScore := 1.0 - float64(score.P95Latency)/float64(10*time.Second)
	if latencyScore < 0 {
		latencyScore = 0
	}

	score.Score = latencyScore * e.config.LatencyWeight

	// Geo boost.
	if ns.geoLabel == preferredGeo && preferredGeo != "" {
		score.Score += e.config.GeoBoost
	}

	// Proof penalty.
	if ns.missedProofs > e.config.ProofGraceMisses {
		score.Score *= 0.5
	}

	return score
}

// NeedsProofCheck returns true if the node's proof TTL has expired.
func (e *Engine) NeedsProofCheck(nodeID string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ns, ok := e.nodes[nodeID]
	if !ok {
		return true
	}
	return time.Since(ns.lastProof) > e.config.ProofTTL
}

func (e *Engine) getOrCreate(nodeID string) *nodeState {
	ns, ok := e.nodes[nodeID]
	if !ok {
		ns = &nodeState{}
		e.nodes[nodeID] = ns
	}
	return ns
}

// p95 computes the P95 latency from a slice of durations.
func p95(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	// Simple: sort a copy and pick the 95th percentile index.
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)
	// Insertion sort (small N).
	for i := 1; i < len(sorted); i++ {
		for j := i; j > 0 && sorted[j] < sorted[j-1]; j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}
	idx := int(float64(len(sorted)) * 0.95)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
