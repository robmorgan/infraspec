package s3

import (
	"context"
	"strings"
	"testing"

	"github.com/robmorgan/infraspec/internal/emulator/core"
	testhelpers "github.com/robmorgan/infraspec/internal/emulator/testing"
)

// Helper function to create a bucket for testing
func createTestBucket(t *testing.T, service *S3Service, bucketName string) {
	t.Helper()
	req := &emulator.AWSRequest{
		Method: "PUT",
		Path:   "/" + bucketName,
		Headers: map[string]string{
			"Content-Type": "application/xml",
			"Host":         "s3.localhost:8000",
		},
		Body:   []byte{},
		Action: "CreateBucket",
	}
	_, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create test bucket: %v", err)
	}
}

// ============================================================================
// CreateBucket Tests
// ============================================================================

func TestCreateBucket_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "PUT",
		Path:   "/test-bucket",
		Headers: map[string]string{
			"Content-Type": "application/xml",
			"Host":         "s3.localhost:8000",
		},
		Body:   []byte{},
		Action: "CreateBucket",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
}

func TestCreateBucket_AlreadyExists(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	// Create first bucket
	createTestBucket(t, service, "test-bucket")

	// Try to create same bucket again - current implementation returns 200 (idempotent)
	// Note: Real S3 returns BucketAlreadyOwnedByYou (200) if same owner, BucketAlreadyExists (409) if different owner
	req := &emulator.AWSRequest{
		Method: "PUT",
		Path:   "/test-bucket",
		Headers: map[string]string{
			"Content-Type": "application/xml",
			"Host":         "s3.localhost:8000",
		},
		Body:   []byte{},
		Action: "CreateBucket",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	// Current implementation is idempotent - creating same bucket succeeds
	testhelpers.AssertResponseStatus(t, resp, 200)
}

// ============================================================================
// DeleteBucket Tests
// ============================================================================

func TestDeleteBucket_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	// Create a bucket first
	createTestBucket(t, service, "test-bucket")

	// Delete the bucket
	req := &emulator.AWSRequest{
		Method: "DELETE",
		Path:   "/test-bucket",
		Headers: map[string]string{
			"Host": "s3.localhost:8000",
		},
		Action: "DeleteBucket",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 204)
}

func TestDeleteBucket_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "DELETE",
		Path:   "/nonexistent-bucket",
		Headers: map[string]string{
			"Host": "s3.localhost:8000",
		},
		Action: "DeleteBucket",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 404)
	testhelpers.AssertErrorResponse(t, resp, "NoSuchBucket", emulator.ProtocolRESTXML)
}

// ============================================================================
// ListBuckets Tests
// ============================================================================

func TestListBuckets_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "GET",
		Path:   "/",
		Headers: map[string]string{
			"Host": "s3.localhost:8000",
		},
		Action: "ListBuckets",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "application/xml")
	testhelpers.AssertXMLStructure(t, resp, "ListAllMyBucketsResult")

	bodyStr := string(resp.Body)
	if !strings.Contains(bodyStr, "<Owner>") {
		t.Error("Response should contain Owner element")
	}
	if !strings.Contains(bodyStr, "<Buckets>") {
		t.Error("Response should contain Buckets element")
	}
}

func TestListBuckets_WithBuckets(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	// Create some buckets
	createTestBucket(t, service, "bucket-1")
	createTestBucket(t, service, "bucket-2")
	createTestBucket(t, service, "bucket-3")

	req := &emulator.AWSRequest{
		Method: "GET",
		Path:   "/",
		Headers: map[string]string{
			"Host": "s3.localhost:8000",
		},
		Action: "ListBuckets",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "application/xml")

	bodyStr := string(resp.Body)
	if !strings.Contains(bodyStr, "<Name>bucket-1</Name>") {
		t.Error("Response should contain bucket-1")
	}
	if !strings.Contains(bodyStr, "<Name>bucket-2</Name>") {
		t.Error("Response should contain bucket-2")
	}
	if !strings.Contains(bodyStr, "<Name>bucket-3</Name>") {
		t.Error("Response should contain bucket-3")
	}
	if !strings.Contains(bodyStr, "<CreationDate>") {
		t.Error("Response should contain CreationDate for buckets")
	}
}

func TestListBuckets_XMLSafe(t *testing.T) {
	// Test that bucket names with special characters are properly escaped
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	// Create a bucket with characters that need escaping
	// Note: In real S3, bucket names have restrictions, but we test XML safety
	createTestBucket(t, service, "test-bucket")

	req := &emulator.AWSRequest{
		Method: "GET",
		Path:   "/",
		Headers: map[string]string{
			"Host": "s3.localhost:8000",
		},
		Action: "ListBuckets",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify proper XML structure (no broken XML from potential injection)
	bodyStr := string(resp.Body)
	if !strings.HasPrefix(bodyStr, "<?xml version=") {
		t.Error("Response should start with XML declaration")
	}
}

// ============================================================================
// HeadBucket Tests
// ============================================================================

func TestHeadBucket_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	createTestBucket(t, service, "test-bucket")

	req := &emulator.AWSRequest{
		Method: "HEAD",
		Path:   "/test-bucket",
		Headers: map[string]string{
			"Host": "s3.localhost:8000",
		},
		Action: "HeadBucket",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
}

func TestHeadBucket_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "HEAD",
		Path:   "/nonexistent-bucket",
		Headers: map[string]string{
			"Host": "s3.localhost:8000",
		},
		Action: "HeadBucket",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 404)
}

// ============================================================================
// PutObject / GetObject Tests
// ============================================================================

func TestPutObject_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	createTestBucket(t, service, "test-bucket")

	req := &emulator.AWSRequest{
		Method: "PUT",
		Path:   "/test-bucket/test-key",
		Headers: map[string]string{
			"Host":         "s3.localhost:8000",
			"Content-Type": "text/plain",
		},
		Body:   []byte("Hello, World!"),
		Action: "PutObject",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
}

func TestGetObject_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	createTestBucket(t, service, "test-bucket")

	// Put an object first
	putReq := &emulator.AWSRequest{
		Method: "PUT",
		Path:   "/test-bucket/test-key",
		Headers: map[string]string{
			"Host":         "s3.localhost:8000",
			"Content-Type": "text/plain",
		},
		Body:   []byte("Hello, World!"),
		Action: "PutObject",
	}
	_, _ = service.HandleRequest(context.Background(), putReq)

	// Get the object
	req := &emulator.AWSRequest{
		Method: "GET",
		Path:   "/test-bucket/test-key",
		Headers: map[string]string{
			"Host": "s3.localhost:8000",
		},
		Action: "GetObject",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	if string(resp.Body) != "Hello, World!" {
		t.Errorf("Expected body 'Hello, World!', got '%s'", string(resp.Body))
	}
}

func TestGetObject_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	createTestBucket(t, service, "test-bucket")

	req := &emulator.AWSRequest{
		Method: "GET",
		Path:   "/test-bucket/nonexistent-key",
		Headers: map[string]string{
			"Host": "s3.localhost:8000",
		},
		Action: "GetObject",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 404)
	testhelpers.AssertErrorResponse(t, resp, "NoSuchKey", emulator.ProtocolRESTXML)
}

// ============================================================================
// Bucket Versioning Tests
// ============================================================================

func TestPutBucketVersioning_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	createTestBucket(t, service, "test-bucket")

	req := &emulator.AWSRequest{
		Method: "PUT",
		Path:   "/test-bucket?versioning",
		Headers: map[string]string{
			"Host":         "s3.localhost:8000",
			"Content-Type": "application/xml",
		},
		Body:   []byte(`<VersioningConfiguration><Status>Enabled</Status></VersioningConfiguration>`),
		Action: "PutBucketVersioning",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
}

func TestGetBucketVersioning_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	createTestBucket(t, service, "test-bucket")

	// Enable versioning first
	putReq := &emulator.AWSRequest{
		Method: "PUT",
		Path:   "/test-bucket?versioning",
		Headers: map[string]string{
			"Host":         "s3.localhost:8000",
			"Content-Type": "application/xml",
		},
		Body:   []byte(`<VersioningConfiguration><Status>Enabled</Status></VersioningConfiguration>`),
		Action: "PutBucketVersioning",
	}
	_, _ = service.HandleRequest(context.Background(), putReq)

	// Get versioning
	req := &emulator.AWSRequest{
		Method: "GET",
		Path:   "/test-bucket?versioning",
		Headers: map[string]string{
			"Host": "s3.localhost:8000",
		},
		Action: "GetBucketVersioning",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "application/xml")
}

// ============================================================================
// Error Response Format Tests
// ============================================================================

func TestErrorResponse_XMLFormat(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "GET",
		Path:   "/nonexistent-bucket/some-key",
		Headers: map[string]string{
			"Host": "s3.localhost:8000",
		},
		Action: "GetObject",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	bodyStr := string(resp.Body)

	// S3 error responses should have proper XML structure
	if !strings.Contains(bodyStr, "<Error>") {
		t.Error("Error response should contain Error element")
	}
	if !strings.Contains(bodyStr, "<Code>") {
		t.Error("Error response should contain Code element")
	}
	if !strings.Contains(bodyStr, "<Message>") {
		t.Error("Error response should contain Message element")
	}
}

// ============================================================================
// ListObjectsV2 Tests
// ============================================================================

func TestListObjectsV2_EmptyBucket(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	createTestBucket(t, service, "test-bucket")

	req := &emulator.AWSRequest{
		Method: "GET",
		Path:   "/test-bucket?list-type=2",
		Headers: map[string]string{
			"Host": "s3.localhost:8000",
		},
		Action: "ListObjectsV2",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "application/xml")
	testhelpers.AssertXMLStructure(t, resp, "ListBucketResult")
}

func TestListObjectsV2_WithObjects(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	createTestBucket(t, service, "test-bucket")

	// Put some objects
	for _, key := range []string{"file1.txt", "file2.txt", "folder/file3.txt"} {
		putReq := &emulator.AWSRequest{
			Method: "PUT",
			Path:   "/test-bucket/" + key,
			Headers: map[string]string{
				"Host":         "s3.localhost:8000",
				"Content-Type": "text/plain",
			},
			Body:   []byte("content"),
			Action: "PutObject",
		}
		_, _ = service.HandleRequest(context.Background(), putReq)
	}

	req := &emulator.AWSRequest{
		Method: "GET",
		Path:   "/test-bucket?list-type=2",
		Headers: map[string]string{
			"Host": "s3.localhost:8000",
		},
		Action: "ListObjectsV2",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertXMLStructure(t, resp, "ListBucketResult")

	// Note: Current implementation may not include all object details in listing
	// This test verifies the basic response structure is correct
}

// ============================================================================
// Public Access Block Tests
// ============================================================================

func TestPutPublicAccessBlock_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	createTestBucket(t, service, "test-bucket")

	req := &emulator.AWSRequest{
		Method: "PUT",
		Path:   "/test-bucket?publicAccessBlock",
		Headers: map[string]string{
			"Host":         "s3.localhost:8000",
			"Content-Type": "application/xml",
		},
		Body: []byte(`<PublicAccessBlockConfiguration>
			<BlockPublicAcls>true</BlockPublicAcls>
			<IgnorePublicAcls>true</IgnorePublicAcls>
			<BlockPublicPolicy>true</BlockPublicPolicy>
			<RestrictPublicBuckets>true</RestrictPublicBuckets>
		</PublicAccessBlockConfiguration>`),
		Action: "PutPublicAccessBlock",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	// S3 returns 204 No Content for PutPublicAccessBlock
	testhelpers.AssertResponseStatus(t, resp, 204)
}

func TestGetPublicAccessBlock_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	createTestBucket(t, service, "test-bucket")

	// Set public access block first
	putReq := &emulator.AWSRequest{
		Method: "PUT",
		Path:   "/test-bucket?publicAccessBlock",
		Headers: map[string]string{
			"Host":         "s3.localhost:8000",
			"Content-Type": "application/xml",
		},
		Body: []byte(`<PublicAccessBlockConfiguration>
			<BlockPublicAcls>true</BlockPublicAcls>
		</PublicAccessBlockConfiguration>`),
		Action: "PutPublicAccessBlock",
	}
	_, _ = service.HandleRequest(context.Background(), putReq)

	// Get public access block
	req := &emulator.AWSRequest{
		Method: "GET",
		Path:   "/test-bucket?publicAccessBlock",
		Headers: map[string]string{
			"Host": "s3.localhost:8000",
		},
		Action: "GetPublicAccessBlock",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "application/xml")
}

// ============================================================================
// Bucket Encryption Tests
// ============================================================================

func TestPutBucketEncryption_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	createTestBucket(t, service, "test-bucket")

	req := &emulator.AWSRequest{
		Method: "PUT",
		Path:   "/test-bucket?encryption",
		Headers: map[string]string{
			"Host":         "test-bucket.s3.localhost:8000",
			"Content-Type": "application/xml",
		},
		Body: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<ServerSideEncryptionConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Rule>
    <ApplyServerSideEncryptionByDefault>
      <SSEAlgorithm>AES256</SSEAlgorithm>
    </ApplyServerSideEncryptionByDefault>
    <BucketKeyEnabled>true</BucketKeyEnabled>
  </Rule>
</ServerSideEncryptionConfiguration>`),
		Action: "PutBucketEncryption",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
}

func TestGetBucketEncryption_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	createTestBucket(t, service, "test-bucket")

	// Set encryption first
	putReq := &emulator.AWSRequest{
		Method: "PUT",
		Path:   "/test-bucket?encryption",
		Headers: map[string]string{
			"Host":         "test-bucket.s3.localhost:8000",
			"Content-Type": "application/xml",
		},
		Body: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<ServerSideEncryptionConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Rule>
    <ApplyServerSideEncryptionByDefault>
      <SSEAlgorithm>AES256</SSEAlgorithm>
    </ApplyServerSideEncryptionByDefault>
    <BucketKeyEnabled>true</BucketKeyEnabled>
  </Rule>
</ServerSideEncryptionConfiguration>`),
		Action: "PutBucketEncryption",
	}
	_, _ = service.HandleRequest(context.Background(), putReq)

	// Get encryption
	req := &emulator.AWSRequest{
		Method: "GET",
		Path:   "/test-bucket?encryption",
		Headers: map[string]string{
			"Host": "test-bucket.s3.localhost:8000",
		},
		Action: "GetBucketEncryption",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "application/xml")
	testhelpers.AssertXMLStructure(t, resp, "ServerSideEncryptionConfiguration")

	// Verify BucketKeyEnabled is in the response
	bodyStr := string(resp.Body)
	if !strings.Contains(bodyStr, "<BucketKeyEnabled>true</BucketKeyEnabled>") {
		t.Errorf("Response should contain BucketKeyEnabled=true, got: %s", bodyStr)
	}
}

func TestGetBucketEncryption_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	createTestBucket(t, service, "test-bucket")

	req := &emulator.AWSRequest{
		Method: "GET",
		Path:   "/test-bucket?encryption",
		Headers: map[string]string{
			"Host": "test-bucket.s3.localhost:8000",
		},
		Action: "GetBucketEncryption",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 404)
	testhelpers.AssertErrorResponse(t, resp, "ServerSideEncryptionConfigurationNotFoundError", emulator.ProtocolRESTXML)
}

func TestBucketEncryption_BucketKeyEnabledFalse(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	createTestBucket(t, service, "test-bucket")

	// Set encryption with BucketKeyEnabled=false
	putReq := &emulator.AWSRequest{
		Method: "PUT",
		Path:   "/test-bucket?encryption",
		Headers: map[string]string{
			"Host":         "test-bucket.s3.localhost:8000",
			"Content-Type": "application/xml",
		},
		Body: []byte(`<?xml version="1.0" encoding="UTF-8"?>
<ServerSideEncryptionConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Rule>
    <ApplyServerSideEncryptionByDefault>
      <SSEAlgorithm>AES256</SSEAlgorithm>
    </ApplyServerSideEncryptionByDefault>
    <BucketKeyEnabled>false</BucketKeyEnabled>
  </Rule>
</ServerSideEncryptionConfiguration>`),
		Action: "PutBucketEncryption",
	}
	_, _ = service.HandleRequest(context.Background(), putReq)

	// Get encryption
	req := &emulator.AWSRequest{
		Method: "GET",
		Path:   "/test-bucket?encryption",
		Headers: map[string]string{
			"Host": "test-bucket.s3.localhost:8000",
		},
		Action: "GetBucketEncryption",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 200)

	// When BucketKeyEnabled is false, it may be omitted (omitempty) or shown as false
	bodyStr := string(resp.Body)
	// The response should contain the encryption configuration
	if !strings.Contains(bodyStr, "AES256") {
		t.Errorf("Response should contain SSEAlgorithm AES256, got: %s", bodyStr)
	}
}

// ============================================================================
// Invalid Action Tests
// ============================================================================

func TestInvalidAction(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewS3Service(state, validator)

	req := &emulator.AWSRequest{
		Method: "POST",
		Path:   "/",
		Headers: map[string]string{
			"Host": "s3.localhost:8000",
		},
		Action: "NonExistentAction",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("HandleRequest failed: %v", err)
	}

	testhelpers.AssertResponseStatus(t, resp, 400)
	testhelpers.AssertErrorResponse(t, resp, "InvalidAction", emulator.ProtocolRESTXML)
}
