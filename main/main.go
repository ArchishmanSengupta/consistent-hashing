package main

import (
	"context"
	"fmt"
	ch "github.com/ArchishmanSengupta/consistent-hashing"
	"github.com/spaolacci/murmur3"
	"hash"
	"log"
)

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

	// update load on a host and display loads
	fmt.Println("\nUpdate load on 127.0.0.2 to 4")
	err = hashRing.UpdateLoad(ctx, "127.0.0.2", 4)
	if err != nil {
		log.Fatalf("Failed to update load: %v", err)
	}
	printLoads(hashRing)

	// Decrease Load on a particular host
	fmt.Println("Decrease load on 127.0.0.2")
	err = hashRing.DecreaseLoad(ctx, "127.0.0.2")
	if err != nil {
		log.Fatalf("Failed to decrease load: %v", err)
	}
	printLoads(hashRing)

	// get least loaded host for a user
	fmt.Println("\nGetting least loaded host for user 'archie'")
	leastLoadedHost, err := hashRing.GetLeast(ctx, "archie")
	if err != nil {
		log.Fatalf("Failed to get least loaded host for user 'archie': %v", err)
	}
	fmt.Printf("Least loaded host for user 'archie': %s\n", leastLoadedHost)

	// remove a host and display hosts and loads
	fmt.Println("\nRemoving 127.0.0.4")
	err = hashRing.Remove(ctx, "127.0.0.4")
	if err != nil {
		log.Fatalf("Failed to remove host: %v", err)
	}
	printLoads(hashRing)
}
