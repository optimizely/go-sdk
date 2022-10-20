package odp

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMinConfig(t *testing.T) {
	cache := NewLRUCache(1000, 2000)
	assert.Equal(t, 1000, cache.maxSize)
	assert.Equal(t, int64(2000), cache.timeoutInSecs)

	cache = NewLRUCache(0, 0)
	assert.Equal(t, 0, cache.maxSize)
	assert.Equal(t, int64(0), cache.timeoutInSecs)
}

func TestSaveAndLookupConfig(t *testing.T) {
	maxSize := 2
	cache := NewLRUCache(maxSize, 1000)

	cache.Save("1", 100) // [1]
	cache.Save("2", 200) // [1, 2]
	cache.Save("3", 300) // [2, 3]
	assert.Nil(t, cache.Lookup("1"))
	assert.Equal(t, 200, cache.Lookup("2")) // [3, 2]
	assert.Equal(t, 300, cache.Lookup("3")) // [2, 3]

	cache.Save("2", 201) // [3, 2]
	cache.Save("1", 101) // [2, 1]
	assert.Equal(t, 101, cache.Lookup("1"))
	assert.Equal(t, 201, cache.Lookup("2")) // [1, 2]
	assert.Nil(t, cache.Lookup("3"))

	cache.Save("3", 302) // [2, 3]
	assert.Nil(t, cache.Lookup("1"))
	assert.Equal(t, 201, cache.Lookup("2")) // [3, 2]
	assert.Equal(t, 302, cache.Lookup("3")) // [2, 3]

	cache.Save("1", 103) // [3, 1]
	assert.Equal(t, 103, cache.Lookup("1"))
	assert.Nil(t, cache.Lookup("2"))
	assert.Equal(t, 302, cache.Lookup("3")) // [1, 3]

	// Check if old items were deleted
	assert.Equal(t, maxSize, cache.queue.Len())
	assert.Equal(t, maxSize, len(cache.items))
}

func TestReset(t *testing.T) {
	maxSize := 2
	cache := NewLRUCache(maxSize, 1000)

	cache.Save("1", 100) // [1]
	cache.Save("2", 200) // [1, 2]

	assert.Equal(t, maxSize, cache.queue.Len())

	// cache reset
	cache.Reset()
	assert.Equal(t, 0, cache.queue.Len())
	assert.Equal(t, 0, len(cache.items))

	// validate cache fully functional after reset
	cache.Save("1", 100) // [1]
	cache.Save("2", 200) // [1, 2]
	cache.Save("3", 300) // [2, 3]
	assert.Nil(t, cache.Lookup("1"))
	assert.Equal(t, 200, cache.Lookup("2")) // [3, 2]
	assert.Equal(t, 300, cache.Lookup("3")) // [2, 3]

	cache.Save("2", 201) // [3, 2]
	cache.Save("1", 101) // [2, 1]
	assert.Equal(t, 101, cache.Lookup("1"))
	assert.Equal(t, 201, cache.Lookup("2")) // [1, 2]
	assert.Nil(t, cache.Lookup("3"))
}

func TestSizeZero(t *testing.T) {
	cache := NewLRUCache(0, 1000)
	cache.Save("1", 100)
	assert.Nil(t, cache.Lookup("1"))
	cache.Save("2", 200)
	assert.Nil(t, cache.Lookup("2"))
	cache.Reset()
	assert.Nil(t, cache.Lookup("1"))
	assert.Nil(t, cache.Lookup("2"))
}

func TestThreadSafe(t *testing.T) {
	maxSize := 1000
	cache := NewLRUCache(maxSize, 1000)
	wg := sync.WaitGroup{}

	save := func(k int, v interface{}, wg *sync.WaitGroup) {
		defer wg.Done()
		strKey := fmt.Sprintf("%d", k)
		cache.Save(strKey, v)
	}
	lookup := func(k int, wg *sync.WaitGroup, checkValue bool) {
		defer wg.Done()
		strKey := fmt.Sprintf("%d", k)
		v := cache.Lookup(strKey)
		if checkValue {
			assert.Equal(t, k*100, v)
		}
	}
	reset := func(wg *sync.WaitGroup) {
		defer wg.Done()
		cache.Reset()
	}

	// Add entries
	wg.Add(maxSize)
	for i := 1; i <= maxSize; i++ {
		go save(i, i*100, &wg)
	}
	wg.Wait()

	// Lookup previous entries
	wg.Add(maxSize)
	for i := 1; i <= maxSize; i++ {
		go lookup(i, &wg, true)
	}
	wg.Wait()

	// Add more entries then the max size
	wg.Add(maxSize)
	for i := maxSize + 1; i <= maxSize*2; i++ {
		go save(i, i*100, &wg)
	}
	wg.Wait()

	// Check if new entries replaced the old ones
	wg.Add(maxSize)
	for i := maxSize + 1; i <= maxSize*2; i++ {
		go lookup(i, &wg, true)
	}
	wg.Wait()

	// Check if old items were deleted
	assert.Equal(t, maxSize, cache.queue.Len())
	assert.Equal(t, maxSize, len(cache.items))

	wg.Add(maxSize * 3)
	// Check all api's simultaneously for race conditions
	for i := 1; i <= maxSize; i++ {
		go save(i, i*100, &wg)
		go lookup(i, &wg, false)
		go reset(&wg)
	}
	wg.Wait()
}

func TestTimeout(t *testing.T) {
	var maxTimeout int64 = 1
	// cache with timeout
	cache1 := NewLRUCache(1000, maxTimeout)
	// Zero timeout cache
	cache2 := NewLRUCache(1000, 0)

	cache1.Save("1", 100) // [1]
	cache1.Save("2", 200) // [1, 2]
	cache1.Save("3", 300) // [1,2,3]
	cache2.Save("1", 100) // [1]
	cache2.Save("2", 200) // [1, 2]
	cache2.Save("3", 300) // [1,2,3]

	// cache1 should expire while cache2 should not
	time.Sleep(1 * time.Second)

	// cache1 should expire
	assert.Nil(t, cache1.Lookup("1"))
	assert.Nil(t, cache1.Lookup("2"))
	assert.Nil(t, cache1.Lookup("3"))
	cache1.Save("1", 100)                    // [1]
	cache1.Save("4", 400)                    // [1,4]
	assert.Equal(t, 100, cache1.Lookup("1")) // [4,1]
	assert.Equal(t, 400, cache1.Lookup("4")) // [1,4]

	// cache2 should not expire
	assert.Equal(t, 100, cache2.Lookup("1"))
	assert.Equal(t, 200, cache2.Lookup("2"))
	assert.Equal(t, 300, cache2.Lookup("3"))
}
