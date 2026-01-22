package iam

import (
	"context"
	"strings"
	"testing"

	emulator "github.com/robmorgan/infraspec/internal/emulator/core"
	testhelpers "github.com/robmorgan/infraspec/internal/emulator/testing"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// SAML Provider Tests
// ============================================================================

func TestCreateSAMLProvider_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateSAMLProvider&Name=TestSAMLProvider&SAMLMetadataDocument=<xml>test</xml>"),
		Action:  "CreateSAMLProvider",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")

	body := string(resp.Body)
	require.Contains(t, body, "<SAMLProviderArn>arn:aws:iam::123456789012:saml-provider/TestSAMLProvider</SAMLProviderArn>")
}

func TestCreateSAMLProvider_MissingName(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateSAMLProvider&SAMLMetadataDocument=<xml>test</xml>"),
		Action:  "CreateSAMLProvider",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "Name is required")
}

func TestCreateSAMLProvider_MissingMetadata(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateSAMLProvider&Name=TestProvider"),
		Action:  "CreateSAMLProvider",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "SAMLMetadataDocument is required")
}

func TestCreateSAMLProvider_AlreadyExists(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create first provider
	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateSAMLProvider&Name=DuplicateProvider&SAMLMetadataDocument=<xml>test</xml>"),
		Action:  "CreateSAMLProvider",
	}
	_, _ = service.HandleRequest(context.Background(), req)

	// Try to create again
	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 409, resp.StatusCode)
	require.Contains(t, string(resp.Body), "EntityAlreadyExists")
}

func TestGetSAMLProvider_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create provider
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateSAMLProvider&Name=GetTestProvider&SAMLMetadataDocument=<xml>metadata</xml>"),
		Action:  "CreateSAMLProvider",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Get provider
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetSAMLProvider&SAMLProviderArn=arn:aws:iam::123456789012:saml-provider/GetTestProvider"),
		Action:  "GetSAMLProvider",
	}
	resp, err := service.HandleRequest(context.Background(), getReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	body := string(resp.Body)
	// XML content is properly escaped in the response
	require.Contains(t, body, "<SAMLMetadataDocument>")
	require.Contains(t, body, "metadata")
	require.Contains(t, body, "<CreateDate>")
}

func TestGetSAMLProvider_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetSAMLProvider&SAMLProviderArn=arn:aws:iam::123456789012:saml-provider/NonexistentProvider"),
		Action:  "GetSAMLProvider",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestUpdateSAMLProvider_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create provider
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateSAMLProvider&Name=UpdateTestProvider&SAMLMetadataDocument=<xml>original</xml>"),
		Action:  "CreateSAMLProvider",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Update provider
	updateReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateSAMLProvider&SAMLProviderArn=arn:aws:iam::123456789012:saml-provider/UpdateTestProvider&SAMLMetadataDocument=<xml>updated</xml>"),
		Action:  "UpdateSAMLProvider",
	}
	resp, err := service.HandleRequest(context.Background(), updateReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	require.Contains(t, string(resp.Body), "<SAMLProviderArn>arn:aws:iam::123456789012:saml-provider/UpdateTestProvider</SAMLProviderArn>")

	// Verify update by getting
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetSAMLProvider&SAMLProviderArn=arn:aws:iam::123456789012:saml-provider/UpdateTestProvider"),
		Action:  "GetSAMLProvider",
	}
	resp, _ = service.HandleRequest(context.Background(), getReq)
	// XML content is properly escaped in the response
	require.Contains(t, string(resp.Body), "updated")
}

func TestDeleteSAMLProvider_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create provider
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateSAMLProvider&Name=DeleteTestProvider&SAMLMetadataDocument=<xml>test</xml>"),
		Action:  "CreateSAMLProvider",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Delete provider
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteSAMLProvider&SAMLProviderArn=arn:aws:iam::123456789012:saml-provider/DeleteTestProvider"),
		Action:  "DeleteSAMLProvider",
	}
	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify deletion
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetSAMLProvider&SAMLProviderArn=arn:aws:iam::123456789012:saml-provider/DeleteTestProvider"),
		Action:  "GetSAMLProvider",
	}
	resp, _ = service.HandleRequest(context.Background(), getReq)
	require.Equal(t, 404, resp.StatusCode)
}

func TestDeleteSAMLProvider_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteSAMLProvider&SAMLProviderArn=arn:aws:iam::123456789012:saml-provider/NonexistentProvider"),
		Action:  "DeleteSAMLProvider",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestListSAMLProviders_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create two providers
	for _, name := range []string{"Provider1", "Provider2"} {
		req := &emulator.AWSRequest{
			Method:  "POST",
			Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			Body:    []byte("Action=CreateSAMLProvider&Name=" + name + "&SAMLMetadataDocument=<xml>test</xml>"),
			Action:  "CreateSAMLProvider",
		}
		_, _ = service.HandleRequest(context.Background(), req)
	}

	// List providers
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListSAMLProviders"),
		Action:  "ListSAMLProviders",
	}
	resp, err := service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	body := string(resp.Body)
	require.Contains(t, body, "Provider1")
	require.Contains(t, body, "Provider2")
}

func TestListSAMLProviders_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListSAMLProviders"),
		Action:  "ListSAMLProviders",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	// Empty list can be rendered as either <SAMLProviderList/> or <SAMLProviderList></SAMLProviderList>
	require.Contains(t, string(resp.Body), "<SAMLProviderList")
}

// ============================================================================
// OIDC Provider Tests
// ============================================================================

func TestCreateOpenIDConnectProvider_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateOpenIDConnectProvider&Url=https://accounts.google.com&ThumbprintList.member.1=1234567890123456789012345678901234567890"),
		Action:  "CreateOpenIDConnectProvider",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	testhelpers.AssertContentType(t, resp, "text/xml")

	body := string(resp.Body)
	require.Contains(t, body, "<OpenIDConnectProviderArn>arn:aws:iam::123456789012:oidc-provider/accounts.google.com</OpenIDConnectProviderArn>")
}

func TestCreateOpenIDConnectProvider_MissingUrl(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateOpenIDConnectProvider&ThumbprintList.member.1=1234567890123456789012345678901234567890"),
		Action:  "CreateOpenIDConnectProvider",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "Url is required")
}

func TestCreateOpenIDConnectProvider_InvalidUrl(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateOpenIDConnectProvider&Url=http://insecure.example.com&ThumbprintList.member.1=1234567890123456789012345678901234567890"),
		Action:  "CreateOpenIDConnectProvider",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "Url must start with https://")
}

func TestCreateOpenIDConnectProvider_MissingThumbprint(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateOpenIDConnectProvider&Url=https://example.com"),
		Action:  "CreateOpenIDConnectProvider",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "At least one thumbprint is required")
}

func TestCreateOpenIDConnectProvider_InvalidThumbprint(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateOpenIDConnectProvider&Url=https://example.com&ThumbprintList.member.1=short"),
		Action:  "CreateOpenIDConnectProvider",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "Invalid thumbprint")
}

func TestCreateOpenIDConnectProvider_AlreadyExists(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateOpenIDConnectProvider&Url=https://duplicate.example.com&ThumbprintList.member.1=1234567890123456789012345678901234567890"),
		Action:  "CreateOpenIDConnectProvider",
	}

	// Create first
	_, _ = service.HandleRequest(context.Background(), req)

	// Try to create again
	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 409, resp.StatusCode)
	require.Contains(t, string(resp.Body), "EntityAlreadyExists")
}

func TestGetOpenIDConnectProvider_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create provider
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateOpenIDConnectProvider&Url=https://get.example.com&ThumbprintList.member.1=abcdef1234567890123456789012345678901234&ClientIDList.member.1=client123"),
		Action:  "CreateOpenIDConnectProvider",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Get provider
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetOpenIDConnectProvider&OpenIDConnectProviderArn=arn:aws:iam::123456789012:oidc-provider/get.example.com"),
		Action:  "GetOpenIDConnectProvider",
	}
	resp, err := service.HandleRequest(context.Background(), getReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	body := string(resp.Body)
	require.Contains(t, body, "<Url>https://get.example.com</Url>")
	require.Contains(t, body, "abcdef1234567890123456789012345678901234")
	require.Contains(t, body, "client123")
}

func TestGetOpenIDConnectProvider_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetOpenIDConnectProvider&OpenIDConnectProviderArn=arn:aws:iam::123456789012:oidc-provider/nonexistent.example.com"),
		Action:  "GetOpenIDConnectProvider",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestDeleteOpenIDConnectProvider_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create provider
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateOpenIDConnectProvider&Url=https://delete.example.com&ThumbprintList.member.1=1234567890123456789012345678901234567890"),
		Action:  "CreateOpenIDConnectProvider",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Delete provider
	deleteReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteOpenIDConnectProvider&OpenIDConnectProviderArn=arn:aws:iam::123456789012:oidc-provider/delete.example.com"),
		Action:  "DeleteOpenIDConnectProvider",
	}
	resp, err := service.HandleRequest(context.Background(), deleteReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify deletion
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetOpenIDConnectProvider&OpenIDConnectProviderArn=arn:aws:iam::123456789012:oidc-provider/delete.example.com"),
		Action:  "GetOpenIDConnectProvider",
	}
	resp, _ = service.HandleRequest(context.Background(), getReq)
	require.Equal(t, 404, resp.StatusCode)
}

func TestDeleteOpenIDConnectProvider_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=DeleteOpenIDConnectProvider&OpenIDConnectProviderArn=arn:aws:iam::123456789012:oidc-provider/nonexistent.example.com"),
		Action:  "DeleteOpenIDConnectProvider",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestListOpenIDConnectProviders_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create two providers
	for i, url := range []string{"https://provider1.example.com", "https://provider2.example.com"} {
		req := &emulator.AWSRequest{
			Method:  "POST",
			Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			Body:    []byte("Action=CreateOpenIDConnectProvider&Url=" + url + "&ThumbprintList.member.1=1234567890123456789012345678901234567890"),
			Action:  "CreateOpenIDConnectProvider",
		}
		_, _ = service.HandleRequest(context.Background(), req)
		_ = i
	}

	// List providers
	listReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListOpenIDConnectProviders"),
		Action:  "ListOpenIDConnectProviders",
	}
	resp, err := service.HandleRequest(context.Background(), listReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	body := string(resp.Body)
	require.Contains(t, body, "provider1.example.com")
	require.Contains(t, body, "provider2.example.com")
}

func TestListOpenIDConnectProviders_Empty(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=ListOpenIDConnectProviders"),
		Action:  "ListOpenIDConnectProviders",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)
	// Empty list can be rendered as either <OpenIDConnectProviderList/> or <OpenIDConnectProviderList></OpenIDConnectProviderList>
	require.Contains(t, string(resp.Body), "<OpenIDConnectProviderList")
}

func TestUpdateOpenIDConnectProviderThumbprint_Success(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create provider
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateOpenIDConnectProvider&Url=https://update.example.com&ThumbprintList.member.1=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		Action:  "CreateOpenIDConnectProvider",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Update thumbprint
	updateReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateOpenIDConnectProviderThumbprint&OpenIDConnectProviderArn=arn:aws:iam::123456789012:oidc-provider/update.example.com&ThumbprintList.member.1=bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
		Action:  "UpdateOpenIDConnectProviderThumbprint",
	}
	resp, err := service.HandleRequest(context.Background(), updateReq)
	require.NoError(t, err)

	testhelpers.AssertResponseStatus(t, resp, 200)

	// Verify update
	getReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=GetOpenIDConnectProvider&OpenIDConnectProviderArn=arn:aws:iam::123456789012:oidc-provider/update.example.com"),
		Action:  "GetOpenIDConnectProvider",
	}
	resp, _ = service.HandleRequest(context.Background(), getReq)
	body := string(resp.Body)
	require.Contains(t, body, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	require.NotContains(t, body, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
}

func TestUpdateOpenIDConnectProviderThumbprint_NotFound(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	req := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateOpenIDConnectProviderThumbprint&OpenIDConnectProviderArn=arn:aws:iam::123456789012:oidc-provider/nonexistent.example.com&ThumbprintList.member.1=1234567890123456789012345678901234567890"),
		Action:  "UpdateOpenIDConnectProviderThumbprint",
	}

	resp, err := service.HandleRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	require.Contains(t, string(resp.Body), "NoSuchEntity")
}

func TestUpdateOpenIDConnectProviderThumbprint_InvalidThumbprint(t *testing.T) {
	state := emulator.NewMemoryStateManager()
	validator := emulator.NewSchemaValidator()
	service := NewIAMService(state, validator)

	// Create provider
	createReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=CreateOpenIDConnectProvider&Url=https://invalid-update.example.com&ThumbprintList.member.1=1234567890123456789012345678901234567890"),
		Action:  "CreateOpenIDConnectProvider",
	}
	_, _ = service.HandleRequest(context.Background(), createReq)

	// Update with invalid thumbprint
	updateReq := &emulator.AWSRequest{
		Method:  "POST",
		Headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		Body:    []byte("Action=UpdateOpenIDConnectProviderThumbprint&OpenIDConnectProviderArn=arn:aws:iam::123456789012:oidc-provider/invalid-update.example.com&ThumbprintList.member.1=invalid"),
		Action:  "UpdateOpenIDConnectProviderThumbprint",
	}
	resp, err := service.HandleRequest(context.Background(), updateReq)
	require.NoError(t, err)
	require.Equal(t, 400, resp.StatusCode)
	require.Contains(t, string(resp.Body), "Invalid thumbprint")
}

// ============================================================================
// Helper function tests
// ============================================================================

func TestIsValidSAMLProviderName(t *testing.T) {
	// Valid names
	require.True(t, isValidSAMLProviderName("MyProvider"))
	require.True(t, isValidSAMLProviderName("My_Provider"))
	require.True(t, isValidSAMLProviderName("My-Provider"))
	require.True(t, isValidSAMLProviderName("My.Provider"))
	require.True(t, isValidSAMLProviderName("a"))

	// Invalid names
	require.False(t, isValidSAMLProviderName(""))                       // too short
	require.False(t, isValidSAMLProviderName(strings.Repeat("a", 129))) // too long
	require.False(t, isValidSAMLProviderName("My Provider"))            // contains space
}

func TestIsValidThumbprint(t *testing.T) {
	// Valid thumbprints (40 hex chars)
	require.True(t, isValidThumbprint("1234567890123456789012345678901234567890"))
	require.True(t, isValidThumbprint("abcdef1234567890123456789012345678901234"))
	require.True(t, isValidThumbprint("ABCDEF1234567890123456789012345678901234"))

	// Invalid thumbprints
	require.False(t, isValidThumbprint("short"))
	require.False(t, isValidThumbprint("12345678901234567890123456789012345678901")) // 41 chars
	require.False(t, isValidThumbprint("123456789012345678901234567890123456789"))   // 39 chars
	require.False(t, isValidThumbprint("gggggggggggggggggggggggggggggggggggggggg"))  // non-hex
}
