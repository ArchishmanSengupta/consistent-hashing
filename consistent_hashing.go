package consistent_hashing

import (
	"context"
	"errors"
	"fmt"
	"hash"
	"hash/fnv"
	"log"
	"math"
	"sort"
	"sync"
	"sync/atomic"
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

// Get retrieves the host that should handle the given key in the consistent hashing ring.
// It returns the host name and nil error if successful. If no hosts are added, it returns ErrNoHost.
// If there's an error generating the hash value or searching for it, it returns an appropriate error.
// If the host associated with the hash value is not found, it returns ErrHostNotFound.
func (c *ConsistentHashing) Get(ctx context.Context, key string) (string, error) {
	// Acquire a read lock to ensure thread safety during read operations.
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return error if no hosts are added
	if len(c.hostList) == 0 {
		return "", ErrNoHost
	}

	// Generate hash value for the given key using the configured hash function.
	h, err := c.Hash(key)
	if err != nil {
		return "", err
	}

	// Find the closest index in the sorted set for the generated hash value.
	index, err := c.Search(h)
	if err != nil {
		return "", err
	}

	// Retrieve the host associated with the hash value from the hosts map.
	if host, ok := c.hosts.Load(c.sortedSet[index]); ok {
		return host.(string), nil
	}

	// Return an error if the host associated with the hash value is not found.
	return "", ErrHostNotFound
}

// GetLeast retrieves the host that should handle the given key in the consistent hashing ring
// with the least current load. It returns the host name and nil error if successful.
// If no hosts are added, it returns ErrNoHost. If there's an error generating the hash value
// or searching for it, it returns an appropriate error. If no host with acceptable load is found,
// it falls back to returning the initially found host. If no suitable host is found at all,
// it returns ErrHostNotFound.
// Bounded Loads: Research Paper: https://research.googleblog.com/2017/04/consistent-hashing-with-bounded-loads.html
func (c *ConsistentHashing) GetLeast(ctx context.Context, key string) (string, error) {
	// Acquire a read lock to ensure thread safety during read operations.
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return error if no hosts are added
	if len(c.hostList) == 0 {
		return "", ErrNoHost
	}

	// Generate hash value for the given key using the configured hash function.
	h, err := c.Hash(key)
	if err != nil {
		return "", err
	}

	// Find the closest index in the sorted set for the generated hash value.
	index, err := c.Search(h)
	if err != nil {
		return "", err
	}

	// Initialize variables to track the host with the least load.
	var leastLoadedHost string
	var minLoad int64 = math.MaxInt64

	// Iterate through the sorted set to find the host with the least load.
	for i := 0; i < len(c.sortedSet); i++ {
		nextIndex := (index + i) % len(c.sortedSet)
		if host, ok := c.hosts.Load(c.sortedSet[nextIndex]); ok {
			// Check if the host's load is acceptable.
			if c.LoadOk(host.(string)) {
				// Retrieve the load for the host.
				if h, ok := c.loadMap.Load(host.(string)); ok {
					load := h.(*Host).Load
					// Update the least loaded host if found.
					if load < minLoad {
						minLoad = load
						leastLoadedHost = host.(string)
					}
				}
			}
		}
	}

	// If no suitable host with acceptable load is found, return the initially found host.
	if leastLoadedHost == "" {
		if host, ok := c.hosts.Load(c.sortedSet[index]); ok {
			return host.(string), nil
		}
	}

	// Return an error if no suitable host is found.
	if leastLoadedHost == "" {
		return "", ErrHostNotFound
	}

	return leastLoadedHost, nil
}

// IncreaseLoad increments the load for a specific host.
func (c *ConsistentHashing) IncreaseLoad(ctx context.Context, host string) error {
	// Check if the host exists in the loadMap.
	if h, ok := c.loadMap.Load(host); ok {
		// Retrieve the host data from the loaded value.
		hostData := h.(*Host)

		// Atomically increment the load for the host by 1.
		atomic.AddInt64(&hostData.Load, 1)

		// Atomically increment the total load across all hosts by 1.
		atomic.AddInt64(&c.totalLoad, 1)

		// Return nil to indicate successful load increment.
		return nil
	}

	// Return an error indicating the host was not found.
	return ErrHostNotFound
}

// DecreaseLoad decreases the Load for a specific host.
func (c *ConsistentHashing) DecreaseLoad(ctx context.Context, host string) error {
	// Check if the host exists in the loadMap.
	if h, ok := c.loadMap.Load(host); ok {
		// Retrieve the host data from the loaded value.
		hostData := h.(*Host)

		// Atomically decrement the Load for the host by 1.
		atomic.AddInt64(&hostData.Load, -1)

		// Atomically decrement the total load across all hosts by 1.
		atomic.AddInt64(&c.totalLoad, -1)

		// Return nil to indicate successful load decrement.
		return nil
	}

	// Return an error indicating the host was not found.
	return ErrHostNotFound
}

// UpdateLoad updates the load for a specific host
func (c *ConsistentHashing) UpdateLoad(ctx context.Context, host string, load int64) error {
	// Check if the host exists in the load map
	if h, ok := c.loadMap.Load(host); ok {
		// Type assert the retrieved value to *Host
		hostData := h.(*Host)

		// Update the total load atomically
		atomic.AddInt64(&c.totalLoad, -hostData.Load+load)

		// Store the new load value for the host atomically
		atomic.StoreInt64(&hostData.Load, load)

		// Successfully updated the load, return nil error
		return nil
	}

	// If the host is not found, return an error
	return ErrHostNotFound
}

// Remove removes a host from the hash ring
func (c *ConsistentHashing) Remove(ctx context.Context, host string) error {
	// Acquire the mutex lock to ensure thread-safety
	c.mu.Lock()
	// Ensure the mutex is unlocked at the end of the function
	defer c.mu.Unlock()

	// Check if the host exists in the load map
	if _, ok := c.loadMap.Load(host); !ok {
		// If the host is not found, return an error
		return ErrHostNotFound
	}

	// Remove the virtual nodes associated with the host
	for i := 0; i < c.config.ReplicationFactor; i++ {
		// Generate a hash for the virtual node
		h, err := c.Hash(fmt.Sprintf("%s%d", host, i))
		if err != nil {
			// Log an error and exit if hashing fails
			log.Fatal("key hashing failed", err)
		}
		// Delete the virtual node from the hosts map
		c.hosts.Delete(h)
		// Remove the virtual node from the sorted set
		c.removeFromSortedSet(h)
	}
	// Delete the host from the load map
	c.loadMap.Delete(host)

	// Remove the host from the host list
	for i, h := range c.hostList {
		if h == host {
			// Remove the host from the list by creating a new slice without the host
			c.hostList = append(c.hostList[:i], c.hostList[i+1:]...)
			break
		}
	}
	// Return nil indicating successful removal
	return nil
}

// --------------------------------- Helper Functions ---------------------------------

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

// Search finds the closest index in the sorted set where the given hash key should be placed.
// It uses binary search to efficiently locate the index.
// For example, if c.sortedSet = [10, 20, 30, 40, 50] and key = 25,
// sort.Search determines that key should be inserted after 20 and before 30, returning index 2.
// The modulo operation (index % len(c.sortedSet)) ensures correct placement within the ring structure
func (c *ConsistentHashing) Search(key uint64) (int, error) {
	// Perform a binary search on the sorted set to find the index where key should be inserted.
	index := sort.Search(len(c.sortedSet), func(i int) bool {
		return c.sortedSet[i] >= key
	})

	// Wrap around the index using modulo operation to ensure it stays within bounds.
	// This is necessary for consistent hashing to handle the circular nature of the ring.
	index = index % len(c.sortedSet)

	// Return the calculated index where the key should be placed.
	return index, nil
}

// LoadOk checks if the host's current load is below the maximum allowed load.
// It returns true if the host's load is acceptable, otherwise false.
func (c *ConsistentHashing) LoadOk(host string) bool {
	// Retrieve the host's load data from the loadMap.
	if h, ok := c.loadMap.Load(host); ok {
		hostData := h.(*Host)
		// Compare the host's current load with the maximum allowed load.
		return hostData.Load < c.MaxLoad()
	}
	// Return false if host data is not found.
	return false
}

// MaxLoad calculates and returns the maximum allowed load per host based on the current
// total load across all hosts and the configured load factor.
func (c *ConsistentHashing) MaxLoad() int64 {
	// Retrieve the current total load across all hosts.
	totalLoad := atomic.LoadInt64(&c.totalLoad)

	// Ensure totalLoad is at least 1 to avoid division by zero.
	if totalLoad == 0 {
		totalLoad = 1
	}

	// Calculate the average load per host.
	avgLoadPerNode := float64(totalLoad) / float64(len(c.hostList))

	// Ensure avgLoadPerNode is at least 1 to avoid division by zero.
	if avgLoadPerNode == 0 {
		avgLoadPerNode = 1
	}

	// Calculate and return the maximum allowed load per host based on the load factor.
	return int64(math.Ceil(avgLoadPerNode * c.config.LoadFactor))
}

// GetLoads returns the current load for all hosts
func (c *ConsistentHashing) GetLoads() map[string]int64 {
	loads := make(map[string]int64)
	c.loadMap.Range(func(key, value interface{}) bool {
		loads[key.(string)] = value.(*Host).Load
		return true
	})
	return loads
}

func (c *ConsistentHashing) removeFromSortedSet(val uint64) {
	// Use binary search to find the index of the value
	index := sort.Search(len(c.sortedSet), func(i int) bool {
		return c.sortedSet[i] >= val
	})

	// If the value is found, remove it
	if index < len(c.sortedSet) && c.sortedSet[index] == val {
		// Remove the element by slicing
		c.sortedSet = append(c.sortedSet[:index], c.sortedSet[index+1:]...)
	}
}

// Hosts returns the list of current hosts
func (c *ConsistentHashing) Hosts() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return append([]string(nil), c.hostList...)
}
