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
	"container/list"
	"sync"
	"time"
)

// Cache is used for caching ODP segments
type Cache interface {
	Save(key string, value interface{})
	Lookup(key string) interface{}
	Reset()
}

// CacheWithRemove extends the Cache interface with removal capability
type CacheWithRemove interface {
	Cache
	Remove(key string)
}

type cacheElement struct {
	data   interface{}
	time   time.Time
	keyPtr *list.Element
}

// LRUCache a Least Recently Used in-memory cache
type LRUCache struct {
	queue   *list.List
	items   map[string]*cacheElement
	maxSize int
	timeout time.Duration
	lock    sync.RWMutex
}

// NewLRUCache returns a new instance of Least Recently Used in-memory cache
func NewLRUCache(size int, timeout time.Duration) *LRUCache {
	return &LRUCache{queue: list.New(), items: make(map[string]*cacheElement), maxSize: size, timeout: timeout}
}

// Save stores a new element into the cache
func (l *LRUCache) Save(key string, value interface{}) {
	if l.maxSize <= 0 {
		return
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	if item, ok := l.items[key]; !ok {
		// remove the last object if queue is full
		if l.maxSize == len(l.items) {
			back := l.queue.Back()
			l.queue.Remove(back)
			delete(l.items, back.Value.(string))
		}
		// push the new object to the front of the queue
		l.items[key] = &cacheElement{data: value, keyPtr: l.queue.PushFront(key), time: time.Now()}
	} else {
		item.data = value
		l.items[key] = item
		l.queue.MoveToFront(item.keyPtr)
	}
}

// Lookup retrieves an element from the cache, reordering the elements
func (l *LRUCache) Lookup(key string) interface{} {
	if l.maxSize <= 0 {
		return nil
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	if item, ok := l.items[key]; ok {
		if l.isValid(item) {
			l.queue.MoveToFront(item.keyPtr)
			return item.data
		}
		l.queue.Remove(item.keyPtr)
		delete(l.items, item.keyPtr.Value.(string))
	}
	return nil
}

// Reset clears all the elements from the cache
func (l *LRUCache) Reset() {
	if l.maxSize <= 0 {
		return
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	l.queue = list.New()
	l.items = make(map[string]*cacheElement)
}

// Remove deletes an element from the cache by key
func (l *LRUCache) Remove(key string) {
	if l.maxSize <= 0 {
		return
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	if item, ok := l.items[key]; ok {
		l.queue.Remove(item.keyPtr)
		delete(l.items, key)
	}
}

func (l *LRUCache) isValid(e *cacheElement) bool {
	if l.timeout <= 0 {
		return true
	}
	currenttime := time.Now()
	elapsedtime := currenttime.Sub(e.time)
	return l.timeout > elapsedtime
}
