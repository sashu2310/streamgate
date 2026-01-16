package ingest

import (
	"bufio"
	"io"
	"log"
	"net"
	"streamgate/pkg/engine"
)

// TCPIngestor listens for TCP connections and pushes logs to the buffer.
type TCPIngestor struct {
	addr   string
	buffer *engine.RingBuffer
}

func NewTCPIngestor(addr string, buffer *engine.RingBuffer) *TCPIngestor {
	return &TCPIngestor{
		addr:   addr,
		buffer: buffer,
	}
}

// Start begins listening on the TCP address. Blocking call.
func (t *TCPIngestor) Start() error {
	listener, err := net.Listen("tcp", t.addr)
	if err != nil {
		return err
	}
	log.Printf("TCP Ingestor listening on %s", t.addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		// Handle each connection in a lightweight goroutine
		go t.handleConnection(conn)
	}
}

func (t *TCPIngestor) handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		// ReadLine is lower level than ReadString, avoids some allocations but be careful with line size.
		// For simplicity in V1, we use ReadBytes('\n').
		// In a real high-perf scenario, we'd use a sync.Pool for the buffer and io.Read.
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("Read error: %v", err)
			}
			return
		}

		// Trim newline if necessary, or keep it.
		// Push to buffer. On buffer full, silently drop (tail drop strategy).
		// Logging every drop would kill performance.
		_ = t.buffer.Push(line)
	}
}
