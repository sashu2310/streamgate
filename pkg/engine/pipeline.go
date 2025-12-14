package engine

import (
	"context"
	"log"
	"streamgate/pkg/output"
	"sync/atomic"
	"time"
)

// Pipeline connects the Ingest Buffer -> ProcessorChain -> Output.
type Pipeline struct {
	buffer *RingBuffer
	chain  atomic.Pointer[ProcessorChain] // Hot-swappable chain
	output atomic.Value                   // Hot-swappable output (stores output.Output)

	// Config
	batchSize int
	workers   int
}

func NewPipeline(buf *RingBuffer, chain *ProcessorChain, out output.Output) *Pipeline {
	p := &Pipeline{
		buffer:    buf,
		batchSize: 100, // naive batching
		workers:   1,   // single consumer for now to ensure strict ordering if needed
	}
	p.chain.Store(chain)
	
	// CRITICAL FIX: atomic.Value must always store the same concrete type!
	// We will always store *output.FanOutOutput. 
	// Even if 'out' is a single ConsoleOutput, we wrap it.
	// Check if it's already FanOut to avoid double wrapping (optional, but safer to just wrap)
	fanOut := output.NewFanOutOutput(out)
	p.output.Store(fanOut)
	
	return p
}

// UpdateChain hot-swaps the processor chain safely.
func (p *Pipeline) UpdateChain(chain *ProcessorChain) {
	p.chain.Store(chain)
	log.Println("Pipeline: Processor chain hot-swapped.")
}

// UpdateOutput hot-swaps the output provider safely.
// Note: We expect 'out' to effectively be a *FanOutOutput for consistency,
// or we wrap it here if we want to be generous.
func (p *Pipeline) UpdateOutput(out output.Output) {
	// Ensure type consistency for atomic.Value
	if _, ok := out.(*output.FanOutOutput); !ok {
		// Wrap it if it's not already a FanOut
		out = output.NewFanOutOutput(out)
	}
	p.output.Store(out)
	log.Println("Pipeline: Output provider hot-swapped.")
}

func (p *Pipeline) Start(ctx context.Context) {
	log.Println("Starting Processing Pipeline...")
	for i := 0; i < p.workers; i++ {
		go p.worker(ctx)
	}
}

func (p *Pipeline) worker(ctx context.Context) {
	// Reusable batch slice
	batch := make([][]byte, 0, p.batchSize)
	pCtx := &ProcessingContext{Context: ctx}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	flush := func() {
		if len(batch) > 0 {
			// Load current output safely
			currentOutput := p.output.Load().(output.Output)
			if err := currentOutput.WriteBatch(batch); err != nil {
				log.Printf("Output error: %v", err)
			}
			// Reset batch slice (keep capacity)
			batch = batch[:0]
		}
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case <-ticker.C:
			flush()
		default:
			// 1. Pop from Buffer
			item := p.buffer.Pop()
			if item == nil {
				// Buffer empty, tiny sleep to save CPU?
				// Or use a Cond/Signal (better).
				// For V1 spin-wait with sleep is okay-ish but suboptimal.
				time.Sleep(1 * time.Millisecond) // TODO: Replace with sync.Cond
				continue
			}

			// 2. Fail-Open Check (Circuit Breaker)
			// If buffer is > 80% full, bypass processing to drain quicker.
			usage := p.buffer.Usage()
			capacity := p.buffer.Capacity()

			if float64(usage) > float64(capacity)*0.80 {
				// Bypass Mode!
				batch = append(batch, item)
			} else {
				// Normal Mode
				// Load current chain safely
				currentChain := p.chain.Load()
				processed, drop, err := currentChain.Process(pCtx, item)
				if err != nil {
					log.Printf("Process error: %v", err)
					continue
				}
				if drop {
					continue
				}
				batch = append(batch, processed)
			}

			// 4. Add to Batch (handled above)
			if len(batch) >= p.batchSize {
				flush()
			}
		}
	}
}
