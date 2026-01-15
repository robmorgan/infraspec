package lambda

import (
	"context"
	"fmt"
	"net/http"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

// handleDeleteFunction handles the DeleteFunction API
// DELETE /2015-03-31/functions/{FunctionName}
func (s *LambdaService) handleDeleteFunction(ctx context.Context, functionName string, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// Get qualifier from query params (optional - for deleting specific version)
	queryParams := parseQueryParams(req.Path)
	qualifier := queryParams.Get("Qualifier")

	// Load the function
	stateKey := fmt.Sprintf("lambda:functions:%s", functionName)
	var function StoredFunction
	if err := s.state.Get(stateKey, &function); err != nil {
		return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
			fmt.Sprintf("Function not found: %s", functionName)), nil
	}

	if qualifier != "" && qualifier != "$LATEST" {
		// Deleting a specific version
		if _, exists := function.PublishedVersions[qualifier]; !exists {
			return s.errorResponse(http.StatusNotFound, "ResourceNotFoundException",
				fmt.Sprintf("Function version not found: %s:%s", functionName, qualifier)), nil
		}

		// Delete the specific version
		delete(function.PublishedVersions, qualifier)

		// Save the updated function
		if err := s.state.Set(stateKey, &function); err != nil {
			return s.errorResponse(http.StatusInternalServerError, "ServiceException",
				"Failed to delete function version"), nil
		}
	} else {
		// Deleting the entire function

		// First, delete all aliases
		aliasKeys, _ := s.state.List(fmt.Sprintf("lambda:aliases:%s:", functionName))
		for _, key := range aliasKeys {
			s.state.Delete(key)
		}

		// Delete the function URL if it exists
		urlKey := fmt.Sprintf("lambda:function-urls:%s", functionName)
		s.state.Delete(urlKey)

		// Delete the function
		if err := s.state.Delete(stateKey); err != nil {
			return s.errorResponse(http.StatusInternalServerError, "ServiceException",
				"Failed to delete function"), nil
		}
	}

	// Return 204 No Content for successful deletion
	return &emulator.AWSResponse{
		StatusCode: http.StatusNoContent,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: nil,
	}, nil
}
