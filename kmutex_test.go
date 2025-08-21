package kmutex_test

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/graxinc/kmutex"

	ikmutex "github.com/im7mortal/kmutex"
	"github.com/pkg/profile"
)

type kmutexer interface {
	Lock(key int) (unlock func())
}

func TestKmutex(t *testing.T) {
	t.Parallel()

	do := func(km *kmutex.KMutex[int]) {
		resources := make([]int, 10)

		var wg sync.WaitGroup
		for i := range 100 {
			wg.Add(1)
			go func() {
				defer wg.Done()

				rando := rand.New(rand.NewSource(int64(i))) //nolint:gosec

				for range 1000 {
					idx := rando.Intn(len(resources))

					// Relies on Go race detector being enabled.
					unlock := km.Lock(idx)
					resources[idx] = resources[idx] + 1
					unlock()
				}
			}()
		}
		wg.Wait()

		for i := range resources {
			km.Lock(i)
		}
	}
	t.Run("mu", func(t *testing.T) {
		do(kmutex.New[int]())
	})
	t.Run("sema_1", func(t *testing.T) {
		km := kmutex.NewLocker[int](func() kmutex.Locker {
			return kmutex.NewSemaLocker(1)
		})
		do(km)
	})
}

func TestSemaLocker_equal(t *testing.T) {
	t.Parallel()

	equalFatal := func(a, b kmutex.Locker) {
		t.Helper()
		if a != b {
			t.Fatal("not equal")
		}
	}

	s1 := kmutex.NewSemaLocker(1)
	s2 := kmutex.NewSemaLocker(1)
	s3 := kmutex.NewSemaLocker(2)

	equalFatal(s1, s1)
	equalFatal(s2, s2)
	equalFatal(s3, s3)

	if s1 == s2 {
		t.Fatal("should not be equal")
	}
	if s1 == s3 {
		t.Fatal("should not be equal")
	}
}

func BenchmarkKMutex_uniqueKeys(b *testing.B) {
	do := func(b *testing.B, km kmutexer) {
		defer profile.Start(profile.ClockProfile).Stop()

		do := func() {
			var wg sync.WaitGroup
			for i := range 100 {
				wg.Add(1)
				go func() {
					defer wg.Done()

					rando := rand.New(rand.NewSource(int64(i))) //nolint:gosec

					for range 1000 {
						idx := rando.Intn(100)

						unlock := km.Lock(idx)
						time.Sleep(time.Microsecond)
						unlock()
					}
				}()
			}
			wg.Wait()
		}

		for b.Loop() {
			do()
		}
	}

	b.Run("grax mu", func(b *testing.B) {
		km := kmutex.New[int]()
		do(b, km)
	})
	b.Run("grax sema 2", func(b *testing.B) {
		km := kmutex.NewLocker[int](func() kmutex.Locker {
			return kmutex.NewSemaLocker(2)
		})
		do(b, km)
	})
	b.Run("immortal", func(b *testing.B) {
		km := immortalKMutex{ikmutex.New()}
		do(b, km)
	})
}

type immortalKMutex struct {
	m *ikmutex.Kmutex
}

func (km immortalKMutex) Lock(key int) (unlock func()) {
	km.m.Lock(key)
	return func() {
		km.m.Unlock(key)
	}
}
