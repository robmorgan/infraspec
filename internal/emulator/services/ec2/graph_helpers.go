package ec2

import (
	"log"

	"github.com/robmorgan/infraspec/internal/emulator/graph"
)

// ==================== Resource Graph Helper Functions ====================

// registerResource registers a resource in the relationship graph if ResourceManager is available.
// This is a no-op if ResourceManager is nil.
func (s *EC2Service) registerResource(resourceType, resourceID string, metadata map[string]string) {
	if s.resourceManager == nil {
		return
	}
	id := graph.ResourceID{
		Service: "ec2",
		Type:    resourceType,
		ID:      resourceID,
	}
	if err := s.resourceManager.RegisterResource(id, metadata); err != nil {
		log.Printf("Warning: failed to register %s/%s in graph: %v", resourceType, resourceID, err)
	}
}

// unregisterResource removes a resource from the relationship graph if ResourceManager is available.
func (s *EC2Service) unregisterResource(resourceType, resourceID string) error {
	if s.resourceManager == nil {
		return nil
	}
	id := graph.ResourceID{
		Service: "ec2",
		Type:    resourceType,
		ID:      resourceID,
	}
	return s.resourceManager.UnregisterResource(id)
}

// addRelationship creates a relationship between two resources in the graph.
func (s *EC2Service) addRelationship(fromType, fromID, toService, toType, toID string, relType graph.RelationshipType) error {
	if s.resourceManager == nil {
		return nil
	}
	from := graph.ResourceID{Service: "ec2", Type: fromType, ID: fromID}
	to := graph.ResourceID{Service: toService, Type: toType, ID: toID}
	return s.resourceManager.AddRelationship(from, to, relType)
}

// removeRelationship removes a relationship between two resources in the graph.
func (s *EC2Service) removeRelationship(fromType, fromID, toService, toType, toID string, relType graph.RelationshipType) error {
	if s.resourceManager == nil {
		return nil
	}
	from := graph.ResourceID{Service: "ec2", Type: fromType, ID: fromID}
	to := graph.ResourceID{Service: toService, Type: toType, ID: toID}
	return s.resourceManager.RemoveRelationship(from, to, relType)
}

// canDeleteResource checks if a resource can be deleted based on its relationships.
func (s *EC2Service) canDeleteResource(resourceType, resourceID string) (bool, []graph.ResourceID) {
	if s.resourceManager == nil {
		return true, nil
	}
	id := graph.ResourceID{
		Service: "ec2",
		Type:    resourceType,
		ID:      resourceID,
	}
	canDelete, dependents, _ := s.resourceManager.CanDelete(id)
	return canDelete, dependents
}

// isStrictMode returns true if graph validation is in strict mode.
// In strict mode, relationship failures should cause operations to fail and rollback.
func (s *EC2Service) isStrictMode() bool {
	if s.resourceManager == nil {
		return false
	}
	return s.resourceManager.IsStrictMode()
}
