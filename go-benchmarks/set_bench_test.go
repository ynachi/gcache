package go_benchmarks

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"sync"
	"testing"
)

var ctxSet = context.Background()
var rdbSet *redis.Client

func init() {
	// Initialize a new Redis connection.
	rdbSet = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

func BenchmarkRedisSet(b *testing.B) {
	// Use a wait group to wait for all goroutines to finish.
	var wg sync.WaitGroup

	for i := 0; i < b.N; i++ {
		// Increment the wait group counter.
		wg.Add(1)

		j := i
		go func() {
			defer wg.Done()

			_, err := rdbSet.Set(ctxSet, fmt.Sprintf("key%d", j), fmt.Sprintf("key%d", j), 0).Result()
			if err != nil {
				b.Error(err)
				return
			}
		}()
	}
	// Wait for all goroutines to finish.
	wg.Wait()
}
