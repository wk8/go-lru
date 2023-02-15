package lru

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicUsageWithNoKeepPeriod(t *testing.T) {
	lru := New[int, int](3, 0)

	// set(i, 2 * i) and check the contents after each
	expectedKeys := make([]int, 0, 10)
	expectedValues := make([]int, 0, 10)
	for i := 0; i < 10; i++ {
		oldValue, present := lru.Set(i, 2*i)
		assert.Equal(t, 0, oldValue)
		assert.False(t, present)

		expectedKeys = append(expectedKeys, i)
		expectedValues = append(expectedValues, 2*i)

		assert.True(t, assertCacheContains(t, lru, tail(expectedKeys, 3), tail(expectedValues, 3)),
			"unexpected content at iteration %d", i)
	}

	value, present := lru.Get(10)
	assert.Equal(t, 0, value)
	assert.False(t, present)

	value, present = lru.Get(8)
	assert.Equal(t, 16, value)
	assert.True(t, present)
	assertCacheContains(t, lru, []int{7, 9, 8}, []int{14, 18, 16})

	value, present = lru.Set(7, 99)
	assert.Equal(t, 14, value)
	assert.True(t, present)
	assertCacheContains(t, lru, []int{9, 8, 7}, []int{18, 16, 99})
}

func TestBasicUsageWithKeepPeriod(t *testing.T) {
	lru := New[int, int](3, 3*time.Minute+30*time.Second)

	clock := &testClock{}

	// set(i, 2 * i) and check the contents after each
	// 1 minute in between each insert
	expectedKeys := make([]int, 0, 10)
	expectedValues := make([]int, 0, 10)
	for i := 0; i < 10; i++ {
		clock.ticks(t, time.Minute)

		oldValue, present := lru.Set(i, 2*i)
		assert.Equal(t, 0, oldValue)
		assert.False(t, present)

		expectedKeys = append(expectedKeys, i)
		expectedValues = append(expectedValues, 2*i)

		// we should have up to 4 entries in there, because of the keep period
		assert.True(t, assertCacheContains(t, lru, tail(expectedKeys, 4), tail(expectedValues, 4)),
			"unexpected content at iteration %d", i)
	}

	clock.ticks(t, 30*time.Minute)
	value, present := lru.Get(7)
	assert.Equal(t, 14, value)
	assert.True(t, present)
	assertCacheContains(t, lru, []int{8, 9, 7}, []int{16, 18, 14})
}

// Helpers below

// assertCacheContains asserts that the cache contains the given keys and values
// from oldest to newest.
func assertCacheContains[K comparable, V any](
	t *testing.T, lru *LRU[K, V], expectedKeys []K, expectedValues []V,
) bool {
	if assert.Equal(t, len(expectedKeys), len(expectedValues), "key & values have different lengths") &&
		assert.Equal(t, len(expectedKeys), lru.om.Len(), "unexpected length") {
		i := 0
		success := true

		for pair := lru.om.Oldest(); pair != nil; pair = pair.Next() {
			success = assert.Equal(t, expectedKeys[i], pair.Key, "key index=%d", i) && success
			success = assert.Equal(t, expectedValues[i], pair.Value.value, "value index=%d", i) && success
			i++
		}

		return success
	}

	return false
}

// returns the last n items of a
func tail[T any](a []T, n int) []T {
	if n >= len(a) {
		return a
	}
	return a[len(a)-n:]
}

var t0 = time.Now()

// allows overriding the now function to return the given "times" in order.
// The times are in fact durations, relative to an arbitrary t0 time - makes it easier
// to read in tests.
// Also checks at the end of the test that all times have been exhausted - fails the test
// if not.
func withTimes(t *testing.T, times ...time.Duration) {
	previous := now

	nextIndex := 0
	n := len(times)
	now = func() time.Time {
		require.Less(t, nextIndex, n, "no more times!")
		timestamp := t0.Add(times[nextIndex])
		nextIndex++
		return timestamp
	}

	t.Cleanup(func() {
		now = previous

		assert.Equal(t, n, nextIndex, "%d unused times", n-nextIndex)
	})
}

type testClock struct {
	elapsed time.Duration
}

func (c *testClock) ticks(t *testing.T, deltas ...time.Duration) {
	times := make([]time.Duration, len(deltas))
	for i, delta := range deltas {
		c.elapsed += delta
		times[i] = c.elapsed
	}

	withTimes(t, times...)
}
