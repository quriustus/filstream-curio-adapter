package test

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/quriustus/filstream-curio-adapter/internal/mock"
	"github.com/quriustus/filstream-curio-adapter/pkg/policy"
)

func TestFullRetrieval(t *testing.T) {
	b := mock.NewBackend()
	ctx := context.Background()

	rc, err := b.Get(ctx, "bafydeadbeef")
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()

	data, _ := io.ReadAll(rc)
	if string(data) != "hello filstream" {
		t.Fatalf("unexpected data: %q", data)
	}
}

func TestRangeRetrieval(t *testing.T) {
	b := mock.NewBackend()
	ctx := context.Background()

	// bafy5678chunk is 256KB of 'C'. Get first 1024 bytes: [0, 1024)
	rc, err := b.GetRange(ctx, "bafy5678chunk", 0, 1024)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()

	data, _ := io.ReadAll(rc)
	if len(data) != 1024 {
		t.Fatalf("expected 1024 bytes, got %d", len(data))
	}
}

func TestRangeInvalid(t *testing.T) {
	b := mock.NewBackend()
	ctx := context.Background()

	_, err := b.GetRange(ctx, "bafydeadbeef", 5, 3) // start >= end
	if err == nil {
		t.Fatal("expected error for invalid range")
	}
}

func TestHealthCheck(t *testing.T) {
	b := mock.NewBackend()
	ctx := context.Background()

	hs, err := b.CheckHealth(ctx, "node-us-east-1")
	if err != nil {
		t.Fatal(err)
	}
	if !hs.Healthy {
		t.Fatal("expected healthy node")
	}
	if hs.GeoLabel != "us-east" {
		t.Fatalf("expected geo us-east, got %s", hs.GeoLabel)
	}
}

func TestHealthCheckUnknown(t *testing.T) {
	b := mock.NewBackend()
	ctx := context.Background()

	hs, err := b.CheckHealth(ctx, "node-nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if hs.Healthy {
		t.Fatal("expected unhealthy for unknown node")
	}
}

func TestProofVerification(t *testing.T) {
	b := mock.NewBackend()
	ctx := context.Background()

	ok, err := b.VerifyProof(ctx, "bafy1234video", []byte("valid-proof-1234"))
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected valid proof")
	}

	ok, err = b.VerifyProof(ctx, "bafy1234video", []byte("wrong-proof"))
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected invalid proof")
	}
}

func TestProofTTL(t *testing.T) {
	b := mock.NewBackend()
	ttl := b.ProofTTL()
	if ttl != 24*time.Hour {
		t.Fatalf("expected 24h TTL, got %v", ttl)
	}
}

func TestScoringEngine(t *testing.T) {
	eng := policy.NewEngine(policy.DefaultConfig())

	// Record enough samples to pass grace period.
	for i := 0; i < 15; i++ {
		eng.RecordLatency("node-1", 20*time.Millisecond)
	}
	eng.SetGeoLabel("node-1", "us-east")

	score := eng.Score("node-1", "us-east")
	if score.SampleCount != 15 {
		t.Fatalf("expected 15 samples, got %d", score.SampleCount)
	}
	if score.Score <= 0 {
		t.Fatal("expected positive score")
	}
}

func TestScoringGracePeriod(t *testing.T) {
	eng := policy.NewEngine(policy.DefaultConfig())

	// Only 3 samples — should get grace score.
	for i := 0; i < 3; i++ {
		eng.RecordLatency("node-1", 100*time.Millisecond)
	}
	score := eng.Score("node-1", "")
	if score.Score != 0.5 {
		t.Fatalf("expected grace score 0.5, got %f", score.Score)
	}
}

func TestScoringProofPenalty(t *testing.T) {
	eng := policy.NewEngine(policy.DefaultConfig())

	for i := 0; i < 15; i++ {
		eng.RecordLatency("node-1", 20*time.Millisecond)
	}

	scoreGood := eng.Score("node-1", "")

	// Miss 3 proofs (grace is 2), should penalize.
	eng.RecordProofResult("node-1", false)
	eng.RecordProofResult("node-1", false)
	eng.RecordProofResult("node-1", false)

	scoreBad := eng.Score("node-1", "")
	if scoreBad.Score >= scoreGood.Score {
		t.Fatalf("expected penalty: good=%f bad=%f", scoreGood.Score, scoreBad.Score)
	}
}
