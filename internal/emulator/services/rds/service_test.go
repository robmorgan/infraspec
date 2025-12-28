package rds

import (
	"context"
	"testing"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	testhelpers "github.com/robmorgan/infraspec/internal/emulator/testing"
)

func TestCreateDBInstance_ResponseFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewRDSService(state, validator)

	// Create response registry and validator
	responseRegistry := emulator.NewResponseRegistry()
	emulator.RegisterRDSResponses(responseRegistry)
	responseValidator := emulator.NewResponseValidator(responseRegistry)

	// Test request
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateDBInstance&DBInstanceIdentifier=test-db&DBInstanceClass=db.t3.micro&Engine=mysql"),
		Action: "CreateDBInstance",
	}

	// Execute request
	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	// Validate response status code
	testhelpers.AssertResponseStatus(t, resp, 200)

	// Validate Content-Type header
	testhelpers.AssertContentType(t, resp, "text/xml")

	// Validate RequestId is present
	testhelpers.AssertRequestID(t, resp)

	// Validate response structure using validator
	testhelpers.ValidateResponse(t, responseValidator, "rds", "CreateDBInstance", resp)

	// Validate XML structure
	testhelpers.AssertXMLStructure(t, resp, "CreateDBInstanceResponse")
}

func TestCreateDBInstance_ErrorResponse(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewRDSService(state, validator)

	// Create first instance
	req1 := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateDBInstance&DBInstanceIdentifier=test-db&DBInstanceClass=db.t3.micro&Engine=mysql"),
		Action: "CreateDBInstance",
	}
	service.HandleRequest(context.Background(), req1)

	// Try to create duplicate - should return error
	req2 := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateDBInstance&DBInstanceIdentifier=test-db&DBInstanceClass=db.t3.micro&Engine=mysql"),
		Action: "CreateDBInstance",
	}

	resp, err := service.HandleRequest(context.Background(), req2)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	// Validate error response
	testhelpers.AssertResponseStatus(t, resp, 409) // Conflict
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "DBInstanceAlreadyExistsFault", emulator.ProtocolQuery)
}

func TestDescribeDBInstances_ResponseFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewRDSService(state, validator)

	// Create a test instance first
	createReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateDBInstance&DBInstanceIdentifier=test-db&DBInstanceClass=db.t3.micro&Engine=mysql"),
		Action: "CreateDBInstance",
	}
	service.HandleRequest(context.Background(), createReq)

	// Describe instances
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=DescribeDBInstances"),
		Action: "DescribeDBInstances",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	// Validate response
	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)
	testhelpers.AssertXMLStructure(t, resp, "DescribeDBInstancesResponse")
}

func TestDeleteDBInstance_ResponseFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewRDSService(state, validator)

	// Create instance first
	createReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateDBInstance&DBInstanceIdentifier=test-db&DBInstanceClass=db.t3.micro&Engine=mysql"),
		Action: "CreateDBInstance",
	}
	service.HandleRequest(context.Background(), createReq)

	// Delete instance
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=DeleteDBInstance&DBInstanceIdentifier=test-db"),
		Action: "DeleteDBInstance",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	// Validate response
	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)
}

func TestDeleteDBInstance_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewRDSService(state, validator)

	// Try to delete non-existent instance
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=DeleteDBInstance&DBInstanceIdentifier=nonexistent"),
		Action: "DeleteDBInstance",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	// Validate error response - AWS returns HTTP 404 for DBInstanceNotFound
	testhelpers.AssertResponseStatus(t, resp, 404)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "DBInstanceNotFound", emulator.ProtocolQuery)
}

func TestDescribeDBInstances_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewRDSService(state, validator)

	// Try to describe non-existent instance with specific identifier
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=DescribeDBInstances&DBInstanceIdentifier=nonexistent"),
		Action: "DescribeDBInstances",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	// Validate error response - should return DBInstanceNotFound, not empty list.
	// This is required for Terraform AWS provider compatibility.
	// AWS returns HTTP 404 for DBInstanceNotFound.
	testhelpers.AssertResponseStatus(t, resp, 404)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "DBInstanceNotFound", emulator.ProtocolQuery)
}

func TestListTagsForResource_ResponseFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewRDSService(state, validator)

	// Create instance with tags
	createReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateDBInstance&DBInstanceIdentifier=test-db&DBInstanceClass=db.t3.micro&Engine=mysql&Tags.member.1.Key=Environment&Tags.member.1.Value=test&Tags.member.2.Key=Project&Tags.member.2.Value=infratest"),
		Action: "CreateDBInstance",
	}
	service.HandleRequest(context.Background(), createReq)

	// List tags
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=ListTagsForResource&ResourceName=arn:aws:rds:us-east-1:123456789012:db:test-db"),
		Action: "ListTagsForResource",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	// Validate response
	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)
	testhelpers.AssertXMLStructure(t, resp, "ListTagsForResourceResponse")
}

func TestListTagsForResource_NoTags(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewRDSService(state, validator)

	// Create instance without tags
	createReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateDBInstance&DBInstanceIdentifier=test-db&DBInstanceClass=db.t3.micro&Engine=mysql"),
		Action: "CreateDBInstance",
	}
	service.HandleRequest(context.Background(), createReq)

	// List tags
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=ListTagsForResource&ResourceName=arn:aws:rds:us-east-1:123456789012:db:test-db"),
		Action: "ListTagsForResource",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	// Validate response - should return empty tag list
	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)
}

// Example: Compare response with golden file
// Uncomment and use when you want to compare against golden files
/*
func TestCreateDBInstance_GoldenFile(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewRDSService(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body: []byte("Action=CreateDBInstance&DBInstanceIdentifier=my-database&DBInstanceClass=db.t3.micro&Engine=mysql"),
		Action: "CreateDBInstance",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	// Compare with golden file
	testing.CompareWithGoldenFile(t, resp, "testdata/responses/rds/CreateDBInstance_success.xml")
}
*/
