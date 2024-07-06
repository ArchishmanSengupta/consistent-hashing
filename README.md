# Consistent-Hashing
A Go library for distributed load balancing using consistent hashing (with bounded loads)

## Installation

To install the package, use the following command:

```bash
go get github.com/ArchishmanSengupta/consistent-hashing
```

## Usage

Here's a simple example of how to use the consistent_hashing package:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/yourusername/consistent-hashing"
)

func main() {
	// Create a new ConsistentHashing instance with default configuration
	ch, err := consistent_hashing.NewWithConfig(consistent_hashing.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Add hosts to the consistent hash ring
	hosts := []string{"host1", "host2", "host3", "host4"}
	ctx := context.Background()
	for _, host := range hosts {
		err := ch.Add(ctx, host)
		if err != nil {
			log.Printf("Error adding host %s: %v", host, err)
		}
	}

	// Distribute some keys
	keys := []string{"key1", "key2", "key3", "key4", "key5"}
	for _, key := range keys {
		host, err := ch.GetLeast(ctx, key)
		if err != nil {
			log.Printf("Error getting host for key %s: %v", key, err)
			continue
		}
		fmt.Printf("Key %s assigned to host %s\n", key, host)
		
		// Increase the load for the assigned host
		err = ch.IncreaseLoad(ctx, host)
		if err != nil {
			log.Printf("Error increasing load for host %s: %v", host, err)
		}
	}

	// Print current loads
	loads := ch.GetLoads()
	fmt.Println("Current loads:")
	for host, load := range loads {
		fmt.Printf("%s: %d\n", host, load)
	}
}
```

## Features

- Consistent hashing with bounded loads
- Customizable replication factor and load factor
- Thread-safe operations
- Efficient key distribution and host lookup

## Configuration

You can customize the consistent hashing behavior by providing a `Config` struct when creating a new instance:

```go
cfg := consistent_hashing.Config{
    ReplicationFactor: 20,    // Number of virtual nodes per host
    LoadFactor:        1.25,  // Maximum load factor before redistribution
    HashFunction:      fnv.New64a, // Custom hash function (optional)
}

ch, err := consistent_hashing.NewWithConfig(cfg)
```

## API Reference

- `NewWithConfig(cfg Config) (*ConsistentHashing, error)`: Create a new ConsistentHashing instance
- `Add(ctx context.Context, host string) error`: Add a new host to the ring
- `Get(ctx context.Context, key string) (string, error)`: Get the host for a given key
- `GetLeast(ctx context.Context, key string) (string, error)`: Get the least loaded host for a given key
- `IncreaseLoad(ctx context.Context, host string) error`: Increase the load for a host
- `DecreaseLoad(ctx context.Context, host string) error`: Decrease the load for a host
- `GetLoads() map[string]int64`: Get the current loads for all hosts
- `Hosts() []string`: Get the list of all hosts in the ring

## Examples

### Adding and Removing Hosts

```go
ch, _ := consistent_hashing.NewWithConfig(consistent_hashing.Config{})
ctx := context.Background()

// Adding hosts
ch.Add(ctx, "host1")
ch.Add(ctx, "host2")
ch.Add(ctx, "host3")

// Removing a host (not implemented in the current version) [Work In Progress]
// ch.Remove(ctx, "host2")

fmt.Println("Current hosts:", ch.Hosts())
```

### Distributing Keys with Load Balancing

```go
ch, _ := consistent_hashing.NewWithConfig(consistent_hashing.Config{})
ctx := context.Background()

// Add hosts
for i := 1; i <= 5; i++ {
    ch.Add(ctx, fmt.Sprintf("host%d", i))
}

// Distribute keys
keys := []string{"user1", "user2", "user3", "user4", "user5", "user6", "user7", "user8", "user9", "user10"}
for _, key := range keys {
    host, _ := ch.GetLeast(ctx, key)
    fmt.Printf("Key %s assigned to %s\n", key, host)
    ch.IncreaseLoad(ctx, host)
}

// Print final loads
loads := ch.GetLoads()
for host, load := range loads {
    fmt.Printf("%s load: %d\n", host, load)
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
