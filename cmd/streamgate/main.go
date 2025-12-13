package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"streamgate/pkg/config"
	"streamgate/pkg/engine"
	"streamgate/pkg/ingest"
	"streamgate/pkg/output"
)

func main() {
	log.Println("Initializing StreamGate...")

	// 1. Config
	cfg := config.DefaultConfig()

	// 2. Buffer (Size 65536)
	buffer, err := engine.NewRingBuffer(65536)
	if err != nil {
		log.Fatalf("Failed to create buffer: %v", err)
	}

	// 3. Processors
	chain := engine.NewProcessorChain(
		engine.NewFilterProcessor("drop_debug", []string{"DEBUG"}),
		engine.NewRedactionProcessor("redact_cc", "4111-1234", "xxxx-xxxx"),
	)

	// 4. Output
	out := output.NewConsoleOutput()

	// 5. Ingestor (TCP)
	addr := fmt.Sprintf(":%d", cfg.Server.TCPPort)
	ingestor := ingest.NewTCPIngestor(addr, buffer)

	// 6. Pipeline
	pipeline := engine.NewPipeline(buffer, chain, out)

	// --- Start ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Pipeline (Consumer)
	pipeline.Start(ctx)

	// Start Ingestor (Producer) - Blocking, so run in goroutine
	go func() {
		if err := ingestor.Start(); err != nil {
			log.Fatalf("Ingestor died: %v", err)
		}
	}()

	// Wait for shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	log.Println("StreamGate running. Press Ctrl+C to stop.")
	
	<-sigChan
	log.Println("Shutting down...")
	cancel()
	time.Sleep(1 * time.Second) // Give workers time to flush
	log.Println("Bye.")
}
