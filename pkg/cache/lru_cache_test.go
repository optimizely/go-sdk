/****************************************************************************
 * Copyright 2022-2025, Optimizely, Inc. and contributors                        *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    https://www.apache.org/licenses/LICENSE-2.0                           *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

// Package cache //
package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMinConfig(t *testing.T) {
	cache := NewLRUCache(1000, 5*time.Second)
	assert.Equal(t, 1000, cache.maxSize)
	assert.Equal(t, 5*time.Second, cache.timeout)

	cache = NewLRUCache(0, 0*time.Second)
	assert.Equal(t, 0, cache.maxSize)
	assert.Equal(t, 0*time.Second, cache.timeout)
}

func TestSaveAndLookupConfig(t *testing.T) {
	maxSize := 2
	cache := NewLRUCache(maxSize, 1000*time.Second)

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
	cache := NewLRUCache(maxSize, 1000*time.Second)

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
	cache := NewLRUCache(0, 1000*time.Second)
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
	cache := NewLRUCache(maxSize, 1000*time.Second)
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
	var maxTimeout = 1 * time.Second
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

func TestRemove(t *testing.T) {
	// Test removing an existing key
	t.Run("Remove existing key", func(t *testing.T) {
		cache := NewLRUCache(3, 1000*time.Second)

		// Add items to cache
		cache.Save("1", 100)
		cache.Save("2", 200)
		cache.Save("3", 300)

		// Verify items exist
		assert.Equal(t, 100, cache.Lookup("1"))
		assert.Equal(t, 200, cache.Lookup("2"))
		assert.Equal(t, 300, cache.Lookup("3"))
		assert.Equal(t, 3, cache.queue.Len())
		assert.Equal(t, 3, len(cache.items))

		// Remove an item
		cache.Remove("2")

		// Verify item was removed
		assert.Equal(t, 100, cache.Lookup("1"))
		assert.Nil(t, cache.Lookup("2"))
		assert.Equal(t, 300, cache.Lookup("3"))
		assert.Equal(t, 2, cache.queue.Len())
		assert.Equal(t, 2, len(cache.items))
	})

	// Test removing a non-existent key
	t.Run("Remove non-existent key", func(t *testing.T) {
		cache := NewLRUCache(3, 1000*time.Second)

		// Add items to cache
		cache.Save("1", 100)
		cache.Save("2", 200)

		// Remove a non-existent key
		cache.Remove("3")

		// Verify state remains unchanged
		assert.Equal(t, 100, cache.Lookup("1"))
		assert.Equal(t, 200, cache.Lookup("2"))
		assert.Equal(t, 2, cache.queue.Len())
		assert.Equal(t, 2, len(cache.items))
	})

	// Test removing from a zero-sized cache
	t.Run("Remove from zero-sized cache", func(t *testing.T) {
		cache := NewLRUCache(0, 1000*time.Second)

		// Try to add and remove items
		cache.Save("1", 100)
		cache.Remove("1")

		// Verify nothing happened
		assert.Nil(t, cache.Lookup("1"))
		assert.Equal(t, 0, cache.queue.Len())
		assert.Equal(t, 0, len(cache.items))
	})

	// Test removing and then adding back
	t.Run("Remove and add back", func(t *testing.T) {
		cache := NewLRUCache(3, 1000*time.Second)

		// Add items to cache
		cache.Save("1", 100)
		cache.Save("2", 200)
		cache.Save("3", 300)

		// Remove an item
		cache.Remove("2")

		// Add it back with a different value
		cache.Save("2", 201)

		// Verify item was added back
		assert.Equal(t, 100, cache.Lookup("1"))
		assert.Equal(t, 201, cache.Lookup("2"))
		assert.Equal(t, 300, cache.Lookup("3"))
		assert.Equal(t, 3, cache.queue.Len())
		assert.Equal(t, 3, len(cache.items))
	})

	// Test thread safety of Remove
	t.Run("Thread safety", func(t *testing.T) {
		maxSize := 100
		cache := NewLRUCache(maxSize, 1000*time.Second)
		wg := sync.WaitGroup{}

		// Add entries
		for i := 1; i <= maxSize; i++ {
			cache.Save(fmt.Sprintf("%d", i), i*100)
		}

		// Concurrently remove half the entries
		wg.Add(maxSize / 2)
		for i := 1; i <= maxSize/2; i++ {
			go func(k int) {
				defer wg.Done()
				cache.Remove(fmt.Sprintf("%d", k))
			}(i)
		}
		wg.Wait()

		// Verify first half is removed, second half remains
		for i := 1; i <= maxSize; i++ {
			if i <= maxSize/2 {
				assert.Nil(t, cache.Lookup(fmt.Sprintf("%d", i)))
			} else {
				assert.Equal(t, i*100, cache.Lookup(fmt.Sprintf("%d", i)))
			}
		}

		// Verify cache size
		assert.Equal(t, maxSize/2, cache.queue.Len())
		assert.Equal(t, maxSize/2, len(cache.items))
	})
}
