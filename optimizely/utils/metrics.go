package utils

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Metrics collects and increment any int numbers, can be used to safely count things
type Metrics struct {
	userLock sync.RWMutex
	userData map[string]int
}

// NewMetrics makes a map to collect any counts/metrics
func NewMetrics() *Metrics {
	return &Metrics{userData: map[string]int{}}
}

// Add increments value for given key and returns new value
func (m *Metrics) Add(key string, delta int) int {
	m.userLock.Lock()
	defer m.userLock.Unlock()
	m.userData[key] += delta
	return m.userData[key]
}

// Inc increments value for given key by one
func (m *Metrics) Inc(key string) int {
	return m.Add(key, 1)
}

// Set value for given key
func (m *Metrics) Set(key string, val int) {
	m.userLock.Lock()
	defer m.userLock.Unlock()
	m.userData[key] = val
}

// Get returns value for given key
func (m *Metrics) Get(key string) int {
	m.userLock.RLock()
	defer m.userLock.RUnlock()

	return m.userData[key]
}

// String returns sorted key:vals string representation of metrics
func (m *Metrics) String() string {

	m.userLock.RLock()
	defer m.userLock.RUnlock()

	sortedKeys := func() (res []string) {
		for k := range m.userData {
			res = append(res, k)
		}
		sort.Strings(res)
		return res
	}()

	var udata []string
	for _, k := range sortedKeys {
		udata = append(udata, fmt.Sprintf("%s:%d", k, m.userData[k]))
	}
	um := ""
	if len(udata) > 0 {
		um = fmt.Sprintf("[%s]", strings.Join(udata, ", "))
	}
	return um
}
