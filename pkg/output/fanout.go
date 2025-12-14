package output

import (
	"sync"
)

// FanOutOutput writes to multiple outputs in parallel.
type FanOutOutput struct {
	outputs []Output
}

func NewFanOutOutput(outputs ...Output) *FanOutOutput {
	return &FanOutOutput{
		outputs: outputs,
	}
}

func (f *FanOutOutput) WriteBatch(entries [][]byte) error {
	var wg sync.WaitGroup
	errs := make([]error, len(f.outputs))

	for i, out := range f.outputs {
		wg.Add(1)
		go func(idx int, o Output) {
			defer wg.Done()
			if err := o.WriteBatch(entries); err != nil {
				errs[idx] = err
			}
		}(i, out)
	}
	wg.Wait()

	// For simple error handling, return the first error found
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}
