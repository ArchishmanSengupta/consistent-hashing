package main

import (
	"errors"
	"hash"
	"sync"
)

// Custom errors
var (
	ErrNoHost       = errors.New("no host added")
	ErrHostNotFound = errors.New("host not found")
)

// Consistent Hashing config parameters
type Config struct {
	ReplicationFact int                // no of virtual_nodes per host
	LoadFactor      float64            // max load factor before redistribution
	HashFunction    func() hash.Hash64 // for the time being lets keep the hash function simple
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
