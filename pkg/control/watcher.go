package control

import (
	"context"
	"encoding/json"
	"log"
	"streamgate/pkg/engine"
	"streamgate/pkg/output"

	"github.com/redis/go-redis/v9"
)

type Manifest struct {
	Version   string           `json:"version"`
	Pipelines []PipelineConfig `json:"pipelines"`
}

type PipelineConfig struct {
	Name       string          `json:"name"`
	Processors []ProcessorRule `json:"processors"`
	Outputs    []OutputTarget  `json:"outputs"`
	BatchSize  int             `json:"batch_size"`
}

type ProcessorRule struct {
	ID     string            `json:"id"`
	Type   string            `json:"type"`
	Params map[string]string `json:"params"`
}

type OutputTarget struct {
	Type    string            `json:"type"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
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
		case "attribute_filter":
			// Params: attribute OR path, operator, value
			cfg := engine.AttributeFilterConfig{
				Name:      rule.ID,
				Attribute: rule.Params["attribute"],
				Path:      rule.Params["path"],
				Operator:  engine.Operator(rule.Params["operator"]),
				Value:     rule.Params["value"],
			}
			proc, err := engine.NewAttributeFilterProcessor(cfg)
			if err != nil {
				log.Printf("Control: Failed to create attribute_filter %s: %v", rule.ID, err)
				continue
			}
			processors = append(processors, proc)
		}
	}

	newChain := engine.NewProcessorChain(processors...)
	w.pipeline.UpdateChain(newChain)

	// Build Outputs
	// Default to Console if none specified
	var outputs []output.Output
	if len(cfg.Outputs) == 0 {
		outputs = append(outputs, output.NewConsoleOutput())
	} else {
		for _, outCfg := range cfg.Outputs {
			switch outCfg.Type {
			case "console":
				outputs = append(outputs, output.NewConsoleOutput())
			case "http":
				if outCfg.URL != "" {
					outputs = append(outputs, output.NewHTTPOutput(outCfg.URL, outCfg.Headers))
				}
			}
		}
	}

	// Use FanOut manager to handle multiple outputs
	w.pipeline.UpdateOutput(output.NewFanOutOutput(outputs...))

	// Update Batch Size
	// If 0 (omitted), default to 100 inside UpdateBatchSize or handle here.
	// We'll pass it directly, Pipeline handles < 1.
	// But let's respect default 100 if missing.
	bz := int64(cfg.BatchSize)
	if bz == 0 {
		bz = 100
	}
	w.pipeline.UpdateBatchSize(bz)
}
