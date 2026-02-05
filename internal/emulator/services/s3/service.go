package s3

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
)

type S3Service struct {
	state     emulator.StateManager
	validator emulator.Validator
}

func NewS3Service(state emulator.StateManager, validator emulator.Validator) *S3Service {
	return &S3Service{
		state:     state,
		validator: validator,
	}
}

func (s *S3Service) ServiceName() string {
	return "s3"
}

func (s *S3Service) HandleRequest(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	if err := s.validator.ValidateRequest(req); err != nil {
		return s.errorResponse(400, "ValidationException", err.Error()), nil
	}

	// Check if this is an S3 Control request (has x-amz-account-id header or /v20180820/ path)
	if s.isS3ControlRequest(req) {
		return s.handleS3ControlRequest(ctx, req)
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
	case "CreateBucket":
		return s.createBucket(ctx, params, req)
	case "CreateBucketMetadataConfiguration":
		return s.createBucketMetadataConfiguration(ctx, params)
	case "CreateBucketMetadataTableConfiguration":
		return s.createBucketMetadataTableConfiguration(ctx, params)
	case "CreateMultipartUpload":
		return s.createMultipartUpload(ctx, params)
	case "CreateSession":
		return s.createSession(ctx, params)
	case "DeleteBucket":
		return s.deleteBucket(ctx, params, req)
	case "ListBuckets":
		return s.listBuckets(ctx, params, req)
	case "PutBucketVersioning":
		return s.putBucketVersioning(ctx, params, req)
	case "GetBucketVersioning":
		return s.getBucketVersioning(ctx, params, req)
	case "PutBucketEncryption":
		return s.putBucketEncryption(ctx, params, req)
	case "GetBucketEncryption":
		return s.getBucketEncryption(ctx, params, req)
	case "PutPublicAccessBlock":
		return s.putPublicAccessBlock(ctx, params, req)
	case "GetPublicAccessBlock":
		return s.getPublicAccessBlock(ctx, params, req)
	case "DeletePublicAccessBlock":
		return s.deletePublicAccessBlock(ctx, params, req)
	case "GetBucketPolicy":
		return s.getBucketPolicy(ctx, params, req)
	case "PutBucketPolicy":
		return s.putBucketPolicy(ctx, params, req)
	case "DeleteBucketPolicy":
		return s.deleteBucketPolicy(ctx, params, req)
	case "GetBucketLogging":
		return s.getBucketLogging(ctx, params, req)
	case "PutBucketLogging":
		return s.putBucketLogging(ctx, params, req)
	case "PutObject":
		return s.putObject(ctx, params, req)
	case "GetObject":
		return s.getObject(ctx, params, req)
	case "HeadBucket":
		return s.headBucket(ctx, params, req)
	case "ListObjectsV2":
		return s.listObjectsV2(ctx, params, req)
	default:
		return s.errorResponse(400, "InvalidAction", fmt.Sprintf("Unknown action: %s", action)), nil
	}
}

// ExtractAction implements the ActionExtractor interface, allowing the handler
// to extract the action before logging. This is necessary for S3 since it's a
// REST-based service that derives actions from HTTP method and path.
func (s *S3Service) ExtractAction(req *emulator.AWSRequest) string {
	return s.extractAction(req)
}

func (s *S3Service) extractAction(req *emulator.AWSRequest) string {
	if req.Action != "" {
		return req.Action
	}

	target := req.Headers["X-Amz-Target"]
	if target != "" {
		parts := strings.Split(target, ".")
		if len(parts) >= 2 {
			action := parts[len(parts)-1]
			return action
		}
	}

	// S3 uses REST API, so derive action from HTTP method and path
	return s.deriveS3ActionFromRequest(req)
}

func (s *S3Service) deriveS3ActionFromRequest(req *emulator.AWSRequest) string {
	// Check if this is virtual-hosted-style (bucket-name in host) or path-style (bucket-name in path)
	// Virtual-hosted style patterns:
	// - bucket-name.s3.infraspec.sh
	// - bucket-name.s3.localhost
	// - bucket-name.localhost (legacy)
	isVirtualHosted := emulator.IsS3VirtualHostedRequest(req.Headers["Host"])

	// Parse the path - remove leading "/" and any query string
	pathWithoutQuery := req.Path
	if idx := strings.Index(pathWithoutQuery, "?"); idx >= 0 {
		pathWithoutQuery = pathWithoutQuery[:idx]
	}
	path := strings.TrimPrefix(pathWithoutQuery, "/")

	// Check for query parameters that indicate specific operations
	if strings.Contains(req.Path, "?") {
		queryString := ""
		if idx := strings.Index(req.Path, "?"); idx >= 0 {
			queryString = req.Path[idx+1:]
		}

		// Parse query parameters more carefully
		// AWS S3 uses query parameters like: ?versioning, ?encryption, ?policy, ?logging
		// Not ?versioning=something, but just the parameter name
		query, _ := url.ParseQuery(queryString)

		if query.Has("versioning") || strings.Contains(queryString, "versioning") {
			if req.Method == "PUT" {
				return "PutBucketVersioning"
			}
			return "GetBucketVersioning"
		}
		if query.Has("encryption") || strings.Contains(queryString, "encryption") {
			if req.Method == "PUT" {
				return "PutBucketEncryption"
			}
			return "GetBucketEncryption"
		}
		if query.Has("publicAccessBlock") || strings.Contains(queryString, "publicAccessBlock") {
			if req.Method == "PUT" {
				return "PutPublicAccessBlock"
			} else if req.Method == "DELETE" {
				return "DeletePublicAccessBlock"
			}
			return "GetPublicAccessBlock"
		}
		if query.Has("policy") || strings.Contains(queryString, "policy") {
			if req.Method == "PUT" {
				return "PutBucketPolicy"
			} else if req.Method == "DELETE" {
				return "DeleteBucketPolicy"
			}
			return "GetBucketPolicy"
		}
		if query.Has("logging") || strings.Contains(queryString, "logging") {
			if req.Method == "PUT" {
				return "PutBucketLogging"
			}
			return "GetBucketLogging"
		}
		if query.Has("delete") || strings.Contains(queryString, "delete") {
			return "DeleteObjects"
		}
	}

	// Determine action based on HTTP method and path structure
	switch req.Method {
	case "PUT":
		if isVirtualHosted {
			// Virtual-hosted style: PUT / = CreateBucket, PUT /key = PutObject
			if path == "" {
				return "CreateBucket"
			}
			return "PutObject"
		} else {
			// Path style: PUT /bucket-name = CreateBucket, PUT /bucket-name/key = PutObject
			if path != "" && !strings.Contains(path, "/") {
				return "CreateBucket"
			}
			return "PutObject"
		}
	case "GET":
		if isVirtualHosted {
			// Virtual-hosted style:
			// GET / without query params = HeadBucket (confirms bucket exists)
			// GET / with ?list-type=2 = ListObjectsV2 (lists objects)
			// GET /key = GetObject
			// Note: Changed from ListObjectsV2 to HeadBucket for GET / to prevent
			// Terraform from receiving XML when it expects empty response after CreateBucket
			if path == "" {
				return "HeadBucket"
			}
			return "GetObject"
		} else {
			// Path style:
			// GET / = ListBuckets
			// GET /bucket-name without query = HeadBucket (confirms bucket exists)
			// GET /bucket-name with ?list-type=2 = ListObjectsV2 (lists objects)
			// GET /bucket-name/key = GetObject
			if path == "" {
				return "ListBuckets"
			}
			if path != "" && !strings.Contains(path, "/") {
				return "HeadBucket"
			}
			return "GetObject"
		}
	case "HEAD":
		if isVirtualHosted {
			// Virtual-hosted style: HEAD / = HeadBucket, HEAD /key = HeadObject
			if path == "" {
				return "HeadBucket"
			}
			return "HeadObject"
		} else {
			// Path style: HEAD /bucket-name = HeadBucket, HEAD /bucket-name/key = HeadObject
			if path != "" && !strings.Contains(path, "/") {
				return "HeadBucket"
			}
			return "HeadObject"
		}
	case "DELETE":
		if isVirtualHosted {
			// Virtual-hosted style: DELETE / = DeleteBucket, DELETE /key = DeleteObject
			if path == "" {
				return "DeleteBucket"
			}
			return "DeleteObject"
		} else {
			// Path style: DELETE /bucket-name = DeleteBucket, DELETE /bucket-name/key = DeleteObject
			if path != "" && !strings.Contains(path, "/") {
				return "DeleteBucket"
			}
			return "DeleteObject"
		}
	}

	return ""
}

func (s *S3Service) parseParameters(req *emulator.AWSRequest) (map[string]interface{}, error) {
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

func (s *S3Service) parseFormData(body string) (map[string]interface{}, error) {
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

func (s *S3Service) extractBucketName(req *emulator.AWSRequest) string {
	// For virtual-hosted-style requests, bucket name is in the host header
	// Priority: Check Host header first for virtual-hosted style patterns
	// Patterns supported:
	// - bucket-name.s3.infraspec.sh
	// - bucket-name.s3.localhost
	// - bucket-name.localhost (legacy)
	if bucketName := emulator.ExtractBucketNameFromHost(req.Headers["Host"]); bucketName != "" {
		return bucketName
	}

	// Fallback: For path-style requests, bucket name is the first path component
	// Pattern: s3.infraspec.sh/bucket-name/key
	path := strings.TrimPrefix(req.Path, "/")

	// Remove query string if present (e.g., ?publicAccessBlock, ?versioning)
	if idx := strings.Index(path, "?"); idx >= 0 {
		path = path[:idx]
	}

	if path != "" {
		pathParts := strings.Split(path, "/")
		if len(pathParts) > 0 && pathParts[0] != "" {
			return pathParts[0]
		}
	}

	return ""
}

func (s *S3Service) createBucket(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	bucketName := s.extractBucketName(req)
	if bucketName == "" {
		return s.errorResponse(400, "InvalidBucketName", "Bucket name is required"), nil
	}

	stateKey := "s3:" + bucketName

	// Check if bucket already exists
	var existing map[string]interface{}
	if err := s.state.Get(stateKey, &existing); err == nil {
		// For a testing emulator, we return success (BucketAlreadyOwnedByYou behavior)
		// This makes bucket creation idempotent which is useful for testing
		// Real AWS returns 200 OK if you own the bucket, 409 if someone else does
		return &emulator.AWSResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "application/xml",
				"Location":     "/" + bucketName,
			},
			Body: []byte{},
		}, nil
	}

	// Store bucket in state with proper attributes
	bucket := map[string]interface{}{
		"Name":         bucketName,
		"CreationDate": "2024-01-01T00:00:00Z",
		"Region":       "us-east-1", // Default region
	}

	if err := s.state.Set(stateKey, bucket); err != nil {
		return s.errorResponse(500, "InternalError", "Failed to create bucket"), nil
	}

	// S3 CreateBucket returns an empty response with Location header
	// The Location header is critical for Terraform to identify the resource
	return &emulator.AWSResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/xml",
			"Location":     "/" + bucketName,
		},
		Body: []byte{},
	}, nil
}

func (s *S3Service) createBucketMetadataConfiguration(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// TODO: Implement CreateBucketMetadataConfiguration
	// Required parameter: CreateBucketMetadataConfiguration (map[string]interface{}) - Input for CreateBucketMetadataConfiguration

	return s.errorResponse(501, "NotImplemented", "CreateBucketMetadataConfiguration is not yet implemented"), nil
}

func (s *S3Service) createBucketMetadataTableConfiguration(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// TODO: Implement CreateBucketMetadataTableConfiguration
	// Required parameter: CreateBucketMetadataTableConfiguration (map[string]interface{}) - Input for CreateBucketMetadataTableConfiguration

	return s.errorResponse(501, "NotImplemented", "CreateBucketMetadataTableConfiguration is not yet implemented"), nil
}

func (s *S3Service) createMultipartUpload(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// TODO: Implement CreateMultipartUpload
	// Required parameter: CreateMultipartUpload (map[string]interface{}) - Input for CreateMultipartUpload

	return s.errorResponse(501, "NotImplemented", "CreateMultipartUpload is not yet implemented"), nil
}

func (s *S3Service) createSession(ctx context.Context, params map[string]interface{}) (*emulator.AWSResponse, error) {
	// TODO: Implement CreateSession
	// Required parameter: CreateSession (map[string]interface{}) - Input for CreateSession

	return s.errorResponse(501, "NotImplemented", "CreateSession is not yet implemented"), nil
}

func (s *S3Service) deleteBucket(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	bucketName := s.extractBucketName(req)
	if bucketName == "" {
		return s.errorResponse(400, "InvalidBucketName", "Bucket name is required"), nil
	}

	stateKey := "s3:" + bucketName

	// Delete bucket from state
	if err := s.state.Delete(stateKey); err != nil {
		return s.errorResponse(404, "NoSuchBucket", fmt.Sprintf("Bucket %s does not exist", bucketName)), nil
	}

	return &emulator.AWSResponse{
		StatusCode: 204,
		Headers:    map[string]string{},
		Body:       []byte{},
	}, nil
}

func (s *S3Service) listBuckets(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// List all buckets from state
	keys, err := s.state.List("s3:")
	if err != nil {
		return s.errorResponse(500, "InternalError", "Failed to list buckets"), nil
	}

	// Filter out non-bucket keys (e.g., versioning, encryption configs)
	var buckets []map[string]interface{}
	for _, key := range keys {
		// Only include base bucket keys like "s3:bucket-name", not "s3:bucket-name:versioning"
		if strings.Count(key, ":") == 1 {
			var bucket map[string]interface{}
			if err := s.state.Get(key, &bucket); err == nil {
				buckets = append(buckets, bucket)
			}
		}
	}

	// Build ListBuckets XML response using response builder
	result := ListAllMyBucketsResult{
		Xmlns: "http://s3.amazonaws.com/doc/2006-03-01/",
		Owner: XMLOwner{
			ID:          "infraspec-api",
			DisplayName: "infraspec-api",
		},
		Buckets: XMLBuckets{
			Bucket: make([]XMLBucket, 0, len(buckets)),
		},
	}

	for _, bucket := range buckets {
		name, _ := bucket["Name"].(string)
		creationDate, _ := bucket["CreationDate"].(string)
		result.Buckets.Bucket = append(result.Buckets.Bucket, XMLBucket{
			Name:         name,
			CreationDate: creationDate,
		})
	}

	resp, err := emulator.BuildS3StructResponse(result)
	if err != nil {
		return s.errorResponse(500, "InternalError", "Failed to marshal response"), nil
	}
	return resp, nil
}

func (s *S3Service) putBucketVersioning(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	bucketName := s.extractBucketName(req)
	if bucketName == "" {
		return s.errorResponse(400, "InvalidBucketName", "Bucket name is required"), nil
	}

	// Store versioning configuration
	stateKey := "s3:" + bucketName + ":versioning"
	versioning := map[string]interface{}{
		"Status": "Enabled",
	}

	if err := s.state.Set(stateKey, versioning); err != nil {
		return s.errorResponse(500, "InternalError", "Failed to put bucket versioning"), nil
	}

	return &emulator.AWSResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/xml"},
		Body:       []byte{},
	}, nil
}

func (s *S3Service) getBucketVersioning(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	bucketName := s.extractBucketName(req)
	if bucketName == "" {
		return s.errorResponse(400, "InvalidBucketName", "Bucket name is required"), nil
	}

	stateKey := "s3:" + bucketName + ":versioning"
	var versioning map[string]interface{}
	err := s.state.Get(stateKey, &versioning)

	result := VersioningConfiguration{
		Xmlns: "http://s3.amazonaws.com/doc/2006-03-01/",
	}

	if err == nil {
		if status, ok := versioning["Status"].(string); ok {
			result.Status = status
		}
	}
	// When versioning has never been enabled, Status is omitted (empty)

	resp, err := emulator.BuildS3StructResponse(result)
	if err != nil {
		return s.errorResponse(500, "InternalError", "Failed to marshal response"), nil
	}
	return resp, nil
}

func (s *S3Service) putBucketEncryption(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	bucketName := s.extractBucketName(req)
	if bucketName == "" {
		return s.errorResponse(400, "InvalidBucketName", "Bucket name is required"), nil
	}

	// Parse the encryption configuration from request body
	var encryptionConfig XMLServerSideEncryptionConfiguration
	if err := xml.Unmarshal(req.Body, &encryptionConfig); err != nil {
		return s.errorResponse(400, "MalformedXML", "The XML you provided was not well-formed"), nil
	}

	// Extract encryption settings from parsed config
	encryption := map[string]interface{}{
		"Algorithm":        "AES256",
		"BucketKeyEnabled": false,
		"KMSMasterKeyID":   "",
	}

	if len(encryptionConfig.Rules) > 0 {
		rule := encryptionConfig.Rules[0]
		if rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm != "" {
			encryption["Algorithm"] = rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm
		}
		if rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID != "" {
			encryption["KMSMasterKeyID"] = rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID
		}
		encryption["BucketKeyEnabled"] = rule.BucketKeyEnabled
	}

	// Store encryption configuration
	stateKey := "s3:" + bucketName + ":encryption"
	if err := s.state.Set(stateKey, encryption); err != nil {
		return s.errorResponse(500, "InternalError", "Failed to put bucket encryption"), nil
	}

	// AWS S3 PutBucketEncryption returns 200 OK on success
	return &emulator.AWSResponse{
		StatusCode: 200,
		Headers:    map[string]string{},
		Body:       []byte{},
	}, nil
}

func (s *S3Service) getBucketEncryption(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	bucketName := s.extractBucketName(req)
	if bucketName == "" {
		return s.errorResponse(400, "InvalidBucketName", "Bucket name is required"), nil
	}

	stateKey := "s3:" + bucketName + ":encryption"
	var encryption map[string]interface{}
	err := s.state.Get(stateKey, &encryption)
	if err != nil {
		return s.errorResponse(404, "ServerSideEncryptionConfigurationNotFoundError", "The server side encryption configuration was not found"), nil
	}

	// Get algorithm from stored state, default to AES256
	algorithm := "AES256"
	if alg, ok := encryption["Algorithm"].(string); ok {
		algorithm = alg
	}

	// Get KMS key ID if set
	kmsMasterKeyID := ""
	if keyID, ok := encryption["KMSMasterKeyID"].(string); ok {
		kmsMasterKeyID = keyID
	}

	// Get BucketKeyEnabled flag
	bucketKeyEnabled := false
	if enabled, ok := encryption["BucketKeyEnabled"].(bool); ok {
		bucketKeyEnabled = enabled
	}

	result := XMLServerSideEncryptionConfiguration{
		Xmlns: "http://s3.amazonaws.com/doc/2006-03-01/",
		Rules: []XMLServerSideEncryptionRule{
			{
				ApplyServerSideEncryptionByDefault: XMLApplyServerSideEncryptionByDefault{
					SSEAlgorithm:   algorithm,
					KMSMasterKeyID: kmsMasterKeyID,
				},
				BucketKeyEnabled: bucketKeyEnabled,
			},
		},
	}

	resp, err := emulator.BuildS3StructResponse(result)
	if err != nil {
		return s.errorResponse(500, "InternalError", "Failed to marshal response"), nil
	}
	return resp, nil
}

func (s *S3Service) putPublicAccessBlock(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	bucketName := s.extractBucketName(req)
	if bucketName == "" {
		return s.errorResponse(400, "InvalidBucketName", "Bucket name is required"), nil
	}

	// Store public access block configuration
	stateKey := "s3:" + bucketName + ":publicAccessBlock"
	publicAccessBlock := map[string]interface{}{
		"BlockPublicAcls":       true,
		"BlockPublicPolicy":     true,
		"IgnorePublicAcls":      true,
		"RestrictPublicBuckets": true,
	}

	if err := s.state.Set(stateKey, publicAccessBlock); err != nil {
		return s.errorResponse(500, "InternalError", "Failed to put public access block"), nil
	}

	// AWS S3 PutPublicAccessBlock returns 204 No Content on success
	return &emulator.AWSResponse{
		StatusCode: 204,
		Headers:    map[string]string{},
		Body:       []byte{},
	}, nil
}

func (s *S3Service) deletePublicAccessBlock(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	bucketName := s.extractBucketName(req)
	if bucketName == "" {
		return s.errorResponse(400, "InvalidBucketName", "Bucket name is required"), nil
	}

	// Delete public access block configuration from state
	stateKey := "s3:" + bucketName + ":publicAccessBlock"
	// Ignore errors - AWS returns 204 even if the configuration doesn't exist
	_ = s.state.Delete(stateKey)

	// AWS S3 DeletePublicAccessBlock returns 204 No Content on success
	return &emulator.AWSResponse{
		StatusCode: 204,
		Headers:    map[string]string{},
		Body:       []byte{},
	}, nil
}

func (s *S3Service) getPublicAccessBlock(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	bucketName := s.extractBucketName(req)
	if bucketName == "" {
		return s.errorResponse(400, "InvalidBucketName", "Bucket name is required"), nil
	}

	stateKey := "s3:" + bucketName + ":publicAccessBlock"
	var config map[string]interface{}
	err := s.state.Get(stateKey, &config)
	if err != nil {
		return s.errorResponse(404, "NoSuchPublicAccessBlockConfiguration", "The public access block configuration was not found"), nil
	}

	// Get values from stored state, default to true
	blockPublicAcls := true
	if v, ok := config["BlockPublicAcls"].(bool); ok {
		blockPublicAcls = v
	}
	blockPublicPolicy := true
	if v, ok := config["BlockPublicPolicy"].(bool); ok {
		blockPublicPolicy = v
	}
	ignorePublicAcls := true
	if v, ok := config["IgnorePublicAcls"].(bool); ok {
		ignorePublicAcls = v
	}
	restrictPublicBuckets := true
	if v, ok := config["RestrictPublicBuckets"].(bool); ok {
		restrictPublicBuckets = v
	}

	result := XMLPublicAccessBlockConfiguration{
		Xmlns:                 "http://s3.amazonaws.com/doc/2006-03-01/",
		BlockPublicAcls:       blockPublicAcls,
		BlockPublicPolicy:     blockPublicPolicy,
		IgnorePublicAcls:      ignorePublicAcls,
		RestrictPublicBuckets: restrictPublicBuckets,
	}

	resp, err := emulator.BuildS3StructResponse(result)
	if err != nil {
		return s.errorResponse(500, "InternalError", "Failed to marshal response"), nil
	}
	return resp, nil
}

func (s *S3Service) putObject(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	bucketName := s.extractBucketName(req)
	if bucketName == "" {
		return s.errorResponse(400, "InvalidBucketName", "Bucket name is required"), nil
	}

	// Extract object key from path
	path := strings.TrimPrefix(req.Path, "/")
	pathParts := strings.Split(path, "/")
	var objectKey string
	if len(pathParts) > 1 {
		objectKey = strings.Join(pathParts[1:], "/")
	} else if len(pathParts) == 1 && pathParts[0] != bucketName {
		objectKey = pathParts[0]
	}

	if objectKey == "" {
		return s.errorResponse(400, "InvalidKey", "Object key is required"), nil
	}

	// Store object
	stateKey := "s3:" + bucketName + ":object:" + objectKey
	object := map[string]interface{}{
		"Key":          objectKey,
		"Bucket":       bucketName,
		"Size":         len(req.Body),
		"LastModified": "2024-01-01T00:00:00Z",
		"ETag":         fmt.Sprintf("\"%s\"", uuid.New().String()[:8]),
		"Body":         string(req.Body),
	}

	if err := s.state.Set(stateKey, object); err != nil {
		return s.errorResponse(500, "InternalError", "Failed to put object"), nil
	}

	return &emulator.AWSResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/xml",
			"ETag":         object["ETag"].(string),
		},
		Body: []byte{},
	}, nil
}

func (s *S3Service) getObject(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	bucketName := s.extractBucketName(req)
	if bucketName == "" {
		return s.errorResponse(400, "InvalidBucketName", "Bucket name is required"), nil
	}

	// Extract object key from path
	path := strings.TrimPrefix(req.Path, "/")
	pathParts := strings.Split(path, "/")
	var objectKey string
	if len(pathParts) > 1 {
		objectKey = strings.Join(pathParts[1:], "/")
	} else if len(pathParts) == 1 && pathParts[0] != bucketName {
		objectKey = pathParts[0]
	}

	if objectKey == "" {
		return s.errorResponse(400, "InvalidKey", "Object key is required"), nil
	}

	stateKey := "s3:" + bucketName + ":object:" + objectKey
	var objMap map[string]interface{}
	err := s.state.Get(stateKey, &objMap)
	if err != nil {
		return s.errorResponse(404, "NoSuchKey", "The specified key does not exist"), nil
	}

	body := []byte(objMap["Body"].(string))

	return &emulator.AWSResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":   "application/octet-stream",
			"Content-Length": fmt.Sprintf("%d", len(body)),
			"ETag":           objMap["ETag"].(string),
		},
		Body: body,
	}, nil
}

func (s *S3Service) headBucket(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	bucketName := s.extractBucketName(req)
	if bucketName == "" {
		return s.errorResponse(400, "InvalidBucketName", "Bucket name is required"), nil
	}

	stateKey := "s3:" + bucketName

	// Check if bucket exists
	var bucket map[string]interface{}
	if err := s.state.Get(stateKey, &bucket); err != nil {
		return s.errorResponse(404, "NoSuchBucket", "The specified bucket does not exist"), nil
	}

	return &emulator.AWSResponse{
		StatusCode: 200,
		Headers:    map[string]string{},
		Body:       []byte{},
	}, nil
}

func (s *S3Service) listObjectsV2(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	bucketName := s.extractBucketName(req)
	if bucketName == "" {
		return s.errorResponse(400, "InvalidBucketName", "Bucket name is required"), nil
	}

	// Build ListBucketResult using struct-based response
	result := ListBucketResult{
		Xmlns:       "http://s3.amazonaws.com/doc/2006-03-01/",
		Name:        bucketName,
		Prefix:      "",
		KeyCount:    0,
		MaxKeys:     1000,
		IsTruncated: false,
		Contents:    []XMLObject{},
	}

	resp, err := emulator.BuildS3StructResponse(result)
	if err != nil {
		return s.errorResponse(500, "InternalError", "Failed to marshal response"), nil
	}
	return resp, nil
}

func (s *S3Service) errorResponse(statusCode int, code, message string) *emulator.AWSResponse {
	return emulator.BuildRESTXMLErrorResponse(statusCode, code, message)
}

// getBucketPolicy returns the bucket policy (NoSuchBucketPolicy if not set)
func (s *S3Service) getBucketPolicy(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	bucketName := s.extractBucketName(req)
	if bucketName == "" {
		return s.errorResponse(400, "InvalidBucketName", "Bucket name is required"), nil
	}

	stateKey := "s3:" + bucketName
	var bucket map[string]interface{}
	if err := s.state.Get(stateKey, &bucket); err != nil {
		return s.errorResponse(404, "NoSuchBucket", "The specified bucket does not exist"), nil
	}

	// Check if policy exists
	if policy, ok := bucket["Policy"].(string); ok && policy != "" {
		return &emulator.AWSResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: []byte(policy),
		}, nil
	}

	// No policy set - return NoSuchBucketPolicy error
	return s.errorResponse(404, "NoSuchBucketPolicy", "The bucket policy does not exist"), nil
}

// putBucketPolicy sets the bucket policy
func (s *S3Service) putBucketPolicy(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	bucketName := s.extractBucketName(req)
	if bucketName == "" {
		return s.errorResponse(400, "InvalidBucketName", "Bucket name is required"), nil
	}

	stateKey := "s3:" + bucketName
	var bucket map[string]interface{}
	if err := s.state.Get(stateKey, &bucket); err != nil {
		return s.errorResponse(404, "NoSuchBucket", "The specified bucket does not exist"), nil
	}

	// Store policy
	bucket["Policy"] = string(req.Body)
	if err := s.state.Set(stateKey, bucket); err != nil {
		return s.errorResponse(500, "InternalError", "Failed to set bucket policy"), nil
	}

	return &emulator.AWSResponse{
		StatusCode: 204,
		Headers:    map[string]string{},
		Body:       []byte{},
	}, nil
}

// deleteBucketPolicy deletes the bucket policy
func (s *S3Service) deleteBucketPolicy(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	bucketName := s.extractBucketName(req)
	if bucketName == "" {
		return s.errorResponse(400, "InvalidBucketName", "Bucket name is required"), nil
	}

	stateKey := "s3:" + bucketName
	var bucket map[string]interface{}
	if err := s.state.Get(stateKey, &bucket); err != nil {
		return s.errorResponse(404, "NoSuchBucket", "The specified bucket does not exist"), nil
	}

	// Remove policy
	delete(bucket, "Policy")
	if err := s.state.Set(stateKey, bucket); err != nil {
		return s.errorResponse(500, "InternalError", "Failed to delete bucket policy"), nil
	}

	return &emulator.AWSResponse{
		StatusCode: 204,
		Headers:    map[string]string{},
		Body:       []byte{},
	}, nil
}

// getBucketLogging returns the bucket logging configuration
func (s *S3Service) getBucketLogging(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	bucketName := s.extractBucketName(req)
	if bucketName == "" {
		return s.errorResponse(400, "InvalidBucketName", "Bucket name is required"), nil
	}

	stateKey := "s3:" + bucketName
	var bucket map[string]interface{}
	if err := s.state.Get(stateKey, &bucket); err != nil {
		return s.errorResponse(404, "NoSuchBucket", "The specified bucket does not exist"), nil
	}

	result := BucketLoggingStatus{
		Xmlns: "http://s3.amazonaws.com/doc/2006-03-01/",
	}

	// Check if logging is configured
	if logging, ok := bucket["Logging"].(map[string]interface{}); ok {
		targetBucket := ""
		targetPrefix := ""
		if tb, ok := logging["TargetBucket"].(string); ok {
			targetBucket = tb
		}
		if tp, ok := logging["TargetPrefix"].(string); ok {
			targetPrefix = tp
		}

		result.LoggingEnabled = &XMLLoggingEnabled{
			TargetBucket: targetBucket,
			TargetPrefix: targetPrefix,
		}
	}
	// When logging is not configured, LoggingEnabled is omitted

	resp, err := emulator.BuildS3StructResponse(result)
	if err != nil {
		return s.errorResponse(500, "InternalError", "Failed to marshal response"), nil
	}
	return resp, nil
}

// putBucketLogging sets the bucket logging configuration
func (s *S3Service) putBucketLogging(ctx context.Context, params map[string]interface{}, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	bucketName := s.extractBucketName(req)
	if bucketName == "" {
		return s.errorResponse(400, "InvalidBucketName", "Bucket name is required"), nil
	}

	stateKey := "s3:" + bucketName
	var bucket map[string]interface{}
	if err := s.state.Get(stateKey, &bucket); err != nil {
		return s.errorResponse(404, "NoSuchBucket", "The specified bucket does not exist"), nil
	}

	// Parse logging configuration from XML body
	type LoggingEnabled struct {
		TargetBucket string `xml:"TargetBucket"`
		TargetPrefix string `xml:"TargetPrefix"`
	}
	type BucketLoggingStatus struct {
		XMLName        xml.Name       `xml:"BucketLoggingStatus"`
		LoggingEnabled LoggingEnabled `xml:"LoggingEnabled"`
	}

	var loggingConfig BucketLoggingStatus
	if err := xml.Unmarshal(req.Body, &loggingConfig); err == nil {
		// Store logging configuration
		bucket["Logging"] = map[string]interface{}{
			"TargetBucket": loggingConfig.LoggingEnabled.TargetBucket,
			"TargetPrefix": loggingConfig.LoggingEnabled.TargetPrefix,
		}
	} else {
		// Empty body or invalid XML - disable logging
		delete(bucket, "Logging")
	}

	if err := s.state.Set(stateKey, bucket); err != nil {
		return s.errorResponse(500, "InternalError", "Failed to set bucket logging"), nil
	}

	// AWS S3 PutBucketLogging returns 204 No Content on success
	return &emulator.AWSResponse{
		StatusCode: 204,
		Headers:    map[string]string{},
		Body:       []byte{},
	}, nil
}

// =====================================================
// S3 Control API Support
// =====================================================

// isS3ControlRequest checks if the request is for the S3 Control API
// S3 Control uses the x-amz-account-id header and /v20180820/ path prefix
func (s *S3Service) isS3ControlRequest(req *emulator.AWSRequest) bool {
	// Check for x-amz-account-id header (S3 Control requires this)
	if _, ok := req.Headers["X-Amz-Account-Id"]; ok {
		return true
	}
	if _, ok := req.Headers["x-amz-account-id"]; ok {
		return true
	}

	// Check for S3 Control API path pattern
	return strings.HasPrefix(req.Path, "/v20180820/")
}

// handleS3ControlRequest handles S3 Control API requests
func (s *S3Service) handleS3ControlRequest(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	// S3 Control API uses REST paths like:
	// GET /v20180820/tags/{resourceArn+} - GetBucketTagging / ListTagsForResource
	// PUT /v20180820/tags/{resourceArn+} - PutBucketTagging / PutResourceTagging
	// DELETE /v20180820/tags/{resourceArn+} - DeleteBucketTagging

	path := req.Path

	// Handle tagging operations
	if strings.Contains(path, "/v20180820/tags/") {
		switch req.Method {
		case "GET":
			return s.s3ControlGetResourceTagging(ctx, req)
		case "PUT":
			return s.s3ControlPutResourceTagging(ctx, req)
		case "DELETE":
			return s.s3ControlDeleteResourceTagging(ctx, req)
		}
	}

	return s.errorResponse(400, "InvalidAction", fmt.Sprintf("Unknown S3 Control action for path: %s", path)), nil
}

// extractResourceArnFromPath extracts the resource ARN from S3 Control API paths
// Path format: /v20180820/tags/{resourceArn+}
// Example: /v20180820/tags/arn%3Aaws%3As3%3Aus-east-1%3A123456789012%3Abucket%2Fmy-bucket
func (s *S3Service) extractResourceArnFromPath(path string) string {
	// Remove the /v20180820/tags/ prefix
	if idx := strings.Index(path, "/v20180820/tags/"); idx >= 0 {
		arnEncoded := path[idx+len("/v20180820/tags/"):]
		// URL decode the ARN
		arnDecoded, err := url.PathUnescape(arnEncoded)
		if err != nil {
			return arnEncoded
		}
		return arnDecoded
	}
	return ""
}

// extractBucketNameFromArn extracts the bucket name from an S3 ARN
// ARN formats:
// - arn:aws:s3:::bucket-name
// - arn:aws:s3:us-east-1:123456789012:bucket/bucket-name
func (s *S3Service) extractBucketNameFromArn(arn string) string {
	// Format: arn:aws:s3:::bucket-name
	if strings.HasPrefix(arn, "arn:aws:s3:::") {
		return strings.TrimPrefix(arn, "arn:aws:s3:::")
	}

	// Format: arn:aws:s3:region:account:bucket/bucket-name
	parts := strings.Split(arn, ":")
	if len(parts) >= 6 && parts[2] == "s3" {
		resource := parts[5]
		// Handle "bucket/bucket-name" format
		if strings.HasPrefix(resource, "bucket/") {
			return strings.TrimPrefix(resource, "bucket/")
		}
		return resource
	}

	return ""
}

// s3ControlGetResourceTagging handles GetBucketTagging / ListTagsForResource
func (s *S3Service) s3ControlGetResourceTagging(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	resourceArn := s.extractResourceArnFromPath(req.Path)
	bucketName := s.extractBucketNameFromArn(resourceArn)

	if bucketName == "" {
		return s.errorResponse(400, "InvalidRequest", "Could not extract bucket name from resource ARN"), nil
	}

	// Get tags from state
	stateKey := "s3:" + bucketName + ":tags"
	var tags map[string]string
	err := s.state.Get(stateKey, &tags)
	if err != nil {
		// No tags - return empty tag set
		tags = make(map[string]string)
	}

	// Build S3 Control tagging response XML using type-safe marshaling
	result := XMLGetBucketTaggingOutput{
		Xmlns: "http://s3.amazonaws.com/doc/2006-03-01/",
		TagSet: XMLTagSet{
			Tags: make([]XMLTag, 0, len(tags)),
		},
	}

	for key, value := range tags {
		result.TagSet.Tags = append(result.TagSet.Tags, XMLTag{
			Key:   key,
			Value: value,
		})
	}

	resp, err := emulator.BuildS3ControlStructResponse(result)
	if err != nil {
		return s.errorResponse(500, "InternalError", "Failed to marshal response"), nil
	}
	return resp, nil
}

// s3ControlPutResourceTagging handles PutBucketTagging / PutResourceTagging
func (s *S3Service) s3ControlPutResourceTagging(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	resourceArn := s.extractResourceArnFromPath(req.Path)
	bucketName := s.extractBucketNameFromArn(resourceArn)

	if bucketName == "" {
		return s.errorResponse(400, "InvalidRequest", "Could not extract bucket name from resource ARN"), nil
	}

	// Parse tags from XML body
	type Tag struct {
		Key   string `xml:"Key"`
		Value string `xml:"Value"`
	}
	type TagSet struct {
		Tags []Tag `xml:"Tag"`
	}
	type Tagging struct {
		XMLName xml.Name `xml:"Tagging"`
		TagSet  TagSet   `xml:"TagSet"`
	}

	var tagging Tagging
	if err := xml.Unmarshal(req.Body, &tagging); err != nil {
		return s.errorResponse(400, "MalformedXML", "The XML you provided was not well-formed"), nil
	}

	// Convert to map
	tags := make(map[string]string)
	for _, tag := range tagging.TagSet.Tags {
		tags[tag.Key] = tag.Value
	}

	// Store tags in state
	stateKey := "s3:" + bucketName + ":tags"
	if err := s.state.Set(stateKey, tags); err != nil {
		return s.errorResponse(500, "InternalError", "Failed to store tags"), nil
	}

	return &emulator.AWSResponse{
		StatusCode: 204,
		Headers:    map[string]string{},
		Body:       []byte{},
	}, nil
}

// s3ControlDeleteResourceTagging handles DeleteBucketTagging
func (s *S3Service) s3ControlDeleteResourceTagging(ctx context.Context, req *emulator.AWSRequest) (*emulator.AWSResponse, error) {
	resourceArn := s.extractResourceArnFromPath(req.Path)
	bucketName := s.extractBucketNameFromArn(resourceArn)

	if bucketName == "" {
		return s.errorResponse(400, "InvalidRequest", "Could not extract bucket name from resource ARN"), nil
	}

	// Delete tags from state
	stateKey := "s3:" + bucketName + ":tags"
	_ = s.state.Delete(stateKey) // Ignore error if tags don't exist

	return &emulator.AWSResponse{
		StatusCode: 204,
		Headers:    map[string]string{},
		Body:       []byte{},
	}, nil
}
