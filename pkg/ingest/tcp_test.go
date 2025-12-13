package ingest

import (
	"fmt"
	"net"
	"streamgate/pkg/engine"
	"testing"
	"time"
)

func TestTCPIngestor_Integration(t *testing.T) {
	// 1. Setup Buffer
	rb, _ := engine.NewRingBuffer(1024)

	// 2. Start Ingestor on random port
	port := 54321 // Likely free
	addr := fmt.Sprintf("localhost:%d", port)
	ingestor := NewTCPIngestor(addr, rb)

	// Start in background
	go func() {
		if err := ingestor.Start(); err != nil {
			// This might fail if test runs twice fast, but good enough for now
			t.Logf("Ingestor stopped: %v", err)
		}
	}()

	// Give it a moment to bind
	time.Sleep(100 * time.Millisecond)

	// 3. Connect and Send Data
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to connect to ingestor: %v", err)
	}
	defer conn.Close()

	msg := "hello streamgate\n"
	_, err = conn.Write([]byte(msg))
	if err != nil {
		t.Fatalf("Failed to write to TCP: %v", err)
	}

	// 4. Verification (Wait for processing)
	// Poll buffer
	timeout := time.After(1 * time.Second)
	found := false
	for {
		select {
		case <-timeout:
			t.Fatal("Timed out waiting for message in buffer")
		default:
			item := rb.Pop()
			if item != nil {
				if string(item) == msg {
					found = true
					goto DONE
				} else {
					// Might receive partial due to OS buffering, but our code assumes ReadBytes('\n')
					// So it should match exact line
					t.Logf("Got unexpected data: %s", string(item))
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
DONE:
	if !found {
		t.Error("Did not find expected message")
	}
}
