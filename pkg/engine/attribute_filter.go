package engine

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tidwall/gjson"
)

// Operator defines the comparison operation for attribute filtering
type Operator string

const (
	OpEquals   Operator = "equals"
	OpContains Operator = "contains"
	OpRegex    Operator = "regex"
)

// otelSearchPaths defines common locations for well-known OTel attributes.
// When user specifies an "attribute" (not explicit "path"), we search these locations.
var otelSearchPaths = map[string][]string{
	// Service identification
	"service.name": {
		"service.name",
		"resource.attributes.service\\.name",
		"resourceAttributes.service\\.name",
		"resource.service\\.name",
	},
	"service.namespace": {
		"service.namespace",
		"resource.attributes.service\\.namespace",
		"resourceAttributes.service\\.namespace",
	},
	"service.version": {
		"service.version",
		"resource.attributes.service\\.version",
		"resourceAttributes.service\\.version",
	},

	// Deployment
	"deployment.environment": {
		"deployment.environment",
		"resource.attributes.deployment\\.environment",
		"resourceAttributes.deployment\\.environment",
	},

	// HTTP attributes
	"http.status_code": {
		"http.status_code",
		"attributes.http\\.status_code",
		"http\\.status_code",
	},
	"http.method": {
		"http.method",
		"attributes.http\\.method",
		"http\\.method",
	},
	"http.url": {
		"http.url",
		"attributes.http\\.url",
		"http\\.url",
	},
	"http.target": {
		"http.target",
		"attributes.http\\.target",
		"http\\.target",
	},

	// Logging
	"log.level": {
		"log.level",
		"severity",
		"severityText",
		"level",
	},
}

// genericSearchPaths are tried for any attribute not in otelSearchPaths
var genericSearchPaths = []string{
	"%s",                          // top-level as-is
	"attributes.%s",               // OTel log attributes
	"resource.attributes.%s",      // OTel resource attributes
	"resourceAttributes.%s",       // flattened resource attributes
	"body.%s",                     // inside body
}

// AttributeFilterProcessor drops logs based on JSON attribute values.
// Supports both well-known OTel attributes (auto-search) and explicit paths.
type AttributeFilterProcessor struct {
	name     string
	attr     string         // well-known attribute name (auto-search mode)
	path     string         // explicit gjson path (explicit mode)
	operator Operator
	value    string
	regex    *regexp.Regexp // compiled regex if operator is OpRegex
}

// AttributeFilterConfig holds configuration for creating an AttributeFilterProcessor
type AttributeFilterConfig struct {
	Name      string
	Attribute string   // use this for well-known OTel attributes (auto-search)
	Path      string   // use this for explicit gjson path
	Operator  Operator
	Value     string
}

// NewAttributeFilterProcessor creates a new attribute filter processor.
// Either Attribute or Path must be specified, not both.
func NewAttributeFilterProcessor(cfg AttributeFilterConfig) (*AttributeFilterProcessor, error) {
	if cfg.Attribute == "" && cfg.Path == "" {
		return nil, fmt.Errorf("either attribute or path must be specified")
	}
	if cfg.Attribute != "" && cfg.Path != "" {
		return nil, fmt.Errorf("cannot specify both attribute and path")
	}

	p := &AttributeFilterProcessor{
		name:     cfg.Name,
		attr:     cfg.Attribute,
		path:     cfg.Path,
		operator: cfg.Operator,
		value:    cfg.Value,
	}

	// Pre-compile regex if needed
	if cfg.Operator == OpRegex {
		re, err := regexp.Compile(cfg.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
		p.regex = re
	}

	// Default operator
	if p.operator == "" {
		p.operator = OpEquals
	}

	return p, nil
}

func (p *AttributeFilterProcessor) Name() string {
	return p.name
}

// Process checks if the log entry matches the filter criteria.
// Returns (entry, drop=true, nil) if the attribute matches and log should be dropped.
func (p *AttributeFilterProcessor) Process(ctx *ProcessingContext, entry []byte) ([]byte, bool, error) {
	// Fail-open: if entry is not valid JSON, pass through
	if !gjson.ValidBytes(entry) {
		return entry, false, nil
	}

	var value gjson.Result

	if p.path != "" {
		// Explicit path mode - convert user path (using /) to gjson path
		gjsonPath := convertToGjsonPath(p.path)
		value = gjson.GetBytes(entry, gjsonPath)
	} else {
		// Auto-search mode - try well-known paths first, then generic
		value = p.searchAttribute(entry)
	}

	// Attribute not found - pass through (fail-open)
	if !value.Exists() {
		return entry, false, nil
	}

	// Check if value matches based on operator
	matched := p.matchValue(value)

	return entry, matched, nil
}

// searchAttribute looks for the attribute in well-known OTel paths,
// falling back to generic search paths.
func (p *AttributeFilterProcessor) searchAttribute(entry []byte) gjson.Result {
	// First, try well-known paths for this attribute
	if paths, ok := otelSearchPaths[p.attr]; ok {
		for _, path := range paths {
			result := gjson.GetBytes(entry, path)
			if result.Exists() {
				return result
			}
		}
	}

	// Fall back to generic search paths
	// Escape dots in attribute name for gjson
	escapedAttr := strings.ReplaceAll(p.attr, ".", "\\.")
	for _, pathTemplate := range genericSearchPaths {
		path := fmt.Sprintf(pathTemplate, escapedAttr)
		result := gjson.GetBytes(entry, path)
		if result.Exists() {
			return result
		}
	}

	return gjson.Result{} // not found
}

// matchValue checks if the gjson result matches based on the operator
func (p *AttributeFilterProcessor) matchValue(value gjson.Result) bool {
	strValue := value.String()

	switch p.operator {
	case OpEquals:
		// Type-aware comparison for numbers
		if value.Type == gjson.Number {
			return strValue == p.value
		}
		return strValue == p.value

	case OpContains:
		return strings.Contains(strValue, p.value)

	case OpRegex:
		if p.regex == nil {
			return false
		}
		return p.regex.MatchString(strValue)

	default:
		return false
	}
}

// convertToGjsonPath converts user-friendly path (using /) to gjson path.
// Example: "resource/attributes/service.name" -> "resource.attributes.service\.name"
func convertToGjsonPath(userPath string) string {
	parts := strings.Split(userPath, "/")
	for i, part := range parts {
		// Escape dots within each part (they're literal key names)
		parts[i] = strings.ReplaceAll(part, ".", "\\.")
	}
	return strings.Join(parts, ".")
}
