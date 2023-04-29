[![Build Status](https://circleci.com/gh/wk8/go-lru.svg?style=svg)](https://app.circleci.com/pipelines/github/wk8/go-lru)

# LRU for golang

Yet another LRU golang implementation.

This one has the added feature of making capacity a soft constraint: caches will guarantee that items will stay cached for a minimum period of time without getting discarded, even when the cache is over capacity.

## Usage

```go
import (
    "time"

    "github.com/wk8/go-lru"
)

func main() {
    // will try to keep the capacity limited to 2 items, but will
    // keep items for at least 1 hour after their last Get/Set, even
    // if that brings the cache over capacity
    cache := lru.New[string, int](2, time.Hour)

    // then immediately
    cache.Set("foo", 12)
    cache.Set("bar", 28)
    cache.Set("baz", 100)

    // here we go over capacity
    value, present := cache.Get("foo")
    // => 12, true
    length := cache.Len()
    // => 3

    // now the cache contains the 3 keys from oldest to newest
    // "bar", "baz", "foo"

    // now let's say we come back more than an hour later:
    time.Sleep(time.Hour + time.Second)

    // bar will have been evicted, since we're over capacity and it's been
    // last added or accessed more than an hour ago
    value, present = cache.Get("bar")
    // => 0, false
    length = cache.Len()
    // => 2
}

```

Note that this implementation is not thread-safe. If you need concurrent access then protect the caches with appropriate locks.
