package lambda

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

const (
	// Default values for Lambda functions
	DefaultTimeout    = int32(3)
	DefaultMemorySize = int32(128)
	DefaultRuntime    = "nodejs20.x"
	DefaultPackageType = "Zip"

	// State values
	StatePending  = "Pending"
	StateActive   = "Active"
	StateInactive = "Inactive"
	StateFailed   = "Failed"

	// Package types
	PackageTypeZip   = "Zip"
	PackageTypeImage = "Image"

	// Invocation types
	InvocationTypeRequestResponse = "RequestResponse"
	InvocationTypeEvent           = "Event"
	InvocationTypeDryRun          = "DryRun"

	// Auth types for function URLs
	AuthTypeNone   = "NONE"
	AuthTypeIAM    = "AWS_IAM"

	// Invoke modes
	InvokeModeBuffered    = "BUFFERED"
	InvokeModeResponseStream = "RESPONSE_STREAM"

	// Default account ID and region for mock
	DefaultAccountID = "000000000000"
	DefaultRegion    = "us-east-1"
)

// Valid runtimes (subset - not exhaustive)
var validRuntimes = map[string]bool{
	"nodejs20.x":    true,
	"nodejs18.x":    true,
	"nodejs16.x":    true,
	"python3.12":    true,
	"python3.11":    true,
	"python3.10":    true,
	"python3.9":     true,
	"python3.8":     true,
	"java21":        true,
	"java17":        true,
	"java11":        true,
	"dotnet8":       true,
	"dotnet6":       true,
	"go1.x":         true,
	"provided.al2":  true,
	"provided.al2023": true,
	"ruby3.3":       true,
	"ruby3.2":       true,
}

// Valid architectures
var validArchitectures = map[string]bool{
	"x86_64": true,
	"arm64":  true,
}

// functionNamePattern validates function names
var functionNamePattern = regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)

// generateFunctionArn generates a Lambda function ARN
func generateFunctionArn(functionName string) string {
	return fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s",
		DefaultRegion, DefaultAccountID, functionName)
}

// generateVersionArn generates an ARN for a specific function version
func generateVersionArn(functionName, version string) string {
	return fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s:%s",
		DefaultRegion, DefaultAccountID, functionName, version)
}

// generateAliasArn generates an ARN for a function alias
func generateAliasArn(functionName, aliasName string) string {
	return fmt.Sprintf("arn:aws:lambda:%s:%s:function:%s:%s",
		DefaultRegion, DefaultAccountID, functionName, aliasName)
}

// generateLayerArn generates a Lambda layer ARN
func generateLayerArn(layerName string, version int64) string {
	return fmt.Sprintf("arn:aws:lambda:%s:%s:layer:%s:%d",
		DefaultRegion, DefaultAccountID, layerName, version)
}

// generateFunctionUrl generates a function URL
func generateFunctionUrl(functionName string) string {
	// Real AWS format: https://<url-id>.lambda-url.<region>.on.aws/
	// For mock, we use a simpler format
	urlId := strings.ToLower(uuid.New().String()[:12])
	return fmt.Sprintf("https://%s.lambda-url.%s.on.aws/", urlId, DefaultRegion)
}

// generateRevisionId generates a new revision ID
func generateRevisionId() string {
	return uuid.New().String()
}

// generateCodeSha256 generates a SHA256 hash for code
func generateCodeSha256(code *FunctionCode) string {
	var data string
	if code != nil {
		if code.ZipFile != "" {
			data = code.ZipFile
		} else if code.S3Bucket != "" && code.S3Key != "" {
			data = code.S3Bucket + code.S3Key + code.S3ObjectVersion
		} else if code.ImageUri != "" {
			data = code.ImageUri
		}
	}
	if data == "" {
		data = uuid.New().String() // Generate random hash for empty code
	}
	hash := sha256.Sum256([]byte(data))
	return base64.StdEncoding.EncodeToString(hash[:])
}

// estimateCodeSize estimates the code size
func estimateCodeSize(code *FunctionCode) int64 {
	if code == nil {
		return 0
	}
	if code.ZipFile != "" {
		// Base64-encoded, so actual size is ~75% of encoded length
		decoded, err := base64.StdEncoding.DecodeString(code.ZipFile)
		if err == nil {
			return int64(len(decoded))
		}
		return int64(len(code.ZipFile) * 3 / 4)
	}
	// Mock size for S3 or image-based functions
	return 1024 * 100 // 100KB default
}

// validateFunctionName validates the function name
func validateFunctionName(name string) error {
	if name == "" {
		return fmt.Errorf("function name is required")
	}
	if len(name) > 64 {
		return fmt.Errorf("function name must be 64 characters or fewer")
	}
	if !functionNamePattern.MatchString(name) {
		return fmt.Errorf("function name can contain only alphanumeric characters, hyphens, and underscores")
	}
	return nil
}

// validateRuntime validates the runtime
func validateRuntime(runtime string) error {
	if runtime == "" {
		return nil // Runtime is optional for container images
	}
	if !validRuntimes[runtime] {
		return fmt.Errorf("invalid runtime: %s", runtime)
	}
	return nil
}

// validateArchitectures validates the architectures
func validateArchitectures(archs []string) error {
	if len(archs) == 0 {
		return nil
	}
	for _, arch := range archs {
		if !validArchitectures[arch] {
			return fmt.Errorf("invalid architecture: %s", arch)
		}
	}
	return nil
}

// validateRole validates the IAM role ARN
func validateRole(role string) error {
	if role == "" {
		return fmt.Errorf("role is required")
	}
	if !strings.HasPrefix(role, "arn:aws:iam::") {
		return fmt.Errorf("invalid role ARN format")
	}
	return nil
}

// validateTimeout validates the timeout value
func validateTimeout(timeout *int32) error {
	if timeout == nil {
		return nil
	}
	if *timeout < 1 || *timeout > 900 {
		return fmt.Errorf("timeout must be between 1 and 900 seconds")
	}
	return nil
}

// validateMemorySize validates the memory size
func validateMemorySize(memorySize *int32) error {
	if memorySize == nil {
		return nil
	}
	if *memorySize < 128 || *memorySize > 10240 {
		return fmt.Errorf("memory size must be between 128 MB and 10,240 MB")
	}
	// Memory must be a multiple of 64 MB
	if *memorySize%64 != 0 {
		return fmt.Errorf("memory size must be a multiple of 64 MB")
	}
	return nil
}

// validatePackageType validates the package type
func validatePackageType(packageType string) error {
	if packageType == "" {
		return nil
	}
	if packageType != PackageTypeZip && packageType != PackageTypeImage {
		return fmt.Errorf("invalid package type: %s", packageType)
	}
	return nil
}

// validateCode validates the function code
func validateCode(code *FunctionCode, packageType string) error {
	if code == nil {
		return fmt.Errorf("code is required")
	}

	switch packageType {
	case PackageTypeImage:
		if code.ImageUri == "" {
			return fmt.Errorf("ImageUri is required for Image package type")
		}
	default: // Zip
		hasZip := code.ZipFile != ""
		hasS3 := code.S3Bucket != "" && code.S3Key != ""
		if !hasZip && !hasS3 {
			return fmt.Errorf("either ZipFile or S3Bucket/S3Key is required")
		}
		if hasZip && hasS3 {
			return fmt.Errorf("cannot specify both ZipFile and S3Bucket/S3Key")
		}
	}
	return nil
}

// validateAliasName validates an alias name
func validateAliasName(name string) error {
	if name == "" {
		return fmt.Errorf("alias name is required")
	}
	if len(name) > 128 {
		return fmt.Errorf("alias name must be 128 characters or fewer")
	}
	// Aliases can't be named $LATEST
	if name == "$LATEST" {
		return fmt.Errorf("alias name cannot be $LATEST")
	}
	return nil
}

// parseFunctionNameFromArn extracts the function name from an ARN
func parseFunctionNameFromArn(arn string) string {
	// ARN format: arn:aws:lambda:region:account:function:name[:qualifier]
	parts := strings.Split(arn, ":")
	if len(parts) >= 7 && parts[5] == "function" {
		return parts[6]
	}
	return ""
}

// getDefaultArchitectures returns the default architectures
func getDefaultArchitectures() []string {
	return []string{"x86_64"}
}

// coalesce returns the first non-empty string
func coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// coalesceInt32 returns the first non-nil int32 pointer's value, or the default
func coalesceInt32(defaultVal int32, values ...*int32) int32 {
	for _, v := range values {
		if v != nil {
			return *v
		}
	}
	return defaultVal
}

// parseQueryParams extracts query parameters from a path that may contain a query string
func parseQueryParams(path string) url.Values {
	if idx := strings.Index(path, "?"); idx >= 0 {
		queryString := path[idx+1:]
		values, err := url.ParseQuery(queryString)
		if err == nil {
			return values
		}
	}
	return url.Values{}
}

// getPathWithoutQuery returns the path without the query string
func getPathWithoutQuery(path string) string {
	if idx := strings.Index(path, "?"); idx >= 0 {
		return path[:idx]
	}
	return path
}
