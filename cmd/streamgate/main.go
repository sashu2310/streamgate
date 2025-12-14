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
	"streamgate/pkg/control"
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

	// 3. Processors (Dynamic)
	// Start with an empty chain. The Watcher will update it.
	chain := engine.NewProcessorChain()

	// 4. Output
	out := output.NewConsoleOutput()

	// 5. Ingestors
	tcpAddr := fmt.Sprintf(":%d", cfg.Server.TCPPort)
	tcpIngestor := ingest.NewTCPIngestor(tcpAddr, buffer)

	udpAddr := fmt.Sprintf(":%d", cfg.Server.UDPPort)
	udpIngestor := ingest.NewUDPIngestor(udpAddr, buffer)

	// 6. Pipeline
	pipeline := engine.NewPipeline(buffer, chain, out)

	// 7. Control Plane Watcher
	// Use Redis address from config
	watcher := control.NewWatcher(cfg.Redis.Address, pipeline)

	// --- Start ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Pipeline (Consumer)
	pipeline.Start(ctx)

	// Start Watcher
	go watcher.Start(ctx)

	// Start Ingestors (Producers)
	go func() {
		if err := tcpIngestor.Start(); err != nil {
			log.Fatalf("TCP Ingestor died: %v", err)
		}
	}()

	go func() {
		if err := udpIngestor.Start(); err != nil {
			log.Fatalf("UDP Ingestor died: %v", err)
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
