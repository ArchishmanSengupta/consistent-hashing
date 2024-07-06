# Consistent-Hashing
A Go library for distributed load balancing using consistent hashing (with bounded loads)

Running a large-scale web service, such as content hosting, necessarily requires load balancing — distributing clients uniformly across multiple servers such that none get overloaded. Further, it is desirable to find an allocation that does not change very much over time in a dynamic environment in which both clients and servers can be added or removed at any time. In other words, we need the allocation of clients to servers to be consistent over time.

This algorithm was originally published by Mikkel Thorup from the University of Copenhagen in collaboration with Google researchers Vahab Mirrokni and Morteza Zadimoghaddam in the paper [Consistent Hashing With Bounded Loads](https://arxiv.org/pdf/1608.01350)

<img width="1028" alt="Screenshot 2024-07-06 at 6 04 39 PM" src="https://github.com/ArchishmanSengupta/consistent-hashing/assets/71402528/184b8287-bb03-4a51-b0b5-ce6451a33172">


A [Blog](https://medium.com/vimeo-engineering-blog/improving-load-balancing-with-a-new-consistent-hashing-algorithm-9f1bd75709ed) on How this algorithm helped Vimeo decrease the cache bandwidth by a factor of almost 8, eliminating a scaling bottleneck.

