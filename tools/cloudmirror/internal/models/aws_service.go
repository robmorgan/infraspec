package models

// AWSService represents a parsed AWS service model
type AWSService struct {
	Name       string                `json:"name"`
	FullName   string                `json:"full_name"`
	APIVersion string                `json:"api_version"`
	Protocol   string                `json:"protocol"` // query, json, rest-xml, rest-json
	Operations map[string]*Operation `json:"operations"`
	Shapes     map[string]*Shape     `json:"shapes"`
}

// Operation represents an AWS API operation
type Operation struct {
	Name          string      `json:"name"`
	HTTPMethod    string      `json:"http_method"`
	HTTPPath      string      `json:"http_path"`
	InputShape    string      `json:"input_shape"`
	OutputShape   string      `json:"output_shape"`
	Parameters    []Parameter `json:"parameters"`
	Errors        []string    `json:"errors"`
	Documentation string      `json:"documentation"`
	Deprecated    bool        `json:"deprecated"`
	DeprecatedMsg string      `json:"deprecated_message,omitempty"`
}

// Parameter represents an input parameter for an operation
type Parameter struct {
	Name       string `json:"name"`
	ShapeRef   string `json:"shape_ref"`
	Type       string `json:"type"`
	Required   bool   `json:"required"`
	Deprecated bool   `json:"deprecated"`
	Location   string `json:"location"` // header, querystring, uri, body
}

// Shape represents an AWS API shape (type definition)
type Shape struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"` // string, integer, boolean, list, map, structure, timestamp, blob, long, double, float
	Required   []string          `json:"required,omitempty"`
	Members    map[string]*Shape `json:"members,omitempty"`
	Enum       []string          `json:"enum,omitempty"`
	Min        *int64            `json:"min,omitempty"`
	Max        *int64            `json:"max,omitempty"`
	Pattern    string            `json:"pattern,omitempty"`
	Deprecated bool              `json:"deprecated"`
}
