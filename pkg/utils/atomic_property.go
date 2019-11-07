/****************************************************************************
 * Copyright 2019, Optimizely, Inc. and contributors                        *
 *                                                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");          *
 * you may not use this file except in compliance with the License.         *
 * You may obtain a copy of the License at                                  *
 *                                                                          *
 *    http://www.apache.org/licenses/LICENSE-2.0                            *
 *                                                                          *
 * Unless required by applicable law or agreed to in writing, software      *
 * distributed under the License is distributed on an "AS IS" BASIS,        *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. *
 * See the License for the specific language governing permissions and      *
 * limitations under the License.                                           *
 ***************************************************************************/

// Package utils //
package utils

import "sync"

// AtomicProperty used to wrap a value with a read write lock.
type AtomicProperty struct {
	property interface{}
	lock sync.RWMutex
}

// Get gets the property used at creation using a read lock
func (p *AtomicProperty)Get() interface{} {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.property
}

// Set sets the property after creation using a write lock
func (p *AtomicProperty)Set(value interface{}) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.property = value
}

// NewAtomicPropertyWrapper creates a new atomic property holding the value passed in
func NewAtomicPropertyWrapper(value interface{}) *AtomicProperty {
	return &AtomicProperty{property:value}
}