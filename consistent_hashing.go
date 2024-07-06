package consistent_hashing

import (
	"context"
	"errors"
	"fmt"
	"hash"
	"hash/fnv"
	"log"
	"sort"
	"sync"
)

// Custom errors
var (
	ErrNoHost       = errors.New("no host added")
	ErrHostNotFound = errors.New("host not found")
)

// Consistent Hashing config parameters
type Config struct {
	ReplicationFactor int                // no of virtual_nodes per host
	LoadFactor        float64            // max load factor before redistribution
	HashFunction      func() hash.Hash64 // for the time being lets keep the hash function simple
}

// Host is a physical node in the CH hashing ring
type Host struct {
	Name string // HostName or identifier
	Load int64  // current load on the host
}

// CH with bounded loads
type ConsistentHashing struct {
	config    Config
	hosts     sync.Map     // Map of hash value to host
	sortedSet []uint64     // sorted slice of hash values
	loadMap   sync.Map     // map of host to Host struct
	totalLoad int64        // total load across all hosts
	hostList  []string     // list of all hosts ['uat-server.something.com', 'be-server.something.com']
	mu        sync.RWMutex // Mutex for synchronizing access
}

// New CH instance
func NewWithConfig(cfg Config) (*ConsistentHashing, error) {
	if cfg.ReplicationFactor <= 0 {
		cfg.ReplicationFactor = 10
	}

	if cfg.LoadFactor <= 1 {
		cfg.LoadFactor = 1.25
	}

	if cfg.HashFunction == nil {
		cfg.HashFunction = fnv.New64a
	}

	return &ConsistentHashing{
		config:    cfg,
		sortedSet: make([]uint64, 0),
	}, nil
}

// adds a host to the hash ring
// Add adds a new host to the consistent hashing ring, including its virtual nodes,
// and updates the internal data structures accordingly. It returns an error if the operation fails.
func (c *ConsistentHashing) Add(ctx context.Context, host string) error {
	// Acquire a lock to ensure thread safety during the update.
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if the host already exists in the loadMap.
	if _, ok := c.loadMap.Load(host); ok {
		return nil // Host already exists, no further action needed.
	}

	// Add the new host with an initial load of 0.
	c.loadMap.Store(host, &Host{Name: host, Load: 0})
	c.hostList = append(c.hostList, host)

	// Add virtual nodes for the host based on the replication factor.
	for i := 0; i < c.config.ReplicationFactor; i++ {
		// Generate a hash value for the virtual node.
		h, err := c.Hash(fmt.Sprintf("%s%d", host, i))
		if err != nil {
			log.Fatal("key hashing failed", err) // Log fatal error and exit if hashing fails.
		}
		// Store the virtual node hash and map it to the host.
		c.hosts.Store(h, host)
		// Append the virtual node hash to the sorted set.
		c.sortedSet = append(c.sortedSet, h)
	}

	// Sort the hash values in the sorted set.
	// This allows efficient key lookups using binary search.
	sort.Slice(c.sortedSet, func(i, j int) bool { return c.sortedSet[i] < c.sortedSet[j] })

	// Return nil to indicate the host was added successfully.
	return nil
}

// Helper Functions

// hash generates a 64-bit hash value for a given key using the configured hash function.
// It returns the computed hash value and an error, if any occurred during the hashing process.
func (c *ConsistentHashing) Hash(key string) (uint64, error) {
	// Create a new hash object using the configured hash function.
	h := c.config.HashFunction()

	// Write the key to the hash object. If an error occurs, panic.
	if _, err := h.Write([]byte(key)); err != nil {
		panic(err)
	}

	// Compute and return the hash value as a 64-bit unsigned integer.
	return uint64(h.Sum64()), nil
}

// Hosts returns the list of current hosts
func (c *ConsistentHashing) Hosts() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]string(nil), c.hostList...)
}
