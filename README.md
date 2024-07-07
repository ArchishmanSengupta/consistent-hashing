---

# Consistent-Hashing

Consistent-Hashing is a Go library designed for distributed load balancing using consistent hashing, enhanced with bounded loads. It provides efficient key distribution across a set of hosts while ensuring that no single host becomes overloaded beyond a specified limit.

## Installation

To install the package, execute the following command:

```bash
go get github.com/ArchishmanSengupta/consistent-hashing
```

## Usage

Here's a straightforward example of how to utilize the `consistent_hashing` package:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ArchishmanSengupta/consistent-hashing"
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

	// Remove a host from the ring
	hostToRemove := "host2"
	err = ch.Remove(ctx, hostToRemove)
	if err != nil {
		log.Printf("Error removing host %s: %v", hostToRemove, err)
	}

	// Print updated list of hosts
	fmt.Println("Remaining hosts after removal:", ch.Hosts())
}
```
## Output
```
Key key1 assigned to host host4
Key key2 assigned to host host3
Key key3 assigned to host host2
Key key4 assigned to host host1
Key key5 assigned to host host4
Current loads:
host2: 1
host3: 1
host4: 2
host1: 1
Remaining hosts after removal: [host1 host3 host4]
```

## Features

- **Consistent Hashing with Bounded Loads**: Distributes load evenly across hosts while limiting maximum host load.
- **Customizable Configuration**: Adjust replication factor, load factor, and hash function to suit specific requirements.
- **Thread-Safe Operations**: Ensures safe concurrent access for adding hosts, distributing keys, and managing loads.
- **Efficient Key Distribution**: Uses consistent hashing principles for efficient key assignment and lookup.

## Configuration

Customize consistent hashing behavior by providing a `Config` struct during instance creation:

```go
cfg := consistent_hashing.Config{
    ReplicationFactor: 20,    // Number of virtual nodes per host
    LoadFactor:        1.25,  // Maximum load factor before redistribution
    HashFunction:      fnv.New64a, // Custom hash function (optional)
}

ch, err := consistent_hashing.NewWithConfig(cfg)
```

## Benchmarking

Use the following command to run benchmarks:

```bash
go test -bench=. -benchmem
```

Example benchmark results:

```
goos: darwin
goarch: arm64
pkg: github.com/ArchishmanSengupta/consistent-hashing
BenchmarkAdd-10                             4626          19088168 ns/op           23748 B/op        895 allocs/op
BenchmarkGet-10                          6359968               186.5 ns/op            47 B/op          4 allocs/op
BenchmarkGetLeast-10                         180           6606643 ns/op              24 B/op          3 allocs/op
BenchmarkIncreaseLoad-10                13727469                83.96 ns/op           13 B/op          1 allocs/op
BenchmarkRemove-10                           139           8671612 ns/op           17696 B/op       1306 allocs/op
BenchmarkParallelOperations-10               884           1297025 ns/op             123 B/op          9 allocs/op
PASS
ok      github.com/ArchishmanSengupta/consistent-hashing        160.723s
```

## API Reference

### Methods

- `NewWithConfig(cfg Config) (*ConsistentHashing, error)`: Creates a new instance of ConsistentHashing with specified configuration.
- `Add(ctx context.Context, host string) error`: Adds a new host to the consistent hash ring.
- `Get(ctx context.Context, key string) (string, error)`: Retrieves the host responsible for a given key.
- `GetLeast(ctx context.Context, key string) (string, error)`: Retrieves the least loaded host for a given key.
- `IncreaseLoad(ctx context.Context, host string) error`: Increases the load for a specified host.
- `DecreaseLoad(ctx context.Context, host string) error`: Decreases the load for a specified host.
- `GetLoads() map[string]int64`: Retrieves the current load for all hosts.
- `Hosts() []string`: Retrieves the list of all hosts in the ring.
- `Remove(ctx context.Context, host string) error`: Removes a host from the ring.

## Examples

### Adding and Removing Hosts

```go
ch, _ := consistent_hashing.NewWithConfig(consistent_hashing.Config{})
ctx := context.Background()

// Adding hosts
ch.Add(ctx, "host1")
ch.Add(ctx, "host2")
ch.Add(ctx, "host3")

// Removing a host
err := ch.Remove(ctx, "host2")
if err != nil {
    log.Printf("Error removing host: %v", err)
}

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

Contributions are welcome! Feel free to submit a Pull Request with your enhancements or bug fixes.

--- 

This README now includes detailed instructions on adding, removing hosts, and showcases example usage scenarios for distributing keys and benchmarking performance. It should provide comprehensive guidance for users looking to integrate and leverage the Consistent-Hashing library effectively.