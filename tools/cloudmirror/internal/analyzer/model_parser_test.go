package analyzer

import (
	"testing"
)

func TestExtractLocalName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"com.amazonaws.rds#CreateDBInstance", "CreateDBInstance"},
		{"com.amazonaws.s3#PutObject", "PutObject"},
		{"com.amazonaws.dynamodb#GetItem", "GetItem"},
		{"SimpleShape", "SimpleShape"},
		{"#OnlyLocalName", "OnlyLocalName"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := extractLocalName(tt.input)
			if got != tt.expected {
				t.Errorf("extractLocalName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractNamespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"com.amazonaws.rds#CreateDBInstance", "com.amazonaws.rds"},
		{"com.amazonaws.s3#PutObject", "com.amazonaws.s3"},
		{"SimpleShape", ""},
		{"#OnlyLocalName", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := extractNamespace(tt.input)
			if got != tt.expected {
				t.Errorf("extractNamespace(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCleanDocumentation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Remove paragraph tags",
			input:    "<p>This is a paragraph.</p>",
			expected: "This is a paragraph.",
		},
		{
			name:     "Convert code tags to backticks",
			input:    "Use <code>CreateDBInstance</code> to create a database.",
			expected: "Use `CreateDBInstance` to create a database.",
		},
		{
			name:     "Convert italic tags",
			input:    "This is <i>important</i>.",
			expected: "This is _important_.",
		},
		{
			name:     "Convert bold tags",
			input:    "This is <b>very important</b>.",
			expected: "This is **very important**.",
		},
		{
			name:     "Convert HTML entities",
			input:    "Use &amp; &quot;quote&quot;",
			expected: "Use & \"quote\"",
		},
		{
			name:     "Remove arbitrary HTML tags",
			input:    "Text with <span class=\"foo\">span</span> content.",
			expected: "Text with span content.",
		},
		{
			name:     "Collapse multiple spaces",
			input:    "Too    many    spaces",
			expected: "Too many spaces",
		},
		{
			name:     "Truncate long documentation",
			input:    "This is a very long documentation string that exceeds the maximum allowed length of two hundred characters. We need to ensure that it gets properly truncated at the right position with an ellipsis appended to indicate that more content exists but was cut off for brevity.",
			expected: "This is a very long documentation string that exceeds the maximum allowed length of two hundred characters. We need to ensure that it gets properly truncated at the right position with an ellipsis ...",
		},
		{
			name:     "Handle newlines",
			input:    "Line 1\nLine 2\nLine 3",
			expected: "Line 1 Line 2 Line 3",
		},
		{
			name:     "Convert list items",
			input:    "<ul><li>Item 1</li><li>Item 2</li></ul>",
			expected: "- Item 1 - Item 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanDocumentation(tt.input)
			if got != tt.expected {
				t.Errorf("cleanDocumentation() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAWSModelParser_ExtractProtocol(t *testing.T) {
	parser := &AWSModelParser{}

	tests := []struct {
		name     string
		traits   map[string]interface{}
		expected string
	}{
		{
			name:     "Query protocol",
			traits:   map[string]interface{}{"aws.protocols#awsQuery": struct{}{}},
			expected: "query",
		},
		{
			name:     "JSON 1.0 protocol",
			traits:   map[string]interface{}{"aws.protocols#awsJson1_0": struct{}{}},
			expected: "json",
		},
		{
			name:     "JSON 1.1 protocol",
			traits:   map[string]interface{}{"aws.protocols#awsJson1_1": struct{}{}},
			expected: "json",
		},
		{
			name:     "REST JSON protocol",
			traits:   map[string]interface{}{"aws.protocols#restJson1": struct{}{}},
			expected: "rest-json",
		},
		{
			name:     "REST XML protocol",
			traits:   map[string]interface{}{"aws.protocols#restXml": struct{}{}},
			expected: "rest-xml",
		},
		{
			name:     "EC2 Query protocol",
			traits:   map[string]interface{}{"aws.protocols#ec2Query": struct{}{}},
			expected: "ec2",
		},
		{
			name:     "Unknown protocol",
			traits:   map[string]interface{}{},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.extractProtocol(tt.traits)
			if got != tt.expected {
				t.Errorf("extractProtocol() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAWSModelParser_IsDeprecated(t *testing.T) {
	parser := &AWSModelParser{}

	tests := []struct {
		name     string
		traits   map[string]interface{}
		expected bool
	}{
		{
			name:     "Not deprecated",
			traits:   map[string]interface{}{},
			expected: false,
		},
		{
			name:     "Deprecated",
			traits:   map[string]interface{}{"smithy.api#deprecated": struct{}{}},
			expected: true,
		},
		{
			name:     "Deprecated with message",
			traits:   map[string]interface{}{"smithy.api#deprecated": map[string]interface{}{"message": "Use NewAPI instead"}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.isDeprecated(tt.traits)
			if got != tt.expected {
				t.Errorf("isDeprecated() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAWSModelParser_ExtractDeprecatedMessage(t *testing.T) {
	parser := &AWSModelParser{}

	tests := []struct {
		name     string
		traits   map[string]interface{}
		expected string
	}{
		{
			name:     "No deprecated trait",
			traits:   map[string]interface{}{},
			expected: "",
		},
		{
			name:     "Deprecated without message",
			traits:   map[string]interface{}{"smithy.api#deprecated": struct{}{}},
			expected: "",
		},
		{
			name: "Deprecated with message",
			traits: map[string]interface{}{
				"smithy.api#deprecated": map[string]interface{}{
					"message": "Use NewAPI instead",
				},
			},
			expected: "Use NewAPI instead",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.extractDeprecatedMessage(tt.traits)
			if got != tt.expected {
				t.Errorf("extractDeprecatedMessage() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAWSModelParser_ExtractHTTPMethod(t *testing.T) {
	parser := &AWSModelParser{}

	tests := []struct {
		name     string
		traits   map[string]interface{}
		expected string
	}{
		{
			name:     "Default POST for query protocol",
			traits:   map[string]interface{}{},
			expected: "POST",
		},
		{
			name: "GET method",
			traits: map[string]interface{}{
				"smithy.api#http": map[string]interface{}{
					"method": "GET",
					"uri":    "/resources",
				},
			},
			expected: "GET",
		},
		{
			name: "PUT method",
			traits: map[string]interface{}{
				"smithy.api#http": map[string]interface{}{
					"method": "PUT",
					"uri":    "/resources/{id}",
				},
			},
			expected: "PUT",
		},
		{
			name: "DELETE method",
			traits: map[string]interface{}{
				"smithy.api#http": map[string]interface{}{
					"method": "DELETE",
					"uri":    "/resources/{id}",
				},
			},
			expected: "DELETE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.extractHTTPMethod(tt.traits)
			if got != tt.expected {
				t.Errorf("extractHTTPMethod() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAWSModelParser_ExtractHTTPPath(t *testing.T) {
	parser := &AWSModelParser{}

	tests := []struct {
		name     string
		traits   map[string]interface{}
		expected string
	}{
		{
			name:     "Default path",
			traits:   map[string]interface{}{},
			expected: "/",
		},
		{
			name: "Custom path",
			traits: map[string]interface{}{
				"smithy.api#http": map[string]interface{}{
					"method": "GET",
					"uri":    "/resources/{resourceId}",
				},
			},
			expected: "/resources/{resourceId}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.extractHTTPPath(tt.traits)
			if got != tt.expected {
				t.Errorf("extractHTTPPath() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAWSModelParser_ExtractLocation(t *testing.T) {
	parser := &AWSModelParser{}

	tests := []struct {
		name     string
		traits   map[string]interface{}
		expected string
	}{
		{
			name:     "No location trait",
			traits:   map[string]interface{}{},
			expected: "",
		},
		{
			name:     "Header location",
			traits:   map[string]interface{}{"smithy.api#httpHeader": "X-Custom-Header"},
			expected: "header",
		},
		{
			name:     "Query location",
			traits:   map[string]interface{}{"smithy.api#httpQuery": "param"},
			expected: "querystring",
		},
		{
			name:     "URI location",
			traits:   map[string]interface{}{"smithy.api#httpLabel": struct{}{}},
			expected: "uri",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.extractLocation(tt.traits)
			if got != tt.expected {
				t.Errorf("extractLocation() = %q, want %q", got, tt.expected)
			}
		})
	}
}
