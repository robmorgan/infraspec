package graph

import (
	"fmt"
)

// SchemaEntry defines the properties of a relationship between two resource types.
type SchemaEntry struct {
	// Type is the semantic relationship type.
	Type RelationshipType

	// Cardinality defines the multiplicity constraint.
	Cardinality Cardinality

	// DeleteBehavior specifies what happens when the target is deleted.
	DeleteBehavior DeleteBehavior

	// Required indicates if the relationship is required for the source to exist.
	// If true, creating the source resource requires establishing this relationship.
	Required bool

	// Description provides documentation for this relationship.
	Description string
}

// RelationshipSchema defines allowed relationships between resource types.
// The schema is optional - when not set or when StrictValidation is false,
// any relationship is allowed.
type RelationshipSchema struct {
	// Relationships maps relationship keys to their schema entries.
	// Key format: "service:type -> service:type"
	// Example: "ec2:subnet -> ec2:vpc"
	Relationships map[string]SchemaEntry
}

// NewRelationshipSchema creates a new empty schema.
func NewRelationshipSchema() *RelationshipSchema {
	return &RelationshipSchema{
		Relationships: make(map[string]SchemaEntry),
	}
}

// AddRelationship registers a relationship in the schema.
func (s *RelationshipSchema) AddRelationship(
	fromService, fromType, toService, toType string,
	entry SchemaEntry,
) {
	key := fmt.Sprintf("%s:%s -> %s:%s", fromService, fromType, toService, toType)
	s.Relationships[key] = entry
}

// GetRelationship returns the schema entry for a relationship, if defined.
func (s *RelationshipSchema) GetRelationship(fromService, fromType, toService, toType string) (SchemaEntry, bool) {
	key := fmt.Sprintf("%s:%s -> %s:%s", fromService, fromType, toService, toType)
	entry, exists := s.Relationships[key]
	return entry, exists
}

// GetRelationshipByKey returns the schema entry for a relationship key.
func (s *RelationshipSchema) GetRelationshipByKey(key string) (SchemaEntry, bool) {
	entry, exists := s.Relationships[key]
	return entry, exists
}

// HasRelationship returns true if the relationship is defined in the schema.
func (s *RelationshipSchema) HasRelationship(fromService, fromType, toService, toType string) bool {
	key := fmt.Sprintf("%s:%s -> %s:%s", fromService, fromType, toService, toType)
	_, exists := s.Relationships[key]
	return exists
}

// GetRelationshipsForSource returns all relationships where the given type is the source.
func (s *RelationshipSchema) GetRelationshipsForSource(service, resourceType string) map[string]SchemaEntry {
	prefix := fmt.Sprintf("%s:%s -> ", service, resourceType)
	result := make(map[string]SchemaEntry)

	for key, entry := range s.Relationships {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			result[key] = entry
		}
	}
	return result
}

// GetRelationshipsForTarget returns all relationships where the given type is the target.
func (s *RelationshipSchema) GetRelationshipsForTarget(service, resourceType string) map[string]SchemaEntry {
	suffix := fmt.Sprintf(" -> %s:%s", service, resourceType)
	result := make(map[string]SchemaEntry)

	for key, entry := range s.Relationships {
		if len(key) >= len(suffix) && key[len(key)-len(suffix):] == suffix {
			result[key] = entry
		}
	}
	return result
}

// ValidateRelationship checks if a relationship is valid according to the schema.
// Returns nil if valid, or an error describing the violation.
func (s *RelationshipSchema) ValidateRelationship(from, to ResourceID, relType RelationshipType) error {
	key := fmt.Sprintf("%s:%s -> %s:%s", from.Service, from.Type, to.Service, to.Type)
	entry, exists := s.Relationships[key]

	if !exists {
		return &SchemaValidationError{
			Relationship: key,
			Message:      "relationship not defined in schema: " + key,
		}
	}

	if entry.Type != relType {
		return &SchemaValidationError{
			Relationship: key,
			Message: fmt.Sprintf("relationship type mismatch for %s: expected %s, got %s",
				key, entry.Type, relType),
		}
	}

	return nil
}

// GetDeleteBehavior returns the delete behavior for a relationship.
// Returns DeleteRestrict as default if the relationship is not defined.
func (s *RelationshipSchema) GetDeleteBehavior(fromService, fromType, toService, toType string) DeleteBehavior {
	entry, exists := s.GetRelationship(fromService, fromType, toService, toType)
	if !exists {
		return DeleteRestrict
	}
	return entry.DeleteBehavior
}

// RelationshipCount returns the number of relationships defined in the schema.
func (s *RelationshipSchema) RelationshipCount() int {
	return len(s.Relationships)
}
