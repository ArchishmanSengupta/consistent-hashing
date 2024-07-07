package consistent_hashing

import (
	"context"
	"fmt"
	"hash/fnv"
	"math"
	"sync"
	"testing"
)

func TestNewWithConfig(t *testing.T) {
	cfg := Config{
		ReplicationFactor: 10,
		LoadFactor:        1.25,
		HashFunction:      fnv.New64a,
	}
	ch, err := NewWithConfig(cfg)
	if err != nil {
		t.Errorf("NewWithConfig failed: %v", err)
	}
	if ch.config.ReplicationFactor != 10 {
		t.Errorf("Expected ReplicationFactor 10, got %d", ch.config.ReplicationFactor)
	}
	if ch.config.LoadFactor != 1.25 {
		t.Errorf("Expected LoadFactor 1.25, got %f", ch.config.LoadFactor)
	}
}

func TestAdd(t *testing.T) {
	ch, _ := NewWithConfig(Config{ReplicationFactor: 3, LoadFactor: 1.25, HashFunction: fnv.New64a})
	ctx := context.Background()
	err := ch.Add(ctx, "host1")
	if err != nil {
		t.Errorf("Add failed: %v", err)
	}
	if len(ch.hostList) != 1 {
		t.Errorf("Expected 1 host, got %d", len(ch.hostList))
	}
	if len(ch.sortedSet) != 3 {
		t.Errorf("Expected 3 virtual nodes, got %d", len(ch.sortedSet))
	}
}

func TestGet(t *testing.T) {
	ch, _ := NewWithConfig(Config{ReplicationFactor: 3, LoadFactor: 1.25, HashFunction: fnv.New64a})
	ctx := context.Background()
	ch.Add(ctx, "host1")
	ch.Add(ctx, "host2")
	host, err := ch.Get(ctx, "key1")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if host != "host1" && host != "host2" {
		t.Errorf("Expected host1 or host2, got %s", host)
	}
}

func TestGetLeast(t *testing.T) {
	ch, _ := NewWithConfig(Config{ReplicationFactor: 3, LoadFactor: 1.25, HashFunction: fnv.New64a})
	ctx := context.Background()
	ch.Add(ctx, "host1")
	ch.Add(ctx, "host2")
	ch.IncreaseLoad(ctx, "host1")
	host, err := ch.GetLeast(ctx, "key1")
	if err != nil {
		t.Errorf("GetLeast failed: %v", err)
	}
	if host != "host2" {
		t.Errorf("Expected host2, got %s", host)
	}
}

func TestIncreaseLoad(t *testing.T) {
	ch, _ := NewWithConfig(Config{ReplicationFactor: 3, LoadFactor: 1.25, HashFunction: fnv.New64a})
	ctx := context.Background()
	ch.Add(ctx, "host1")
	err := ch.IncreaseLoad(ctx, "host1")
	if err != nil {
		t.Errorf("IncreaseLoad failed: %v", err)
	}
	loads := ch.GetLoads()
	if loads["host1"] != 1 {
		t.Errorf("Expected load 1, got %d", loads["host1"])
	}
}

func TestDecreaseLoad(t *testing.T) {
	ch, _ := NewWithConfig(Config{ReplicationFactor: 3, LoadFactor: 1.25, HashFunction: fnv.New64a})
	ctx := context.Background()
	ch.Add(ctx, "host1")
	ch.IncreaseLoad(ctx, "host1")
	ch.DecreaseLoad(ctx, "host1")
	loads := ch.GetLoads()
	if loads["host1"] != 0 {
		t.Errorf("Expected load 0, got %d", loads["host1"])
	}
}

func TestUpdateLoad(t *testing.T) {
	ch, _ := NewWithConfig(Config{ReplicationFactor: 3, LoadFactor: 1.25, HashFunction: fnv.New64a})
	ctx := context.Background()
	ch.Add(ctx, "host1")
	err := ch.UpdateLoad(ctx, "host1", 5)
	if err != nil {
		t.Errorf("UpdateLoad failed: %v", err)
	}
	loads := ch.GetLoads()
	if loads["host1"] != 5 {
		t.Errorf("Expected load 5, got %d", loads["host1"])
	}
}

func TestRemove(t *testing.T) {
	ch, _ := NewWithConfig(Config{ReplicationFactor: 3, LoadFactor: 1.25, HashFunction: fnv.New64a})
	ctx := context.Background()
	ch.Add(ctx, "host1")
	ch.Add(ctx, "host2")
	err := ch.Remove(ctx, "host1")
	if err != nil {
		t.Errorf("Remove failed: %v", err)
	}
	if len(ch.hostList) != 1 {
		t.Errorf("Expected 1 host, got %d", len(ch.hostList))
	}
	if ch.hostList[0] != "host2" {
		t.Errorf("Expected host2, got %s", ch.hostList[0])
	}
}

func TestConcurrency(t *testing.T) {
	ch, _ := NewWithConfig(Config{ReplicationFactor: 3, LoadFactor: 1.25, HashFunction: fnv.New64a})
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			host := fmt.Sprintf("host%d", i)
			ch.Add(ctx, host)
			ch.Get(ctx, host)
			ch.GetLeast(ctx, host)
			ch.IncreaseLoad(ctx, host)
			ch.DecreaseLoad(ctx, host)
			ch.UpdateLoad(ctx, host, int64(i))
			if i%2 == 0 {
				ch.Remove(ctx, host)
			}
		}(i)
	}
	wg.Wait()
}

func TestLoadBalancing(t *testing.T) {
	ch, _ := NewWithConfig(Config{ReplicationFactor: 100, LoadFactor: 1.25, HashFunction: fnv.New64a})
	ctx := context.Background()
	hosts := []string{"host1", "host2", "host3", "host4", "host5"}

	// Add hosts to the consistent hashing instance
	for _, host := range hosts {
		if err := ch.Add(ctx, host); err != nil {
			t.Fatalf("Error adding host %s: %v", host, err)
		}
	}

	keyCount := 10000
	hostCounts := make(map[string]int)

	// Generate keys and assign them to hosts
	for i := 0; i < keyCount; i++ {
		key := fmt.Sprintf("key%d", i)
		host, err := ch.GetLeast(ctx, key)
		if err != nil {
			t.Fatalf("Error getting host for key %s: %v", key, err)
		}
		hostCounts[host]++
		if err := ch.IncreaseLoad(ctx, host); err != nil {
			t.Fatalf("Error increasing load for host %s: %v", host, err)
		}
	}

	// Check if the load is reasonably balanced
	expectedCount := keyCount / len(hosts)
	tolerance := float64(expectedCount) * 0.1 // 10% tolerance

	for host, count := range hostCounts {
		if math.Abs(float64(count-expectedCount)) > tolerance {
			t.Errorf("Load for %s is not balanced. Expected around %d, got %d", host, expectedCount, count)
		}
	}
}
