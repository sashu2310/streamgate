package engine

import (
	"errors"
	"sync/atomic"
)

var (
	ErrBufferFull = errors.New("buffer is full")
)

// RingBuffer is a fixed-size circular buffer for byte slices.
// It is safe for a single writer (Ingestor) and single reader (Processor).
// For multiple writers, we would need a mutex or a channel-based approach.
type RingBuffer struct {
	data [][]byte
	head uint64
	tail uint64
	mask uint64
	size uint64

	// Metrics
	dropped uint64
}

// NewRingBuffer creates a ring buffer with the specified size (must be power of 2).
func NewRingBuffer(size uint64) (*RingBuffer, error) {
	if size == 0 || (size&(size-1)) != 0 {
		return nil, errors.New("size must be a power of 2")
	}
	return &RingBuffer{
		data: make([][]byte, size),
		mask: size - 1,
		size: size,
	}, nil
}

// Push adds an item to the buffer.
// If the buffer is full, it drops the item and returns ErrBufferFull.
func (rb *RingBuffer) Push(item []byte) error {
	head := atomic.LoadUint64(&rb.head)
	tail := atomic.LoadUint64(&rb.tail)

	if head-tail >= rb.size {
		atomic.AddUint64(&rb.dropped, 1)
		return ErrBufferFull
	}

	rb.data[head&rb.mask] = item
	atomic.StoreUint64(&rb.head, head+1)
	return nil
}

// Pop removes an item from the buffer.
// Returns nil if empty.
func (rb *RingBuffer) Pop() []byte {
	tail := rb.tail
	head := atomic.LoadUint64(&rb.head)

	if tail == head {
		return nil
	}

	item := rb.data[tail&rb.mask]
	// Help GC? In our zero-alloc design, we might want to recycle this buffer back to a pool later.
	// rb.data[tail&rb.mask] = nil 
	
	atomic.StoreUint64(&rb.tail, tail+1)
	return item
}

// DroppedCount returns the number of dropped events.
func (rb *RingBuffer) DroppedCount() uint64 {
	return atomic.LoadUint64(&rb.dropped)
}

// Usage returns the number of items currently in the buffer.
func (rb *RingBuffer) Usage() uint64 {
	return atomic.LoadUint64(&rb.head) - atomic.LoadUint64(&rb.tail)
}

// Capacity returns the total size of the buffer.
func (rb *RingBuffer) Capacity() uint64 {
	return rb.size
}
