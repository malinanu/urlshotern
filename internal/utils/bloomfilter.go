package utils

import (
	"hash/fnv"
	"math"
	"sync"
)

// BloomFilter is a space-efficient probabilistic data structure for set membership testing
type BloomFilter struct {
	bits      []bool
	size      uint
	hashFuncs int
	mu        sync.RWMutex
}

// NewBloomFilter creates a new Bloom filter with optimal parameters
// expectedItems: expected number of items to be added
// falsePositiveRate: desired false positive probability (e.g., 0.01 for 1%)
func NewBloomFilter(expectedItems int, falsePositiveRate float64) *BloomFilter {
	// Calculate optimal bit array size
	size := optimalSize(expectedItems, falsePositiveRate)

	// Calculate optimal number of hash functions
	hashFuncs := optimalHashFuncs(size, expectedItems)

	return &BloomFilter{
		bits:      make([]bool, size),
		size:      size,
		hashFuncs: hashFuncs,
	}
}

// Add adds an item to the Bloom filter
func (bf *BloomFilter) Add(item string) {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	hashes := bf.getHashes(item)
	for _, hash := range hashes {
		bf.bits[hash%bf.size] = true
	}
}

// MightContain checks if an item might be in the set
// Returns true if the item might be present (with some false positive rate)
// Returns false if the item is definitely not present (no false negatives)
func (bf *BloomFilter) MightContain(item string) bool {
	bf.mu.RLock()
	defer bf.mu.RUnlock()

	hashes := bf.getHashes(item)
	for _, hash := range hashes {
		if !bf.bits[hash%bf.size] {
			return false
		}
	}
	return true
}

// getHashes generates multiple hash values for an item
func (bf *BloomFilter) getHashes(item string) []uint {
	hashes := make([]uint, bf.hashFuncs)

	// Use double hashing technique for generating multiple hashes
	h1 := hash1(item)
	h2 := hash2(item)

	for i := 0; i < bf.hashFuncs; i++ {
		hashes[i] = uint(h1 + uint64(i)*h2)
	}

	return hashes
}

// hash1 generates the first hash using FNV-1a
func hash1(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

// hash2 generates the second hash using a simple polynomial rolling hash
func hash2(s string) uint64 {
	var hash uint64 = 5381
	for _, c := range s {
		hash = ((hash << 5) + hash) + uint64(c)
	}
	return hash
}

// optimalSize calculates the optimal bit array size
func optimalSize(n int, p float64) uint {
	// m = -(n * ln(p)) / (ln(2)^2)
	m := -float64(n) * math.Log(p) / (math.Ln2 * math.Ln2)
	return uint(math.Ceil(m))
}

// optimalHashFuncs calculates the optimal number of hash functions
func optimalHashFuncs(m uint, n int) int {
	// k = (m/n) * ln(2)
	k := float64(m) / float64(n) * math.Ln2
	return int(math.Ceil(k))
}

// Clear resets the Bloom filter
func (bf *BloomFilter) Clear() {
	bf.mu.Lock()
	defer bf.mu.Unlock()

	for i := range bf.bits {
		bf.bits[i] = false
	}
}

// EstimateFalsePositiveRate estimates the current false positive rate
func (bf *BloomFilter) EstimateFalsePositiveRate(itemsAdded int) float64 {
	bf.mu.RLock()
	defer bf.mu.RUnlock()

	// p = (1 - e^(-k*n/m))^k
	// where k = number of hash functions, n = items added, m = bit array size
	exponent := -float64(bf.hashFuncs) * float64(itemsAdded) / float64(bf.size)
	return math.Pow(1-math.Exp(exponent), float64(bf.hashFuncs))
}

// URLBloomFilter is a singleton Bloom filter for URL existence checks
var (
	urlBloomFilter     *BloomFilter
	urlBloomFilterOnce sync.Once
)

// GetURLBloomFilter returns the singleton URL Bloom filter
func GetURLBloomFilter() *BloomFilter {
	urlBloomFilterOnce.Do(func() {
		// Expected 1 million URLs with 1% false positive rate
		urlBloomFilter = NewBloomFilter(1000000, 0.01)
	})
	return urlBloomFilter
}
