package event

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNSQQueue_Add_Size_Remove(t *testing.T) {
	q := NewNSQueue(5)

	q.Add(1)
	q.Add(2)
	q.Add(3)

	time.Sleep(2000 * time.Millisecond)

	items1 := q.Get(1)

	assert.Equal(t, 1, len(items1))

	q.Remove(1)

	items2 := q.Get(5)

	assert.True(t, len(items2) != 0)

	allItems := q.Remove(3)

	assert.True(t,len(allItems) > 0)

	assert.Equal(t, 0, q.Size())
}
