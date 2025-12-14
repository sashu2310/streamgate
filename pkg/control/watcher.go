package control

import (
	"context"
	"encoding/json"
	"log"
	"streamgate/pkg/engine"

	"github.com/redis/go-redis/v9"
)

type Manifest struct {
	Version   string           `json:"version"`
	Pipelines []PipelineConfig `json:"pipelines"`
}

type PipelineConfig struct {
	Name       string          `json:"name"`
	Processors []ProcessorRule `json:"processors"`
}

type ProcessorRule struct {
	ID     string            `json:"id"`
	Type   string            `json:"type"`
	Params map[string]string `json:"params"`
}

type Watcher struct {
	redisClient *redis.Client
	pipeline    *engine.Pipeline
}

func NewWatcher(addr string, pipeline *engine.Pipeline) *Watcher {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &Watcher{
		redisClient: rdb,
		pipeline:    pipeline,
	}
}

func (w *Watcher) Start(ctx context.Context) {
	log.Println("Control: Starting Config Watcher...")

	// 1. Initial Load
	w.reload()

	// 2. Subscribe to updates
	pubsub := w.redisClient.Subscribe(ctx, "streamgate_updates")
	ch := pubsub.Channel()

	go func() {
		defer pubsub.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				log.Printf("Control: Received update signal: %s", msg.Payload)
				w.reload()
			}
		}
	}()
}

func (w *Watcher) reload() {
	ctx := context.Background()
	val, err := w.redisClient.Get(ctx, "streamgate_config").Result()
	if err == redis.Nil {
		log.Println("Control: No config found in Redis. Keeping current state.")
		return
	} else if err != nil {
		log.Printf("Control: Failed to fetch config: %v", err)
		return
	}

	var manifest Manifest
	if err := json.Unmarshal([]byte(val), &manifest); err != nil {
		log.Printf("Control: Invalid config JSON: %v", err)
		return
	}

	// For Prototype: We only support one pipeline named "default_pipeline" or the first one
	if len(manifest.Pipelines) == 0 {
		return
	}
	cfg := manifest.Pipelines[0]

	// Build Chain
	var processors []engine.Processor
	for _, rule := range cfg.Processors {
		switch rule.Type {
		case "filter":
			// Params: key, value
			// In V1 filter implementation, we might only support 'contains' on the whole body or simple checks.
			// The current engine.FilterProcessor logic checks if the body contains any of the keywords.
			if val, ok := rule.Params["value"]; ok {
				processors = append(processors, engine.NewFilterProcessor(rule.ID, []string{val}))
			}
		case "redact":
			// Params: pattern, replacement
			pat := rule.Params["pattern"]
			rep := rule.Params["replacement"]
			if pat != "" && rep != "" {
				processors = append(processors, engine.NewRedactionProcessor(rule.ID, pat, rep))
			}
		}
	}

	newChain := engine.NewProcessorChain(processors...)
	w.pipeline.UpdateChain(newChain)
}
