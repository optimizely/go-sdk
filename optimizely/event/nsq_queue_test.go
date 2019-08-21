package event

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNSQQueue_Add_Size_Remove(t *testing.T) {
	q := NewNSQueue(5)

	impression := BuildTestImpressionEvent()
	conversion := BuildTestConversionEvent()

	q.Add(impression)
	q.Add(impression)
	q.Add(conversion)

	time.Sleep(2000 * time.Millisecond)

	items1 := q.Get(2)

	assert.Equal(t, 2, len(items1))

	q.Remove(1)

	items2 := q.Get(1)

	assert.True(t, len(items2) != 0)

	allItems := q.Remove(3)

	assert.True(t,len(allItems) > 0)

	assert.Equal(t, 0, q.Size())
}
