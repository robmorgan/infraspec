package iam

import (
	"log"

	"github.com/robmorgan/infraspec/internal/emulator/graph"
)

// registerResource registers a resource in the relationship graph if ResourceManager is available.
// This is a no-op if ResourceManager is nil.
func (s *IAMService) registerResource(resourceType, resourceID string, metadata map[string]string) {
	if s.resourceManager == nil {
		return
	}
	id := graph.ResourceID{
		Service: "iam",
		Type:    resourceType,
		ID:      resourceID,
	}
	if err := s.resourceManager.RegisterResource(id, metadata); err != nil {
		log.Printf("Warning: failed to register %s/%s in graph: %v", resourceType, resourceID, err)
	}
}

// unregisterResource removes a resource from the relationship graph if ResourceManager is available.
func (s *IAMService) unregisterResource(resourceType, resourceID string) error {
	if s.resourceManager == nil {
		return nil
	}
	id := graph.ResourceID{
		Service: "iam",
		Type:    resourceType,
		ID:      resourceID,
	}
	return s.resourceManager.UnregisterResource(id)
}

// addRelationship creates a relationship between two resources in the graph.
func (s *IAMService) addRelationship(fromType, fromID, toType, toID string, relType graph.RelationshipType) error {
	if s.resourceManager == nil {
		return nil
	}
	from := graph.ResourceID{Service: "iam", Type: fromType, ID: fromID}
	to := graph.ResourceID{Service: "iam", Type: toType, ID: toID}
	return s.resourceManager.AddRelationship(from, to, relType)
}

// removeRelationship removes a relationship between two resources in the graph.
func (s *IAMService) removeRelationship(fromType, fromID, toType, toID string, relType graph.RelationshipType) error {
	if s.resourceManager == nil {
		return nil
	}
	from := graph.ResourceID{Service: "iam", Type: fromType, ID: fromID}
	to := graph.ResourceID{Service: "iam", Type: toType, ID: toID}
	return s.resourceManager.RemoveRelationship(from, to, relType)
}

// isStrictMode returns true if graph validation is in strict mode.
// In strict mode, relationship failures should cause operations to fail and rollback.
func (s *IAMService) isStrictMode() bool {
	if s.resourceManager == nil {
		return false
	}
	return s.resourceManager.IsStrictMode()
}
