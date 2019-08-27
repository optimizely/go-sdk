package event

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInMemoryQueue_Add_Size_Remove(t *testing.T) {
	q := NewInMemoryQueue(5)

	q.Add(1)
	q.Add(2)
	q.Add(3)

	assert.Equal(t,3, q.Size())

	items1 := q.Get(1)

	assert.Equal(t, 1, len(items1))
	assert.Equal(t, 1, items1[0])

	items2 := q.Get(5)

	assert.Equal(t, 3, len(items2))
	assert.Equal(t, 3, items2[2])

	empty := q.Get(0)
	assert.Equal(t, 0, len(empty))

	allItems := q.Remove(3)

	assert.Equal(t, 3, len(allItems))

	assert.Equal(t, 0, q.Size())
}

func TestInMemoryQueue_Concurrent(t *testing.T) {

	q := NewInMemoryQueue(5)

	quit := make(chan int)

	go func() {
		i := 5
		for  i > 0 {
			q.Add(i)
			i--
		}

		quit <- 0
	}()

	go func() {
		i := 5
		for  i > 0 {
			q.Add(i)
			i--
		}

		quit <- 0
	}()

	<- quit

	q.Remove(1)
	q.Remove(1)

	<- quit

	assert.Equal(t, 8, q.Size())
}
