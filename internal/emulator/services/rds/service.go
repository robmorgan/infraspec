package rds

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/helpers"
)

type RDSService struct {
	state          emulator.StateManager
	validator      emulator.Validator
	stateMachine   *ResourceStateManager
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
}

func NewRDSService(state emulator.StateManager, validator emulator.Validator) *RDSService {
	ctx, cancel := context.WithCancel(context.Background())
	return &RDSService{
		state:          state,
		validator:      validator,
		stateMachine:   NewResourceStateManager(),
		shutdownCtx:    ctx,
		shutdownCancel: cancel,
	}
}

// Shutdown cancels all pending state transitions
func (s *RDSService) Shutdown() {
	s.shutdownCancel()
}

func (s *RDSService) ServiceName() string {
	return "rds"
}

// SupportedActions returns the list of AWS API actions this service handles.
// Used by the router to determine which service handles a given Query Protocol request.
func (s *RDSService) SupportedActions() []string {
	return []string{
		"CreateDBInstance",
		"DescribeDBInstances",
		"DeleteDBInstance",
		"ModifyDBInstance",
		"StartDBInstance",
		"StopDBInstance",
		"RebootDBInstance",
		"ListTagsForResource",
		"AddTagsToResource",
	}
}

func (s *RDSService) HandleRequest(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	if err := s.validator.ValidateRequest(req); err != nil {
		return s.errorResponse(400, "ValidationException", err.Error()), nil
	}

	action := s.extractAction(req)
	if action == "" {
		return s.errorResponse(400, "InvalidAction", "Missing or invalid action"), nil
	}

	params, err := s.parseParameters(req)
	if err != nil {
		return s.errorResponse(400, "InvalidParameterValue", err.Error()), nil
	}

	if err := s.validator.ValidateAction(action, params); err != nil {
		return s.errorResponse(400, "ValidationException", err.Error()), nil
	}

	switch action {
	case "CreateDBInstance":
		return s.createDBInstance(ctx, params)
	case "DescribeDBInstances":
		return s.describeDBInstances(ctx, params)
	case "DeleteDBInstance":
		return s.deleteDBInstance(ctx, params)
	case "ModifyDBInstance":
		return s.modifyDBInstance(ctx, params)
	case "StartDBInstance":
		return s.startDBInstance(ctx, params)
	case "StopDBInstance":
		return s.stopDBInstance(ctx, params)
	case "RebootDBInstance":
		return s.rebootDBInstance(ctx, params)
	case "ListTagsForResource":
		return s.listTagsForResource(ctx, params)
	case "AddTagsToResource":
		return s.addTagsToResource(ctx, params)
	default:
		return s.errorResponse(400, "InvalidAction", fmt.Sprintf("Unknown action: %s", action)), nil
	}
}

func (s *RDSService) extractAction(req *emulator.AWSRequest) string {
	if req.Action != "" {
		return req.Action
	}

	target := req.Headers["X-Amz-Target"]
	if target != "" {
		parts := strings.Split(target, ".")
		if len(parts) >= 2 {
			return parts[len(parts)-1]
		}
	}

	return ""
}

func (s *RDSService) parseParameters(req *emulator.AWSRequest) (map[string]interface{}, error) {
	if req.Parameters != nil {
		return req.Parameters, nil
	}

	contentType := req.Headers["Content-Type"]
	if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		return s.parseFormData(string(req.Body))
	}

	if strings.Contains(contentType, "application/json") {
		var params map[string]interface{}
		if err := json.Unmarshal(req.Body, &params); err != nil {
			return nil, fmt.Errorf("failed to parse JSON body: %w", err)
		}
		return params, nil
	}

	return make(map[string]interface{}), nil
}

func (s *RDSService) parseFormData(body string) (map[string]interface{}, error) {
	values, err := url.ParseQuery(body)
	if err != nil {
		return nil, err
	}

	params := make(map[string]interface{})
	for key, vals := range values {
		if len(vals) == 1 {
			params[key] = vals[0]
		} else {
			params[key] = vals
		}
	}

	return params, nil
}

func (s *RDSService) createDBInstance(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	identifier, ok := params["DBInstanceIdentifier"].(string)
	if !ok {
		return s.errorResponse(400, "InvalidParameterValue", "DBInstanceIdentifier is required"), nil
	}

	if s.state.(*emulator.MemoryStateManager).Exists(fmt.Sprintf("rds:db-instance:%s", identifier)) {
		return s.errorResponse(409, "DBInstanceAlreadyExistsFault", fmt.Sprintf("DB instance %s already exists", identifier)), nil
	}

	dbInstance := &DBInstance{
		DBInstanceIdentifier:             &identifier,
		DBInstanceClass:                  getStringParam(params, "DBInstanceClass", "db.t3.micro"),
		Engine:                           getStringParam(params, "Engine", "mysql"),
		EngineVersion:                    getStringParam(params, "EngineVersion", "8.0.35"),
		DBInstanceStatus:                 helpers.StringPtr("creating"),
		AllocatedStorage:                 getInt32Param(params, "AllocatedStorage", 20),
		StorageType:                      getStringParam(params, "StorageType", "gp2"),
		MasterUsername:                   getStringParam(params, "MasterUsername", "admin"),
		MultiAZ:                          getBoolParam(params, "MultiAZ", false),
		PubliclyAccessible:               getBoolParam(params, "PubliclyAccessible", false),
		StorageEncrypted:                 getBoolParam(params, "StorageEncrypted", false),
		AutoMinorVersionUpgrade:          getBoolParam(params, "AutoMinorVersionUpgrade", true),
		CopyTagsToSnapshot:               getBoolParam(params, "CopyTagsToSnapshot", false),
		DeletionProtection:               getBoolParam(params, "DeletionProtection", false),
		IAMDatabaseAuthenticationEnabled: getBoolParam(params, "IAMDatabaseAuthenticationEnabled", false),
		PerformanceInsightsEnabled:       getBoolParam(params, "PerformanceInsightsEnabled", false),
		DBInstanceArn:                    helpers.StringPtr(fmt.Sprintf("arn:aws:rds:us-east-1:123456789012:db:%s", identifier)),
		DbiResourceId:                    helpers.StringPtr(fmt.Sprintf("db-%s", uuid.New().String()[:8])),
		InstanceCreateTime:               &time.Time{},
	}

	if port := getInt32Param(params, "Port", 0); port != nil && *port > 0 {
		dbInstance.Endpoint = &Endpoint{
			Address: helpers.StringPtr(fmt.Sprintf("%s.cluster-xyz.us-east-1.rds.amazonaws.com", identifier)),
			Port:    port,
		}
		dbInstance.DbInstancePort = port
	}

	if dbName := getStringParam(params, "DBName", ""); dbName != nil && *dbName != "" {
		dbInstance.DBName = dbName
	}

	if err := s.state.Set(fmt.Sprintf("rds:db-instance:%s", identifier), dbInstance); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store DB instance"), nil
	}

	// Store a secondary index mapping DbiResourceId -> DBInstanceIdentifier
	// This allows lookups by resource ID (used by Terraform AWS provider)
	if dbInstance.DbiResourceId != nil {
		if err := s.state.Set(fmt.Sprintf("rds:db-instance-by-resource-id:%s", *dbInstance.DbiResourceId), identifier); err != nil {
			return s.errorResponse(500, "InternalFailure", "Failed to store resource ID index"), nil
		}
	}

	// Parse and store tags if present
	tags := s.parseTags(params)
	if len(tags) > 0 {
		arn := *dbInstance.DBInstanceArn
		if err := s.state.Set(fmt.Sprintf("rds:db-instance-tags:%s", arn), tags); err != nil {
			return s.errorResponse(500, "InternalFailure", "Failed to store tags"), nil
		}
	}

	// Schedule transition to "available" after delay
	s.scheduleDBInstanceTransition(identifier, DBInstanceStateAvailable, 5*time.Second)

	return s.successResponse("CreateDBInstance", CreateDBInstanceResult{DBInstance: dbInstance})
}

func (s *RDSService) describeDBInstances(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	identifier, hasIdentifier := params["DBInstanceIdentifier"].(string)

	var instances []DBInstance

	if hasIdentifier {
		var instance DBInstance
		if err := s.state.Get(fmt.Sprintf("rds:db-instance:%s", identifier), &instance); err != nil {
			// If lookup by identifier fails, try looking up by DbiResourceId
			// The Terraform AWS provider sometimes uses the resource ID for lookups
			var actualIdentifier string
			if err := s.state.Get(fmt.Sprintf("rds:db-instance-by-resource-id:%s", identifier), &actualIdentifier); err == nil {
				// Found the actual identifier via resource ID index
				if err := s.state.Get(fmt.Sprintf("rds:db-instance:%s", actualIdentifier), &instance); err == nil {
					instances = append(instances, instance)
					goto returnResponse
				}
			}
			// Return DBInstanceNotFound error when a specific instance is requested but doesn't exist.
			// Real AWS returns HTTP 404 with this error code.
			// The AWS SDK delete waiter treats DBInstanceNotFound as success (instance deleted).
			// Returning an empty list causes Terraform to fail with "empty result" error.
			return s.errorResponse(404, "DBInstanceNotFound", fmt.Sprintf("DBInstance %s not found.", identifier)), nil
		} else {
			instances = append(instances, instance)
		}
	} else {
		keys, err := s.state.List("rds:db-instance:")
		if err != nil {
			return s.errorResponse(500, "InternalFailure", "Failed to list DB instances"), nil
		}

		for _, key := range keys {
			var instance DBInstance
			if err := s.state.Get(key, &instance); err == nil {
				instances = append(instances, instance)
			}
		}
	}

returnResponse:
	return s.successResponse("DescribeDBInstances", DescribeDBInstancesResult{DBInstances: instances})
}

func (s *RDSService) deleteDBInstance(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	identifier, ok := params["DBInstanceIdentifier"].(string)
	if !ok {
		return s.errorResponse(400, "InvalidParameterValue", "DBInstanceIdentifier is required"), nil
	}

	var instance DBInstance
	if err := s.state.Get(fmt.Sprintf("rds:db-instance:%s", identifier), &instance); err != nil {
		// AWS returns HTTP 404 for DBInstanceNotFound
		return s.errorResponse(404, "DBInstanceNotFound", fmt.Sprintf("DBInstance %s not found.", identifier)), nil
	}

	// Cancel any pending state transitions before deleting
	resourceKey := "db-instance:" + identifier
	s.stateMachine.CancelPendingTransition(resourceKey)

	instance.DBInstanceStatus = helpers.StringPtr(string(DBInstanceStateDeleting))

	if err := s.state.Set(fmt.Sprintf("rds:db-instance:%s", identifier), &instance); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update DB instance"), nil
	}

	// Delete tags associated with this instance
	if instance.DBInstanceArn != nil {
		s.state.Delete(fmt.Sprintf("rds:db-instance-tags:%s", *instance.DBInstanceArn))
	}

	// Schedule removal of the instance after delay
	// Use 120 seconds to allow Terraform's polling to complete
	// Terraform polls every 10s and may take 60-90s to complete delete wait
	s.removeDBInstanceAfterDelay(identifier, 120*time.Second)

	return s.successResponse("DeleteDBInstance", DeleteDBInstanceResult{DBInstance: &instance})
}

func (s *RDSService) modifyDBInstance(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	identifier, ok := params["DBInstanceIdentifier"].(string)
	if !ok {
		return s.errorResponse(400, "InvalidParameterValue", "DBInstanceIdentifier is required"), nil
	}

	var instance DBInstance
	if err := s.state.Get(fmt.Sprintf("rds:db-instance:%s", identifier), &instance); err != nil {
		return s.errorResponse(404, "DBInstanceNotFound", fmt.Sprintf("DBInstance %s not found.", identifier)), nil
	}

	// Validate instance is in a modifiable state
	currentState := DBInstanceState(*instance.DBInstanceStatus)
	if currentState != DBInstanceStateAvailable && currentState != DBInstanceStateStorageFull {
		return s.errorResponse(400, "InvalidDBInstanceStateFault",
			fmt.Sprintf("DB instance %s is not in a modifiable state. Current state: %s", identifier, currentState)), nil
	}

	// Cancel any pending transitions
	resourceKey := "db-instance:" + identifier
	s.stateMachine.CancelPendingTransition(resourceKey)

	if allocatedStorage := getInt32Param(params, "AllocatedStorage", 0); allocatedStorage != nil && *allocatedStorage > 0 {
		instance.AllocatedStorage = allocatedStorage
	}

	if instanceClass := getStringParam(params, "DBInstanceClass", ""); instanceClass != nil && *instanceClass != "" {
		instance.DBInstanceClass = instanceClass
	}

	instance.DBInstanceStatus = helpers.StringPtr(string(DBInstanceStateModifying))

	if err := s.state.Set(fmt.Sprintf("rds:db-instance:%s", identifier), &instance); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update DB instance"), nil
	}

	// Schedule transition to available after modifications complete
	s.scheduleDBInstanceTransition(identifier, DBInstanceStateAvailable, 3*time.Second)

	return s.successResponse("ModifyDBInstance", ModifyDBInstanceResult{DBInstance: &instance})
}

func (s *RDSService) startDBInstance(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	identifier, ok := params["DBInstanceIdentifier"].(string)
	if !ok {
		return s.errorResponse(400, "InvalidParameterValue", "DBInstanceIdentifier is required"), nil
	}

	var instance DBInstance
	if err := s.state.Get(fmt.Sprintf("rds:db-instance:%s", identifier), &instance); err != nil {
		return s.errorResponse(404, "DBInstanceNotFound", fmt.Sprintf("DBInstance %s not found.", identifier)), nil
	}

	// Validate instance is stopped
	currentState := DBInstanceState(*instance.DBInstanceStatus)
	if currentState != DBInstanceStateStopped {
		return s.errorResponse(400, "InvalidDBInstanceStateFault",
			fmt.Sprintf("DB instance %s is not stopped. Current state: %s", identifier, currentState)), nil
	}

	instance.DBInstanceStatus = helpers.StringPtr(string(DBInstanceStateStarting))

	if err := s.state.Set(fmt.Sprintf("rds:db-instance:%s", identifier), &instance); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update DB instance"), nil
	}

	// Schedule transition to available after startup completes
	s.scheduleDBInstanceTransition(identifier, DBInstanceStateAvailable, 5*time.Second)

	return s.successResponse("StartDBInstance", StartDBInstanceResult{DBInstance: &instance})
}

func (s *RDSService) stopDBInstance(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	identifier, ok := params["DBInstanceIdentifier"].(string)
	if !ok {
		return s.errorResponse(400, "InvalidParameterValue", "DBInstanceIdentifier is required"), nil
	}

	var instance DBInstance
	if err := s.state.Get(fmt.Sprintf("rds:db-instance:%s", identifier), &instance); err != nil {
		return s.errorResponse(404, "DBInstanceNotFound", fmt.Sprintf("DBInstance %s not found.", identifier)), nil
	}

	// Validate instance is available
	currentState := DBInstanceState(*instance.DBInstanceStatus)
	if currentState != DBInstanceStateAvailable {
		return s.errorResponse(400, "InvalidDBInstanceStateFault",
			fmt.Sprintf("DB instance %s is not available. Current state: %s", identifier, currentState)), nil
	}

	// Cancel any pending transitions (e.g., if rebooting)
	resourceKey := "db-instance:" + identifier
	s.stateMachine.CancelPendingTransition(resourceKey)

	instance.DBInstanceStatus = helpers.StringPtr(string(DBInstanceStateStopping))

	if err := s.state.Set(fmt.Sprintf("rds:db-instance:%s", identifier), &instance); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update DB instance"), nil
	}

	// Schedule transition to stopped after shutdown completes
	s.scheduleDBInstanceTransition(identifier, DBInstanceStateStopped, 5*time.Second)

	return s.successResponse("StopDBInstance", StopDBInstanceResult{DBInstance: &instance})
}

func (s *RDSService) rebootDBInstance(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	identifier, ok := params["DBInstanceIdentifier"].(string)
	if !ok {
		return s.errorResponse(400, "InvalidParameterValue", "DBInstanceIdentifier is required"), nil
	}

	var instance DBInstance
	if err := s.state.Get(fmt.Sprintf("rds:db-instance:%s", identifier), &instance); err != nil {
		return s.errorResponse(404, "DBInstanceNotFound", fmt.Sprintf("DBInstance %s not found.", identifier)), nil
	}

	// Validate instance is available
	currentState := DBInstanceState(*instance.DBInstanceStatus)
	if currentState != DBInstanceStateAvailable {
		return s.errorResponse(400, "InvalidDBInstanceStateFault",
			fmt.Sprintf("DB instance %s is not available. Current state: %s", identifier, currentState)), nil
	}

	instance.DBInstanceStatus = helpers.StringPtr(string(DBInstanceStateRebooting))

	if err := s.state.Set(fmt.Sprintf("rds:db-instance:%s", identifier), &instance); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to update DB instance"), nil
	}

	// Schedule transition to available after reboot completes
	s.scheduleDBInstanceTransition(identifier, DBInstanceStateAvailable, 3*time.Second)

	return s.successResponse("RebootDBInstance", RebootDBInstanceResult{DBInstance: &instance})
}

func (s *RDSService) listTagsForResource(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	resourceName, ok := params["ResourceName"].(string)
	if !ok {
		return s.errorResponse(400, "InvalidParameterValue", "ResourceName is required"), nil
	}

	// Retrieve tags from state
	var tags []Tag
	if err := s.state.Get(fmt.Sprintf("rds:db-instance-tags:%s", resourceName), &tags); err != nil {
		// If no tags found, return empty list
		tags = []Tag{}
	}

	return s.successResponse("ListTagsForResource", ListTagsForResourceResult{TagList: tags})
}

func (s *RDSService) addTagsToResource(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	resourceName, ok := params["ResourceName"].(string)
	if !ok {
		return s.errorResponse(400, "InvalidParameterValue", "ResourceName is required"), nil
	}

	// Parse new tags from the request
	newTags := s.parseTags(params)
	if len(newTags) == 0 {
		return s.errorResponse(400, "InvalidParameterValue", "At least one tag must be provided"), nil
	}

	// Retrieve existing tags from state
	var existingTags []Tag
	stateKey := fmt.Sprintf("rds:db-instance-tags:%s", resourceName)
	if err := s.state.Get(stateKey, &existingTags); err != nil {
		// If no existing tags, start with empty list
		existingTags = []Tag{}
	}

	// Merge tags - update existing keys or add new ones
	tagMap := make(map[string]string)
	for _, tag := range existingTags {
		if tag.Key != nil && tag.Value != nil {
			tagMap[*tag.Key] = *tag.Value
		}
	}
	for _, tag := range newTags {
		if tag.Key != nil && tag.Value != nil {
			tagMap[*tag.Key] = *tag.Value
		}
	}

	// Convert back to tag list
	mergedTags := make([]Tag, 0, len(tagMap))
	for k, v := range tagMap {
		key := k
		value := v
		mergedTags = append(mergedTags, Tag{
			Key:   &key,
			Value: &value,
		})
	}

	// Store merged tags
	if err := s.state.Set(stateKey, mergedTags); err != nil {
		return s.errorResponse(500, "InternalFailure", "Failed to store tags"), nil
	}

	// Return empty response (AWS returns minimal response for AddTagsToResource)
	return s.successResponse("AddTagsToResource", AddTagsToResourceResult{})
}

// parseTags extracts tags from request parameters
// AWS sends tags as Tag.1.Key, Tag.1.Value, Tag.2.Key, Tag.2.Value, etc.
func (s *RDSService) parseTags(params map[string]interface{}) []Tag {
	tags := []Tag{}
	tagIndex := 1

	for {
		// Try both formats: Tags.Tag.N.Key (used by Terraform AWS provider)
		// and Tags.member.N.Key (used by AWS SDK directly)
		var keyParam, valueParam string
		var key, value string
		var hasKey, hasValue bool

		// Try Tags.Tag.N format first (Terraform)
		keyParam = fmt.Sprintf("Tags.Tag.%d.Key", tagIndex)
		valueParam = fmt.Sprintf("Tags.Tag.%d.Value", tagIndex)
		key, hasKey = params[keyParam].(string)
		value, hasValue = params[valueParam].(string)

		// If not found, try Tags.member.N format (AWS SDK)
		if !hasKey || !hasValue {
			keyParam = fmt.Sprintf("Tags.member.%d.Key", tagIndex)
			valueParam = fmt.Sprintf("Tags.member.%d.Value", tagIndex)
			key, hasKey = params[keyParam].(string)
			value, hasValue = params[valueParam].(string)
		}

		if !hasKey || !hasValue {
			break
		}

		tags = append(tags, Tag{
			Key:   &key,
			Value: &value,
		})

		tagIndex++
	}

	return tags
}

func (s *RDSService) successResponse(action string, data interface{}) (*emulator.AWSResponse, error) {
	return emulator.BuildQueryResponse(action, data, emulator.ResponseBuilderConfig{
		ServiceName: "rds",
		Version:     "2014-10-31",
	})
}

func (s *RDSService) errorResponse(statusCode int, code, message string) *emulator.AWSResponse {
	return emulator.BuildErrorResponse("rds", statusCode, code, message)
}

func getStringParam(params map[string]interface{}, key, defaultValue string) *string {
	if val, ok := params[key].(string); ok {
		return &val
	}
	if defaultValue != "" {
		return &defaultValue
	}
	return nil
}

func getInt32Param(params map[string]interface{}, key string, defaultValue int32) *int32 {
	if val, ok := params[key].(float64); ok {
		result := int32(val)
		return &result
	}
	if val, ok := params[key].(int); ok {
		result := int32(val)
		return &result
	}
	if val, ok := params[key].(int32); ok {
		return &val
	}
	// Handle string values (from form-encoded data)
	if val, ok := params[key].(string); ok {
		var parsed int32
		if _, err := fmt.Sscanf(val, "%d", &parsed); err == nil {
			return &parsed
		}
	}
	if defaultValue != 0 {
		return &defaultValue
	}
	return nil
}

func getBoolParam(params map[string]interface{}, key string, defaultValue bool) *bool {
	if val, ok := params[key].(bool); ok {
		return &val
	}
	if val, ok := params[key].(string); ok {
		result := val == "true"
		return &result
	}
	return &defaultValue
}
