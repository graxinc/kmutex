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
	km := kmutex.New[int]()

	resources := make([]int, 10)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		i := i

		wg.Add(1)
		go func() {
			defer wg.Done()

			rando := rand.New(rand.NewSource(int64(i))) //nolint:gosec

			for j := 0; j < 1000; j++ {
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

func BenchmarkKMutex_uniqueKeys(b *testing.B) {
	do := func(b *testing.B, km kmutexer) {
		defer profile.Start(profile.ClockProfile).Stop()

		do := func() {
			var wg sync.WaitGroup
			for i := 0; i < 100; i++ {
				i := i
				wg.Add(1)
				go func() {
					defer wg.Done()

					rando := rand.New(rand.NewSource(int64(i))) //nolint:gosec

					for j := 0; j < 1000; j++ {
						idx := rando.Intn(100)

						unlock := km.Lock(idx)
						time.Sleep(time.Microsecond)
						unlock()
					}
				}()
			}
			wg.Wait()
		}

		for i := 0; i < b.N; i++ {
			do()
		}
	}

	b.Run("grax", func(b *testing.B) {
		km := kmutex.New[int]()
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
