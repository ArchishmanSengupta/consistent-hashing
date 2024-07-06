package main

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
