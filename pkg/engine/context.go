package engine

import (
	"context"
)

// ProcessingContext holds per-request state.
// It wraps standard context.Context and adds high-performance helpers.
type ProcessingContext struct {
	context.Context
	// Add other field as needed, e.g., TraceID, UserID
}
