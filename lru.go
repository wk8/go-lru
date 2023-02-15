package lru

import (
	"time"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type LRU[K comparable, V any] struct {
	capacity   int
	keepPeriod time.Duration

	om *orderedmap.OrderedMap[K, *timestampedValue[V]]
}

type timestampedValue[V any] struct {
	value     V
	timestamp time.Time
}

// New creates a new LRU with the given capacity and keep period.
// The keep period defines a minimum duration during which items
// will be kept in the cache after each Get/Set calls, regardless of
// the current length of the cache. In that sense, the capacity is a
// "soft" constraint that may be exceeded to accommodate the keep period.
// If keepPeriod is <= 0, then it is ignored.
func New[K comparable, V any](capacity int, keepPeriod time.Duration) *LRU[K, V] {
	return &LRU[K, V]{
		capacity:   capacity,
		keepPeriod: keepPeriod,
		om:         orderedmap.New[K, *timestampedValue[V]](orderedmap.WithCapacity[K, *timestampedValue[V]](capacity)),
	}
}

var now = time.Now //nolint:gochecknoglobals

// Get looks for the given key, and returns the value associated with it,
// or V's nil value if not found. The boolean it returns says whether the key is present in the cache.
// It marks the key as recently accessed.
func (l *LRU[K, V]) Get(key K) (val V, present bool) {
	timestamp := now()

	if l.keepPeriod > 0 {
		l.prune(timestamp)
	}

	withTimestamp, err := l.om.GetAndMoveToBack(key)
	if err == nil {
		present = true
		val = withTimestamp.value
		withTimestamp.timestamp = timestamp
	}

	return
}

// Set sets the key-value pair, and returns what `Get` would have returned
// on that key prior to the call to `Set`.
// It marks the key as recently accessed.
func (l *LRU[K, V]) Set(key K, value V) (val V, present bool) {
	timestamp := now()

	withTimestamp, present := l.om.Set(key, &timestampedValue[V]{
		value:     value,
		timestamp: timestamp,
	})

	if present {
		val = withTimestamp.value
		_ = l.om.MoveToBack(key)
	}

	l.prune(timestamp)

	return
}

func (l *LRU[K, V]) prune(timestamp time.Time) {
	if l.capacity > 0 {
		cutoff := timestamp.Add(-l.keepPeriod)
		for pair := l.om.Oldest(); pair != nil && l.Len() > l.capacity &&
			(l.keepPeriod <= 0 || pair.Value.timestamp.Before(cutoff)); pair = pair.Next() {
			l.om.Delete(pair.Key)
		}
	}
}

// Len returns the cache's current length.
func (l *LRU[K, V]) Len() int {
	return l.om.Len()
}
