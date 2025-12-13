package engine

import (
	"bytes"
)

// FilterProcessor drops logs that DO NOT match the criteria (AllowList)
// or drops logs that DO match (BlockList).
// For V1, let's implement a BlockList (Drop if contains X).
type FilterProcessor struct {
	name      string
	blockBytes [][]byte // Pre-converted to bytes for zero-alloc comparison
}

func NewFilterProcessor(name string, blockWords []string) *FilterProcessor {
	bb := make([][]byte, len(blockWords))
	for i, w := range blockWords {
		bb[i] = []byte(w)
	}
	return &FilterProcessor{
		name:       name,
		blockBytes: bb,
	}
}

func (f *FilterProcessor) Name() string {
	return f.name
}

func (f *FilterProcessor) Process(ctx *ProcessingContext, entry []byte) ([]byte, bool, error) {
	// Naive O(N*M) check. 
	// Optimization: Aho-Corasick for many patterns.
	for _, word := range f.blockBytes {
		if bytes.Contains(entry, word) {
			return entry, true, nil // DROP
		}
	}
	return entry, false, nil
}
