package rules

// Operator represents a condition operator
type Operator string

const (
	// Existence operators
	OpExists    Operator = "exists"
	OpNotExists Operator = "not_exists"

	// Equality operators
	OpEquals    Operator = "equals"
	OpNotEquals Operator = "not_equals"

	// String/array operators
	OpContains    Operator = "contains"
	OpNotContains Operator = "not_contains"
	OpMatches     Operator = "matches"

	// Numeric operators
	OpGreaterThan Operator = "greater_than"
	OpLessThan    Operator = "less_than"

	// Set operators
	OpOneOf Operator = "one_of"

	// Logical operators (for combining conditions)
	OpAll Operator = "all"
	OpAny Operator = "any"
	OpNot Operator = "not"
)

// Condition represents a rule condition
type Condition struct {
	// Attribute is the path to the attribute to check (e.g., "versioning.enabled")
	Attribute string `yaml:"attribute,omitempty"`

	// Operator is the comparison operator
	Operator Operator `yaml:"operator,omitempty"`

	// Value is the expected value for comparison operators
	Value interface{} `yaml:"value,omitempty"`

	// Conditions is used for logical operators (all, any, not)
	Conditions []Condition `yaml:"conditions,omitempty"`
}

// IsLogical returns true if this is a logical operator (all, any, not)
func (c *Condition) IsLogical() bool {
	switch c.Operator {
	case OpAll, OpAny, OpNot:
		return true
	default:
		return false
	}
}

// Validate checks if the condition is valid
func (c *Condition) Validate() error {
	// Logical operators require conditions
	if c.IsLogical() {
		if len(c.Conditions) == 0 {
			return &ValidationError{
				Field:   "conditions",
				Message: "logical operator requires at least one condition",
			}
		}
		// Validate nested conditions
		for i, sub := range c.Conditions {
			if err := sub.Validate(); err != nil {
				return &ValidationError{
					Field:   "conditions",
					Message: "invalid condition at index " + string(rune('0'+i)) + ": " + err.Error(),
				}
			}
		}
		return nil
	}

	// Non-logical operators require attribute
	if c.Attribute == "" {
		return &ValidationError{
			Field:   "attribute",
			Message: "attribute is required for non-logical operators",
		}
	}

	// Validate operator-specific requirements
	switch c.Operator {
	case OpExists, OpNotExists:
		// No value needed
	case OpEquals, OpNotEquals, OpContains, OpNotContains, OpMatches, OpGreaterThan, OpLessThan, OpOneOf:
		if c.Value == nil {
			return &ValidationError{
				Field:   "value",
				Message: "value is required for operator " + string(c.Operator),
			}
		}
	case "":
		return &ValidationError{
			Field:   "operator",
			Message: "operator is required",
		}
	default:
		return &ValidationError{
			Field:   "operator",
			Message: "unknown operator: " + string(c.Operator),
		}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
