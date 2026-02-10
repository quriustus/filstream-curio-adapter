package adapter

import (
	"context"
	"time"
)

// DefaultProofTTL is the default proof time-to-live (24 hours).
const DefaultProofTTL = 24 * time.Hour

// ProofVerifier verifies storage proofs for content.
type ProofVerifier interface {
	// VerifyProof checks whether the given proof is valid for the CID.
	VerifyProof(ctx context.Context, cid string, proof []byte) (bool, error)

	// ProofTTL returns how long a verified proof remains valid.
	// Default: 24 hours. Configurable via the policy engine.
	ProofTTL() time.Duration
}
