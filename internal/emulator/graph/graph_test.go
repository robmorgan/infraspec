package graph

import (
	"fmt"
	"testing"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceID_String(t *testing.T) {
	id := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-12345"}
	assert.Equal(t, "ec2:vpc:vpc-12345", id.String())
}

func TestResourceID_TypeKey(t *testing.T) {
	id := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-12345"}
	assert.Equal(t, "ec2:vpc", id.TypeKey())
}

func TestResourceID_IsZero(t *testing.T) {
	assert.True(t, ResourceID{}.IsZero())
	assert.False(t, ResourceID{Service: "ec2"}.IsZero())
}

func TestGraph_AddNode(t *testing.T) {
	g := NewRelationshipGraph(DefaultGraphConfig())

	id := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-12345"}
	err := g.AddNode(id, nil)
	require.NoError(t, err)

	// Verify node exists
	assert.True(t, g.HasNode(id))
	assert.Equal(t, 1, g.NodeCount())

	// Duplicate should fail
	err = g.AddNode(id, nil)
	assert.Error(t, err)
	assert.True(t, IsNodeExistsError(err))
}

func TestGraph_AddNode_WithMetadata(t *testing.T) {
	g := NewRelationshipGraph(DefaultGraphConfig())

	id := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-12345"}
	metadata := map[string]string{"name": "my-vpc", "region": "us-east-1"}
	err := g.AddNode(id, metadata)
	require.NoError(t, err)

	node, err := g.GetNode(id)
	require.NoError(t, err)
	assert.Equal(t, "my-vpc", node.Metadata["name"])
	assert.Equal(t, "us-east-1", node.Metadata["region"])
}

func TestGraph_RemoveNode(t *testing.T) {
	g := NewRelationshipGraph(DefaultGraphConfig())

	id := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-12345"}
	err := g.AddNode(id, nil)
	require.NoError(t, err)

	err = g.RemoveNode(id, DeleteRestrict)
	require.NoError(t, err)

	assert.False(t, g.HasNode(id))
	assert.Equal(t, 0, g.NodeCount())
}

func TestGraph_RemoveNode_NotFound(t *testing.T) {
	g := NewRelationshipGraph(DefaultGraphConfig())

	id := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-12345"}
	err := g.RemoveNode(id, DeleteRestrict)
	assert.Error(t, err)
	assert.True(t, IsNodeNotFoundError(err))
}

func TestGraph_AddEdge(t *testing.T) {
	g := NewRelationshipGraph(DefaultGraphConfig())

	vpc := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-12345"}
	subnet := ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-67890"}

	require.NoError(t, g.AddNode(vpc, nil))
	require.NoError(t, g.AddNode(subnet, nil))

	edge := &Edge{From: subnet, To: vpc, Type: RelContains}
	err := g.AddEdge(edge)
	require.NoError(t, err)

	assert.Equal(t, 1, g.EdgeCount())

	// Verify forward and reverse lookups
	deps, err := g.GetDependencies(subnet)
	require.NoError(t, err)
	assert.Len(t, deps, 1)
	assert.Equal(t, vpc, deps[0].To)

	dependents, err := g.GetDependents(vpc)
	require.NoError(t, err)
	assert.Len(t, dependents, 1)
	assert.Equal(t, subnet, dependents[0].From)
}

func TestGraph_AddEdge_Idempotent(t *testing.T) {
	g := NewRelationshipGraph(DefaultGraphConfig())

	vpc := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-12345"}
	subnet := ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-67890"}

	require.NoError(t, g.AddNode(vpc, nil))
	require.NoError(t, g.AddNode(subnet, nil))

	edge := &Edge{From: subnet, To: vpc, Type: RelContains}

	// First addition
	require.NoError(t, g.AddEdge(edge))
	assert.Equal(t, 1, g.EdgeCount())

	// Second addition (same edge) should be idempotent
	require.NoError(t, g.AddEdge(edge))
	assert.Equal(t, 1, g.EdgeCount())
}

func TestGraph_AddEdge_NodeNotFound(t *testing.T) {
	g := NewRelationshipGraph(DefaultGraphConfig())

	vpc := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-12345"}
	subnet := ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-67890"}

	// Only add VPC, not subnet
	require.NoError(t, g.AddNode(vpc, nil))

	edge := &Edge{From: subnet, To: vpc, Type: RelContains}
	err := g.AddEdge(edge)
	assert.Error(t, err)
	assert.True(t, IsNodeNotFoundError(err))
}

func TestGraph_RemoveEdge(t *testing.T) {
	g := NewRelationshipGraph(DefaultGraphConfig())

	vpc := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-12345"}
	subnet := ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-67890"}

	require.NoError(t, g.AddNode(vpc, nil))
	require.NoError(t, g.AddNode(subnet, nil))

	edge := &Edge{From: subnet, To: vpc, Type: RelContains}
	require.NoError(t, g.AddEdge(edge))
	assert.Equal(t, 1, g.EdgeCount())

	err := g.RemoveEdge(subnet, vpc, RelContains)
	require.NoError(t, err)
	assert.Equal(t, 0, g.EdgeCount())
}

func TestGraph_CycleDetection(t *testing.T) {
	config := DefaultGraphConfig()
	config.DetectCycles = true
	g := NewRelationshipGraph(config)

	a := ResourceID{Service: "test", Type: "node", ID: "a"}
	b := ResourceID{Service: "test", Type: "node", ID: "b"}
	c := ResourceID{Service: "test", Type: "node", ID: "c"}

	require.NoError(t, g.AddNode(a, nil))
	require.NoError(t, g.AddNode(b, nil))
	require.NoError(t, g.AddNode(c, nil))

	// a -> b -> c
	require.NoError(t, g.AddEdge(&Edge{From: a, To: b, Type: RelReferences}))
	require.NoError(t, g.AddEdge(&Edge{From: b, To: c, Type: RelReferences}))

	// c -> a would create a cycle
	err := g.AddEdge(&Edge{From: c, To: a, Type: RelReferences})
	assert.Error(t, err)
	assert.True(t, IsCycleError(err))
}

func TestGraph_CycleDetection_Disabled(t *testing.T) {
	config := DefaultGraphConfig()
	config.DetectCycles = false
	g := NewRelationshipGraph(config)

	a := ResourceID{Service: "test", Type: "node", ID: "a"}
	b := ResourceID{Service: "test", Type: "node", ID: "b"}

	require.NoError(t, g.AddNode(a, nil))
	require.NoError(t, g.AddNode(b, nil))

	// a -> b -> a (cycle, but detection is disabled)
	require.NoError(t, g.AddEdge(&Edge{From: a, To: b, Type: RelReferences}))
	require.NoError(t, g.AddEdge(&Edge{From: b, To: a, Type: RelReferences}))
}

func TestGraph_RemoveNode_WithDependents_Restrict(t *testing.T) {
	g := NewRelationshipGraph(DefaultGraphConfig())

	vpc := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-12345"}
	subnet := ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-67890"}

	require.NoError(t, g.AddNode(vpc, nil))
	require.NoError(t, g.AddNode(subnet, nil))
	require.NoError(t, g.AddEdge(&Edge{From: subnet, To: vpc, Type: RelContains}))

	// Try to delete VPC (has dependent subnet)
	err := g.RemoveNode(vpc, DeleteRestrict)
	assert.Error(t, err)
	assert.True(t, IsDependencyError(err))

	depErr := err.(*DependencyError)
	assert.Len(t, depErr.Dependents, 1)
	assert.Equal(t, subnet, depErr.Dependents[0])
}

func TestGraph_RemoveNode_WithDependents_Cascade(t *testing.T) {
	g := NewRelationshipGraph(DefaultGraphConfig())

	vpc := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-12345"}
	subnet := ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-67890"}

	require.NoError(t, g.AddNode(vpc, nil))
	require.NoError(t, g.AddNode(subnet, nil))
	require.NoError(t, g.AddEdge(&Edge{From: subnet, To: vpc, Type: RelContains}))

	// Delete VPC with cascade (should delete subnet too)
	err := g.RemoveNode(vpc, DeleteCascade)
	require.NoError(t, err)

	assert.False(t, g.HasNode(vpc))
	assert.False(t, g.HasNode(subnet))
	assert.Equal(t, 0, g.NodeCount())
}

func TestGraph_RemoveNode_WithDependents_SetNull(t *testing.T) {
	g := NewRelationshipGraph(DefaultGraphConfig())

	vpc := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-12345"}
	subnet := ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-67890"}

	require.NoError(t, g.AddNode(vpc, nil))
	require.NoError(t, g.AddNode(subnet, nil))
	require.NoError(t, g.AddEdge(&Edge{From: subnet, To: vpc, Type: RelContains}))

	// Delete VPC with SetNull (removes edge but keeps subnet)
	err := g.RemoveNode(vpc, DeleteSetNull)
	require.NoError(t, err)

	assert.False(t, g.HasNode(vpc))
	assert.True(t, g.HasNode(subnet))

	// Subnet should have no dependencies now
	deps, err := g.GetDependencies(subnet)
	require.NoError(t, err)
	assert.Len(t, deps, 0)
}

func TestGraph_GetAllDependents_BFS(t *testing.T) {
	g := NewRelationshipGraph(DefaultGraphConfig())

	// Create hierarchy: vpc <- subnet <- instance
	vpc := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-1"}
	subnet := ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-1"}
	instance := ResourceID{Service: "ec2", Type: "instance", ID: "i-1"}

	require.NoError(t, g.AddNode(vpc, nil))
	require.NoError(t, g.AddNode(subnet, nil))
	require.NoError(t, g.AddNode(instance, nil))

	require.NoError(t, g.AddEdge(&Edge{From: subnet, To: vpc, Type: RelContains}))
	require.NoError(t, g.AddEdge(&Edge{From: instance, To: subnet, Type: RelReferences}))

	// Get all dependents of VPC (should include subnet and instance)
	allDeps, err := g.GetAllDependents(vpc)
	require.NoError(t, err)
	assert.Len(t, allDeps, 2)

	// Verify both subnet and instance are in the result
	ids := make(map[string]bool)
	for _, d := range allDeps {
		ids[d.String()] = true
	}
	assert.True(t, ids[subnet.String()])
	assert.True(t, ids[instance.String()])
}

func TestGraph_GetAllDependencies_BFS(t *testing.T) {
	g := NewRelationshipGraph(DefaultGraphConfig())

	// Create hierarchy: vpc <- subnet <- instance
	vpc := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-1"}
	subnet := ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-1"}
	instance := ResourceID{Service: "ec2", Type: "instance", ID: "i-1"}

	require.NoError(t, g.AddNode(vpc, nil))
	require.NoError(t, g.AddNode(subnet, nil))
	require.NoError(t, g.AddNode(instance, nil))

	require.NoError(t, g.AddEdge(&Edge{From: subnet, To: vpc, Type: RelContains}))
	require.NoError(t, g.AddEdge(&Edge{From: instance, To: subnet, Type: RelReferences}))

	// Get all dependencies of instance (should include subnet and vpc)
	allDeps, err := g.GetAllDependencies(instance)
	require.NoError(t, err)
	assert.Len(t, allDeps, 2)

	ids := make(map[string]bool)
	for _, d := range allDeps {
		ids[d.String()] = true
	}
	assert.True(t, ids[subnet.String()])
	assert.True(t, ids[vpc.String()])
}

func TestGraph_CanDelete(t *testing.T) {
	g := NewRelationshipGraph(DefaultGraphConfig())

	vpc := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-12345"}
	subnet := ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-67890"}

	require.NoError(t, g.AddNode(vpc, nil))
	require.NoError(t, g.AddNode(subnet, nil))
	require.NoError(t, g.AddEdge(&Edge{From: subnet, To: vpc, Type: RelContains}))

	// Subnet can be deleted (no dependents)
	canDelete, deps, err := g.CanDelete(subnet)
	require.NoError(t, err)
	assert.True(t, canDelete)
	assert.Nil(t, deps)

	// VPC cannot be deleted (has dependent)
	canDelete, deps, err = g.CanDelete(vpc)
	require.NoError(t, err)
	assert.False(t, canDelete)
	assert.Len(t, deps, 1)
	assert.Equal(t, subnet, deps[0])
}

func TestGraph_SchemaValidation_Strict(t *testing.T) {
	config := DefaultGraphConfig()
	config.StrictValidation = true
	g := NewRelationshipGraph(config)
	g.SetSchema(NewAWSSchema())

	vpc := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-12345"}
	subnet := ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-67890"}

	require.NoError(t, g.AddNode(vpc, nil))
	require.NoError(t, g.AddNode(subnet, nil))

	// Valid relationship per schema
	err := g.AddEdge(&Edge{From: subnet, To: vpc, Type: RelContains})
	require.NoError(t, err)

	// Invalid relationship type (schema says RelContains, not RelReferences)
	// Actually, let's test an undefined relationship
	unknown := ResourceID{Service: "unknown", Type: "thing", ID: "thing-1"}
	require.NoError(t, g.AddNode(unknown, nil))

	err = g.AddEdge(&Edge{From: unknown, To: vpc, Type: RelReferences})
	assert.Error(t, err)
	assert.True(t, IsSchemaValidationError(err))
}

func TestGraph_Cardinality_OneToOne(t *testing.T) {
	config := DefaultGraphConfig()
	config.StrictValidation = true
	g := NewRelationshipGraph(config)
	g.SetSchema(NewAWSSchema())

	vpc := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-1"}
	igw1 := ResourceID{Service: "ec2", Type: "internet-gateway", ID: "igw-1"}
	igw2 := ResourceID{Service: "ec2", Type: "internet-gateway", ID: "igw-2"}

	require.NoError(t, g.AddNode(vpc, nil))
	require.NoError(t, g.AddNode(igw1, nil))
	require.NoError(t, g.AddNode(igw2, nil))

	// First IGW attachment works
	err := g.AddEdge(&Edge{From: igw1, To: vpc, Type: RelAttachedTo})
	require.NoError(t, err)

	// Second IGW to same VPC should fail (one-to-one cardinality)
	err = g.AddEdge(&Edge{From: igw2, To: vpc, Type: RelAttachedTo})
	assert.Error(t, err)
	assert.True(t, IsCardinalityError(err))
}

func TestGraph_Cardinality_ManyToOne(t *testing.T) {
	config := DefaultGraphConfig()
	config.StrictValidation = true
	g := NewRelationshipGraph(config)
	g.SetSchema(NewAWSSchema())

	vpc := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-1"}
	subnet1 := ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-1"}
	subnet2 := ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-2"}

	require.NoError(t, g.AddNode(vpc, nil))
	require.NoError(t, g.AddNode(subnet1, nil))
	require.NoError(t, g.AddNode(subnet2, nil))

	// Multiple subnets can reference the same VPC (many-to-one)
	require.NoError(t, g.AddEdge(&Edge{From: subnet1, To: vpc, Type: RelContains}))
	require.NoError(t, g.AddEdge(&Edge{From: subnet2, To: vpc, Type: RelContains}))

	assert.Equal(t, 2, g.EdgeCount())
}

func TestGraph_Cardinality_ManyToMany(t *testing.T) {
	config := DefaultGraphConfig()
	config.StrictValidation = true
	g := NewRelationshipGraph(config)
	g.SetSchema(NewAWSSchema())

	role := ResourceID{Service: "iam", Type: "role", ID: "role-1"}
	policy1 := ResourceID{Service: "iam", Type: "policy", ID: "policy-1"}
	policy2 := ResourceID{Service: "iam", Type: "policy", ID: "policy-2"}

	require.NoError(t, g.AddNode(role, nil))
	require.NoError(t, g.AddNode(policy1, nil))
	require.NoError(t, g.AddNode(policy2, nil))

	// Policy attachments point to roles (many-to-many)
	// Edge direction: policy -> role means deleting the role fails if policies are attached
	require.NoError(t, g.AddEdge(&Edge{From: policy1, To: role, Type: RelAssociatedWith}))
	require.NoError(t, g.AddEdge(&Edge{From: policy2, To: role, Type: RelAssociatedWith}))

	assert.Equal(t, 2, g.EdgeCount())
}

func TestGraph_AllNodes(t *testing.T) {
	g := NewRelationshipGraph(DefaultGraphConfig())

	vpc := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-1"}
	subnet := ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-1"}

	require.NoError(t, g.AddNode(vpc, nil))
	require.NoError(t, g.AddNode(subnet, nil))

	nodes := g.AllNodes()
	assert.Len(t, nodes, 2)
}

func TestGraph_AllEdges(t *testing.T) {
	g := NewRelationshipGraph(DefaultGraphConfig())

	vpc := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-1"}
	subnet1 := ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-1"}
	subnet2 := ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-2"}

	require.NoError(t, g.AddNode(vpc, nil))
	require.NoError(t, g.AddNode(subnet1, nil))
	require.NoError(t, g.AddNode(subnet2, nil))

	require.NoError(t, g.AddEdge(&Edge{From: subnet1, To: vpc, Type: RelContains}))
	require.NoError(t, g.AddEdge(&Edge{From: subnet2, To: vpc, Type: RelContains}))

	edges := g.AllEdges()
	assert.Len(t, edges, 2)
}

func TestSchema_AddAndGetRelationship(t *testing.T) {
	schema := NewRelationshipSchema()

	schema.AddRelationship("ec2", "subnet", "ec2", "vpc", SchemaEntry{
		Type:           RelContains,
		Cardinality:    CardManyToOne,
		DeleteBehavior: DeleteRestrict,
		Required:       true,
	})

	entry, exists := schema.GetRelationship("ec2", "subnet", "ec2", "vpc")
	assert.True(t, exists)
	assert.Equal(t, RelContains, entry.Type)
	assert.Equal(t, CardManyToOne, entry.Cardinality)
	assert.Equal(t, DeleteRestrict, entry.DeleteBehavior)
	assert.True(t, entry.Required)

	// Non-existent relationship
	_, exists = schema.GetRelationship("ec2", "instance", "s3", "bucket")
	assert.False(t, exists)
}

func TestSchema_GetRelationshipsForSource(t *testing.T) {
	schema := NewAWSSchema()

	// EC2 instance has multiple relationships as source
	rels := schema.GetRelationshipsForSource("ec2", "instance")
	assert.NotEmpty(t, rels)

	// Should include instance -> subnet, instance -> security-group, etc.
	assert.Contains(t, rels, "ec2:instance -> ec2:subnet")
	assert.Contains(t, rels, "ec2:instance -> ec2:security-group")
}

func TestSchema_GetRelationshipsForTarget(t *testing.T) {
	schema := NewAWSSchema()

	// EC2 VPC is target of multiple relationships
	rels := schema.GetRelationshipsForTarget("ec2", "vpc")
	assert.NotEmpty(t, rels)

	// Should include subnet -> vpc, security-group -> vpc, etc.
	assert.Contains(t, rels, "ec2:subnet -> ec2:vpc")
	assert.Contains(t, rels, "ec2:security-group -> ec2:vpc")
}

func TestAWSSchema_RelationshipCount(t *testing.T) {
	schema := NewAWSSchema()
	// AWS schema should have a reasonable number of relationships defined
	assert.Greater(t, schema.RelationshipCount(), 20)
}

func TestGraph_SetConfig(t *testing.T) {
	config := DefaultGraphConfig()
	config.StrictValidation = false
	g := NewRelationshipGraph(config)

	// Verify initial config
	assert.False(t, g.Config().StrictValidation)

	// Update config
	newConfig := g.Config()
	newConfig.StrictValidation = true
	g.SetConfig(newConfig)

	// Verify config was updated
	assert.True(t, g.Config().StrictValidation)
}

func TestResourceManager_SetValidationMode(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	config := DefaultResourceManagerConfig()
	config.StrictValidation = false
	rm := NewResourceManager(state, config)

	// Verify initial mode
	assert.False(t, rm.IsStrictMode())

	// Change to strict mode
	rm.SetValidationMode(true)
	assert.True(t, rm.IsStrictMode())

	// Verify the graph's config was also updated
	// by attempting to add an invalid relationship in strict mode
	vpc := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-123"}
	subnet := ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-456"}
	rm.RegisterResource(vpc, nil)
	rm.RegisterResource(subnet, nil)

	// Invalid relationship type should fail in strict mode with schema
	rm.SetSchema(NewAWSSchema())
	err := rm.AddRelationship(vpc, subnet, RelAssociatedWith) // vpc doesn't associate with subnet
	assert.Error(t, err) // Should fail because schema doesn't allow this

	// Change back to lenient mode
	rm.SetValidationMode(false)
	assert.False(t, rm.IsStrictMode())

	// Same relationship should now succeed (lenient mode ignores schema violations)
	err = rm.AddRelationship(vpc, subnet, RelAssociatedWith)
	assert.NoError(t, err)
}

// failingStateManager wraps a StateManager and fails Delete operations
type failingStateManager struct {
	emulator.StateManager
	failDelete bool
}

func (f *failingStateManager) Delete(key string) error {
	if f.failDelete {
		return fmt.Errorf("simulated delete failure")
	}
	return f.StateManager.Delete(key)
}

func TestResourceManager_DeleteResource_RollbackOnStateFailure(t *testing.T) {
	// Create a state manager that will fail on Delete
	baseState := emulator.NewMemoryStateManager()
	failingState := &failingStateManager{StateManager: baseState, failDelete: false}

	config := DefaultResourceManagerConfig()
	rm := NewResourceManager(failingState, config)

	// Create a VPC resource
	vpcId := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-test"}
	vpcStateKey := "ec2:vpc:vpc-test"
	vpcMetadata := map[string]string{"cidr": "10.0.0.0/16"}

	// Create a subnet resource that depends on the VPC
	subnetId := ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-test"}
	subnetStateKey := "ec2:subnet:subnet-test"

	// Store in state and register in graph
	err := baseState.Set(vpcStateKey, map[string]string{"id": "vpc-test"})
	require.NoError(t, err)
	err = baseState.Set(subnetStateKey, map[string]string{"id": "subnet-test"})
	require.NoError(t, err)

	err = rm.RegisterResource(vpcId, vpcMetadata)
	require.NoError(t, err)
	err = rm.RegisterResource(subnetId, nil)
	require.NoError(t, err)

	// Create relationship: subnet -> vpc (subnet depends on vpc)
	err = rm.AddRelationship(subnetId, vpcId, RelContains)
	require.NoError(t, err)

	// Verify resources and relationship exist
	assert.True(t, rm.HasResource(vpcId))
	assert.True(t, rm.HasResource(subnetId))

	// Verify subnet has VPC as dependency
	deps, err := rm.GetDependencies(subnetId)
	require.NoError(t, err)
	assert.Len(t, deps, 1)
	assert.Equal(t, vpcId, deps[0])

	// Now make Delete fail and try to delete the subnet
	failingState.failDelete = true

	// Attempt to delete subnet - should fail but rollback the graph including edges
	err = rm.DeleteResource(subnetId, subnetStateKey)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete resource data")
	assert.Contains(t, err.Error(), "graph restored")

	// Verify subnet was restored in graph after rollback
	assert.True(t, rm.HasResource(subnetId), "Subnet should be restored in graph after rollback")

	// Verify the edge (relationship) was also restored
	deps, err = rm.GetDependencies(subnetId)
	require.NoError(t, err)
	assert.Len(t, deps, 1, "Edge should be restored after rollback")
	assert.Equal(t, vpcId, deps[0], "Subnet should still depend on VPC after rollback")

	// Also verify VPC still has subnet as dependent
	dependents, err := rm.GetDependents(vpcId)
	require.NoError(t, err)
	assert.Len(t, dependents, 1, "VPC should still have subnet as dependent after rollback")
	assert.Equal(t, subnetId, dependents[0])
}

func TestResourceManager_DeleteResource_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	config := DefaultResourceManagerConfig()
	rm := NewResourceManager(state, config)

	// Create a resource
	id := ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-test"}
	stateKey := "ec2:vpc:vpc-test"

	// Use CreateResource to atomically add to state and graph
	err := rm.CreateResource(id, stateKey, map[string]string{"id": "vpc-test"})
	require.NoError(t, err)

	// Verify resource exists
	assert.True(t, rm.HasResource(id))
	assert.True(t, state.Exists(stateKey))

	// Delete resource
	err = rm.DeleteResource(id, stateKey)
	require.NoError(t, err)

	// Verify resource is gone from both graph and state
	assert.False(t, rm.HasResource(id))
	assert.False(t, state.Exists(stateKey))
}
