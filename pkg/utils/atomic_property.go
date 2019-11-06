package utils

import "sync"

type AtomicProperty struct {
	property interface{}
	lock sync.RWMutex
}

func (p *AtomicProperty)Get() interface{} {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.property
}

func (p *AtomicProperty)Set(value interface{}) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.property = value
}

func NewAtomicPropertyWrapper(value interface{}) *AtomicProperty {
	return &AtomicProperty{property:value}
}