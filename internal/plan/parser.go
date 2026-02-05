package plan

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParsePlanFile reads a Terraform plan JSON file and parses it into a Plan struct.
func ParsePlanFile(path string) (*Plan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read plan file: %w", err)
	}
	return ParsePlanBytes(data)
}

// ParsePlanBytes parses Terraform plan JSON bytes into a Plan struct.
func ParsePlanBytes(data []byte) (*Plan, error) {
	var plan Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan JSON: %w", err)
	}
	return &plan, nil
}

// ResourcesByType returns all resource changes that match the given resource type.
func (p *Plan) ResourcesByType(resourceType string) []*ResourceChange {
	var result []*ResourceChange
	for _, rc := range p.ResourceChanges {
		if rc.Type == resourceType {
			result = append(result, rc)
		}
	}
	return result
}

// ResourceByAddress returns the resource change with the given address, or nil if not found.
func (p *Plan) ResourceByAddress(addr string) *ResourceChange {
	for _, rc := range p.ResourceChanges {
		if rc.Address == addr {
			return rc
		}
	}
	return nil
}

// ResourcesByModule returns all resource changes within the given module address.
// Pass empty string for root module resources.
func (p *Plan) ResourcesByModule(moduleAddr string) []*ResourceChange {
	var result []*ResourceChange
	for _, rc := range p.ResourceChanges {
		if rc.ModuleAddress == moduleAddr {
			result = append(result, rc)
		}
	}
	return result
}

// GetAfter returns the value of a top-level attribute from the Change.After map.
func (rc *ResourceChange) GetAfter(key string) (interface{}, bool) {
	if rc.Change == nil || rc.Change.After == nil {
		return nil, false
	}
	val, ok := rc.Change.After[key]
	return val, ok
}

// GetAfterNested returns a nested value from the Change.After map using a path.
// Supports both map keys and array indices (as string integers).
// Example: GetAfterNested("tags", "Name") or GetAfterNested("ingress", "0", "from_port")
func (rc *ResourceChange) GetAfterNested(path ...string) (interface{}, bool) {
	if rc.Change == nil || rc.Change.After == nil || len(path) == 0 {
		return nil, false
	}
	return getNestedValue(rc.Change.After, path)
}

// GetBefore returns the value of a top-level attribute from the Change.Before map.
func (rc *ResourceChange) GetBefore(key string) (interface{}, bool) {
	if rc.Change == nil || rc.Change.Before == nil {
		return nil, false
	}
	val, ok := rc.Change.Before[key]
	return val, ok
}

// GetBeforeNested returns a nested value from the Change.Before map using a path.
func (rc *ResourceChange) GetBeforeNested(path ...string) (interface{}, bool) {
	if rc.Change == nil || rc.Change.Before == nil || len(path) == 0 {
		return nil, false
	}
	return getNestedValue(rc.Change.Before, path)
}

// GetAfterString returns the value as a string, or empty string if not found or not a string.
func (rc *ResourceChange) GetAfterString(key string) string {
	val, ok := rc.GetAfter(key)
	if !ok {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// GetAfterBool returns the value as a bool, or false if not found or not a bool.
func (rc *ResourceChange) GetAfterBool(key string) bool {
	val, ok := rc.GetAfter(key)
	if !ok {
		return false
	}
	if b, ok := val.(bool); ok {
		return b
	}
	return false
}

// GetAfterInt returns the value as an int, or 0 if not found or not convertible to int.
// Handles JSON numbers which are typically float64.
func (rc *ResourceChange) GetAfterInt(key string) int {
	val, ok := rc.GetAfter(key)
	if !ok {
		return 0
	}
	return toInt(val)
}

// GetAfterFloat returns the value as a float64, or 0 if not found or not a number.
func (rc *ResourceChange) GetAfterFloat(key string) float64 {
	val, ok := rc.GetAfter(key)
	if !ok {
		return 0
	}
	if f, ok := val.(float64); ok {
		return f
	}
	return 0
}

// GetAfterStringSlice returns the value as a []string, or nil if not found or not a slice.
func (rc *ResourceChange) GetAfterStringSlice(key string) []string {
	val, ok := rc.GetAfter(key)
	if !ok {
		return nil
	}
	if slice, ok := val.([]interface{}); ok {
		result := make([]string, 0, len(slice))
		for _, item := range slice {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

// GetAfterMap returns the value as a map[string]interface{}, or nil if not found or not a map.
func (rc *ResourceChange) GetAfterMap(key string) map[string]interface{} {
	val, ok := rc.GetAfter(key)
	if !ok {
		return nil
	}
	if m, ok := val.(map[string]interface{}); ok {
		return m
	}
	return nil
}

// IsCreate returns true if the resource will be created (and not replaced).
func (rc *ResourceChange) IsCreate() bool {
	if rc.Change == nil || len(rc.Change.Actions) == 0 {
		return false
	}
	hasCreate := false
	hasDelete := false
	for _, a := range rc.Change.Actions {
		if a == ActionCreate {
			hasCreate = true
		}
		if a == ActionDelete {
			hasDelete = true
		}
	}
	return hasCreate && !hasDelete
}

// IsUpdate returns true if the resource will be updated in-place.
func (rc *ResourceChange) IsUpdate() bool {
	if rc.Change == nil || len(rc.Change.Actions) == 0 {
		return false
	}
	for _, a := range rc.Change.Actions {
		if a == ActionUpdate {
			return true
		}
	}
	return false
}

// IsDelete returns true if the resource will be deleted (and not replaced).
func (rc *ResourceChange) IsDelete() bool {
	if rc.Change == nil || len(rc.Change.Actions) == 0 {
		return false
	}
	hasCreate := false
	hasDelete := false
	for _, a := range rc.Change.Actions {
		if a == ActionCreate {
			hasCreate = true
		}
		if a == ActionDelete {
			hasDelete = true
		}
	}
	return hasDelete && !hasCreate
}

// IsReplace returns true if the resource will be replaced (deleted and recreated).
func (rc *ResourceChange) IsReplace() bool {
	if rc.Change == nil || len(rc.Change.Actions) == 0 {
		return false
	}
	hasCreate := false
	hasDelete := false
	for _, a := range rc.Change.Actions {
		if a == ActionCreate {
			hasCreate = true
		}
		if a == ActionDelete {
			hasDelete = true
		}
	}
	return hasCreate && hasDelete
}

// IsNoOp returns true if no changes will be made to the resource.
func (rc *ResourceChange) IsNoOp() bool {
	if rc.Change == nil || len(rc.Change.Actions) == 0 {
		return true
	}
	for _, a := range rc.Change.Actions {
		if a != ActionNoOp {
			return false
		}
	}
	return true
}

// IsRead returns true if this is a data source read operation.
func (rc *ResourceChange) IsRead() bool {
	if rc.Change == nil || len(rc.Change.Actions) == 0 {
		return false
	}
	for _, a := range rc.Change.Actions {
		if a == ActionRead {
			return true
		}
	}
	return false
}

// HasAction returns true if the resource change includes the specified action.
func (rc *ResourceChange) HasAction(action Action) bool {
	if rc.Change == nil {
		return false
	}
	for _, a := range rc.Change.Actions {
		if a == action {
			return true
		}
	}
	return false
}

// getNestedValue traverses a map using the given path and returns the value.
func getNestedValue(data map[string]interface{}, path []string) (interface{}, bool) {
	current := interface{}(data)

	for _, part := range path {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[part]
			if !ok {
				return nil, false
			}
			current = val
		case []interface{}:
			// Try to parse as array index
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(v) {
				return nil, false
			}
			current = v[idx]
		default:
			return nil, false
		}
	}

	return current, true
}

// toInt converts an interface{} to int, handling JSON float64 numbers.
func toInt(val interface{}) int {
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return 0
}

// IsDataSource returns true if this resource change represents a data source.
func (rc *ResourceChange) IsDataSource() bool {
	return rc.Mode == "data"
}

// IsManaged returns true if this resource change represents a managed resource.
func (rc *ResourceChange) IsManaged() bool {
	return rc.Mode == "managed"
}

// FullType returns the full resource type including provider prefix if available.
func (rc *ResourceChange) FullType() string {
	if rc.ProviderName != "" && !strings.Contains(rc.Type, "_") {
		return rc.ProviderName + "_" + rc.Type
	}
	return rc.Type
}
