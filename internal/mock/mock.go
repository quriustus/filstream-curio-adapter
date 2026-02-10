// Package mock provides in-memory mock implementations of all adapter interfaces.
package mock

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/quriustus/filstream-curio-adapter/pkg/adapter"
)

// Backend is an in-memory mock backend implementing all adapter interfaces.
type Backend struct {
	mu       sync.RWMutex
	objects  map[string][]byte // cid -> data
	proofs   map[string][]byte // cid -> valid proof
	nodes    map[string]adapter.HealthStatus
	proofTTL time.Duration
}

// NewBackend creates a new mock backend with some fake data.
func NewBackend() *Backend {
	b := &Backend{
		objects:  make(map[string][]byte),
		proofs:   make(map[string][]byte),
		nodes:    make(map[string]adapter.HealthStatus),
		proofTTL: adapter.DefaultProofTTL,
	}

	// Seed with fake CIDs.
	b.objects["bafy1234video"] = bytes.Repeat([]byte("V"), 1024*1024)   // 1MB fake video
	b.objects["bafy5678chunk"] = bytes.Repeat([]byte("C"), 256*1024)    // 256KB chunk
	b.objects["bafydeadbeef"] = []byte("hello filstream")

	b.proofs["bafy1234video"] = []byte("valid-proof-1234")
	b.proofs["bafy5678chunk"] = []byte("valid-proof-5678")

	b.nodes["node-us-east-1"] = adapter.HealthStatus{
		NodeID: "node-us-east-1", Healthy: true, Latency: 15 * time.Millisecond,
		GeoLabel: "us-east", CheckedAt: time.Now(),
	}
	b.nodes["node-eu-west-1"] = adapter.HealthStatus{
		NodeID: "node-eu-west-1", Healthy: true, Latency: 45 * time.Millisecond,
		GeoLabel: "eu-west", CheckedAt: time.Now(),
	}

	return b
}

// AddObject adds a CID with data to the mock store.
func (b *Backend) AddObject(cid string, data []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.objects[cid] = data
}

// AddProof sets the valid proof for a CID.
func (b *Backend) AddProof(cid string, proof []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.proofs[cid] = proof
}

// --- RetrieverAPI ---

func (b *Backend) Get(ctx context.Context, cid string) (io.ReadCloser, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	data, ok := b.objects[cid]
	if !ok {
		return nil, fmt.Errorf("cid not found: %s", cid)
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (b *Backend) GetRange(ctx context.Context, cid string, start, end uint64) (io.ReadCloser, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	data, ok := b.objects[cid]
	if !ok {
		return nil, fmt.Errorf("cid not found: %s", cid)
	}
	if start >= uint64(len(data)) || end > uint64(len(data)) || start >= end {
		return nil, fmt.Errorf("invalid range [%d, %d) for object of size %d", start, end, len(data))
	}
	return io.NopCloser(bytes.NewReader(data[start:end])), nil
}

// --- HealthChecker ---

func (b *Backend) CheckHealth(ctx context.Context, nodeID string) (adapter.HealthStatus, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	hs, ok := b.nodes[nodeID]
	if !ok {
		return adapter.HealthStatus{NodeID: nodeID, Healthy: false, Message: "unknown node"}, nil
	}
	hs.CheckedAt = time.Now()
	return hs, nil
}

// --- ProofVerifier ---

func (b *Backend) VerifyProof(ctx context.Context, cid string, proof []byte) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	validProof, ok := b.proofs[cid]
	if !ok {
		return false, fmt.Errorf("no proof registered for cid: %s", cid)
	}
	return bytes.Equal(validProof, proof), nil
}

func (b *Backend) ProofTTL() time.Duration {
	return b.proofTTL
}

// Compile-time interface checks.
var (
	_ adapter.RetrieverAPI  = (*Backend)(nil)
	_ adapter.HealthChecker = (*Backend)(nil)
	_ adapter.ProofVerifier = (*Backend)(nil)
)
