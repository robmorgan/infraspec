package graph

import (
	"log"
	"sync"
	"time"
)

// RelationshipGraph manages AWS resource relationships using an adjacency list representation.
// It provides efficient queries for dependencies and dependents, cycle detection, and
// configurable validation.
type RelationshipGraph struct {
	mu sync.RWMutex

	// nodes stores all registered resources: ResourceID.String() -> *Node
	nodes map[string]*Node

	// outEdges stores forward relationships: what this node depends on (points to)
	// Key: source ResourceID.String(), Value: slice of edges from this node
	outEdges map[string][]*Edge

	// inEdges stores reverse relationships: what depends on this node (points to it)
	// Key: target ResourceID.String(), Value: slice of edges to this node
	inEdges map[string][]*Edge

	// schema defines allowed relationships (optional)
	schema *RelationshipSchema

	// config holds graph configuration
	config GraphConfig
}

// NewRelationshipGraph creates a new graph instance with the given configuration.
func NewRelationshipGraph(config GraphConfig) *RelationshipGraph {
	return &RelationshipGraph{
		nodes:    make(map[string]*Node),
		outEdges: make(map[string][]*Edge),
		inEdges:  make(map[string][]*Edge),
		config:   config,
	}
}

// SetSchema sets the relationship schema for validation.
func (g *RelationshipGraph) SetSchema(schema *RelationshipSchema) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.schema = schema
}

// Lock acquires an exclusive lock on the graph.
// Use this when performing multiple operations that need to be atomic.
// Must be paired with Unlock().
func (g *RelationshipGraph) Lock() {
	g.mu.Lock()
}

// Unlock releases the exclusive lock on the graph.
func (g *RelationshipGraph) Unlock() {
	g.mu.Unlock()
}

// RLock acquires a read lock on the graph.
// Use this when performing multiple read operations that need to be consistent.
// Must be paired with RUnlock().
func (g *RelationshipGraph) RLock() {
	g.mu.RLock()
}

// RUnlock releases the read lock on the graph.
func (g *RelationshipGraph) RUnlock() {
	g.mu.RUnlock()
}

// GetSchema returns the current relationship schema.
func (g *RelationshipGraph) GetSchema() *RelationshipSchema {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.schema
}

// Config returns the current graph configuration.
func (g *RelationshipGraph) Config() GraphConfig {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.config
}

// SetConfig updates the graph configuration.
func (g *RelationshipGraph) SetConfig(config GraphConfig) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.config = config
}

// AddNode registers a resource in the graph.
func (g *RelationshipGraph) AddNode(id ResourceID, metadata map[string]string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.addNodeInternal(id, metadata)
}

// addNodeInternal registers a resource in the graph (assumes lock held).
// This is used by ResourceManager to avoid double-locking.
func (g *RelationshipGraph) addNodeInternal(id ResourceID, metadata map[string]string) error {
	key := id.String()
	if _, exists := g.nodes[key]; exists {
		return &NodeExistsError{Resource: id}
	}

	g.nodes[key] = &Node{
		ID:        id,
		CreatedAt: time.Now(),
		Metadata:  metadata,
	}

	// Initialize empty adjacency lists
	g.outEdges[key] = []*Edge{}
	g.inEdges[key] = []*Edge{}

	return nil
}

// RemoveNode removes a resource and its edges from the graph.
// Respects the specified delete behavior.
func (g *RelationshipGraph) RemoveNode(id ResourceID, behavior DeleteBehavior) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.removeNodeWithBehavior(id, behavior)
}

// removeNodeWithBehavior removes a node with specified behavior (assumes lock held).
// This is used by ResourceManager to avoid double-locking.
func (g *RelationshipGraph) removeNodeWithBehavior(id ResourceID, behavior DeleteBehavior) error {
	key := id.String()
	if _, exists := g.nodes[key]; !exists {
		return &NodeNotFoundError{Resource: id}
	}

	// Check for dependents (resources that have edges pointing TO this node)
	dependents := g.inEdges[key]

	switch behavior {
	case DeleteRestrict:
		if len(dependents) > 0 {
			depIDs := make([]ResourceID, len(dependents))
			for i, e := range dependents {
				depIDs[i] = e.From
			}
			return &DependencyError{
				Resource:   id,
				Dependents: depIDs,
			}
		}

	case DeleteCascade:
		// Recursively delete dependents first (dangerous!)
		// Make a copy to avoid modifying slice while iterating
		depsCopy := make([]*Edge, len(dependents))
		copy(depsCopy, dependents)
		for _, e := range depsCopy {
			if err := g.removeNodeInternal(e.From, DeleteCascade); err != nil {
				return err
			}
		}

	case DeleteSetNull:
		// Just remove the edges, don't delete dependents
		// The dependent resources remain but with broken references
	}

	return g.removeNodeInternal(id, behavior)
}

// removeNodeInternal removes a node and cleans up edges (assumes lock held).
func (g *RelationshipGraph) removeNodeInternal(id ResourceID, _ DeleteBehavior) error {
	key := id.String()

	// Remove all outgoing edges from this node
	for _, edge := range g.outEdges[key] {
		toKey := edge.To.String()
		g.inEdges[toKey] = removeEdgeFrom(g.inEdges[toKey], key)
	}

	// Remove all incoming edges to this node
	for _, edge := range g.inEdges[key] {
		fromKey := edge.From.String()
		g.outEdges[fromKey] = removeEdgeTo(g.outEdges[fromKey], key)
	}

	// Delete the node and its adjacency lists
	delete(g.nodes, key)
	delete(g.outEdges, key)
	delete(g.inEdges, key)

	return nil
}

// AddEdge creates a relationship between two resources.
func (g *RelationshipGraph) AddEdge(edge *Edge) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.addEdgeInternal(edge)
}

// addEdgeInternal creates a relationship between two resources (assumes lock held).
// This is used by ResourceManager to avoid double-locking.
func (g *RelationshipGraph) addEdgeInternal(edge *Edge) error {
	fromKey := edge.From.String()
	toKey := edge.To.String()

	// Validate nodes exist
	if _, exists := g.nodes[fromKey]; !exists {
		return &NodeNotFoundError{Resource: edge.From}
	}
	if _, exists := g.nodes[toKey]; !exists {
		return &NodeNotFoundError{Resource: edge.To}
	}

	// Schema validation (if schema exists)
	if g.schema != nil {
		if err := g.validateEdgeAgainstSchema(edge); err != nil {
			if g.config.StrictValidation {
				return err
			}
			// Log warning in lenient mode so validation issues are visible
			log.Printf("[WARN] Relationship validation failed (lenient mode, allowing): %v", err)
		}
	}

	// Cycle detection (if enabled)
	if g.config.DetectCycles {
		if g.wouldCreateCycle(edge) {
			return &CycleError{From: edge.From, To: edge.To}
		}
	}

	// Check for duplicate edge (idempotent operation)
	for _, e := range g.outEdges[fromKey] {
		if e.To.String() == toKey && e.Type == edge.Type {
			return nil // Edge already exists
		}
	}

	// Add to adjacency lists
	g.outEdges[fromKey] = append(g.outEdges[fromKey], edge)
	g.inEdges[toKey] = append(g.inEdges[toKey], edge)

	return nil
}

// RemoveEdge removes a relationship between two resources.
func (g *RelationshipGraph) RemoveEdge(from, to ResourceID, relType RelationshipType) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.removeEdgeInternal(from, to, relType)
}

// removeEdgeInternal removes a relationship between two resources (assumes lock held).
// This is used by ResourceManager to avoid double-locking.
func (g *RelationshipGraph) removeEdgeInternal(from, to ResourceID, relType RelationshipType) error {
	fromKey := from.String()
	toKey := to.String()

	// Remove from forward links
	g.outEdges[fromKey] = removeEdgeToWithType(g.outEdges[fromKey], toKey, relType)

	// Remove from reverse links
	g.inEdges[toKey] = removeEdgeFromWithType(g.inEdges[toKey], fromKey, relType)

	return nil
}

// GetNode returns a node by its ResourceID.
func (g *RelationshipGraph) GetNode(id ResourceID) (*Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.getNodeInternal(id)
}

// getNodeInternal returns a node by its ResourceID (assumes lock held).
// This is used by ResourceManager to avoid double-locking.
func (g *RelationshipGraph) getNodeInternal(id ResourceID) (*Node, error) {
	node, exists := g.nodes[id.String()]
	if !exists {
		return nil, &NodeNotFoundError{Resource: id}
	}
	return node, nil
}

// HasNode returns true if the node exists in the graph.
func (g *RelationshipGraph) HasNode(id ResourceID) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	_, exists := g.nodes[id.String()]
	return exists
}

// GetDependents returns all resources that directly depend on this one
// (resources that have edges pointing TO this node).
func (g *RelationshipGraph) GetDependents(id ResourceID) ([]*Edge, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.getDependentsInternal(id)
}

// getDependentsInternal returns all resources that directly depend on this one (assumes lock held).
// This is used by ResourceManager to avoid double-locking.
func (g *RelationshipGraph) getDependentsInternal(id ResourceID) ([]*Edge, error) {
	key := id.String()
	if _, exists := g.nodes[key]; !exists {
		return nil, &NodeNotFoundError{Resource: id}
	}

	// Return a copy to avoid race conditions
	edges := g.inEdges[key]
	result := make([]*Edge, len(edges))
	copy(result, edges)
	return result, nil
}

// GetDependencies returns all resources this one depends on
// (resources this node has edges pointing TO).
func (g *RelationshipGraph) GetDependencies(id ResourceID) ([]*Edge, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.getDependenciesInternal(id)
}

// getDependenciesInternal returns all resources this one depends on (assumes lock held).
// This is used by ResourceManager to avoid double-locking.
func (g *RelationshipGraph) getDependenciesInternal(id ResourceID) ([]*Edge, error) {
	key := id.String()
	if _, exists := g.nodes[key]; !exists {
		return nil, &NodeNotFoundError{Resource: id}
	}

	edges := g.outEdges[key]
	result := make([]*Edge, len(edges))
	copy(result, edges)
	return result, nil
}

// GetAllDependents returns the full dependency tree using BFS traversal.
// Returns all resources that would be affected if this resource is deleted.
func (g *RelationshipGraph) GetAllDependents(id ResourceID) ([]ResourceID, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	key := id.String()
	if _, exists := g.nodes[key]; !exists {
		return nil, &NodeNotFoundError{Resource: id}
	}

	visited := make(map[string]bool)
	result := []ResourceID{}
	queue := []string{key}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, edge := range g.inEdges[current] {
			fromKey := edge.From.String()
			if !visited[fromKey] {
				visited[fromKey] = true
				result = append(result, edge.From)
				queue = append(queue, fromKey)
			}
		}
	}

	return result, nil
}

// GetAllDependencies returns all resources this one transitively depends on using BFS.
func (g *RelationshipGraph) GetAllDependencies(id ResourceID) ([]ResourceID, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	key := id.String()
	if _, exists := g.nodes[key]; !exists {
		return nil, &NodeNotFoundError{Resource: id}
	}

	visited := make(map[string]bool)
	result := []ResourceID{}
	queue := []string{key}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, edge := range g.outEdges[current] {
			toKey := edge.To.String()
			if !visited[toKey] {
				visited[toKey] = true
				result = append(result, edge.To)
				queue = append(queue, toKey)
			}
		}
	}

	return result, nil
}

// CanDelete checks if a resource can be deleted based on relationships.
// Returns (canDelete, listOfDependentResourceIDs, error).
func (g *RelationshipGraph) CanDelete(id ResourceID) (bool, []ResourceID, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	key := id.String()
	if _, exists := g.nodes[key]; !exists {
		return false, nil, &NodeNotFoundError{Resource: id}
	}

	dependents := g.inEdges[key]
	if len(dependents) == 0 {
		return true, nil, nil
	}

	depList := make([]ResourceID, len(dependents))
	for i, e := range dependents {
		depList[i] = e.From
	}
	return false, depList, nil
}

// NodeCount returns the number of nodes in the graph.
func (g *RelationshipGraph) NodeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.nodes)
}

// EdgeCount returns the total number of edges in the graph.
func (g *RelationshipGraph) EdgeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	count := 0
	for _, edges := range g.outEdges {
		count += len(edges)
	}
	return count
}

// AllNodes returns all nodes in the graph.
func (g *RelationshipGraph) AllNodes() []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make([]*Node, 0, len(g.nodes))
	for _, node := range g.nodes {
		result = append(result, node)
	}
	return result
}

// AllEdges returns all edges in the graph.
func (g *RelationshipGraph) AllEdges() []*Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := []*Edge{}
	for _, edges := range g.outEdges {
		result = append(result, edges...)
	}
	return result
}

// wouldCreateCycle checks if adding an edge would create a cycle using DFS.
// Assumes lock is already held.
func (g *RelationshipGraph) wouldCreateCycle(edge *Edge) bool {
	// Check if we can reach edge.From starting from edge.To (would mean cycle)
	visited := make(map[string]bool)
	return g.canReach(edge.To.String(), edge.From.String(), visited)
}

// canReach performs DFS to check reachability (internal, assumes lock held).
func (g *RelationshipGraph) canReach(from, target string, visited map[string]bool) bool {
	if from == target {
		return true
	}
	if visited[from] {
		return false
	}
	visited[from] = true

	for _, edge := range g.outEdges[from] {
		if g.canReach(edge.To.String(), target, visited) {
			return true
		}
	}
	return false
}

// validateEdgeAgainstSchema validates an edge against the schema (assumes lock held).
func (g *RelationshipGraph) validateEdgeAgainstSchema(edge *Edge) error {
	if g.schema == nil {
		return nil
	}

	relKey := edge.From.TypeKey() + " -> " + edge.To.TypeKey()
	entry, exists := g.schema.Relationships[relKey]
	if !exists {
		return &SchemaValidationError{
			Relationship: relKey,
			Message:      "relationship not defined in schema: " + relKey,
		}
	}

	// Validate relationship type matches
	if entry.Type != edge.Type {
		return &SchemaValidationError{
			Relationship: relKey,
			Message:      "relationship type mismatch: expected " + string(entry.Type) + ", got " + string(edge.Type),
		}
	}

	// Validate cardinality constraints
	if err := g.checkCardinalityForLink(entry, edge); err != nil {
		return err
	}

	return nil
}

// checkCardinalityForLink checks if adding an edge would violate cardinality constraints.
func (g *RelationshipGraph) checkCardinalityForLink(entry SchemaEntry, edge *Edge) error {
	fromKey := edge.From.String()
	toKey := edge.To.String()
	relKey := edge.From.TypeKey() + " -> " + edge.To.TypeKey()

	switch entry.Cardinality {
	case CardOneToOne:
		// Source can have only one target
		for _, e := range g.outEdges[fromKey] {
			if e.Type == edge.Type && e.To.TypeKey() == edge.To.TypeKey() {
				return &CardinalityError{
					Relationship: relKey,
					Expected:     CardOneToOne,
					Message:      edge.From.String() + " already has a linked " + edge.To.TypeKey(),
				}
			}
		}
		// Target can have only one source
		for _, e := range g.inEdges[toKey] {
			if e.Type == edge.Type && e.From.TypeKey() == edge.From.TypeKey() {
				return &CardinalityError{
					Relationship: relKey,
					Expected:     CardOneToOne,
					Message:      edge.To.String() + " is already linked from another " + edge.From.TypeKey(),
				}
			}
		}

	case CardManyToOne:
		// Source can have only one target of this type
		for _, e := range g.outEdges[fromKey] {
			if e.Type == edge.Type && e.To.TypeKey() == edge.To.TypeKey() {
				return &CardinalityError{
					Relationship: relKey,
					Expected:     CardManyToOne,
					Message:      edge.From.String() + " already has a linked " + edge.To.TypeKey(),
				}
			}
		}

	case CardOneToMany:
		// Target can have only one source of this type
		for _, e := range g.inEdges[toKey] {
			if e.Type == edge.Type && e.From.TypeKey() == edge.From.TypeKey() {
				return &CardinalityError{
					Relationship: relKey,
					Expected:     CardOneToMany,
					Message:      edge.To.String() + " is already linked from another " + edge.From.TypeKey(),
				}
			}
		}

	case CardManyToMany:
		// No cardinality constraints
	}

	return nil
}

// Helper functions for edge list manipulation

func removeEdgeFrom(edges []*Edge, fromKey string) []*Edge {
	result := make([]*Edge, 0, len(edges))
	for _, e := range edges {
		if e.From.String() != fromKey {
			result = append(result, e)
		}
	}
	return result
}

func removeEdgeTo(edges []*Edge, toKey string) []*Edge {
	result := make([]*Edge, 0, len(edges))
	for _, e := range edges {
		if e.To.String() != toKey {
			result = append(result, e)
		}
	}
	return result
}

func removeEdgeToWithType(edges []*Edge, toKey string, relType RelationshipType) []*Edge {
	result := make([]*Edge, 0, len(edges))
	for _, e := range edges {
		if !(e.To.String() == toKey && e.Type == relType) {
			result = append(result, e)
		}
	}
	return result
}

func removeEdgeFromWithType(edges []*Edge, fromKey string, relType RelationshipType) []*Edge {
	result := make([]*Edge, 0, len(edges))
	for _, e := range edges {
		if !(e.From.String() == fromKey && e.Type == relType) {
			result = append(result, e)
		}
	}
	return result
}
