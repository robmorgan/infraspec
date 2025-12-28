// Package graph provides a graph-based resource relationship system for modeling
// AWS resource dependencies in the emulator.
package graph

import (
	"fmt"
	"time"
)

// ResourceID uniquely identifies a resource across all services.
type ResourceID struct {
	Service string // e.g., "ec2", "iam", "rds"
	Type    string // e.g., "vpc", "instance", "role"
	ID      string // e.g., "vpc-12345", "i-abc123"
}

// String returns the canonical string representation of the ResourceID.
func (r ResourceID) String() string {
	return fmt.Sprintf("%s:%s:%s", r.Service, r.Type, r.ID)
}

// TypeKey returns the service:type portion for schema lookups.
func (r ResourceID) TypeKey() string {
	return fmt.Sprintf("%s:%s", r.Service, r.Type)
}

// IsZero returns true if the ResourceID is empty.
func (r ResourceID) IsZero() bool {
	return r.Service == "" && r.Type == "" && r.ID == ""
}

// RelationshipType defines the semantic meaning of an edge between resources.
type RelationshipType string

const (
	// RelContains indicates parent-child containment (parent owns child lifecycle).
	// Example: VPC contains Subnet
	RelContains RelationshipType = "contains"

	// RelReferences indicates a resource references another resource.
	// Example: Instance references SecurityGroup
	RelReferences RelationshipType = "references"

	// RelAttachedTo indicates a bidirectional attachment association.
	// Example: InternetGateway attached to VPC
	RelAttachedTo RelationshipType = "attached_to"

	// RelAssociatedWith indicates a loose coupling association.
	// Example: Role associated with Policy
	RelAssociatedWith RelationshipType = "associated_with"
)

// Cardinality specifies the relationship multiplicity.
type Cardinality int

const (
	// CardOneToOne allows exactly one source to one target.
	// Example: Instance -> KeyPair
	CardOneToOne Cardinality = iota

	// CardOneToMany allows one source to multiple targets.
	// Example: VPC -> Subnets (from VPC's perspective)
	CardOneToMany

	// CardManyToOne allows multiple sources to one target.
	// Example: Subnets -> VPC (from Subnet's perspective)
	CardManyToOne

	// CardManyToMany allows multiple sources to multiple targets.
	// Example: Roles <-> Policies
	CardManyToMany
)

// String returns a human-readable cardinality description.
func (c Cardinality) String() string {
	switch c {
	case CardOneToOne:
		return "one-to-one"
	case CardOneToMany:
		return "one-to-many"
	case CardManyToOne:
		return "many-to-one"
	case CardManyToMany:
		return "many-to-many"
	default:
		return "unknown"
	}
}

// DeleteBehavior specifies what happens when deleting a node with dependents.
type DeleteBehavior int

const (
	// DeleteRestrict blocks deletion if dependents exist.
	DeleteRestrict DeleteBehavior = iota

	// DeleteCascade recursively deletes dependents (dangerous, rarely used).
	DeleteCascade

	// DeleteSetNull removes relationship references but keeps dependent resources.
	DeleteSetNull
)

// String returns a human-readable delete behavior description.
func (d DeleteBehavior) String() string {
	switch d {
	case DeleteRestrict:
		return "restrict"
	case DeleteCascade:
		return "cascade"
	case DeleteSetNull:
		return "set-null"
	default:
		return "unknown"
	}
}

// ValidationMode controls strictness of relationship validation.
type ValidationMode string

const (
	// ValidationStrict enforces all schema constraints and fails on violations.
	ValidationStrict ValidationMode = "strict"

	// ValidationLenient logs warnings but allows operations (backward compatible).
	ValidationLenient ValidationMode = "lenient"
)

// Edge represents a directed relationship between two resources.
type Edge struct {
	From     ResourceID
	To       ResourceID
	Type     RelationshipType
	Metadata map[string]string // Optional metadata (e.g., attachment IDs, timestamps)
}

// String returns a human-readable edge description.
func (e *Edge) String() string {
	return fmt.Sprintf("%s -[%s]-> %s", e.From.String(), e.Type, e.To.String())
}

// Node represents a resource in the graph.
type Node struct {
	ID        ResourceID
	CreatedAt time.Time
	Metadata  map[string]string // Service-specific metadata
}

// String returns a human-readable node description.
func (n *Node) String() string {
	return n.ID.String()
}

// GraphConfig holds configuration for the RelationshipGraph.
type GraphConfig struct {
	// StrictValidation enforces schema rules when true.
	StrictValidation bool

	// DefaultDeleteBehavior specifies what happens when deleting nodes with dependents.
	DefaultDeleteBehavior DeleteBehavior

	// DetectCycles enables cycle detection when adding edges.
	DetectCycles bool
}

// DefaultGraphConfig returns a default configuration suitable for most use cases.
func DefaultGraphConfig() GraphConfig {
	return GraphConfig{
		StrictValidation:      false, // Lenient by default for backward compatibility
		DefaultDeleteBehavior: DeleteRestrict,
		DetectCycles:          true,
	}
}
