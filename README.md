# filstream-curio-adapter

Go adapter connecting [FilStream](https://filstream.io) (decentralized video streaming) to [Curio](https://curiostorage.org) (Filecoin storage pipeline) for retrieval, health checks, and proof verification.

## Architecture

```
┌──────────────┐     ┌─────────────────────┐     ┌──────────────┐
│   FilStream   │────▶│  filstream-curio-   │────▶│    Curio     │
│   (video)     │◀────│     adapter         │◀────│  (storage)   │
└──────────────┘     └─────────────────────┘     └──────────────┘
                       │ pkg/adapter/  │
                       │  RetrieverAPI │
                       │  HealthChecker│
                       │  ProofVerifier│
                       │               │
                       │ pkg/policy/   │
                       │  Scoring      │
                       └───────────────┘
```

### Core Interfaces (`pkg/adapter/`)

| Interface | Methods | Purpose |
|-----------|---------|---------|
| **RetrieverAPI** | `Get(ctx, cid)`, `GetRange(ctx, cid, start, end)` | Retrieve content from Curio by CID |
| **HealthChecker** | `CheckHealth(ctx, nodeID)` | Monitor Curio storage node health |
| **ProofVerifier** | `VerifyProof(ctx, cid, proof)`, `ProofTTL()` | Verify storage proofs |

### Policy Engine (`pkg/policy/`)

Node scoring framework with configurable weights:

- **Sliding P95 latency window** — last 100 samples, insertion-sorted
- **Min-samples grace period** — nodes with <10 samples get neutral score (0.5)
- **Geo label boost** — additive bonus for geo-matching nodes
- **Half-open proof probes** — degraded nodes get periodic probe attempts
- **Configurable weights** — `LatencyWeight`, `GeoBoost`, `ProofGraceMisses`, etc.

### Mock Backend (`internal/mock/`)

In-memory implementation of all interfaces with pre-seeded fake CIDs for testing.

## Pinned Endpoint Semantics

### Range Reads

- **Start and End are both required** for range reads
- **Semantics: `[Start, End)` — End is EXCLUSIVE (half-open)**
- Full object of size N: `Start=0, End=N`
- Example: 1MB video, first 256KB → `GetRange(ctx, cid, 0, 262144)`

### Proof TTL

- **Default: 24 hours**, configurable via policy engine `Config.ProofTTL`
- **2 missed proofs grace period** before scoring penalty applies
- After grace exceeded: node score halved, marked half-open
- **Re-verify triggered on next health check** after TTL expiry
- `Engine.NeedsProofCheck(nodeID)` returns true when TTL has elapsed

### Examples

```go
// Full retrieval
rc, err := retriever.Get(ctx, "bafy1234video")

// Range read: bytes [1024, 2048)
rc, err := retriever.GetRange(ctx, "bafy1234video", 1024, 2048)

// Health check (triggers re-verify if proof TTL expired)
status, err := checker.CheckHealth(ctx, "node-us-east-1")

// Proof verification
valid, err := verifier.VerifyProof(ctx, "bafy1234video", proofBytes)
ttl := verifier.ProofTTL() // 24h default
```

## Division of Labor

| Person | Scope |
|--------|-------|
| **Rick** | Interfaces, skeleton, policy engine, mock backend, tests |
| **Capri** | Curio implementation (real storage backend, proof logic) |

## Branch Conventions

- **`main`** — stable, always passes tests
- **Feature branches** — `feature/<name>` or `<author>/<description>`
- **Pull requests required** for merging to main

## Development

```bash
# Run tests
go test ./...

# Run integration tests only
go test ./test/
```

## License

TBD
