package emulator

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ProtocolType represents the AWS API protocol type
type ProtocolType string

const (
	ProtocolQuery    ProtocolType = "query"     // Query Protocol (RDS, EC2, SQS)
	ProtocolJSON     ProtocolType = "json"      // JSON Protocol (DynamoDB, CloudWatch)
	ProtocolRESTXML  ProtocolType = "rest-xml"  // REST-XML Protocol (S3)
	ProtocolRESTJSON ProtocolType = "rest-json" // REST-JSON Protocol (Lambda, API Gateway)
)

// ResponseBuilderConfig holds configuration for building responses
type ResponseBuilderConfig struct {
	ServiceName string
	Namespace   string // XML namespace URL
	Version     string // API version
}

// BuildQueryResponse builds a Query Protocol (XML) response
// Used by RDS, EC2, SQS, IAM services
//
// The data parameter should be a struct that marshals to the <ActionResult> element.
// For example, for CreateRole, pass CreateRoleResult{Role: role} which marshals to
// <CreateRoleResult><Role>...</Role></CreateRoleResult>
//
// The response will be wrapped with <ActionResponse> and <ResponseMetadata>.
func BuildQueryResponse(action string, data interface{}, config ResponseBuilderConfig) (*AWSResponse, error) {
	// Marshal the data to XML - this should produce the <ActionResult> element
	dataXML, err := xml.MarshalIndent(data, "  ", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response data: %w", err)
	}

	// Build namespace URL if not provided
	namespace := config.Namespace
	if namespace == "" && config.ServiceName != "" {
		namespace = fmt.Sprintf("http://%s.amazonaws.com/doc/%s/", config.ServiceName, config.Version)
	}
	if namespace == "" {
		namespace = "http://rds.amazonaws.com/doc/2014-10-31/" // Default
	}

	// Construct the proper AWS Query Protocol XML response
	// Note: The data should already marshal to <ActionResult>, we just wrap with
	// <ActionResponse> and add <ResponseMetadata>
	requestID := uuid.New().String()
	responseXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<%sResponse xmlns="%s">
  %s
  <ResponseMetadata>
    <RequestId>%s</RequestId>
  </ResponseMetadata>
</%sResponse>`, action, namespace, string(dataXML), requestID, action)

	return &AWSResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "text/xml",
		},
		Body: []byte(responseXML),
	}, nil
}

// BuildEC2Response builds an EC2 Query Protocol response
// EC2 uses a different format than RDS - no <ActionResult> wrapper
// The response data should already have the correct root element (e.g., RunInstancesResponse)
// Adds requestId element inside the response for AWS SDK compatibility
func BuildEC2Response(data interface{}, config ResponseBuilderConfig) (*AWSResponse, error) {
	// Marshal the data to XML
	dataXML, err := xml.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response data: %w", err)
	}

	// Build namespace URL if not provided
	namespace := config.Namespace
	if namespace == "" && config.ServiceName != "" {
		namespace = fmt.Sprintf("http://%s.amazonaws.com/doc/%s/", config.ServiceName, config.Version)
	}
	if namespace == "" {
		namespace = "http://ec2.amazonaws.com/doc/2016-11-15/"
	}

	// Generate RequestId
	requestID := uuid.New().String()

	// For EC2, the data struct should have the XMLName set to the action response name
	// We just need to add the XML declaration and namespace
	xmlStr := string(dataXML)

	// Add namespace to root element if not present
	if !strings.Contains(xmlStr, "xmlns=") {
		// Find the first > and insert namespace before it
		idx := strings.Index(xmlStr, ">")
		if idx > 0 {
			xmlStr = xmlStr[:idx] + fmt.Sprintf(` xmlns="%s"`, namespace) + xmlStr[idx:]
		}
	}

	// Add requestId before the closing tag (EC2 uses lowercase requestId)
	// Find the last </ sequence to insert requestId before it
	lastClose := strings.LastIndex(xmlStr, "</")
	if lastClose > 0 {
		xmlStr = xmlStr[:lastClose] + fmt.Sprintf("  <requestId>%s</requestId>\n", requestID) + xmlStr[lastClose:]
	}

	responseXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
%s`, xmlStr)

	return &AWSResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "text/xml",
		},
		Body: []byte(responseXML),
	}, nil
}

// BuildJSONResponse builds a JSON Protocol response
// Used by DynamoDB, CloudWatch services
func BuildJSONResponse(statusCode int, data interface{}) (*AWSResponse, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON response: %w", err)
	}

	requestID := uuid.New().String()
	headers := map[string]string{
		"Content-Type":     "application/x-amz-json-1.0",
		"x-amzn-RequestId": requestID,
	}

	// Add CRC32 header for DynamoDB (optional)
	// headers["x-amz-crc32"] = "0"

	return &AWSResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       body,
	}, nil
}

// BuildRESTXMLResponse builds a REST-XML Protocol response
// Used by S3 service
func BuildRESTXMLResponse(rootElement string, data interface{}, namespace string) (*AWSResponse, error) {
	// Default namespace for S3
	if namespace == "" {
		namespace = "http://s3.amazonaws.com/doc/2006-03-01/"
	}

	// Marshal the data to XML
	dataXML, err := xml.MarshalIndent(data, "    ", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal XML response: %w", err)
	}

	// Construct REST-XML response (no wrapper, direct element)
	responseXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<%s xmlns="%s">
    %s
</%s>`, rootElement, namespace, string(dataXML), rootElement)

	return &AWSResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/xml",
		},
		Body: []byte(responseXML),
	}, nil
}

// BuildS3StructResponse builds an S3 XML response from a struct.
// This is the ONLY supported method for building S3 responses - do NOT use
// manual XML string construction with fmt.Sprintf.
//
// The struct MUST have:
//   - XMLName field with the root element name
//   - Xmlns field with the S3 namespace as attribute
//
// Example struct:
//
//	type VersioningConfiguration struct {
//	    XMLName xml.Name `xml:"VersioningConfiguration"`
//	    Xmlns   string   `xml:"xmlns,attr"`
//	    Status  string   `xml:"Status,omitempty"`
//	}
//
// Usage:
//
//	result := VersioningConfiguration{
//	    Xmlns:  "http://s3.amazonaws.com/doc/2006-03-01/",
//	    Status: "Enabled",
//	}
//	resp, err := BuildS3StructResponse(result)
func BuildS3StructResponse(data interface{}) (*AWSResponse, error) {
	xmlBytes, err := xml.MarshalIndent(data, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal S3 response: %w", err)
	}

	return &AWSResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/xml",
		},
		Body: append([]byte(xml.Header), xmlBytes...),
	}, nil
}

// BuildS3ControlStructResponse builds an S3 Control XML response from a struct.
// This is the ONLY supported method for building S3 Control responses.
// Used for S3 Control operations like tagging (GetBucketTagging, PutBucketTagging, etc.)
//
// The struct MUST have XMLName and Xmlns fields, similar to BuildS3StructResponse.
func BuildS3ControlStructResponse(data interface{}) (*AWSResponse, error) {
	xmlBytes, err := xml.MarshalIndent(data, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal S3 Control response: %w", err)
	}

	return &AWSResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/xml",
		},
		Body: append([]byte(xml.Header), xmlBytes...),
	}, nil
}

// BuildRESTJSONResponse builds a REST-JSON Protocol response
// Used by Lambda, API Gateway services
func BuildRESTJSONResponse(statusCode int, data interface{}) (*AWSResponse, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON response: %w", err)
	}

	requestID := uuid.New().String()

	return &AWSResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":     "application/json",
			"x-amzn-RequestId": requestID,
		},
		Body: body,
	}, nil
}

// BuildQueryErrorResponse builds a Query Protocol error response
func BuildQueryErrorResponse(statusCode int, code, message string) *AWSResponse {
	requestID := uuid.New().String()
	errorXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ErrorResponse>
    <Error>
        <Code>%s</Code>
        <Message>%s</Message>
    </Error>
    <RequestId>%s</RequestId>
</ErrorResponse>`, code, message, requestID)

	return &AWSResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "text/xml",
		},
		Body: []byte(errorXML),
	}
}

// BuildEC2ErrorResponse builds an EC2-specific error response
// EC2 uses a different error format: <Response><Errors><Error>...</Error></Errors></Response>
func BuildEC2ErrorResponse(statusCode int, code, message string) *AWSResponse {
	requestID := uuid.New().String()
	errorXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Response>
  <Errors>
    <Error>
      <Code>%s</Code>
      <Message>%s</Message>
    </Error>
  </Errors>
  <RequestId>%s</RequestId>
</Response>`, code, message, requestID)

	return &AWSResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "text/xml",
		},
		Body: []byte(errorXML),
	}
}

// BuildJSONErrorResponse builds a JSON Protocol error response
func BuildJSONErrorResponse(statusCode int, code, message string) *AWSResponse {
	errorData := map[string]interface{}{
		"__type":  code,
		"message": message,
	}

	body, _ := json.Marshal(errorData)
	requestID := uuid.New().String()

	return &AWSResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":     "application/x-amz-json-1.0",
			"x-amzn-RequestId": requestID,
			"x-amzn-ErrorType": code,
		},
		Body: body,
	}
}

// BuildRESTXMLErrorResponse builds a REST-XML Protocol error response
// Used by S3 - includes x-amz-request-id and x-amz-id-2 headers
func BuildRESTXMLErrorResponse(statusCode int, code, message string) *AWSResponse {
	requestID := uuid.New().String()
	hostID := uuid.New().String()

	errorXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Error>
    <Code>%s</Code>
    <Message>%s</Message>
    <RequestId>%s</RequestId>
    <HostId>%s</HostId>
</Error>`, code, message, requestID, hostID)

	return &AWSResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":     "application/xml",
			"x-amz-request-id": requestID,
			"x-amz-id-2":       hostID,
		},
		Body: []byte(errorXML),
	}
}

// BuildRESTJSONErrorResponse builds a REST-JSON Protocol error response
func BuildRESTJSONErrorResponse(statusCode int, code, message string) *AWSResponse {
	errorData := map[string]interface{}{
		"Type":    "User",
		"message": message,
	}

	body, _ := json.Marshal(errorData)
	requestID := uuid.New().String()

	return &AWSResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":     "application/json",
			"x-amzn-RequestId": requestID,
		},
		Body: body,
	}
}

// GetProtocolForService returns the protocol type for a given service name
func GetProtocolForService(serviceName string) ProtocolType {
	serviceName = strings.ToLower(serviceName)

	switch serviceName {
	case "rds", "ec2":
		return ProtocolQuery
	case "dynamodb", "cloudwatch", "sqs":
		// SQS uses JSON protocol in AWS SDK v2
		return ProtocolJSON
	case "s3":
		return ProtocolRESTXML
	case "lambda", "apigateway":
		return ProtocolRESTJSON
	default:
		return ProtocolQuery // Default to Query protocol
	}
}

// BuildResponse builds a response using the appropriate protocol based on service name
func BuildResponse(serviceName, action string, data interface{}, config ResponseBuilderConfig) (*AWSResponse, error) {
	protocol := GetProtocolForService(serviceName)

	switch protocol {
	case ProtocolQuery:
		return BuildQueryResponse(action, data, config)
	case ProtocolJSON:
		return BuildJSONResponse(200, data)
	case ProtocolRESTXML:
		// For REST-XML, we need the root element name
		// Default to action name if not specified
		rootElement := action + "Result"
		return BuildRESTXMLResponse(rootElement, data, config.Namespace)
	case ProtocolRESTJSON:
		return BuildRESTJSONResponse(200, data)
	default:
		return BuildQueryResponse(action, data, config)
	}
}

// BuildErrorResponse builds an error response using the appropriate protocol
func BuildErrorResponse(serviceName string, statusCode int, code, message string) *AWSResponse {
	protocol := GetProtocolForService(serviceName)

	switch protocol {
	case ProtocolQuery:
		return BuildQueryErrorResponse(statusCode, code, message)
	case ProtocolJSON:
		return BuildJSONErrorResponse(statusCode, code, message)
	case ProtocolRESTXML:
		return BuildRESTXMLErrorResponse(statusCode, code, message)
	case ProtocolRESTJSON:
		return BuildRESTJSONErrorResponse(statusCode, code, message)
	default:
		return BuildQueryErrorResponse(statusCode, code, message)
	}
}
