package moderation

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"math"
	"sync"
)

// DenylistBloom is a compact Bloom filter for seeder-side denylist checking.
// Seeders call MayContain before serving every segment — this must be fast.
// The filter is designed to be small enough (<1KB for 10K items) for frequent
// network sync via BroadcastBloom.
type DenylistBloom struct {
	mu       sync.RWMutex
	bits     []byte
	numHash  uint32 // number of hash functions (k)
	numBits  uint32 // total bits (m)
	count    uint32 // items added
}

// NewDenylistBloom creates a Bloom filter sized for the given capacity and
// false-positive rate. For 10,000 items at 1% FP rate, this produces a
// ~12KB filter. For <1KB targeting, use estimatedItems=10000 with fpRate=0.5
// or reduce estimatedItems. A reasonable default: 10000 items at 10% ≈ 6KB.
//
// For truly compact filters (~1KB), use estimatedItems=1000, fpRate=0.01.
// The filter still works with more items — the FP rate just increases.
func NewDenylistBloom(estimatedItems uint32, fpRate float64) *DenylistBloom {
	if estimatedItems == 0 {
		estimatedItems = 1000
	}
	if fpRate <= 0 || fpRate >= 1 {
		fpRate = 0.01
	}

	// m = -n*ln(p) / (ln2)^2
	m := uint32(math.Ceil(-float64(estimatedItems) * math.Log(fpRate) / (math.Ln2 * math.Ln2)))
	// Round up to nearest byte
	m = ((m + 7) / 8) * 8

	// k = (m/n) * ln2
	k := uint32(math.Ceil(float64(m) / float64(estimatedItems) * math.Ln2))
	if k == 0 {
		k = 1
	}

	return &DenylistBloom{
		bits:    make([]byte, m/8),
		numHash: k,
		numBits: m,
	}
}

// Add inserts a content hash into the Bloom filter.
func (b *DenylistBloom) Add(contentHash string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, idx := range b.hashIndices(contentHash) {
		b.bits[idx/8] |= 1 << (idx % 8)
	}
	b.count++
}

// MayContain returns true if the content hash might be in the denylist.
// False means definitely not denied. True means probably denied (check the
// authoritative denylist to confirm if needed).
func (b *DenylistBloom) MayContain(contentHash string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, idx := range b.hashIndices(contentHash) {
		if b.bits[idx/8]&(1<<(idx%8)) == 0 {
			return false
		}
	}
	return true
}

// Count returns the number of items added.
func (b *DenylistBloom) Count() uint32 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.count
}

// Serialize encodes the Bloom filter to bytes for network transmission.
// Format: [numBits:4][numHash:4][count:4][bits...]
func (b *DenylistBloom) Serialize() []byte {
	b.mu.RLock()
	defer b.mu.RUnlock()

	buf := make([]byte, 12+len(b.bits))
	binary.LittleEndian.PutUint32(buf[0:4], b.numBits)
	binary.LittleEndian.PutUint32(buf[4:8], b.numHash)
	binary.LittleEndian.PutUint32(buf[8:12], b.count)
	copy(buf[12:], b.bits)
	return buf
}

// Deserialize reconstructs a Bloom filter from bytes produced by Serialize.
func Deserialize(data []byte) (*DenylistBloom, error) {
	if len(data) < 12 {
		return nil, ErrInvalidBloomData
	}

	numBits := binary.LittleEndian.Uint32(data[0:4])
	numHash := binary.LittleEndian.Uint32(data[4:8])
	count := binary.LittleEndian.Uint32(data[8:12])

	expectedLen := 12 + int(numBits/8)
	if len(data) != expectedLen {
		return nil, ErrInvalidBloomData
	}

	bits := make([]byte, numBits/8)
	copy(bits, data[12:])

	return &DenylistBloom{
		bits:    bits,
		numHash: numHash,
		numBits: numBits,
		count:   count,
	}, nil
}

// Merge combines another Bloom filter into this one (bitwise OR).
// Both filters must have the same dimensions. This is useful for combining
// denylist updates from multiple moderation sources.
func (b *DenylistBloom) Merge(other *DenylistBloom) error {
	if other == nil {
		return nil
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	other.mu.RLock()
	defer other.mu.RUnlock()

	if b.numBits != other.numBits || b.numHash != other.numHash {
		return ErrBloomDimensionMismatch
	}

	for i := range b.bits {
		b.bits[i] |= other.bits[i]
	}
	// Count is approximate after merge — we can't know exact unique count
	b.count += other.count
	return nil
}

// SizeBytes returns the serialized size in bytes.
func (b *DenylistBloom) SizeBytes() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return 12 + len(b.bits)
}

// hashIndices computes k bit positions for the given key using double hashing.
func (b *DenylistBloom) hashIndices(key string) []uint32 {
	var h1, h2 hash.Hash64
	h1 = fnv.New64()
	h2 = fnv.New64a()

	h1.Write([]byte(key))
	h2.Write([]byte(key))

	a := h1.Sum64()
	bb := h2.Sum64()

	indices := make([]uint32, b.numHash)
	for i := uint32(0); i < b.numHash; i++ {
		indices[i] = uint32((a + uint64(i)*bb) % uint64(b.numBits))
	}
	return indices
}

// Sentinel errors for Bloom filter operations.
var (
	ErrInvalidBloomData       = &bloomError{"invalid bloom filter data"}
	ErrBloomDimensionMismatch = &bloomError{"bloom filter dimension mismatch: numBits and numHash must match"}
)

type bloomError struct {
	msg string
}

func (e *bloomError) Error() string { return e.msg }
