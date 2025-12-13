package engine

import (
	"context"
	"log"
	"streamgate/pkg/output"
	"time"
)

// Pipeline connects the Ingest Buffer -> ProcessorChain -> Output.
type Pipeline struct {
	buffer *RingBuffer
	chain  *ProcessorChain
	output output.Output

	// Config
	batchSize int
	workers   int
}

func NewPipeline(buf *RingBuffer, chain *ProcessorChain, out output.Output) *Pipeline {
	return &Pipeline{
		buffer:    buf,
		chain:     chain,
		output:    out,
		batchSize: 100, // naive batching
		workers:   1,   // single consumer for now to ensure strict ordering if needed
	}
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
			if err := p.output.WriteBatch(batch); err != nil {
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
				processed, drop, err := p.chain.Process(pCtx, item)
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
