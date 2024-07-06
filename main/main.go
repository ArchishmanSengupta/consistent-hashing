package main

import (
	"context"
	"fmt"
	ch "github.com/ArchishmanSengupta/consistent-hashing"
	"github.com/spaolacci/murmur3"
	"hash"
	"log"
)

// Consistent Hashing with Bounded Loads Implementation

// Todo: Core Data Structures
// - 1. CH config Struct
//		- replication factor -> number of virtual nodes for each host
// 		- load factor -> max load factor before redistribution
// - 2. Host struct
//		- Name -> Host Name or identifier
// - 3. Consistent Hashing Struct
// 		- config Config
//		- hosts sync.Map -> Map Hash Values to the host values
// 		- sortedSet []uint64 -> Sorted Slice of hash values
//		- loadMap sync.Map -> map of host to Host struct
// 		- totalLoad int64 -> Total Load accross all hosts
//		- hostList []string -> List of all hosts ['uat-server.something.com', 'be-server.something.com']
// 		- mu sync.RWMutex -> mutex for synchronizing access

// Todo: Core functions:
/*
-> NewWithConfig: Initializes a new Consistent Hashing instance
1. Add: Adds Hosts to the Hash Ring
2. Get: Map a key(user) to a host
3. GetLeast: Gets least loaded host for a key
4. IncreaseLead: Increase Load of a particular server
5. UpdateLoad: Update load of a particular server
6. DecreaseLoad: Decrease load of a particular server
7. Remove: Remove a server from the Hash Ring
8. GetLoads: Get current loads of all hosts
*/
//

// Todo: Helper functions:
/*
 - hash:
hash function for keys and host names(can use murmur hash)
 - search:
	Binary search on the sorted set of hashes
	(influencer mislead karte reh gaye baccho ko that
	DSA is not required)
 - loadOk: check host's load is within bounds
 - maxLoad: maximum allowed load computation
 - removeSlice: remove a hash from the sorted set
*/

// Todo: Error Handling
/*
 - ErrNoHost: "no host added"
 - ErrHostNotFound: "host not found"
*/

// stand by for now
func customMurmurHash() hash.Hash64 {
	return murmur3.New64()
}

// printLoads prints the current load of all hosts
func printLoads(c *ch.ConsistentHashing) {
	fmt.Println("Current loads:")
	for host, load := range c.GetLoads() {
		fmt.Printf("Host: %s -> Load: %d\n", host, load)
	}
}

func main() {
	// create a context for managing request-scoped values, cancellation, and deadlines.
	// for controlling the lifecycle of a request.
	ctx := context.Background()

	// create a new ch instance with default config
	cfg := ch.Config{
		ReplicationFactor: 3,
		LoadFactor:        1.5,
		HashFunction:      customMurmurHash,
	}

	hashRing, err := ch.NewWithConfig(cfg)
	if err != nil {
		fmt.Printf("Error creating hash ring: %v\n", err)
		return
	}

	// add hosts to the hash ring
	hosts := []string{"127.0.0.1", "127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5", "127.0.0.6"}
	for _, host := range hosts {
		err := hashRing.Add(ctx, host)
		if err != nil {
			log.Fatalf("Failed to add host %s: %v", host, err)
		}
	}

	// show current hosts
	fmt.Println("HOSTs added to the hash ring: ", hashRing.Hosts())

	// add keys to the hash ring and display the mapped host
	users := []string{"striver", "arpitbhayani", "piyushgarg", "hkiratsingh", "archie", "sergeybin"}
	fmt.Println("User to Host Mapping: ")
	for _, user := range users {
		host, err := hashRing.Get(ctx, user)
		if err != nil {
			log.Fatalf("Failed to get host %s: %v", user, err)
		}
		fmt.Printf("User: %s -> Host: %s\n", user, host)
	}

	// Increment Load on a particular host
	fmt.Println("Incrementting load on 127.0.0.2")
	err = hashRing.IncreaseLoad(ctx, "127.0.0.2")
	if err != nil {
		log.Fatalf("Failed to increment load: %v", err)
	}
	printLoads(hashRing)
}
