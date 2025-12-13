package output

import (
	"fmt"
	"os"
)

// Output defines where the processed logs go.
type Output interface {
	WriteBatch(entries [][]byte) error
}

// ConsoleOutput writes logs to stdout.
type ConsoleOutput struct{}

func NewConsoleOutput() *ConsoleOutput {
	return &ConsoleOutput{}
}

func (c *ConsoleOutput) WriteBatch(entries [][]byte) error {
	for _, entry := range entries {
		// In production, we'd use a buffered writer or similar.
		// For console, fmt.Println is fine.
		_, err := fmt.Fprintf(os.Stdout, "%s", entry)
		if err != nil {
			return err
		}
	}
	return nil
}
