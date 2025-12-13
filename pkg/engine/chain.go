package engine

// ProcessorChain manages a sequential list of processors.
type ProcessorChain struct {
	processors []Processor
}

// NewProcessorChain creates a chain with the given list of processors.
func NewProcessorChain(processors ...Processor) *ProcessorChain {
	return &ProcessorChain{
		processors: processors,
	}
}

// Process runs the entry through all processors in the chain.
// It stops if a processor returns drop=true or an error.
func (c *ProcessorChain) Process(ctx *ProcessingContext, entry []byte) ([]byte, bool, error) {
	var drop bool
	var err error

	for _, p := range c.processors {
		entry, drop, err = p.Process(ctx, entry)
		if err != nil {
			return entry, false, err
		}
		if drop {
			return entry, true, nil
		}
	}

	return entry, false, nil
}
