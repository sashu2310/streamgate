package engine

import (
	"context"
	"testing"
)

func BenchmarkPipeline_NoAlloc(b *testing.B) {
	// Setup
	chain := NewProcessorChain(
		NewFilterProcessor("drop_debug", []string{"DEBUG"}),
		// Redaction *does* alloc if it finds a match, so we test the "happy path" (no sensitive data) first:
		NewRedactionProcessor("redact_cc", "4111-1234", "xxxx"),
	)
	
	ctx := &ProcessingContext{Context: context.Background()}
	data := []byte("INFO: User login successful for ID 9999")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// We expect 0 allocs here because filter doesn't match and redact doesn't match
		_, _, _ = chain.Process(ctx, data)
	}
}

func BenchmarkPipeline_WithRedaction(b *testing.B) {
	chain := NewProcessorChain(
		NewRedactionProcessor("redact_secret", "SECRET", "XXXXXX"),
	)
	ctx := &ProcessingContext{Context: context.Background()}
	data := []byte("This log contains a SECRET value")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// This WILL alloc because bytes.ReplaceAll generates a new slice
		_, _, _ = chain.Process(ctx, data)
	}
}
