package engine

// Processor defines the interface for any component that transforms or filters logs.
type Processor interface {
	// Process applies logic to the entry.
	// It returns the (potentially modified) entry, a bool indicating if the entry should be DROPPED, and any error.
	// If drop is true, the pipeline stops processing this entry.
	// Implementations MUST try to modify 'entry' in-place or return a slice of it to avoid allocation.
	Process(ctx *ProcessingContext, entry []byte) ([]byte, bool, error)

	// Name returns the identifier of the processor (for metrics/logging).
	Name() string
}
