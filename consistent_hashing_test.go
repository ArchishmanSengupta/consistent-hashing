package consistent_hashing

import (
	"context"
	"fmt"
	"hash/fnv"
	"testing"
)

func BenchmarkAdd(b *testing.B) {
	ch, _ := NewWithConfig(Config{ReplicationFactor: 100, LoadFactor: 1.25, HashFunction: fnv.New64a})
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		host := fmt.Sprintf("host-%d", i)
		_ = ch.Add(ctx, host)
	}
}

func BenchmarkGet(b *testing.B) {
	ch, _ := NewWithConfig(Config{ReplicationFactor: 100, LoadFactor: 1.25, HashFunction: fnv.New64a})
	ctx := context.Background()

	// Add some hosts
	for i := 0; i < 1000; i++ {
		host := fmt.Sprintf("host-%d", i)
		_ = ch.Add(ctx, host)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		_, _ = ch.Get(ctx, key)
	}
}

func BenchmarkGetLeast(b *testing.B) {
	ch, _ := NewWithConfig(Config{ReplicationFactor: 100, LoadFactor: 1.25, HashFunction: fnv.New64a})
	ctx := context.Background()

	// Add some hosts
	for i := 0; i < 1000; i++ {
		host := fmt.Sprintf("host-%d", i)
		_ = ch.Add(ctx, host)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		_, _ = ch.GetLeast(ctx, key)
	}
}

func BenchmarkIncreaseLoad(b *testing.B) {
	ch, _ := NewWithConfig(Config{ReplicationFactor: 100, LoadFactor: 1.25, HashFunction: fnv.New64a})
	ctx := context.Background()

	// Add some hosts
	for i := 0; i < 1000; i++ {
		host := fmt.Sprintf("host-%d", i)
		_ = ch.Add(ctx, host)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		host := fmt.Sprintf("host-%d", i%1000)
		_ = ch.IncreaseLoad(ctx, host)
	}
}

func BenchmarkRemove(b *testing.B) {
	ch, _ := NewWithConfig(Config{ReplicationFactor: 100, LoadFactor: 1.25, HashFunction: fnv.New64a})
	ctx := context.Background()

	// Add hosts
	for i := 0; i < b.N; i++ {
		host := fmt.Sprintf("host-%d", i)
		_ = ch.Add(ctx, host)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		host := fmt.Sprintf("host-%d", i)
		_ = ch.Remove(ctx, host)
	}
}

func BenchmarkParallelOperations(b *testing.B) {
	ch, _ := NewWithConfig(Config{ReplicationFactor: 100, LoadFactor: 1.25, HashFunction: fnv.New64a})
	ctx := context.Background()

	// Add initial hosts
	for i := 0; i < 1000; i++ {
		host := fmt.Sprintf("host-%d", i)
		_ = ch.Add(ctx, host)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 5 {
			case 0:
				host := fmt.Sprintf("host-%d", i)
				_ = ch.Add(ctx, host)
			case 1:
				key := fmt.Sprintf("key-%d", i)
				_, _ = ch.Get(ctx, key)
			case 2:
				key := fmt.Sprintf("key-%d", i)
				_, _ = ch.GetLeast(ctx, key)
			case 3:
				host := fmt.Sprintf("host-%d", i%1000)
				_ = ch.IncreaseLoad(ctx, host)
			case 4:
				host := fmt.Sprintf("host-%d", i%1000)
				_ = ch.Remove(ctx, host)
			}
			i++
		}
	})
}
