package ingest

import (
	"log"
	"net"
	"streamgate/pkg/engine"
)

// UDPIngestor listens for UDP packets and pushes logs to the buffer.
type UDPIngestor struct {
	addr   string
	buffer *engine.RingBuffer
}

func NewUDPIngestor(addr string, buffer *engine.RingBuffer) *UDPIngestor {
	return &UDPIngestor{
		addr:   addr,
		buffer: buffer,
	}
}

// Start begins listening on the UDP address. Blocking call.
func (u *UDPIngestor) Start() error {
	addr, err := net.ResolveUDPAddr("udp", u.addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	log.Printf("UDP Ingestor listening on %s", u.addr)

	// Reuse a buffer for reading packets to minimize allocations.
	// In high-perf, we might have a pool of these buffers or multiple reader goroutines.
	// Max UDP packet size is usually 65535, but typical MTU is 1500.
	buf := make([]byte, 65535)

	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("UDP Read error: %v", err)
			continue
		}

		// Copy the data because 'buf' is reused in the next loop iteration.
		// This allocation is unavoidable unless we use a "Buffer Pool" pattern where we pass
		// the pool-owned buffer to the RingBuffer and get a new one.
		// For V1, a simple copy is acceptable.
		packet := make([]byte, n)
		copy(packet, buf[:n])

		// On buffer full, silently drop (tail drop strategy).
		_ = u.buffer.Push(packet)
	}
}
