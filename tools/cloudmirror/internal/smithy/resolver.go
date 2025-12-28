package smithy

import (
	"fmt"
	"strings"
	"unicode"
)

// Resolver resolves Smithy types to Go types and collects dependencies
type Resolver struct {
	parser   *Parser
	protocol string
	resolved map[string]bool // Track which shapes have been resolved
}

// NewResolver creates a new type resolver
func NewResolver(parser *Parser, protocol string) *Resolver {
	return &Resolver{
		parser:   parser,
		protocol: protocol,
		resolved: make(map[string]bool),
	}
}

// ResolvedType represents a resolved Go type
type ResolvedType struct {
	Name          string          // Go type name
	ShapeName     string          // Original Smithy shape name
	ShapeType     string          // Smithy shape type
	GoType        string          // Go type string (e.g., "string", "[]Vpc", "*time.Time")
	Fields        []ResolvedField // Fields for structure types
	EnumValues    []EnumValue     // Values for enum types
	ListItemType  string          // Item type for list types
	MapKeyType    string          // Key type for map types
	MapValueType  string          // Value type for map types
	Documentation string          // Documentation from traits
	IsDeprecated  bool            // From deprecated trait
}

// ResolvedField represents a resolved struct field
type ResolvedField struct {
	Name          string           // Go field name (PascalCase)
	MemberName    string           // Original Smithy member name
	GoType        string           // Go type string
	XMLName       string           // XML element name (from xmlName trait)
	XMLTag        string           // Full XML tag (e.g., "vpcId" or "tagSet>item")
	IsRequired    bool             // From required trait
	IsFlattened   bool             // From xmlFlattened trait
	IsAttribute   bool             // From xmlAttribute trait
	Documentation string           // Documentation
	TargetShape   string           // Target shape name for nested types
	Validation    ValidationTraits // Validation constraints (length, pattern, range)
	HTTP          HTTPTraits       // HTTP location traits (header, query, uri, payload)
}

// EnumValue represents an enum value
type EnumValue struct {
	Name  string // Go const name
	Value string // String value
}

// ResolveShape resolves a shape by name and returns all dependent types
func (r *Resolver) ResolveShape(shapeName string) (*ResolvedType, []string, error) {
	shape, ok := r.parser.GetShape(shapeName)
	if !ok {
		return nil, nil, fmt.Errorf("shape not found: %s", shapeName)
	}

	return r.resolveShapeInternal(shapeName, shape)
}

// resolveShapeInternal resolves a shape and collects dependencies
func (r *Resolver) resolveShapeInternal(shapeName string, shape *Shape) (*ResolvedType, []string, error) {
	resolved := &ResolvedType{
		Name:          shapeName,
		ShapeName:     shapeName,
		ShapeType:     shape.Type,
		Documentation: GetDocumentation(shape.Traits),
		IsDeprecated:  IsDeprecated(shape.Traits),
	}

	var dependencies []string

	switch shape.Type {
	case ShapeTypeStructure:
		fields, deps, err := r.resolveStructure(shape)
		if err != nil {
			return nil, nil, err
		}
		resolved.Fields = fields
		resolved.GoType = shapeName
		dependencies = deps

	case ShapeTypeList, ShapeTypeSet:
		itemType, deps, err := r.resolveList(shape)
		if err != nil {
			return nil, nil, err
		}
		resolved.ListItemType = itemType
		resolved.GoType = "[]" + itemType
		dependencies = deps

	case ShapeTypeMap:
		keyType, valueType, deps, err := r.resolveMap(shape)
		if err != nil {
			return nil, nil, err
		}
		resolved.MapKeyType = keyType
		resolved.MapValueType = valueType
		resolved.GoType = fmt.Sprintf("map[%s]%s", keyType, valueType)
		dependencies = deps

	case ShapeTypeEnum:
		values := r.resolveEnum(shape)
		resolved.EnumValues = values
		resolved.GoType = "string" // Enums are represented as strings

	case ShapeTypeString, ShapeTypeInteger, ShapeTypeLong, ShapeTypeShort, ShapeTypeByte,
		ShapeTypeFloat, ShapeTypeDouble, ShapeTypeBoolean, ShapeTypeTimestamp, ShapeTypeBlob,
		ShapeTypeBigInt, ShapeTypeBigDec, ShapeTypeDocument:
		resolved.GoType = r.primitiveToGoType(shape.Type)

	default:
		resolved.GoType = "interface{}"
	}

	return resolved, dependencies, nil
}

// resolveStructure resolves a structure shape's fields
func (r *Resolver) resolveStructure(shape *Shape) ([]ResolvedField, []string, error) {
	var fields []ResolvedField
	var dependencies []string

	for memberName, member := range shape.Members {
		field, deps, err := r.resolveMember(memberName, &member)
		if err != nil {
			return nil, nil, err
		}
		fields = append(fields, *field)
		dependencies = append(dependencies, deps...)
	}

	return fields, dependencies, nil
}

// resolveMember resolves a structure member
func (r *Resolver) resolveMember(memberName string, member *Member) (*ResolvedField, []string, error) {
	field := &ResolvedField{
		Name:          capitalizeFirst(memberName), // Capitalize for Go export
		MemberName:    memberName,                  // Keep original for JSON/XML tags
		IsRequired:    IsRequired(member.Traits),
		Documentation: GetDocumentation(member.Traits),
	}

	// Extract XML traits
	xmlTraits := ExtractXMLTraits(member.Traits)
	field.IsFlattened = xmlTraits.IsFlattened
	field.IsAttribute = xmlTraits.IsAttribute

	// Extract validation traits
	field.Validation = ExtractValidationTraits(member.Traits)

	// Extract HTTP location traits
	field.HTTP = ExtractHTTPTraits(member.Traits)

	// Determine XML element name
	field.XMLName = GetXMLElementName(memberName, member.Traits, r.protocol)

	var dependencies []string

	// Resolve target type
	targetName := ResolveTarget(member.Target)
	field.TargetShape = targetName

	// Check if target is a smithy primitive
	if strings.HasPrefix(member.Target, "smithy.api#") {
		field.GoType = r.primitiveToGoType(targetName)
		field.XMLTag = field.XMLName
	} else {
		// Look up the target shape
		targetShape, ok := r.parser.GetShape(targetName)
		if !ok {
			// If not found, treat as string
			field.GoType = "string"
			field.XMLTag = field.XMLName
		} else {
			goType, xmlTag, deps := r.resolveTargetType(targetName, targetShape, field.XMLName, xmlTraits)
			field.GoType = goType
			field.XMLTag = xmlTag
			dependencies = deps

			// Merge validation traits from target shape (for type aliases like UsernameType)
			// Member traits take precedence over shape traits
			if targetShape.Traits != nil {
				shapeValidation := ExtractValidationTraits(targetShape.Traits)
				if field.Validation.LengthMin == nil {
					field.Validation.LengthMin = shapeValidation.LengthMin
				}
				if field.Validation.LengthMax == nil {
					field.Validation.LengthMax = shapeValidation.LengthMax
				}
				if field.Validation.Pattern == "" {
					field.Validation.Pattern = shapeValidation.Pattern
				}
				if field.Validation.RangeMin == nil {
					field.Validation.RangeMin = shapeValidation.RangeMin
				}
				if field.Validation.RangeMax == nil {
					field.Validation.RangeMax = shapeValidation.RangeMax
				}
			}
		}
	}

	// Add omitempty for optional fields
	if !field.IsRequired && !field.IsAttribute {
		if !strings.Contains(field.XMLTag, ",") {
			field.XMLTag += ",omitempty"
		}
	}

	return field, dependencies, nil
}

// resolveTargetType resolves the Go type for a target shape
func (r *Resolver) resolveTargetType(targetName string, targetShape *Shape, xmlName string, xmlTraits XMLTraits) (string, string, []string) {
	var dependencies []string

	switch targetShape.Type {
	case ShapeTypeStructure:
		dependencies = append(dependencies, targetName)
		return targetName, xmlName, dependencies

	case ShapeTypeList, ShapeTypeSet:
		// Get the list item type
		var itemTypeName string
		if targetShape.Member != nil {
			itemTypeName = ResolveTarget(targetShape.Member.Target)
		} else if targetShape.Target != "" {
			itemTypeName = ResolveTarget(targetShape.Target)
		} else {
			itemTypeName = "interface{}"
		}

		// Check if item type is a structure
		itemShape, ok := r.parser.GetShape(itemTypeName)
		if ok && itemShape.Type == ShapeTypeStructure {
			dependencies = append(dependencies, itemTypeName)
		}

		goType := "[]" + r.goTypeName(itemTypeName)

		// Build XML tag for list
		if xmlTraits.IsFlattened {
			// Flattened lists don't have wrapper
			return goType, xmlName, dependencies
		}

		// Get item element name from list member traits
		itemXMLName := "item" // default
		if targetShape.Member != nil {
			memberXMLTraits := ExtractXMLTraits(targetShape.Member.Traits)
			if memberXMLTraits.XMLName != "" {
				itemXMLName = memberXMLTraits.XMLName
			}
		}

		return goType, xmlName + ">" + itemXMLName, dependencies

	case ShapeTypeMap:
		return "map[string]string", xmlName, nil

	case ShapeTypeEnum:
		// Return the enum type name so it can be used as a type alias
		// Enum types don't get added as dependencies since they're type aliases, not structs
		return targetName, xmlName, nil

	case ShapeTypeString, ShapeTypeInteger, ShapeTypeLong, ShapeTypeShort, ShapeTypeByte,
		ShapeTypeFloat, ShapeTypeDouble, ShapeTypeBoolean, ShapeTypeTimestamp, ShapeTypeBlob,
		ShapeTypeBigInt, ShapeTypeBigDec, ShapeTypeDocument:
		return r.primitiveToGoType(targetShape.Type), xmlName, nil

	default:
		return "interface{}", xmlName, nil
	}
}

// resolveList resolves a list shape's item type
func (r *Resolver) resolveList(shape *Shape) (string, []string, error) {
	var itemTypeName string
	if shape.Member != nil {
		itemTypeName = ResolveTarget(shape.Member.Target)
	} else if shape.Target != "" {
		itemTypeName = ResolveTarget(shape.Target)
	} else {
		return "interface{}", nil, nil
	}

	var dependencies []string

	// Check if item type is a structure
	itemShape, ok := r.parser.GetShape(itemTypeName)
	if ok && itemShape.Type == ShapeTypeStructure {
		dependencies = append(dependencies, itemTypeName)
		return itemTypeName, dependencies, nil
	}

	return r.goTypeName(itemTypeName), dependencies, nil
}

// resolveMap resolves a map shape's key and value types
func (r *Resolver) resolveMap(shape *Shape) (string, string, []string, error) {
	keyType := "string"
	valueType := "string"
	var dependencies []string

	if shape.Key != nil {
		keyTypeName := ResolveTarget(shape.Key.Target)
		keyType = r.goTypeName(keyTypeName)
	}

	if shape.Value != nil {
		valueTypeName := ResolveTarget(shape.Value.Target)
		valueShape, ok := r.parser.GetShape(valueTypeName)
		if ok && valueShape.Type == ShapeTypeStructure {
			dependencies = append(dependencies, valueTypeName)
			valueType = valueTypeName
		} else {
			valueType = r.goTypeName(valueTypeName)
		}
	}

	return keyType, valueType, dependencies, nil
}

// resolveEnum resolves an enum shape's values
func (r *Resolver) resolveEnum(shape *Shape) []EnumValue {
	var values []EnumValue

	for name, member := range shape.Members {
		value := GetEnumValue(member.Traits)
		if value == "" {
			value = name
		}
		values = append(values, EnumValue{
			Name:  name,
			Value: value,
		})
	}

	return values
}

// primitiveToGoType converts a Smithy primitive type to a Go type
func (r *Resolver) primitiveToGoType(smithyType string) string {
	switch smithyType {
	case ShapeTypeString, "String":
		return "string"
	case ShapeTypeInteger, "Integer":
		return "int32"
	case ShapeTypeLong, "Long":
		return "int64"
	case ShapeTypeShort, "Short":
		return "int16"
	case ShapeTypeByte, "Byte":
		return "int8"
	case ShapeTypeFloat, "Float":
		return "float32"
	case ShapeTypeDouble, "Double":
		return "float64"
	case ShapeTypeBoolean, "Boolean":
		return "bool"
	case ShapeTypeTimestamp, "Timestamp":
		return "time.Time"
	case ShapeTypeBlob, "Blob":
		return "[]byte"
	case ShapeTypeBigInt, "BigInteger":
		return "string" // Big integers as strings
	case ShapeTypeBigDec, "BigDecimal":
		return "string" // Big decimals as strings
	case ShapeTypeDocument, "Document":
		return "interface{}"
	default:
		return "string"
	}
}

// goTypeName returns the Go type name for a resolved shape
func (r *Resolver) goTypeName(shapeName string) string {
	// Check if it's a known primitive type name
	primitiveType := r.primitiveToGoType(shapeName)
	if primitiveType != "string" || shapeName == "String" || shapeName == "string" {
		return primitiveType
	}

	// Look up the shape to see if it's a string/enum alias
	shape, ok := r.parser.GetShape(shapeName)
	if ok {
		switch shape.Type {
		case ShapeTypeString:
			return "string"
		case ShapeTypeInteger:
			return "int32"
		case ShapeTypeLong:
			return "int64"
		case ShapeTypeShort:
			return "int16"
		case ShapeTypeByte:
			return "int8"
		case ShapeTypeFloat:
			return "float32"
		case ShapeTypeDouble:
			return "float64"
		case ShapeTypeBoolean:
			return "bool"
		case ShapeTypeTimestamp:
			return "time.Time"
		case ShapeTypeBlob:
			return "[]byte"
		case ShapeTypeEnum:
			return shapeName // Return the enum type name for type alias
		}
	}

	// Otherwise it's a custom type
	return shapeName
}

// capitalizeFirst capitalizes the first letter of a string for Go export.
// This ensures field names are exported (public) in the generated Go structs.
// Examples: "tags" -> "Tags", "queueName" -> "QueueName", "VpcId" -> "VpcId"
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// CollectDependencies recursively collects all type dependencies starting from a shape
func (r *Resolver) CollectDependencies(shapeName string) ([]string, error) {
	visited := make(map[string]bool)
	var result []string

	err := r.collectDepsRecursive(shapeName, visited, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// collectDepsRecursive recursively collects dependencies
func (r *Resolver) collectDepsRecursive(shapeName string, visited map[string]bool, result *[]string) error {
	if visited[shapeName] {
		return nil
	}
	visited[shapeName] = true

	shape, ok := r.parser.GetShape(shapeName)
	if !ok {
		return nil // Skip unknown shapes
	}

	// Only collect structure types
	if shape.Type != ShapeTypeStructure {
		return nil
	}

	*result = append(*result, shapeName)

	// Collect dependencies from members
	for _, member := range shape.Members {
		targetName := ResolveTarget(member.Target)
		targetShape, ok := r.parser.GetShape(targetName)
		if !ok {
			continue
		}

		if targetShape.Type == ShapeTypeStructure {
			if err := r.collectDepsRecursive(targetName, visited, result); err != nil {
				return err
			}
		} else if targetShape.Type == ShapeTypeList || targetShape.Type == ShapeTypeSet {
			// Check list item type
			var itemTypeName string
			if targetShape.Member != nil {
				itemTypeName = ResolveTarget(targetShape.Member.Target)
			} else if targetShape.Target != "" {
				itemTypeName = ResolveTarget(targetShape.Target)
			}
			if itemTypeName != "" {
				if err := r.collectDepsRecursive(itemTypeName, visited, result); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
