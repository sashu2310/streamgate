package model

import (
	"time"
)

// LogEntry represents a single log event flowing through the system.
// It is designed to be reused via sync.Pool to minimize allocations.
type LogEntry struct {
	// Timestamp is the time the log was received.
	Timestamp time.Time

	// Raw is the underlying byte slice of the log message.
	// Processors modify this slice in-place if possible.
	Raw []byte

	// Metadata can hold extracted fields (like 'service', 'level') for routing.
	// We use a map but in a high-perf scenario, we might switch to a fixed struct or strict key set.
	Metadata map[string]string
}

// Reset clears the LogEntry for reuse in a sync.Pool.
func (l *LogEntry) Reset() {
	l.Timestamp = time.Time{}
	l.Raw = l.Raw[:0] // Keep capacity
	for k := range l.Metadata {
		delete(l.Metadata, k)
	}
}
