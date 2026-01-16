package engine

import (
	"bytes"
	"testing"
)

func TestRingBuffer_NormalOperation(t *testing.T) {
	// Size 4 (must be power of 2)
	rb, err := NewRingBuffer(4)
	if err != nil {
		t.Fatalf("Failed to create buffer: %v", err)
	}

	data1 := []byte("msg1")
	data2 := []byte("msg2")

	if err := rb.Push(data1); err != nil {
		t.Errorf("Push failed: %v", err)
	}
	if err := rb.Push(data2); err != nil {
		t.Errorf("Push failed: %v", err)
	}

	out1 := rb.Pop()
	if !bytes.Equal(out1, data1) {
		t.Errorf("Expected %s, got %s", data1, out1)
	}
	out2 := rb.Pop()
	if !bytes.Equal(out2, data2) {
		t.Errorf("Expected %s, got %s", data2, out2)
	}

	if out3 := rb.Pop(); out3 != nil {
		t.Errorf("Expected nil (empty), got %s", out3)
	}
}

func TestRingBuffer_FullDrop(t *testing.T) {
	// Small buffer to test overflow easily
	rb, _ := NewRingBuffer(2)

	_ = rb.Push([]byte("1"))
	_ = rb.Push([]byte("2"))

	// Third push should fail (Buffer Full)
	err := rb.Push([]byte("3"))
	if err != ErrBufferFull {
		t.Errorf("Expected ErrBufferFull, got %v", err)
	}

	if dropped := rb.DroppedCount(); dropped != 1 {
		t.Errorf("Expected 1 dropped item, got %d", dropped)
	}

	// Should still read 1 and 2
	if string(rb.Pop()) != "1" {
		t.Error("Order corrupted")
	}
	if string(rb.Pop()) != "2" {
		t.Error("Order corrupted")
	}
}
