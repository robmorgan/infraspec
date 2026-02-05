package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// PolicyDocument represents an IAM policy document
type PolicyDocument struct {
	Version   string            `json:"Version"`
	Id        string            `json:"Id"`
	Statement []PolicyStatement `json:"Statement"`
}

// PolicyStatement represents a single statement in an IAM policy
type PolicyStatement struct {
	Sid       string      `json:"Sid"`
	Effect    string      `json:"Effect"`
	Principal interface{} `json:"Principal"` // Can be string "*" or object {"Service": "..."}
	Action    string      `json:"Action"`
	Resource  string      `json:"Resource"`
	Condition interface{} `json:"Condition,omitempty"`
}

// handleAddPermission handles AddPermission API
// POST /functions/{FunctionName}/policy
func (s *LambdaService) handleAddPermission(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Parse input
	var input AddPermissionInput
	if err := json.Unmarshal(req.Body, &input); err != nil {
		return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
			fmt.Sprintf("Invalid request body: %v", err)), nil
	}

	// Validate required fields
	if input.StatementId == "" {
		return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
			"StatementId is required"), nil
	}
	if input.Action == "" {
		return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
			"Action is required"), nil
	}
	if input.Principal == "" {
		return s.errorResponse(http.StatusBadRequest, "InvalidParameterValueException",
			"Principal is required"), nil
	}

	// Get the function
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	// Parse existing policy or create new one
	var policy PolicyDocument
	if function.Policy != "" {
		if err := json.Unmarshal([]byte(function.Policy), &policy); err != nil {
			// If policy is malformed, start fresh
			policy = PolicyDocument{
				Version:   "2012-10-17",
				Id:        "default",
				Statement: []PolicyStatement{},
			}
		}
	} else {
		policy = PolicyDocument{
			Version:   "2012-10-17",
			Id:        "default",
			Statement: []PolicyStatement{},
		}
	}

	// Check for duplicate statement ID
	for _, stmt := range policy.Statement {
		if stmt.Sid == input.StatementId {
			return s.errorResponse(http.StatusConflict, "ResourceConflictException",
				fmt.Sprintf("The statement id (%s) provided already exists. Please provide a unique statement id.", input.StatementId)), nil
		}
	}

	// Build the principal value
	var principal interface{}
	if input.Principal == "*" {
		principal = "*"
	} else {
		// Check if it's a service principal or AWS account
		principal = map[string]string{"Service": input.Principal}
	}

	// Build the new statement
	newStatement := PolicyStatement{
		Sid:       input.StatementId,
		Effect:    "Allow",
		Principal: principal,
		Action:    input.Action,
		Resource:  function.FunctionArn,
	}

	// Add condition if source ARN or account is specified
	if input.SourceArn != "" || input.SourceAccount != "" {
		condition := make(map[string]map[string]string)
		if input.SourceArn != "" {
			condition["ArnLike"] = map[string]string{"AWS:SourceArn": input.SourceArn}
		}
		if input.SourceAccount != "" {
			if condition["StringEquals"] == nil {
				condition["StringEquals"] = make(map[string]string)
			}
			condition["StringEquals"]["AWS:SourceAccount"] = input.SourceAccount
		}
		newStatement.Condition = condition
	}

	// Add the statement
	policy.Statement = append(policy.Statement, newStatement)

	// Serialize the policy
	policyBytes, err := json.Marshal(policy)
	if err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to serialize policy"), nil
	}

	// Update the function
	function.Policy = string(policyBytes)
	function.RevisionId = uuid.New().String()

	if err := s.state.Set(stateKey, function); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to update function"), nil
	}

	// Return response
	response := map[string]interface{}{
		"Statement": string(policyBytes),
	}
	return s.successResponse(http.StatusCreated, response)
}

// handleRemovePermission handles RemovePermission API
// DELETE /functions/{FunctionName}/policy/{StatementId}
func (s *LambdaService) handleRemovePermission(ctx context.Context, functionName, statementId string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Get the function
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	// Check if policy exists
	if function.Policy == "" {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			"No policy is associated with the function"), nil
	}

	// Parse the policy
	var policy PolicyDocument
	if err := json.Unmarshal([]byte(function.Policy), &policy); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to parse policy"), nil
	}

	// Find and remove the statement
	found := false
	newStatements := []PolicyStatement{}
	for _, stmt := range policy.Statement {
		if stmt.Sid == statementId {
			found = true
		} else {
			newStatements = append(newStatements, stmt)
		}
	}

	if !found {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("No policy statement with id (%s) found", statementId)), nil
	}

	// Update the policy
	policy.Statement = newStatements

	// Serialize the policy
	policyBytes, err := json.Marshal(policy)
	if err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to serialize policy"), nil
	}

	// Update the function
	function.Policy = string(policyBytes)
	function.RevisionId = uuid.New().String()

	if err := s.state.Set(stateKey, function); err != nil {
		return s.errorResponse(http.StatusInternalServerError, "ServiceException",
			"Failed to update function"), nil
	}

	// Return 204 No Content
	return &emulator.AWSResponse{
		StatusCode: http.StatusNoContent,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}
