package kmutex

import (
	"sync"

	"github.com/graxinc/syncmap"
)

// Must compare equal per instance.
type Locker interface {
	Lock()
	Unlock()
}

type SemaLocker struct {
	ch chan struct{}
}

// n must be > 0.
func NewSemaLocker(n int) SemaLocker {
	if n <= 0 {
		panic("kmutex: sema locker n <= 0")
	}
	return SemaLocker{make(chan struct{}, n)}
}

func (l SemaLocker) Lock() {
	l.ch <- struct{}{}
}

func (l SemaLocker) Unlock() {
	<-l.ch
}

// Concurrent safe.
type KMutex[T comparable] struct {
	m         syncmap.Map[T, Locker]
	newLocker func() Locker
}

func New[T comparable]() *KMutex[T] {
	l := func() Locker {
		return &sync.Mutex{}
	}
	return NewLocker[T](l)
}

func NewLocker[T comparable](newLocker func() Locker) *KMutex[T] {
	return &KMutex[T]{newLocker: newLocker}
}

// Takes an exclusive lock for key.
// unlock must be called when finished with key.
func (km *KMutex[T]) Lock(key T) (unlock func()) {
	// Some overhead here on the atomic ops in the Map,
	// but the goal is low contention when most Lock(key)
	// are with unique keys. Shared keys will contend anyways
	// on their underlying, since that is the point.
	for {
		mu, _ := km.m.LoadOrStore(key, km.newLocker())
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
