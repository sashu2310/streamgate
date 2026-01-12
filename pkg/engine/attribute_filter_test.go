package engine

import (
	"testing"
)

func TestAttributeFilter_EqualsOperator(t *testing.T) {
	proc, err := NewAttributeFilterProcessor(AttributeFilterConfig{
		Name:      "test",
		Attribute: "service.name",
		Operator:  OpEquals,
		Value:     "test-service",
	})
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		wantDrop bool
	}{
		{
			name:     "exact match - drop",
			input:    `{"service.name": "test-service", "message": "hello"}`,
			wantDrop: true,
		},
		{
			name:     "no match - pass",
			input:    `{"service.name": "prod-service", "message": "hello"}`,
			wantDrop: false,
		},
		{
			name:     "partial match - pass (equals is exact)",
			input:    `{"service.name": "test-service-v2", "message": "hello"}`,
			wantDrop: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, drop, err := proc.Process(nil, []byte(tt.input))
			if err != nil {
				t.Errorf("Process() error = %v", err)
			}
			if drop != tt.wantDrop {
				t.Errorf("Process() drop = %v, want %v", drop, tt.wantDrop)
			}
		})
	}
}

func TestAttributeFilter_ContainsOperator(t *testing.T) {
	proc, err := NewAttributeFilterProcessor(AttributeFilterConfig{
		Name:      "test",
		Attribute: "service.name",
		Operator:  OpContains,
		Value:     "test",
	})
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		wantDrop bool
	}{
		{
			name:     "contains match - drop",
			input:    `{"service.name": "test-service", "message": "hello"}`,
			wantDrop: true,
		},
		{
			name:     "contains match v2 - drop",
			input:    `{"service.name": "my-test-app", "message": "hello"}`,
			wantDrop: true,
		},
		{
			name:     "no match - pass",
			input:    `{"service.name": "prod-service", "message": "hello"}`,
			wantDrop: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, drop, err := proc.Process(nil, []byte(tt.input))
			if err != nil {
				t.Errorf("Process() error = %v", err)
			}
			if drop != tt.wantDrop {
				t.Errorf("Process() drop = %v, want %v", drop, tt.wantDrop)
			}
		})
	}
}

func TestAttributeFilter_RegexOperator(t *testing.T) {
	proc, err := NewAttributeFilterProcessor(AttributeFilterConfig{
		Name:      "test",
		Attribute: "service.name",
		Operator:  OpRegex,
		Value:     "^test-.*$",
	})
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		wantDrop bool
	}{
		{
			name:     "regex match - drop",
			input:    `{"service.name": "test-service", "message": "hello"}`,
			wantDrop: true,
		},
		{
			name:     "regex match v2 - drop",
			input:    `{"service.name": "test-api", "message": "hello"}`,
			wantDrop: true,
		},
		{
			name:     "no match - pass",
			input:    `{"service.name": "prod-service", "message": "hello"}`,
			wantDrop: false,
		},
		{
			name:     "partial no match - pass",
			input:    `{"service.name": "my-test-service", "message": "hello"}`,
			wantDrop: false, // doesn't start with test-
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, drop, err := proc.Process(nil, []byte(tt.input))
			if err != nil {
				t.Errorf("Process() error = %v", err)
			}
			if drop != tt.wantDrop {
				t.Errorf("Process() drop = %v, want %v", drop, tt.wantDrop)
			}
		})
	}
}

func TestAttributeFilter_OTelNestedPaths(t *testing.T) {
	proc, err := NewAttributeFilterProcessor(AttributeFilterConfig{
		Name:      "test",
		Attribute: "service.name",
		Operator:  OpEquals,
		Value:     "auth-service",
	})
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		wantDrop bool
	}{
		{
			name:     "top-level attribute",
			input:    `{"service.name": "auth-service", "message": "hello"}`,
			wantDrop: true,
		},
		{
			name:     "nested in resource.attributes",
			input:    `{"resource": {"attributes": {"service.name": "auth-service"}}, "body": "hello"}`,
			wantDrop: true,
		},
		{
			name:     "nested in resourceAttributes (flattened)",
			input:    `{"resourceAttributes": {"service.name": "auth-service"}, "body": "hello"}`,
			wantDrop: true,
		},
		{
			name:     "attribute not found - pass through",
			input:    `{"other.field": "value", "body": "hello"}`,
			wantDrop: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, drop, err := proc.Process(nil, []byte(tt.input))
			if err != nil {
				t.Errorf("Process() error = %v", err)
			}
			if drop != tt.wantDrop {
				t.Errorf("Process() drop = %v, want %v", drop, tt.wantDrop)
			}
		})
	}
}

func TestAttributeFilter_ExplicitPath(t *testing.T) {
	proc, err := NewAttributeFilterProcessor(AttributeFilterConfig{
		Name:     "test",
		Path:     "metadata/labels/app.name",
		Operator: OpEquals,
		Value:    "my-app",
	})
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		wantDrop bool
	}{
		{
			name:     "explicit path match - drop",
			input:    `{"metadata": {"labels": {"app.name": "my-app"}}, "message": "hello"}`,
			wantDrop: true,
		},
		{
			name:     "explicit path no match - pass",
			input:    `{"metadata": {"labels": {"app.name": "other-app"}}, "message": "hello"}`,
			wantDrop: false,
		},
		{
			name:     "path not found - pass",
			input:    `{"other": "data"}`,
			wantDrop: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, drop, err := proc.Process(nil, []byte(tt.input))
			if err != nil {
				t.Errorf("Process() error = %v", err)
			}
			if drop != tt.wantDrop {
				t.Errorf("Process() drop = %v, want %v", drop, tt.wantDrop)
			}
		})
	}
}

func TestAttributeFilter_HTTPStatusCode(t *testing.T) {
	proc, err := NewAttributeFilterProcessor(AttributeFilterConfig{
		Name:      "test",
		Attribute: "http.status_code",
		Operator:  OpEquals,
		Value:     "200",
	})
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		wantDrop bool
	}{
		{
			name:     "numeric value match - drop",
			input:    `{"http.status_code": 200, "message": "ok"}`,
			wantDrop: true,
		},
		{
			name:     "string value match - drop",
			input:    `{"http.status_code": "200", "message": "ok"}`,
			wantDrop: true,
		},
		{
			name:     "no match - pass",
			input:    `{"http.status_code": 500, "message": "error"}`,
			wantDrop: false,
		},
		{
			name:     "nested in attributes - drop",
			input:    `{"attributes": {"http.status_code": 200}, "body": "ok"}`,
			wantDrop: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, drop, err := proc.Process(nil, []byte(tt.input))
			if err != nil {
				t.Errorf("Process() error = %v", err)
			}
			if drop != tt.wantDrop {
				t.Errorf("Process() drop = %v, want %v", drop, tt.wantDrop)
			}
		})
	}
}

func TestAttributeFilter_NonJSONInput(t *testing.T) {
	proc, err := NewAttributeFilterProcessor(AttributeFilterConfig{
		Name:      "test",
		Attribute: "service.name",
		Operator:  OpEquals,
		Value:     "test",
	})
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	// Non-JSON input should pass through (fail-open)
	input := []byte("This is plain text, not JSON")
	_, drop, err := proc.Process(nil, input)
	if err != nil {
		t.Errorf("Process() error = %v", err)
	}
	if drop {
		t.Error("Expected non-JSON input to pass through (fail-open)")
	}
}

func TestAttributeFilter_InvalidConfig(t *testing.T) {
	// Neither attribute nor path specified
	_, err := NewAttributeFilterProcessor(AttributeFilterConfig{
		Name:     "test",
		Operator: OpEquals,
		Value:    "test",
	})
	if err == nil {
		t.Error("Expected error when neither attribute nor path is specified")
	}

	// Both attribute and path specified
	_, err = NewAttributeFilterProcessor(AttributeFilterConfig{
		Name:      "test",
		Attribute: "service.name",
		Path:      "some/path",
		Operator:  OpEquals,
		Value:     "test",
	})
	if err == nil {
		t.Error("Expected error when both attribute and path are specified")
	}

	// Invalid regex
	_, err = NewAttributeFilterProcessor(AttributeFilterConfig{
		Name:      "test",
		Attribute: "service.name",
		Operator:  OpRegex,
		Value:     "[invalid",
	})
	if err == nil {
		t.Error("Expected error for invalid regex pattern")
	}
}

func TestAttributeFilter_GenericSearchPaths(t *testing.T) {
	// Test with a non-well-known attribute that should use generic search
	proc, err := NewAttributeFilterProcessor(AttributeFilterConfig{
		Name:      "test",
		Attribute: "custom.field",
		Operator:  OpEquals,
		Value:     "custom-value",
	})
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		wantDrop bool
	}{
		{
			name:     "top-level custom attribute",
			input:    `{"custom.field": "custom-value", "message": "hello"}`,
			wantDrop: true,
		},
		{
			name:     "in attributes",
			input:    `{"attributes": {"custom.field": "custom-value"}, "body": "hello"}`,
			wantDrop: true,
		},
		{
			name:     "in resource.attributes",
			input:    `{"resource": {"attributes": {"custom.field": "custom-value"}}, "body": "hello"}`,
			wantDrop: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, drop, err := proc.Process(nil, []byte(tt.input))
			if err != nil {
				t.Errorf("Process() error = %v", err)
			}
			if drop != tt.wantDrop {
				t.Errorf("Process() drop = %v, want %v", drop, tt.wantDrop)
			}
		})
	}
}

func TestConvertToGjsonPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "resource/attributes/service.name",
			expected: "resource.attributes.service\\.name",
		},
		{
			input:    "metadata/labels/app.name",
			expected: "metadata.labels.app\\.name",
		},
		{
			input:    "simple",
			expected: "simple",
		},
		{
			input:    "deeply/nested/path/with.dots",
			expected: "deeply.nested.path.with\\.dots",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := convertToGjsonPath(tt.input)
			if result != tt.expected {
				t.Errorf("convertToGjsonPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
