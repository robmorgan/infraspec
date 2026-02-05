package graph

import (
	"fmt"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// ResourceManager wraps StateManager and RelationshipGraph to provide
// atomic operations that maintain consistency between resource data and relationships.
// ResourceManager uses the graph's internal lock for synchronization, avoiding double-locking.
type ResourceManager struct {
	state  emulator.StateManager
	graph  *RelationshipGraph
	config ResourceManagerConfig
}

// ResourceManagerConfig holds configuration for the ResourceManager.
type ResourceManagerConfig struct {
	// StrictValidation enforces schema rules when true.
	StrictValidation bool

	// DefaultDeleteBehavior specifies what happens when deleting nodes with dependents.
	DefaultDeleteBehavior DeleteBehavior

	// DetectCycles enables cycle detection when adding edges.
	DetectCycles bool

	// UseAWSSchema loads the pre-defined AWS relationship schema.
	UseAWSSchema bool
}

// DefaultResourceManagerConfig returns a default configuration.
func DefaultResourceManagerConfig() ResourceManagerConfig {
	return ResourceManagerConfig{
		StrictValidation:      false, // Lenient by default
		DefaultDeleteBehavior: DeleteRestrict,
		DetectCycles:          true,
		UseAWSSchema:          true,
	}
}

// NewResourceManager creates a new ResourceManager wrapping the given StateManager.
func NewResourceManager(state emulator.StateManager, config ResourceManagerConfig) *ResourceManager {
	graphConfig := GraphConfig{
		StrictValidation:      config.StrictValidation,
		DefaultDeleteBehavior: config.DefaultDeleteBehavior,
		DetectCycles:          config.DetectCycles,
	}

	graph := NewRelationshipGraph(graphConfig)

	// Load AWS schema if configured
	if config.UseAWSSchema {
		graph.SetSchema(NewAWSSchema())
	}

	return &ResourceManager{
		state:  state,
		graph:  graph,
		config: config,
	}
}

// Graph returns the underlying RelationshipGraph for read-only queries.
//
// WARNING: Calling mutating methods (AddNode, RemoveNode, AddEdge, RemoveEdge)
// directly on the graph bypasses ResourceManager's mutex and can cause race
// conditions with ResourceManager operations like CreateResource/DeleteResource.
//
// Safe operations: GetNode, GetDependents, GetDependencies, HasNode, NodeCount, EdgeCount
// Unsafe operations: AddNode, RemoveNode, AddEdge, RemoveEdge (use ResourceManager methods instead)
func (rm *ResourceManager) Graph() *RelationshipGraph {
	return rm.graph
}

// State returns the underlying StateManager for direct access.
// Use with caution - prefer using ResourceManager methods for consistency.
func (rm *ResourceManager) State() emulator.StateManager {
	return rm.state
}

// RegisterResource registers a resource in the graph without storing data.
// Use this when the resource data is already stored in the state manager.
func (rm *ResourceManager) RegisterResource(id ResourceID, metadata map[string]string) error {
	rm.graph.Lock()
	defer rm.graph.Unlock()
	return rm.graph.addNodeInternal(id, metadata)
}

// UnregisterResource removes a resource from the graph without deleting state data.
// Use this when the resource data is managed separately from the graph.
// Returns an error if the resource has dependents (DeleteRestrict behavior).
func (rm *ResourceManager) UnregisterResource(id ResourceID) error {
	rm.graph.Lock()
	defer rm.graph.Unlock()
	return rm.graph.removeNodeWithBehavior(id, rm.config.DefaultDeleteBehavior)
}

// CreateResource creates a resource in the state manager and registers it in the graph.
// This is an atomic operation - if either step fails, neither change is committed.
func (rm *ResourceManager) CreateResource(id ResourceID, stateKey string, data interface{}) error {
	rm.graph.Lock()
	defer rm.graph.Unlock()

	// Store in state manager first
	if err := rm.state.Set(stateKey, data); err != nil {
		return fmt.Errorf("failed to store resource data: %w", err)
	}

	// Register in graph - addNodeInternal returns NodeExistsError if already exists
	if err := rm.graph.addNodeInternal(id, nil); err != nil {
		// Rollback state change
		rm.state.Delete(stateKey)
		return err
	}

	return nil
}

// DeleteResource removes a resource from the state manager and graph.
// Returns an error if the resource has dependents (DeleteRestrict behavior).
func (rm *ResourceManager) DeleteResource(id ResourceID, stateKey string) error {
	rm.graph.Lock()
	defer rm.graph.Unlock()

	// Get node metadata before removal for potential rollback
	node, err := rm.graph.getNodeInternal(id)
	if err != nil {
		return err
	}
	nodeMetadata := node.Metadata

	// Store edges before removal for complete rollback
	// inEdges: edges FROM other nodes TO this node (dependents)
	// outEdges: edges FROM this node TO other nodes (dependencies)
	inEdges, _ := rm.graph.getDependentsInternal(id)
	outEdges, _ := rm.graph.getDependenciesInternal(id)

	// Remove from graph first (this also cleans up edges)
	// removeNodeWithBehavior checks for dependents when using DeleteRestrict behavior
	// and returns DependencyError if any exist - no separate CanDelete check needed
	if err := rm.graph.removeNodeWithBehavior(id, rm.config.DefaultDeleteBehavior); err != nil {
		return err
	}

	// Delete from state manager
	if err := rm.state.Delete(stateKey); err != nil {
		// Attempt to rollback the graph change to maintain consistency
		// First restore the node
		if rollbackErr := rm.graph.addNodeInternal(id, nodeMetadata); rollbackErr != nil {
			return fmt.Errorf("failed to delete resource data and rollback failed: delete error: %w, rollback error: %v", err, rollbackErr)
		}

		// Restore all edges (both incoming and outgoing)
		for _, edge := range inEdges {
			rm.graph.addEdgeInternal(edge)
		}
		for _, edge := range outEdges {
			rm.graph.addEdgeInternal(edge)
		}

		return fmt.Errorf("failed to delete resource data (graph restored): %w", err)
	}

	return nil
}

// AddRelationship creates a relationship between two resources.
// Both resources must already exist in the graph.
func (rm *ResourceManager) AddRelationship(from, to ResourceID, relType RelationshipType) error {
	rm.graph.Lock()
	defer rm.graph.Unlock()

	edge := &Edge{
		From: from,
		To:   to,
		Type: relType,
	}

	return rm.graph.addEdgeInternal(edge)
}

// AddRelationshipWithMetadata creates a relationship with additional metadata.
func (rm *ResourceManager) AddRelationshipWithMetadata(from, to ResourceID, relType RelationshipType, metadata map[string]string) error {
	rm.graph.Lock()
	defer rm.graph.Unlock()

	edge := &Edge{
		From:     from,
		To:       to,
		Type:     relType,
		Metadata: metadata,
	}

	return rm.graph.addEdgeInternal(edge)
}

// RemoveRelationship removes a relationship between two resources.
func (rm *ResourceManager) RemoveRelationship(from, to ResourceID, relType RelationshipType) error {
	rm.graph.Lock()
	defer rm.graph.Unlock()
	return rm.graph.removeEdgeInternal(from, to, relType)
}

// CanDelete checks if a resource can be deleted based on its relationships.
// Returns (canDelete, listOfDependentResourceIDs, error).
func (rm *ResourceManager) CanDelete(id ResourceID) (bool, []ResourceID, error) {
	return rm.graph.CanDelete(id)
}

// GetDependents returns all resources that directly depend on this one.
func (rm *ResourceManager) GetDependents(id ResourceID) ([]ResourceID, error) {
	edges, err := rm.graph.GetDependents(id)
	if err != nil {
		return nil, err
	}

	result := make([]ResourceID, len(edges))
	for i, e := range edges {
		result[i] = e.From
	}
	return result, nil
}

// GetDependencies returns all resources this one directly depends on.
func (rm *ResourceManager) GetDependencies(id ResourceID) ([]ResourceID, error) {
	edges, err := rm.graph.GetDependencies(id)
	if err != nil {
		return nil, err
	}

	result := make([]ResourceID, len(edges))
	for i, e := range edges {
		result[i] = e.To
	}
	return result, nil
}

// GetAllDependents returns the full dependency tree (all resources affected by deletion).
func (rm *ResourceManager) GetAllDependents(id ResourceID) ([]ResourceID, error) {
	return rm.graph.GetAllDependents(id)
}

// GetAllDependencies returns all resources this one transitively depends on.
func (rm *ResourceManager) GetAllDependencies(id ResourceID) ([]ResourceID, error) {
	return rm.graph.GetAllDependencies(id)
}

// HasResource checks if a resource exists in the graph.
func (rm *ResourceManager) HasResource(id ResourceID) bool {
	return rm.graph.HasNode(id)
}

// GetResource returns a resource node by its ID.
func (rm *ResourceManager) GetResource(id ResourceID) (*Node, error) {
	return rm.graph.GetNode(id)
}

// ResourceCount returns the number of resources in the graph.
func (rm *ResourceManager) ResourceCount() int {
	return rm.graph.NodeCount()
}

// RelationshipCount returns the number of relationships in the graph.
func (rm *ResourceManager) RelationshipCount() int {
	return rm.graph.EdgeCount()
}

// SetValidationMode changes the validation mode at runtime.
func (rm *ResourceManager) SetValidationMode(strict bool) {
	rm.graph.Lock()
	defer rm.graph.Unlock()

	rm.config.StrictValidation = strict

	// Update the graph's config (note: we're already holding the lock, so access config directly)
	rm.graph.config.StrictValidation = strict
}

// IsStrictMode returns true if strict validation is enabled.
// In strict mode, relationship failures should cause operations to fail and rollback.
func (rm *ResourceManager) IsStrictMode() bool {
	rm.graph.RLock()
	defer rm.graph.RUnlock()
	return rm.config.StrictValidation
}

// SetSchema sets a custom relationship schema.
func (rm *ResourceManager) SetSchema(schema *RelationshipSchema) {
	rm.graph.SetSchema(schema)
}

// GetSchema returns the current relationship schema.
func (rm *ResourceManager) GetSchema() *RelationshipSchema {
	return rm.graph.GetSchema()
}

// ValidateRelationship checks if a relationship would be valid according to the schema.
func (rm *ResourceManager) ValidateRelationship(from, to ResourceID, relType RelationshipType) error {
	schema := rm.graph.GetSchema()
	if schema == nil {
		return nil // No schema, all relationships are valid
	}
	return schema.ValidateRelationship(from, to, relType)
}
