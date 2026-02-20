package ec2

import (
	"context"
	"strings"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	"github.com/robmorgan/infraspec/internal/emulator/graph"
	testhelpers "github.com/robmorgan/infraspec/internal/emulator/testing"
)

func TestRunInstances_ResponseFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=RunInstances&ImageId=ami-0c55b159cbfafe1f0&MinCount=1&MaxCount=1&InstanceType=t2.micro"),
		Action: "RunInstances",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)

	// Verify response contains instance ID (camelCase per Smithy xmlName)
	if !strings.Contains(string(resp.Body), "<instanceId>i-") {
		t.Error("Response should contain instance ID starting with i-")
	}
}

func TestDescribeInstances_ResponseFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	// First create an instance
	createReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=RunInstances&ImageId=ami-0c55b159cbfafe1f0&MinCount=1&MaxCount=1"),
		Action: "RunInstances",
	}
	service.HandleRequest(context.Background(), createReq)

	// Now describe instances
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=DescribeInstances"),
		Action: "DescribeInstances",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)
}

func TestDescribeInstances_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=DescribeInstances&InstanceId.1=i-nonexistent"),
		Action: "DescribeInstances",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 400)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "InvalidInstanceID.NotFound", emulator.ProtocolQuery)
}

func TestTerminateInstances_ResponseFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	// First create an instance
	createReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=RunInstances&ImageId=ami-0c55b159cbfafe1f0&MinCount=1&MaxCount=1"),
		Action: "RunInstances",
	}
	createResp, _ := service.HandleRequest(context.Background(), createReq)

	// Extract instance ID from response (camelCase per Smithy xmlName)
	bodyStr := string(createResp.Body)
	start := strings.Index(bodyStr, "<instanceId>") + len("<instanceId>")
	end := strings.Index(bodyStr[start:], "</instanceId>")
	if start < len("<instanceId>") || end < 0 {
		t.Fatal("Could not extract instance ID from response")
	}
	instanceId := bodyStr[start : start+end]

	// Terminate the instance
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=TerminateInstances&InstanceId.1=" + instanceId),
		Action: "TerminateInstances",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
}

func TestCreateVpc_ResponseFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateVpc&CidrBlock=10.0.0.0/16"),
		Action: "CreateVpc",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)

	// Verify response contains VPC ID
	if !strings.Contains(string(resp.Body), "vpc-") {
		t.Error("Response should contain VPC ID starting with vpc-")
	}
}

func TestDescribeVpcs_ResponseFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=DescribeVpcs"),
		Action: "DescribeVpcs",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)

	// Should include default VPC
	if !strings.Contains(string(resp.Body), "vpc-default") {
		t.Error("Response should contain default VPC")
	}
}

func TestCreateSecurityGroup_ResponseFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateSecurityGroup&GroupName=test-sg&GroupDescription=Test+security+group"),
		Action: "CreateSecurityGroup",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)

	// Verify response contains security group ID
	if !strings.Contains(string(resp.Body), "<groupId>sg-") {
		t.Error("Response should contain security group ID starting with sg-")
	}
}

func TestCreateSubnet_ResponseFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateSubnet&VpcId=vpc-default&CidrBlock=172.31.48.0/20"),
		Action: "CreateSubnet",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)

	// Verify response contains subnet ID
	if !strings.Contains(string(resp.Body), "subnet-") {
		t.Error("Response should contain subnet ID starting with subnet-")
	}
}

func TestCreateInternetGateway_ResponseFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateInternetGateway"),
		Action: "CreateInternetGateway",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)

	// Verify response contains IGW ID
	if !strings.Contains(string(resp.Body), "igw-") {
		t.Error("Response should contain internet gateway ID starting with igw-")
	}
}

func TestCreateVolume_ResponseFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateVolume&AvailabilityZone=us-east-1a&Size=10&VolumeType=gp2"),
		Action: "CreateVolume",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)

	// Verify response contains volume ID
	if !strings.Contains(string(resp.Body), "vol-") {
		t.Error("Response should contain volume ID starting with vol-")
	}
}

func TestCreateKeyPair_ResponseFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateKeyPair&KeyName=test-key"),
		Action: "CreateKeyPair",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)

	// Verify response contains key pair info
	if !strings.Contains(string(resp.Body), "<keyName>test-key</keyName>") {
		t.Error("Response should contain key name")
	}
	if !strings.Contains(string(resp.Body), "<keyMaterial>") {
		t.Error("Response should contain key material (private key)")
	}
}

func TestCreateKeyPair_Duplicate(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	// Create first key pair
	req1 := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateKeyPair&KeyName=test-key"),
		Action: "CreateKeyPair",
	}
	service.HandleRequest(context.Background(), req1)

	// Try to create duplicate
	req2 := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateKeyPair&KeyName=test-key"),
		Action: "CreateKeyPair",
	}

	resp, err := service.HandleRequest(context.Background(), req2)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 400)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "InvalidKeyPair.Duplicate", emulator.ProtocolQuery)
}

func TestDescribeImages_ResponseFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=DescribeImages"),
		Action: "DescribeImages",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)

	// Should include pre-populated AMIs
	if !strings.Contains(string(resp.Body), "ami-0c55b159cbfafe1f0") {
		t.Error("Response should contain pre-populated Amazon Linux AMI")
	}
}

func TestCreateLaunchTemplate_ResponseFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateLaunchTemplate&LaunchTemplateName=test-template"),
		Action: "CreateLaunchTemplate",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)

	// Verify response contains launch template ID
	if !strings.Contains(string(resp.Body), "lt-") {
		t.Error("Response should contain launch template ID starting with lt-")
	}
}

func TestCreateTags_ResponseFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	// First create an instance
	createReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=RunInstances&ImageId=ami-0c55b159cbfafe1f0&MinCount=1&MaxCount=1"),
		Action: "RunInstances",
	}
	createResp, _ := service.HandleRequest(context.Background(), createReq)

	// Extract instance ID (camelCase per Smithy xmlName)
	bodyStr := string(createResp.Body)
	start := strings.Index(bodyStr, "<instanceId>") + len("<instanceId>")
	end := strings.Index(bodyStr[start:], "</instanceId>")
	if start < len("<instanceId>") || end < 0 {
		t.Fatal("Could not extract instance ID from response")
	}
	instanceId := bodyStr[start : start+end]

	// Create tags
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateTags&ResourceId.1=" + instanceId + "&Tag.1.Key=Name&Tag.1.Value=TestInstance"),
		Action: "CreateTags",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)

	// Verify return true
	if !strings.Contains(string(resp.Body), "<return>true</return>") {
		t.Error("Response should contain return true")
	}
}

func TestModifyVpcAttribute_EnableDnsHostnames(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	// First create a VPC
	createReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateVpc&CidrBlock=10.0.0.0/16"),
		Action: "CreateVpc",
	}
	createResp, err := service.HandleRequest(context.Background(), createReq)
	if err != nil {
		t.Fatalf("CreateVpc failed: %v", err)
	}

	// Extract VPC ID - AWS API uses vpcId (camelCase per Smithy xmlName)
	bodyStr := string(createResp.Body)
	start := strings.Index(bodyStr, "<vpcId>") + len("<vpcId>")
	end := strings.Index(bodyStr[start:], "</vpcId>")
	if start < len("<vpcId>") || end < 0 {
		t.Logf("Response body: %s", bodyStr)
		t.Fatal("Could not extract VPC ID from response")
	}
	vpcId := bodyStr[start : start+end]

	// Modify EnableDnsHostnames attribute
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=ModifyVpcAttribute&VpcId=" + vpcId + "&EnableDnsHostnames.Value=true"),
		Action: "ModifyVpcAttribute",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)

	// Verify return true
	if !strings.Contains(string(resp.Body), "<return>true</return>") {
		t.Error("Response should contain return true")
	}
}

func TestModifyVpcAttribute_EnableDnsSupport(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	// Use default VPC
	vpcId := "vpc-default"

	// Modify EnableDnsSupport attribute
	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=ModifyVpcAttribute&VpcId=" + vpcId + "&EnableDnsSupport.Value=false"),
		Action: "ModifyVpcAttribute",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertRequestID(t, resp)

	// Verify return true
	if !strings.Contains(string(resp.Body), "<return>true</return>") {
		t.Error("Response should contain return true")
	}
}

func TestModifyVpcAttribute_VpcNotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=ModifyVpcAttribute&VpcId=vpc-nonexistent&EnableDnsHostnames.Value=true"),
		Action: "ModifyVpcAttribute",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 400)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "InvalidVpcID.NotFound", emulator.ProtocolQuery)
}

func TestModifyVpcAttribute_MissingVpcId(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewEC2Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=ModifyVpcAttribute&EnableDnsHostnames.Value=true"),
		Action: "ModifyVpcAttribute",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 400)
	testhelpers.AssertContentType(t, resp, "text/xml")
	testhelpers.AssertErrorResponse(t, resp, "MissingParameter", emulator.ProtocolQuery)
}

func TestDeleteVpc_WithSubnet_BlockedByGraph(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()

	// Create service WITH graph support
	rm := createTestResourceManager(state)
	service := NewEC2ServiceWithGraph(state, validator, rm)

	// Create a VPC
	createVpcReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateVpc&CidrBlock=10.0.0.0/16"),
		Action: "CreateVpc",
	}
	createVpcResp, err := service.HandleRequest(context.Background(), createVpcReq)
	if err != nil {
		t.Fatalf("CreateVpc failed: %v", err)
	}
	testhelpers.AssertResponseStatus(t, createVpcResp, 200)

	// Extract VPC ID (camelCase per Smithy xmlName)
	bodyStr := string(createVpcResp.Body)
	start := strings.Index(bodyStr, "<vpcId>") + len("<vpcId>")
	end := strings.Index(bodyStr[start:], "</vpcId>")
	if start < len("<vpcId>") || end < 0 {
		t.Fatalf("Could not extract VPC ID from response: %s", bodyStr)
	}
	vpcId := bodyStr[start : start+end]

	// Create a subnet in the VPC
	createSubnetReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateSubnet&VpcId=" + vpcId + "&CidrBlock=10.0.1.0/24"),
		Action: "CreateSubnet",
	}
	createSubnetResp, err := service.HandleRequest(context.Background(), createSubnetReq)
	if err != nil {
		t.Fatalf("CreateSubnet failed: %v", err)
	}
	testhelpers.AssertResponseStatus(t, createSubnetResp, 200)

	// Try to delete the VPC - should be blocked because subnet exists
	deleteVpcReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=DeleteVpc&VpcId=" + vpcId),
		Action: "DeleteVpc",
	}
	deleteVpcResp, err := service.HandleRequest(context.Background(), deleteVpcReq)
	if err != nil {
		t.Fatalf("DeleteVpc failed: %v", err)
	}

	// Should return DependencyViolation error
	testhelpers.AssertResponseStatus(t, deleteVpcResp, 400)
	testhelpers.AssertContentType(t, deleteVpcResp, "text/xml")
	testhelpers.AssertErrorResponse(t, deleteVpcResp, "DependencyViolation", emulator.ProtocolQuery)

	// Verify error message mentions the subnet
	if !strings.Contains(string(deleteVpcResp.Body), "subnet") {
		t.Errorf("Error message should mention subnet dependency, got: %s", string(deleteVpcResp.Body))
	}
}

func TestDeleteVpc_WithSecurityGroup_BlockedByGraph(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()

	// Create service WITH graph support
	rm := createTestResourceManager(state)
	service := NewEC2ServiceWithGraph(state, validator, rm)

	// Create a VPC
	createVpcReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateVpc&CidrBlock=10.0.0.0/16"),
		Action: "CreateVpc",
	}
	createVpcResp, err := service.HandleRequest(context.Background(), createVpcReq)
	if err != nil {
		t.Fatalf("CreateVpc failed: %v", err)
	}
	testhelpers.AssertResponseStatus(t, createVpcResp, 200)

	// Extract VPC ID (camelCase per Smithy xmlName)
	bodyStr := string(createVpcResp.Body)
	start := strings.Index(bodyStr, "<vpcId>") + len("<vpcId>")
	end := strings.Index(bodyStr[start:], "</vpcId>")
	if start < len("<vpcId>") || end < 0 {
		t.Fatalf("Could not extract VPC ID from response: %s", bodyStr)
	}
	vpcId := bodyStr[start : start+end]

	// Create a security group in the VPC
	createSgReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateSecurityGroup&GroupName=test-sg&GroupDescription=Test&VpcId=" + vpcId),
		Action: "CreateSecurityGroup",
	}
	createSgResp, err := service.HandleRequest(context.Background(), createSgReq)
	if err != nil {
		t.Fatalf("CreateSecurityGroup failed: %v", err)
	}
	testhelpers.AssertResponseStatus(t, createSgResp, 200)

	// Try to delete the VPC - should be blocked because security group exists
	deleteVpcReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=DeleteVpc&VpcId=" + vpcId),
		Action: "DeleteVpc",
	}
	deleteVpcResp, err := service.HandleRequest(context.Background(), deleteVpcReq)
	if err != nil {
		t.Fatalf("DeleteVpc failed: %v", err)
	}

	// Should return DependencyViolation error
	testhelpers.AssertResponseStatus(t, deleteVpcResp, 400)
	testhelpers.AssertContentType(t, deleteVpcResp, "text/xml")
	testhelpers.AssertErrorResponse(t, deleteVpcResp, "DependencyViolation", emulator.ProtocolQuery)

	// Verify error message mentions the security group
	if !strings.Contains(string(deleteVpcResp.Body), "security-group") {
		t.Errorf("Error message should mention security group dependency, got: %s", string(deleteVpcResp.Body))
	}
}

func TestDeleteVpc_NoDependents_Succeeds(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()

	// Create service WITH graph support
	rm := createTestResourceManager(state)
	service := NewEC2ServiceWithGraph(state, validator, rm)

	// Create a VPC
	createVpcReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateVpc&CidrBlock=10.0.0.0/16"),
		Action: "CreateVpc",
	}
	createVpcResp, err := service.HandleRequest(context.Background(), createVpcReq)
	if err != nil {
		t.Fatalf("CreateVpc failed: %v", err)
	}
	testhelpers.AssertResponseStatus(t, createVpcResp, 200)

	// Extract VPC ID (camelCase per Smithy xmlName)
	bodyStr := string(createVpcResp.Body)
	start := strings.Index(bodyStr, "<vpcId>") + len("<vpcId>")
	end := strings.Index(bodyStr[start:], "</vpcId>")
	if start < len("<vpcId>") || end < 0 {
		t.Fatalf("Could not extract VPC ID from response: %s", bodyStr)
	}
	vpcId := bodyStr[start : start+end]

	// Delete the VPC - should succeed (no dependents)
	deleteVpcReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=DeleteVpc&VpcId=" + vpcId),
		Action: "DeleteVpc",
	}
	deleteVpcResp, err := service.HandleRequest(context.Background(), deleteVpcReq)
	if err != nil {
		t.Fatalf("DeleteVpc failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, deleteVpcResp, 200)
	testhelpers.AssertContentType(t, deleteVpcResp, "text/xml")

	// Verify return true
	if !strings.Contains(string(deleteVpcResp.Body), "<return>true</return>") {
		t.Error("Response should contain return true")
	}
}

// createTestResourceManager creates a ResourceManager for testing with graph support
func createTestResourceManager(state emulator.StateManager) *graph.ResourceManager {
	config := graph.ResourceManagerConfig{
		StrictValidation: true,
		DetectCycles:     true,
		UseAWSSchema:     true,
	}
	return graph.NewResourceManager(state, config)
}

func TestDeleteDefaultVpc_BlockedAsDefaultVpc(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()

	// Create service WITH graph support - this will register default resources in graph
	rm := createTestResourceManager(state)
	service := NewEC2ServiceWithGraph(state, validator, rm)

	// Try to delete the default VPC - should be blocked because it's the default VPC
	deleteVpcReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=DeleteVpc&VpcId=vpc-default"),
		Action: "DeleteVpc",
	}
	deleteVpcResp, err := service.HandleRequest(context.Background(), deleteVpcReq)
	if err != nil {
		t.Fatalf("DeleteVpc failed: %v", err)
	}

	// Should return OperationNotPermitted error for default VPC
	testhelpers.AssertResponseStatus(t, deleteVpcResp, 400)
	testhelpers.AssertContentType(t, deleteVpcResp, "text/xml")
	testhelpers.AssertErrorResponse(t, deleteVpcResp, "OperationNotPermitted", emulator.ProtocolQuery)

	// Verify error message mentions default VPC
	body := string(deleteVpcResp.Body)
	if !strings.Contains(body, "default VPC") {
		t.Errorf("Error message should mention default VPC, got: %s", body)
	}
}

func TestDefaultResourcesRegisteredInGraph(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()

	// Create service WITH graph support
	rm := createTestResourceManager(state)
	_ = NewEC2ServiceWithGraph(state, validator, rm)

	// Verify default VPC is registered in graph
	vpcId := graph.ResourceID{Service: "ec2", Type: "vpc", ID: "vpc-default"}
	if !rm.Graph().HasNode(vpcId) {
		t.Error("Default VPC should be registered in graph")
	}

	// Verify default subnet is registered in graph
	subnetId := graph.ResourceID{Service: "ec2", Type: "subnet", ID: "subnet-default"}
	if !rm.Graph().HasNode(subnetId) {
		t.Error("Default subnet should be registered in graph")
	}

	// Verify default security group is registered in graph
	sgId := graph.ResourceID{Service: "ec2", Type: "security-group", ID: "sg-default"}
	if !rm.Graph().HasNode(sgId) {
		t.Error("Default security group should be registered in graph")
	}

	// Verify default route table is registered in graph
	rtbId := graph.ResourceID{Service: "ec2", Type: "route-table", ID: "rtb-default"}
	if !rm.Graph().HasNode(rtbId) {
		t.Error("Default route table should be registered in graph")
	}

	// Verify relationships exist - subnet, SG, network ACL, and route table depend on VPC
	canDelete, dependents, _ := rm.CanDelete(vpcId)
	if canDelete {
		t.Error("Default VPC should not be deletable - has dependents")
	}
	if len(dependents) != 4 {
		t.Errorf("Default VPC should have 4 dependents (subnet, SG, network ACL, and route table), got %d: %v", len(dependents), dependents)
	}
}

func TestStrictMode_SubnetCreation_RollbackOnRelationshipFailure(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()

	// Create service with strict mode enabled
	config := graph.ResourceManagerConfig{
		StrictValidation: true,
		DetectCycles:     true,
		UseAWSSchema:     true,
	}
	rm := graph.NewResourceManager(state, config)
	service := NewEC2ServiceWithGraph(state, validator, rm)

	// Create a VPC first
	createVpcReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateVpc&CidrBlock=10.0.0.0/16"),
		Action: "CreateVpc",
	}
	createVpcResp, err := service.HandleRequest(context.Background(), createVpcReq)
	if err != nil {
		t.Fatalf("CreateVpc failed: %v", err)
	}
	testhelpers.AssertResponseStatus(t, createVpcResp, 200)

	// Extract VPC ID (camelCase per Smithy xmlName)
	bodyStr := string(createVpcResp.Body)
	start := strings.Index(bodyStr, "<vpcId>") + len("<vpcId>")
	end := strings.Index(bodyStr[start:], "</vpcId>")
	if start < len("<vpcId>") || end < 0 {
		t.Fatalf("Could not extract VPC ID from response: %s", bodyStr)
	}
	vpcId := bodyStr[start : start+end]

	// Verify we're in strict mode
	if !rm.IsStrictMode() {
		t.Fatal("Expected strict mode to be enabled")
	}

	// Create a subnet - should succeed normally
	createSubnetReq := &emulator.AWSRequest{
		Method: "POST",
		Headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Body:   []byte("Action=CreateSubnet&VpcId=" + vpcId + "&CidrBlock=10.0.1.0/24"),
		Action: "CreateSubnet",
	}
	createSubnetResp, err := service.HandleRequest(context.Background(), createSubnetReq)
	if err != nil {
		t.Fatalf("CreateSubnet failed: %v", err)
	}
	testhelpers.AssertResponseStatus(t, createSubnetResp, 200)

	// Extract subnet ID (camelCase per Smithy xmlName) and verify it exists in both state and graph
	subnetBodyStr := string(createSubnetResp.Body)
	subnetStart := strings.Index(subnetBodyStr, "<subnetId>") + len("<subnetId>")
	subnetEnd := strings.Index(subnetBodyStr[subnetStart:], "</subnetId>")
	if subnetStart < len("<subnetId>") || subnetEnd < 0 {
		t.Fatalf("Could not extract Subnet ID from response: %s", subnetBodyStr)
	}
	subnetId := subnetBodyStr[subnetStart : subnetStart+subnetEnd]

	// Verify subnet is in state
	var subnet interface{}
	err = state.Get("ec2:subnets:"+subnetId, &subnet)
	if err != nil {
		t.Errorf("Subnet should exist in state: %v", err)
	}

	// Verify subnet is in graph with relationship to VPC
	subnetResourceId := graph.ResourceID{Service: "ec2", Type: "subnet", ID: subnetId}
	if !rm.Graph().HasNode(subnetResourceId) {
		t.Error("Subnet should be registered in graph")
	}

	// Verify VPC has the subnet as dependent
	vpcResourceId := graph.ResourceID{Service: "ec2", Type: "vpc", ID: vpcId}
	canDelete, dependents, _ := rm.CanDelete(vpcResourceId)
	if canDelete {
		t.Error("VPC should not be deletable - has subnet dependent")
	}
	found := false
	for _, dep := range dependents {
		if dep.ID == subnetId {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("VPC dependents should include subnet %s, got: %v", subnetId, dependents)
	}
}
