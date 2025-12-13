package engine

import (
	"bytes"
)

// RedactionProcessor replaces occurrences of a target string with a mask.
type RedactionProcessor struct {
	name   string
	target []byte
	mask   []byte
}

func NewRedactionProcessor(name string, target string, mask string) *RedactionProcessor {
	return &RedactionProcessor{
		name:   name,
		target: []byte(target),
		mask:   []byte(mask),
	}
}

func (r *RedactionProcessor) Name() string {
	return r.name
}

func (r *RedactionProcessor) Process(ctx *ProcessingContext, entry []byte) ([]byte, bool, error) {
	if bytes.Contains(entry, r.target) {
		// bytes.Replace allocates a new slice if changes are made.
		// To be strictly zero-alloc, we'd need a mutable buffer or in-place replacement if lengths match.
		// For now, we accept the allocation of Replace for simplicity in V1, 
		// but note it as a candidate for optimization (allocating a new buffer is safer than in-place if size changes).
		return bytes.ReplaceAll(entry, r.target, r.mask), false, nil
	}
	return entry, false, nil
}
