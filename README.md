# kmutex

[![Go Reference](https://pkg.go.dev/badge/github.com/graxinc/kmutex.svg)](https://pkg.go.dev/github.com/graxinc/kmutex)

## Purpose

Needing to individually lock a set of uniquely identified resources.

A common case arises when caching, since there is a period between the cache `Get` and `Set` where work is done generating the cache item. Depending on the work, duplication could be expensive.

## Usage

Usage is similar to a standard `sync.Mutex`, but with a key argument. Concurrently you could lock key `5` and perform some work:

```
m := kmutex.New[int]()

unlock := m.Lock(5)

// Do work related to 5.

unlock()
```

## Improvements

In cases where a hash can be generated easily of `T`, then instead of the `sync.Map` that is currently used we could use a bucketed slice of `sync.Mutex` guarded standard maps, for a different performance profile.
