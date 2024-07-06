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
