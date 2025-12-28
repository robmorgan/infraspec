// Package smithy provides types and parsing functionality for AWS Smithy 2.0 JSON AST models.
package smithy

// Model represents the top-level Smithy 2.0 JSON AST format from api-models-aws
type Model struct {
	Smithy   string                 `json:"smithy"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Shapes   map[string]Shape       `json:"shapes"`
}

// Shape represents a shape definition in the Smithy model
type Shape struct {
	Type    string                 `json:"type"`
	Members map[string]Member      `json:"members,omitempty"`
	Traits  map[string]interface{} `json:"traits,omitempty"`
	Input   *ShapeRef              `json:"input,omitempty"`
	Output  *ShapeRef              `json:"output,omitempty"`
	Errors  []ShapeRef             `json:"errors,omitempty"`
	Target  string                 `json:"target,omitempty"` // For list/set target
	Key     *Member                `json:"key,omitempty"`    // For map key
	Value   *Member                `json:"value,omitempty"`  // For map value
	Member  *Member                `json:"member,omitempty"` // For list member
}

// Member represents a member of a structure shape
type Member struct {
	Target string                 `json:"target"`
	Traits map[string]interface{} `json:"traits,omitempty"`
}

// ShapeRef is a reference to another shape
type ShapeRef struct {
	Target string `json:"target"`
}

// Shape types in Smithy
const (
	ShapeTypeService   = "service"
	ShapeTypeOperation = "operation"
	ShapeTypeStructure = "structure"
	ShapeTypeList      = "list"
	ShapeTypeSet       = "set"
	ShapeTypeMap       = "map"
	ShapeTypeString    = "string"
	ShapeTypeInteger   = "integer"
	ShapeTypeLong      = "long"
	ShapeTypeShort     = "short"
	ShapeTypeByte      = "byte"
	ShapeTypeFloat     = "float"
	ShapeTypeDouble    = "double"
	ShapeTypeBoolean   = "boolean"
	ShapeTypeTimestamp = "timestamp"
	ShapeTypeBlob      = "blob"
	ShapeTypeBigInt    = "bigInteger"
	ShapeTypeBigDec    = "bigDecimal"
	ShapeTypeEnum      = "enum"
	ShapeTypeUnion     = "union"
	ShapeTypeDocument  = "document"
)

// Smithy trait keys
const (
	// XML serialization traits
	TraitXMLName       = "smithy.api#xmlName"
	TraitXMLAttribute  = "smithy.api#xmlAttribute"
	TraitXMLFlattened  = "smithy.api#xmlFlattened"
	TraitXMLNamespace  = "smithy.api#xmlNamespace"
	TraitEC2QueryName  = "aws.protocols#ec2QueryName"

	// Metadata traits
	TraitRequired      = "smithy.api#required"
	TraitDocumentation = "smithy.api#documentation"
	TraitDeprecated    = "smithy.api#deprecated"
	TraitSensitive     = "smithy.api#sensitive"
	TraitDefault       = "smithy.api#default"
	TraitEnumValue     = "smithy.api#enumValue"

	// Protocol traits
	TraitAWSQuery   = "aws.protocols#awsQuery"
	TraitAWSJSON10  = "aws.protocols#awsJson1_0"
	TraitAWSJSON11  = "aws.protocols#awsJson1_1"
	TraitRestJSON   = "aws.protocols#restJson1"
	TraitRestXML    = "aws.protocols#restXml"
	TraitEC2Query   = "aws.protocols#ec2Query"

	// AWS service traits
	TraitAWSService = "aws.api#service"
	TraitTitle      = "smithy.api#title"

	// HTTP traits
	TraitHTTP        = "smithy.api#http"
	TraitHTTPHeader  = "smithy.api#httpHeader"
	TraitHTTPQuery   = "smithy.api#httpQuery"
	TraitHTTPLabel   = "smithy.api#httpLabel"
	TraitHTTPPayload = "smithy.api#httpPayload"

	// Validation/Constraint traits
	TraitLength  = "smithy.api#length"
	TraitPattern = "smithy.api#pattern"
	TraitRange   = "smithy.api#range"
)

// XMLTraits contains extracted XML serialization information for a member
type XMLTraits struct {
	XMLName     string // Element name from xmlName trait
	EC2Name     string // EC2 Query protocol name
	IsFlattened bool   // List without wrapper element
	IsAttribute bool   // Serialize as XML attribute
}

// ExtractXMLTraits extracts XML serialization traits from a member's traits
func ExtractXMLTraits(traits map[string]interface{}) XMLTraits {
	var result XMLTraits

	if name, ok := traits[TraitXMLName].(string); ok {
		result.XMLName = name
	}

	if name, ok := traits[TraitEC2QueryName].(string); ok {
		result.EC2Name = name
	}

	if _, ok := traits[TraitXMLFlattened]; ok {
		result.IsFlattened = true
	}

	if _, ok := traits[TraitXMLAttribute]; ok {
		result.IsAttribute = true
	}

	return result
}

// ValidationTraits contains extracted validation constraint information for a member
type ValidationTraits struct {
	// Length constraints (for strings, lists, maps)
	LengthMin *int64 // Minimum length (nil if not set)
	LengthMax *int64 // Maximum length (nil if not set)

	// Pattern constraint (for strings)
	Pattern string // Regex pattern (empty if not set)

	// Range constraints (for numeric types)
	RangeMin *float64 // Minimum value (nil if not set)
	RangeMax *float64 // Maximum value (nil if not set)
}

// HasConstraints returns true if any validation constraints are defined
func (v ValidationTraits) HasConstraints() bool {
	return v.LengthMin != nil || v.LengthMax != nil ||
		v.Pattern != "" ||
		v.RangeMin != nil || v.RangeMax != nil
}

// ExtractValidationTraits extracts validation constraint traits from a shape/member's traits
func ExtractValidationTraits(traits map[string]interface{}) ValidationTraits {
	var result ValidationTraits

	// Extract length trait: {"min": N, "max": M}
	if length, ok := traits[TraitLength].(map[string]interface{}); ok {
		if min, ok := length["min"].(float64); ok {
			minInt := int64(min)
			result.LengthMin = &minInt
		}
		if max, ok := length["max"].(float64); ok {
			maxInt := int64(max)
			result.LengthMax = &maxInt
		}
	}

	// Extract pattern trait: "regex-string"
	if pattern, ok := traits[TraitPattern].(string); ok {
		result.Pattern = pattern
	}

	// Extract range trait: {"min": N, "max": M}
	if rangeVal, ok := traits[TraitRange].(map[string]interface{}); ok {
		if min, ok := rangeVal["min"].(float64); ok {
			result.RangeMin = &min
		}
		if max, ok := rangeVal["max"].(float64); ok {
			result.RangeMax = &max
		}
	}

	return result
}

// HTTPTraits contains extracted HTTP binding information for a member
type HTTPTraits struct {
	Location     string // "header", "query", "uri", "payload", or ""
	LocationName string // The header name, query param name, etc.
	IsPayload    bool   // True if this is the request/response body
}

// HasHTTPTraits returns true if any HTTP location trait is present
func (h HTTPTraits) HasHTTPTraits() bool {
	return h.Location != ""
}

// ExtractHTTPTraits extracts HTTP location traits from a member's traits
func ExtractHTTPTraits(traits map[string]interface{}) HTTPTraits {
	var result HTTPTraits

	// httpHeader: "X-Header-Name" (string value)
	if headerName, ok := traits[TraitHTTPHeader].(string); ok {
		result.Location = "header"
		result.LocationName = headerName
		return result
	}

	// httpQuery: "queryParam" (string value)
	if queryName, ok := traits[TraitHTTPQuery].(string); ok {
		result.Location = "query"
		result.LocationName = queryName
		return result
	}

	// httpLabel: {} (empty object - uses member name as URI label)
	if _, ok := traits[TraitHTTPLabel]; ok {
		result.Location = "uri"
		// LocationName left empty - will use member name
		return result
	}

	// httpPayload: {} (empty object - this field is the body)
	if _, ok := traits[TraitHTTPPayload]; ok {
		result.Location = "payload"
		result.IsPayload = true
		return result
	}

	return result
}

// GetXMLElementName returns the XML element name for a member
// For response types, prefer xmlName which has the correct camelCase names for AWS responses.
// ec2QueryName is for request parameters and uses PascalCase.
// When no xmlName trait exists, behavior is protocol-dependent:
// - EC2 protocol: converts to camelCase (e.g., VpcId -> vpcId)
// - Query protocol: uses member name as-is (e.g., DBInstance -> DBInstance)
func GetXMLElementName(memberName string, traits map[string]interface{}, protocol string) string {
	// Prefer explicit xmlName trait
	if name, ok := traits[TraitXMLName].(string); ok && name != "" {
		return name
	}

	// Protocol-specific fallback when no xmlName trait exists
	if protocol == "ec2" {
		// EC2 protocol uses camelCase for XML elements
		return toLowerFirst(memberName)
	}

	// Query protocol (RDS, IAM, STS, etc.) uses member name as-is (PascalCase)
	return memberName
}

// IsRequired checks if a member is required
func IsRequired(traits map[string]interface{}) bool {
	_, ok := traits[TraitRequired]
	return ok
}

// GetDocumentation extracts documentation from traits
func GetDocumentation(traits map[string]interface{}) string {
	if doc, ok := traits[TraitDocumentation].(string); ok {
		return doc
	}
	return ""
}

// IsDeprecated checks if a shape or member is deprecated
func IsDeprecated(traits map[string]interface{}) bool {
	_, ok := traits[TraitDeprecated]
	return ok
}

// GetEnumValue gets the enum value from traits
func GetEnumValue(traits map[string]interface{}) string {
	if val, ok := traits[TraitEnumValue].(string); ok {
		return val
	}
	return ""
}

// toLowerFirst converts the first character to lowercase
func toLowerFirst(s string) string {
	if s == "" {
		return s
	}
	return string(s[0]|32) + s[1:]
}
