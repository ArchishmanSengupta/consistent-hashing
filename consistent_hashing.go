package consistent_hashing

import (
	"context"
	"errors"
	"hash"
	"hash/fnv"
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
func (c *ConsistentHashing) Add(ctx context.Context, host string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// check if the host already exists
	if _, ok := c.loadMap.Load(host); ok {
		return nil
	}

	// add host with 0 load
	c.loadMap.Store(host, &Host{Name: host, Load: 0})
	c.hostList = append(c.hostList, host)

	return nil
}

// Hosts returns the list of current hosts
func (c *ConsistentHashing) Hosts() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]string(nil), c.hostList...)
}
