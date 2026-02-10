package moderation

import (
	"fmt"
	"testing"
)

func TestNewDenylistBloom(t *testing.T) {
	b := NewDenylistBloom(1000, 0.01)
	if b == nil {
		t.Fatal("expected non-nil bloom filter")
	}
	if b.numBits == 0 {
		t.Fatal("expected non-zero numBits")
	}
	if b.numHash == 0 {
		t.Fatal("expected non-zero numHash")
	}
}

func TestAddAndMayContain(t *testing.T) {
	b := NewDenylistBloom(1000, 0.01)

	hashes := []string{
		"QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG",
		"bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3okuber2ce",
		"bafkreihdwdcefirg2gfaiyu7fvh4o2z5bkdntvaosq3",
	}

	for _, h := range hashes {
		b.Add(h)
	}

	if b.Count() != 3 {
		t.Fatalf("expected count=3, got %d", b.Count())
	}

	for _, h := range hashes {
		if !b.MayContain(h) {
			t.Errorf("expected MayContain(%s) = true", h)
		}
	}

	// Something never added should (almost certainly) return false
	if b.MayContain("definitely-not-in-filter-xyz123") {
		t.Log("false positive on single check — possible but unlikely")
	}
}

func TestFalsePositiveRate(t *testing.T) {
	n := uint32(1000)
	fpRate := 0.05
	b := NewDenylistBloom(n, fpRate)

	// Add n items
	for i := uint32(0); i < n; i++ {
		b.Add(fmt.Sprintf("content-%d", i))
	}

	// Test false positives with items NOT in the set
	falsePositives := 0
	tests := 10000
	for i := 0; i < tests; i++ {
		if b.MayContain(fmt.Sprintf("other-%d", i)) {
			falsePositives++
		}
	}

	observedRate := float64(falsePositives) / float64(tests)
	// Allow 2x the target FP rate as margin
	maxAcceptable := fpRate * 2
	if observedRate > maxAcceptable {
		t.Errorf("false positive rate too high: observed %.4f, max acceptable %.4f", observedRate, maxAcceptable)
	}
	t.Logf("False positive rate: %.4f (target: %.4f)", observedRate, fpRate)
}

func TestSerializeDeserialize(t *testing.T) {
	b := NewDenylistBloom(1000, 0.01)
	b.Add("hash-1")
	b.Add("hash-2")
	b.Add("hash-3")

	data := b.Serialize()
	b2, err := Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Verify same results
	for _, h := range []string{"hash-1", "hash-2", "hash-3"} {
		if !b2.MayContain(h) {
			t.Errorf("deserialized filter missing %s", h)
		}
	}
	if b2.Count() != 3 {
		t.Errorf("expected count=3, got %d", b2.Count())
	}
	if b2.numBits != b.numBits || b2.numHash != b.numHash {
		t.Error("dimensions mismatch after deserialization")
	}
}

func TestDeserializeInvalid(t *testing.T) {
	_, err := Deserialize([]byte{1, 2, 3})
	if err != ErrInvalidBloomData {
		t.Errorf("expected ErrInvalidBloomData, got %v", err)
	}
}

func TestMerge(t *testing.T) {
	b1 := NewDenylistBloom(1000, 0.01)
	b2 := NewDenylistBloom(1000, 0.01)

	b1.Add("hash-A")
	b1.Add("hash-B")
	b2.Add("hash-C")
	b2.Add("hash-D")

	if err := b1.Merge(b2); err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	for _, h := range []string{"hash-A", "hash-B", "hash-C", "hash-D"} {
		if !b1.MayContain(h) {
			t.Errorf("merged filter missing %s", h)
		}
	}
}

func TestMergeDimensionMismatch(t *testing.T) {
	b1 := NewDenylistBloom(1000, 0.01)
	b2 := NewDenylistBloom(5000, 0.01)

	err := b1.Merge(b2)
	if err != ErrBloomDimensionMismatch {
		t.Errorf("expected ErrBloomDimensionMismatch, got %v", err)
	}
}

func TestMergeNil(t *testing.T) {
	b := NewDenylistBloom(1000, 0.01)
	if err := b.Merge(nil); err != nil {
		t.Errorf("Merge(nil) should not error, got %v", err)
	}
}

func TestSizeCompact(t *testing.T) {
	// Verify filter for 1000 items is reasonably compact
	b := NewDenylistBloom(1000, 0.01)
	size := b.SizeBytes()
	t.Logf("Bloom filter size for 1000 items @ 1%% FP: %d bytes", size)
	if size > 2048 {
		t.Errorf("filter too large: %d bytes (want <2KB)", size)
	}
}

func BenchmarkMayContain(b *testing.B) {
	bloom := NewDenylistBloom(10000, 0.01)
	for i := 0; i < 10000; i++ {
		bloom.Add(fmt.Sprintf("hash-%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bloom.MayContain("hash-5000")
	}
}
