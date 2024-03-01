package kmutex

import (
	"sync"

	"github.com/shiolier/syncmap"
)

// Concurrent safe.
type KMutex[T comparable] struct {
	m syncmap.Map[T, *sync.Mutex]
}

func New[T comparable]() *KMutex[T] {
	return &KMutex[T]{}
}

// Takes an exclusive lock for key.
// unlock must be called when finished with key.
func (km *KMutex[T]) Lock(key T) (unlock func()) {
	// Some overhead here on the atomic ops in the Map,
	// but the goal is low contention when most Lock(key)
	// are with unique keys. Shared keys will contend anyways
	// on their lock, since that is the point of KMutex.
	for {
		mu, _ := km.m.LoadOrStore(key, &sync.Mutex{})
		mu.Lock()

		// For !ok case, Delete will happen before this Lock.
		// For mu != mu2, Delete plus another LoadOrStore could happen
		// between our LoadOrStore + Lock.
		// For mu we have Locked, another could be able to discover the
		// same case we are in, so must Unlock mu.
		if mu2, ok := km.m.Load(key); !ok || mu != mu2 {
			mu.Unlock()
			continue
		}

		return func() {
			km.m.Delete(key)
			mu.Unlock()
		}
	}
}
