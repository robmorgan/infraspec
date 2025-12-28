package emulator

import (
	"testing"
)

// Test types mimicking Smithy-generated input types
type CreateRoleRequest struct {
	RoleName                 *string `xml:"RoleName"`
	AssumeRolePolicyDocument *string `xml:"AssumeRolePolicyDocument"`
	Description              *string `xml:"Description,omitempty"`
	MaxSessionDuration       *int32  `xml:"MaxSessionDuration,omitempty"`
	Tags                     []Tag   `xml:"Tags>item,omitempty"`
}

type Tag struct {
	Key   *string `xml:"Key"`
	Value *string `xml:"Value"`
}

type SimpleRequest struct {
	Name  *string  `xml:"Name"`
	Count *int32   `xml:"Count"`
	Flag  *bool    `xml:"Flag"`
	Rate  *float64 `xml:"Rate"`
}

type JSONRequest struct {
	TableName string   `json:"TableName"`
	KeySchema []string `json:"KeySchema"`
}

func TestParseQueryRequest_BasicFields(t *testing.T) {
	body := []byte("RoleName=test-role&AssumeRolePolicyDocument=%7B%22Version%22%3A%222012-10-17%22%7D&Description=Test%20role")

	result, err := ParseQueryRequest[CreateRoleRequest](body)
	if err != nil {
		t.Fatalf("ParseQueryRequest failed: %v", err)
	}

	if result.RoleName == nil || *result.RoleName != "test-role" {
		t.Errorf("Expected RoleName='test-role', got %v", result.RoleName)
	}

	if result.AssumeRolePolicyDocument == nil || *result.AssumeRolePolicyDocument != `{"Version":"2012-10-17"}` {
		t.Errorf("Expected decoded policy document, got %v", result.AssumeRolePolicyDocument)
	}

	if result.Description == nil || *result.Description != "Test role" {
		t.Errorf("Expected Description='Test role', got %v", result.Description)
	}
}

func TestParseQueryRequest_NumericFields(t *testing.T) {
	body := []byte("Name=test&Count=42&Flag=true&Rate=3.14")

	result, err := ParseQueryRequest[SimpleRequest](body)
	if err != nil {
		t.Fatalf("ParseQueryRequest failed: %v", err)
	}

	if result.Name == nil || *result.Name != "test" {
		t.Errorf("Expected Name='test', got %v", result.Name)
	}

	if result.Count == nil || *result.Count != 42 {
		t.Errorf("Expected Count=42, got %v", result.Count)
	}

	if result.Flag == nil || *result.Flag != true {
		t.Errorf("Expected Flag=true, got %v", result.Flag)
	}

	if result.Rate == nil || *result.Rate != 3.14 {
		t.Errorf("Expected Rate=3.14, got %v", result.Rate)
	}
}

func TestParseQueryRequest_ListFields(t *testing.T) {
	body := []byte("RoleName=test-role&Tags.member.1.Key=env&Tags.member.1.Value=prod&Tags.member.2.Key=team&Tags.member.2.Value=platform")

	result, err := ParseQueryRequest[CreateRoleRequest](body)
	if err != nil {
		t.Fatalf("ParseQueryRequest failed: %v", err)
	}

	if len(result.Tags) != 2 {
		t.Fatalf("Expected 2 tags, got %d", len(result.Tags))
	}

	if result.Tags[0].Key == nil || *result.Tags[0].Key != "env" {
		t.Errorf("Expected Tag[0].Key='env', got %v", result.Tags[0].Key)
	}

	if result.Tags[0].Value == nil || *result.Tags[0].Value != "prod" {
		t.Errorf("Expected Tag[0].Value='prod', got %v", result.Tags[0].Value)
	}

	if result.Tags[1].Key == nil || *result.Tags[1].Key != "team" {
		t.Errorf("Expected Tag[1].Key='team', got %v", result.Tags[1].Key)
	}
}

func TestParseQueryRequest_EmptyBody(t *testing.T) {
	result, err := ParseQueryRequest[CreateRoleRequest]([]byte{})
	if err != nil {
		t.Fatalf("ParseQueryRequest failed: %v", err)
	}

	if result.RoleName != nil {
		t.Errorf("Expected nil RoleName for empty body, got %v", result.RoleName)
	}
}

func TestParseJSONRequest_BasicFields(t *testing.T) {
	body := []byte(`{"TableName": "test-table", "KeySchema": ["pk", "sk"]}`)

	result, err := ParseJSONRequest[JSONRequest](body)
	if err != nil {
		t.Fatalf("ParseJSONRequest failed: %v", err)
	}

	if result.TableName != "test-table" {
		t.Errorf("Expected TableName='test-table', got %s", result.TableName)
	}

	if len(result.KeySchema) != 2 {
		t.Errorf("Expected 2 key schema elements, got %d", len(result.KeySchema))
	}
}

func TestParseJSONRequest_EmptyBody(t *testing.T) {
	result, err := ParseJSONRequest[JSONRequest]([]byte{})
	if err != nil {
		t.Fatalf("ParseJSONRequest failed: %v", err)
	}

	if result.TableName != "" {
		t.Errorf("Expected empty TableName for empty body, got %s", result.TableName)
	}
}

func TestParseRequest_WithProtocol(t *testing.T) {
	tests := []struct {
		name     string
		protocol ProtocolType
		body     []byte
	}{
		{
			name:     "Query protocol",
			protocol: ProtocolQuery,
			body:     []byte("RoleName=test"),
		},
		{
			name:     "JSON protocol",
			protocol: ProtocolJSON,
			body:     []byte(`{"TableName": "test"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &AWSRequest{Body: tt.body}

			if tt.protocol == ProtocolQuery {
				result, err := ParseRequest[CreateRoleRequest](req, tt.protocol)
				if err != nil {
					t.Fatalf("ParseRequest failed: %v", err)
				}
				if result.RoleName == nil || *result.RoleName != "test" {
					t.Errorf("Expected RoleName='test', got %v", result.RoleName)
				}
			} else {
				result, err := ParseRequest[JSONRequest](req, tt.protocol)
				if err != nil {
					t.Fatalf("ParseRequest failed: %v", err)
				}
				if result.TableName != "test" {
					t.Errorf("Expected TableName='test', got %s", result.TableName)
				}
			}
		})
	}
}
