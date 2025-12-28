package iam

import (
	"encoding/xml"
	"time"
)

// ============================================================================
// XML Response Types for IAM Query Protocol
// These types use XMLName for proper IAM XML responses
// Prefixed with "XML" where they conflict with Smithy-generated types
// ============================================================================

// XMLTag represents an IAM resource tag
type XMLTag struct {
	Key   string `xml:"Key"`
	Value string `xml:"Value"`
}

// XMLRole represents an IAM role (used in single-item responses like GetRole)
type XMLRole struct {
	XMLName                  xml.Name  `xml:"Role"`
	RoleName                 string    `xml:"RoleName"`
	RoleId                   string    `xml:"RoleId"`
	Arn                      string    `xml:"Arn"`
	Path                     string    `xml:"Path"`
	AssumeRolePolicyDocument string    `xml:"AssumeRolePolicyDocument"`
	Description              string    `xml:"Description,omitempty"`
	MaxSessionDuration       int32     `xml:"MaxSessionDuration"`
	CreateDate               time.Time `xml:"CreateDate"`
	Tags                     []XMLTag  `xml:"Tags>member,omitempty"`
}

// XMLRoleListItem represents a role in a list (no XMLName for proper member serialization)
type XMLRoleListItem struct {
	RoleName                 string    `xml:"RoleName"`
	RoleId                   string    `xml:"RoleId"`
	Arn                      string    `xml:"Arn"`
	Path                     string    `xml:"Path"`
	AssumeRolePolicyDocument string    `xml:"AssumeRolePolicyDocument"`
	Description              string    `xml:"Description,omitempty"`
	MaxSessionDuration       int32     `xml:"MaxSessionDuration"`
	CreateDate               time.Time `xml:"CreateDate"`
	Tags                     []XMLTag  `xml:"Tags>member,omitempty"`
}

// XMLPolicy represents an IAM managed policy
type XMLPolicy struct {
	XMLName          xml.Name  `xml:"Policy"`
	PolicyName       string    `xml:"PolicyName"`
	PolicyId         string    `xml:"PolicyId"`
	Arn              string    `xml:"Arn"`
	Path             string    `xml:"Path"`
	Description      string    `xml:"Description,omitempty"`
	DefaultVersionId string    `xml:"DefaultVersionId"`
	CreateDate       time.Time `xml:"CreateDate"`
	UpdateDate       time.Time `xml:"UpdateDate"`
	AttachmentCount  int32     `xml:"AttachmentCount"`
	IsAttachable     bool      `xml:"IsAttachable"`
	Tags             []XMLTag  `xml:"Tags>member,omitempty"`
}

// XMLPolicyVersion represents a version of an IAM policy
type XMLPolicyVersion struct {
	XMLName          xml.Name  `xml:"PolicyVersion"`
	VersionId        string    `xml:"VersionId"`
	Document         string    `xml:"Document"`
	IsDefaultVersion bool      `xml:"IsDefaultVersion"`
	CreateDate       time.Time `xml:"CreateDate"`
}

// XMLInstanceProfile represents an IAM instance profile (used in single-item responses)
type XMLInstanceProfile struct {
	XMLName             xml.Name          `xml:"InstanceProfile"`
	InstanceProfileName string            `xml:"InstanceProfileName"`
	InstanceProfileId   string            `xml:"InstanceProfileId"`
	Arn                 string            `xml:"Arn"`
	Path                string            `xml:"Path"`
	CreateDate          time.Time         `xml:"CreateDate"`
	Roles               []XMLRoleListItem `xml:"Roles>member,omitempty"`
	Tags                []XMLTag          `xml:"Tags>member,omitempty"`
}

// XMLInstanceProfileListItem represents an instance profile in a list (no XMLName)
type XMLInstanceProfileListItem struct {
	InstanceProfileName string            `xml:"InstanceProfileName"`
	InstanceProfileId   string            `xml:"InstanceProfileId"`
	Arn                 string            `xml:"Arn"`
	Path                string            `xml:"Path"`
	CreateDate          time.Time         `xml:"CreateDate"`
	Roles               []XMLRoleListItem `xml:"Roles>member,omitempty"`
	Tags                []XMLTag          `xml:"Tags>member,omitempty"`
}

// XMLAttachedPolicy represents a policy attached to a role or user
type XMLAttachedPolicy struct {
	PolicyName string `xml:"PolicyName"`
	PolicyArn  string `xml:"PolicyArn"`
}

// XMLUser represents an IAM user (used in single-item responses like GetUser)
type XMLUser struct {
	XMLName          xml.Name  `xml:"User"`
	UserName         string    `xml:"UserName"`
	UserId           string    `xml:"UserId"`
	Arn              string    `xml:"Arn"`
	Path             string    `xml:"Path"`
	CreateDate       time.Time `xml:"CreateDate"`
	PasswordLastUsed time.Time `xml:"PasswordLastUsed,omitempty"`
	Tags             []XMLTag  `xml:"Tags>member,omitempty"`
}

// XMLUserListItem represents a user in a list (no XMLName for proper member serialization)
type XMLUserListItem struct {
	UserName         string    `xml:"UserName"`
	UserId           string    `xml:"UserId"`
	Arn              string    `xml:"Arn"`
	Path             string    `xml:"Path"`
	CreateDate       time.Time `xml:"CreateDate"`
	PasswordLastUsed time.Time `xml:"PasswordLastUsed,omitempty"`
	Tags             []XMLTag  `xml:"Tags>member,omitempty"`
}

// XMLLoginProfile represents a user's login profile for console access
type XMLLoginProfile struct {
	XMLName               xml.Name  `xml:"LoginProfile"`
	UserName              string    `xml:"UserName"`
	CreateDate            time.Time `xml:"CreateDate"`
	PasswordResetRequired bool      `xml:"PasswordResetRequired"`
}

// ============================================================================
// Response wrapper types for XML marshaling
// These types marshal to <ActionResult> elements which BuildQueryResponse wraps
// with <ActionResponse> and <ResponseMetadata>
// ============================================================================

// GetRoleResult wraps the Role for GetRole response
type GetRoleResult struct {
	XMLName xml.Name `xml:"GetRoleResult"`
	Role    XMLRole  `xml:"Role"`
}

// CreateRoleResult wraps the Role for CreateRole response
type CreateRoleResult struct {
	XMLName xml.Name `xml:"CreateRoleResult"`
	Role    XMLRole  `xml:"Role"`
}

// ListRolesResult wraps the roles list for ListRoles response
type ListRolesResult struct {
	XMLName     xml.Name          `xml:"ListRolesResult"`
	Roles       []XMLRoleListItem `xml:"Roles>member"`
	IsTruncated bool              `xml:"IsTruncated"`
	Marker      string            `xml:"Marker,omitempty"`
}

// UpdateRoleResult wraps the response for UpdateRole
type UpdateRoleResult struct {
	XMLName xml.Name `xml:"UpdateRoleResult"`
}

// ListRoleTagsResult wraps the tags list for ListRoleTags response
type ListRoleTagsResult struct {
	XMLName     xml.Name `xml:"ListRoleTagsResult"`
	Tags        []XMLTag `xml:"Tags>member"`
	IsTruncated bool     `xml:"IsTruncated"`
	Marker      string   `xml:"Marker,omitempty"`
}

// CreateServiceLinkedRoleResult wraps the Role for CreateServiceLinkedRole response
type CreateServiceLinkedRoleResult struct {
	XMLName xml.Name `xml:"CreateServiceLinkedRoleResult"`
	Role    XMLRole  `xml:"Role"`
}

// DeleteServiceLinkedRoleResult wraps the deletion task ID
type DeleteServiceLinkedRoleResult struct {
	XMLName        xml.Name `xml:"DeleteServiceLinkedRoleResult"`
	DeletionTaskId string   `xml:"DeletionTaskId"`
}

// GetServiceLinkedRoleDeletionStatusResult wraps the deletion status
type GetServiceLinkedRoleDeletionStatusResult struct {
	XMLName xml.Name                    `xml:"GetServiceLinkedRoleDeletionStatusResult"`
	Status  string                      `xml:"Status"`
	Reason  *RoleDeletionFailureReason  `xml:"Reason,omitempty"`
}

// RoleDeletionFailureReason contains details about why deletion failed
type RoleDeletionFailureReason struct {
	Reason            string                    `xml:"Reason,omitempty"`
	RoleUsageList     []RoleUsageInfo           `xml:"RoleUsageList>member,omitempty"`
}

// RoleUsageInfo contains info about resources using the role
type RoleUsageInfo struct {
	Region    string   `xml:"Region,omitempty"`
	Resources []string `xml:"Resources>member,omitempty"`
}

// GetPolicyResult wraps the Policy for GetPolicy response
type GetPolicyResult struct {
	XMLName xml.Name  `xml:"GetPolicyResult"`
	Policy  XMLPolicy `xml:"Policy"`
}

// CreatePolicyResult wraps the Policy for CreatePolicy response
type CreatePolicyResult struct {
	XMLName xml.Name  `xml:"CreatePolicyResult"`
	Policy  XMLPolicy `xml:"Policy"`
}

// ListPoliciesResult wraps the policies list for ListPolicies response
type ListPoliciesResult struct {
	XMLName     xml.Name    `xml:"ListPoliciesResult"`
	Policies    []XMLPolicy `xml:"Policies>member"`
	IsTruncated bool        `xml:"IsTruncated"`
	Marker      string      `xml:"Marker,omitempty"`
}

// GetPolicyVersionResult wraps the PolicyVersion for GetPolicyVersion response
type GetPolicyVersionResult struct {
	XMLName       xml.Name         `xml:"GetPolicyVersionResult"`
	PolicyVersion XMLPolicyVersion `xml:"PolicyVersion"`
}

// ListPolicyVersionsResult wraps the versions list for ListPolicyVersions response
type ListPolicyVersionsResult struct {
	XMLName     xml.Name           `xml:"ListPolicyVersionsResult"`
	Versions    []XMLPolicyVersion `xml:"Versions>member"`
	IsTruncated bool               `xml:"IsTruncated"`
	Marker      string             `xml:"Marker,omitempty"`
}

// GetInstanceProfileResult wraps the InstanceProfile for GetInstanceProfile response
type GetInstanceProfileResult struct {
	XMLName         xml.Name           `xml:"GetInstanceProfileResult"`
	InstanceProfile XMLInstanceProfile `xml:"InstanceProfile"`
}

// CreateInstanceProfileResult wraps the InstanceProfile for CreateInstanceProfile response
type CreateInstanceProfileResult struct {
	XMLName         xml.Name           `xml:"CreateInstanceProfileResult"`
	InstanceProfile XMLInstanceProfile `xml:"InstanceProfile"`
}

// ListInstanceProfilesResult wraps the profiles list for ListInstanceProfiles response
type ListInstanceProfilesResult struct {
	XMLName          xml.Name                     `xml:"ListInstanceProfilesResult"`
	InstanceProfiles []XMLInstanceProfileListItem `xml:"InstanceProfiles>member"`
	IsTruncated      bool                         `xml:"IsTruncated"`
	Marker           string                       `xml:"Marker,omitempty"`
}

// ListInstanceProfilesForRoleResult wraps the profiles list for ListInstanceProfilesForRole response
type ListInstanceProfilesForRoleResult struct {
	XMLName          xml.Name                     `xml:"ListInstanceProfilesForRoleResult"`
	InstanceProfiles []XMLInstanceProfileListItem `xml:"InstanceProfiles>member"`
	IsTruncated      bool                         `xml:"IsTruncated"`
	Marker           string                       `xml:"Marker,omitempty"`
}

// ListAttachedRolePoliciesResult wraps the attached policies list
type ListAttachedRolePoliciesResult struct {
	XMLName          xml.Name            `xml:"ListAttachedRolePoliciesResult"`
	AttachedPolicies []XMLAttachedPolicy `xml:"AttachedPolicies>member"`
	IsTruncated      bool                `xml:"IsTruncated"`
	Marker           string              `xml:"Marker,omitempty"`
}

// EmptyResult for operations that return no data (Delete, Attach, Detach, etc.)
type EmptyResult struct {
	XMLName xml.Name `xml:""`
}

// ListRolePoliciesResult wraps the inline policy names list
type ListRolePoliciesResult struct {
	XMLName     xml.Name `xml:"ListRolePoliciesResult"`
	PolicyNames []string `xml:"PolicyNames>member"`
	IsTruncated bool     `xml:"IsTruncated"`
	Marker      string   `xml:"Marker,omitempty"`
}

// GetRolePolicyResult wraps the inline policy for GetRolePolicy response
type GetRolePolicyResult struct {
	XMLName        xml.Name `xml:"GetRolePolicyResult"`
	RoleName       string   `xml:"RoleName"`
	PolicyName     string   `xml:"PolicyName"`
	PolicyDocument string   `xml:"PolicyDocument"`
}

// ============================================================================
// User Response Types
// ============================================================================

// CreateUserResult wraps the User for CreateUser response
type CreateUserResult struct {
	XMLName xml.Name `xml:"CreateUserResult"`
	User    XMLUser  `xml:"User"`
}

// GetUserResult wraps the User for GetUser response
type GetUserResult struct {
	XMLName xml.Name `xml:"GetUserResult"`
	User    XMLUser  `xml:"User"`
}

// ListUsersResult wraps the users list for ListUsers response
type ListUsersResult struct {
	XMLName     xml.Name          `xml:"ListUsersResult"`
	Users       []XMLUserListItem `xml:"Users>member"`
	IsTruncated bool              `xml:"IsTruncated"`
	Marker      string            `xml:"Marker,omitempty"`
}

// CreateLoginProfileResult wraps the LoginProfile for CreateLoginProfile response
type CreateLoginProfileResult struct {
	XMLName      xml.Name        `xml:"CreateLoginProfileResult"`
	LoginProfile XMLLoginProfile `xml:"LoginProfile"`
}

// GetLoginProfileResult wraps the LoginProfile for GetLoginProfile response
type GetLoginProfileResult struct {
	XMLName      xml.Name        `xml:"GetLoginProfileResult"`
	LoginProfile XMLLoginProfile `xml:"LoginProfile"`
}

// ListAttachedUserPoliciesResult wraps the attached policies list for user
type ListAttachedUserPoliciesResult struct {
	XMLName          xml.Name            `xml:"ListAttachedUserPoliciesResult"`
	AttachedPolicies []XMLAttachedPolicy `xml:"AttachedPolicies>member"`
	IsTruncated      bool                `xml:"IsTruncated"`
	Marker           string              `xml:"Marker,omitempty"`
}

// ListUserPoliciesResult wraps the inline policy names list for user
type ListUserPoliciesResult struct {
	XMLName     xml.Name `xml:"ListUserPoliciesResult"`
	PolicyNames []string `xml:"PolicyNames>member"`
	IsTruncated bool     `xml:"IsTruncated"`
	Marker      string   `xml:"Marker,omitempty"`
}

// GetUserPolicyResult wraps the inline policy for GetUserPolicy response
type GetUserPolicyResult struct {
	XMLName        xml.Name `xml:"GetUserPolicyResult"`
	UserName       string   `xml:"UserName"`
	PolicyName     string   `xml:"PolicyName"`
	PolicyDocument string   `xml:"PolicyDocument"`
}

// ListUserTagsResult wraps the tags list for ListUserTags response
type ListUserTagsResult struct {
	XMLName     xml.Name `xml:"ListUserTagsResult"`
	Tags        []XMLTag `xml:"Tags>member"`
	IsTruncated bool     `xml:"IsTruncated"`
	Marker      string   `xml:"Marker,omitempty"`
}

// ============================================================================
// Group XML Types
// ============================================================================

// XMLGroup represents an IAM group (used in single-item responses like GetGroup)
type XMLGroup struct {
	XMLName    xml.Name  `xml:"Group"`
	GroupName  string    `xml:"GroupName"`
	GroupId    string    `xml:"GroupId"`
	Arn        string    `xml:"Arn"`
	Path       string    `xml:"Path"`
	CreateDate time.Time `xml:"CreateDate"`
}

// XMLGroupListItem represents a group in a list (no XMLName for proper member serialization)
type XMLGroupListItem struct {
	GroupName  string    `xml:"GroupName"`
	GroupId    string    `xml:"GroupId"`
	Arn        string    `xml:"Arn"`
	Path       string    `xml:"Path"`
	CreateDate time.Time `xml:"CreateDate"`
}

// ============================================================================
// Group Response Types
// ============================================================================

// CreateGroupResult wraps the Group for CreateGroup response
type CreateGroupResult struct {
	XMLName xml.Name `xml:"CreateGroupResult"`
	Group   XMLGroup `xml:"Group"`
}

// GetGroupResult wraps the Group and Users for GetGroup response
type GetGroupResult struct {
	XMLName     xml.Name          `xml:"GetGroupResult"`
	Group       XMLGroup          `xml:"Group"`
	Users       []XMLUserListItem `xml:"Users>member,omitempty"`
	IsTruncated bool              `xml:"IsTruncated"`
	Marker      string            `xml:"Marker,omitempty"`
}

// ListGroupsResult wraps the groups list for ListGroups response
type ListGroupsResult struct {
	XMLName     xml.Name           `xml:"ListGroupsResult"`
	Groups      []XMLGroupListItem `xml:"Groups>member"`
	IsTruncated bool               `xml:"IsTruncated"`
	Marker      string             `xml:"Marker,omitempty"`
}

// ListGroupsForUserResult wraps the groups list for ListGroupsForUser response
type ListGroupsForUserResult struct {
	XMLName     xml.Name           `xml:"ListGroupsForUserResult"`
	Groups      []XMLGroupListItem `xml:"Groups>member"`
	IsTruncated bool               `xml:"IsTruncated"`
	Marker      string             `xml:"Marker,omitempty"`
}

// ListAttachedGroupPoliciesResult wraps the attached policies list for group
type ListAttachedGroupPoliciesResult struct {
	XMLName          xml.Name            `xml:"ListAttachedGroupPoliciesResult"`
	AttachedPolicies []XMLAttachedPolicy `xml:"AttachedPolicies>member"`
	IsTruncated      bool                `xml:"IsTruncated"`
	Marker           string              `xml:"Marker,omitempty"`
}

// ListGroupPoliciesResult wraps the inline policy names list for group
type ListGroupPoliciesResult struct {
	XMLName     xml.Name `xml:"ListGroupPoliciesResult"`
	PolicyNames []string `xml:"PolicyNames>member"`
	IsTruncated bool     `xml:"IsTruncated"`
	Marker      string   `xml:"Marker,omitempty"`
}

// GetGroupPolicyResult wraps the inline policy for GetGroupPolicy response
type GetGroupPolicyResult struct {
	XMLName        xml.Name `xml:"GetGroupPolicyResult"`
	GroupName      string   `xml:"GroupName"`
	PolicyName     string   `xml:"PolicyName"`
	PolicyDocument string   `xml:"PolicyDocument"`
}

// ============================================================================
// Access Key XML Types
// ============================================================================

// XMLAccessKey represents an IAM access key (includes secret, used in CreateAccessKey response)
type XMLAccessKey struct {
	XMLName         xml.Name  `xml:"AccessKey"`
	UserName        string    `xml:"UserName"`
	AccessKeyId     string    `xml:"AccessKeyId"`
	Status          string    `xml:"Status"`
	SecretAccessKey string    `xml:"SecretAccessKey"`
	CreateDate      time.Time `xml:"CreateDate"`
}

// XMLAccessKeyMetadata represents access key metadata (no secret, used in list responses)
type XMLAccessKeyMetadata struct {
	UserName    string    `xml:"UserName"`
	AccessKeyId string    `xml:"AccessKeyId"`
	Status      string    `xml:"Status"`
	CreateDate  time.Time `xml:"CreateDate"`
}

// XMLAccessKeyLastUsed represents when an access key was last used
type XMLAccessKeyLastUsed struct {
	LastUsedDate time.Time `xml:"LastUsedDate,omitempty"`
	ServiceName  string    `xml:"ServiceName,omitempty"`
	Region       string    `xml:"Region,omitempty"`
}

// ============================================================================
// Access Key Response Types
// ============================================================================

// CreateAccessKeyResult wraps the AccessKey for CreateAccessKey response
type CreateAccessKeyResult struct {
	XMLName   xml.Name     `xml:"CreateAccessKeyResult"`
	AccessKey XMLAccessKey `xml:"AccessKey"`
}

// ListAccessKeysResult wraps the access keys list for ListAccessKeys response
type ListAccessKeysResult struct {
	XMLName             xml.Name               `xml:"ListAccessKeysResult"`
	AccessKeyMetadata   []XMLAccessKeyMetadata `xml:"AccessKeyMetadata>member"`
	IsTruncated         bool                   `xml:"IsTruncated"`
	Marker              string                 `xml:"Marker,omitempty"`
}

// GetAccessKeyLastUsedResult wraps the last used info for GetAccessKeyLastUsed response
type GetAccessKeyLastUsedResult struct {
	XMLName           xml.Name             `xml:"GetAccessKeyLastUsedResult"`
	UserName          string               `xml:"UserName"`
	AccessKeyLastUsed XMLAccessKeyLastUsed `xml:"AccessKeyLastUsed"`
}

// ============================================================================
// Internal storage types (used for state management, not XML)
// ============================================================================

// RoleAttachments tracks which policies are attached to a role
type RoleAttachments struct {
	PolicyArns []string
}

// RoleInlinePolicies tracks inline policies for a role
type RoleInlinePolicies struct {
	Policies map[string]string // PolicyName -> PolicyDocument
}

// ProfileRoles tracks which roles are in an instance profile
type ProfileRoles struct {
	RoleNames []string
}

// UserAttachments tracks which policies are attached to a user
type UserAttachments struct {
	PolicyArns []string
}

// UserInlinePolicies tracks inline policies for a user
type UserInlinePolicies struct {
	Policies map[string]string // PolicyName -> PolicyDocument
}

// UserLoginProfile stores login profile data for a user
type UserLoginProfile struct {
	PasswordHash          string    // In real AWS, this would be the hashed password
	CreateDate            time.Time
	PasswordResetRequired bool
}

// GroupAttachments tracks which policies are attached to a group
type GroupAttachments struct {
	PolicyArns []string
}

// GroupInlinePolicies tracks inline policies for a group
type GroupInlinePolicies struct {
	Policies map[string]string // PolicyName -> PolicyDocument
}

// GroupMembers tracks which users are members of a group
type GroupMembers struct {
	UserNames []string
}

// UserGroups tracks which groups a user belongs to
type UserGroups struct {
	GroupNames []string
}

// AccessKeyData stores the full access key data including secret
type AccessKeyData struct {
	UserName        string
	AccessKeyId     string
	SecretAccessKey string
	Status          string // "Active" or "Inactive"
	CreateDate      time.Time
	LastUsedDate    time.Time
	LastUsedService string
	LastUsedRegion  string
}

// ServiceLinkedRoleDeletionTask tracks the status of a service-linked role deletion
type ServiceLinkedRoleDeletionTask struct {
	TaskId     string
	RoleName   string
	Status     string // "SUCCEEDED", "IN_PROGRESS", "FAILED", "NOT_STARTED"
	CreateDate time.Time
	Reason     string
}

// ============================================================================
// SAML Provider XML Types
// ============================================================================

// XMLSAMLProvider represents an IAM SAML provider
type XMLSAMLProvider struct {
	XMLName            xml.Name  `xml:"SAMLProviderArn"`
	Arn                string    `xml:"Arn"`
	ValidUntil         time.Time `xml:"ValidUntil,omitempty"`
	CreateDate         time.Time `xml:"CreateDate"`
	Tags               []XMLTag  `xml:"Tags>member,omitempty"`
	SAMLMetadataDocument string  `xml:"SAMLMetadataDocument,omitempty"`
}

// Note: SAMLProviderListEntry is defined in smithy_types.go

// ============================================================================
// SAML Provider Response Types
// ============================================================================

// CreateSAMLProviderResult wraps the ARN for CreateSAMLProvider response
type CreateSAMLProviderResult struct {
	XMLName         xml.Name `xml:"CreateSAMLProviderResult"`
	SAMLProviderArn string   `xml:"SAMLProviderArn"`
	Tags            []XMLTag `xml:"Tags>member,omitempty"`
}

// GetSAMLProviderResult wraps the SAML provider details for GetSAMLProvider response
type GetSAMLProviderResult struct {
	XMLName              xml.Name  `xml:"GetSAMLProviderResult"`
	CreateDate           time.Time `xml:"CreateDate"`
	ValidUntil           time.Time `xml:"ValidUntil,omitempty"`
	SAMLMetadataDocument string    `xml:"SAMLMetadataDocument"`
	Tags                 []XMLTag  `xml:"Tags>member,omitempty"`
}

// UpdateSAMLProviderResult wraps the ARN for UpdateSAMLProvider response
type UpdateSAMLProviderResult struct {
	XMLName         xml.Name `xml:"UpdateSAMLProviderResult"`
	SAMLProviderArn string   `xml:"SAMLProviderArn"`
}

// ListSAMLProvidersResult wraps the SAML providers list
type ListSAMLProvidersResult struct {
	XMLName           xml.Name                `xml:"ListSAMLProvidersResult"`
	SAMLProviderList  []SAMLProviderListEntry `xml:"SAMLProviderList>member"`
}

// ============================================================================
// OIDC Provider XML Types
// ============================================================================

// XMLOpenIDConnectProvider represents an IAM OIDC provider
type XMLOpenIDConnectProvider struct {
	XMLName        xml.Name  `xml:"OpenIDConnectProvider"`
	Arn            string    `xml:"Arn"`
	Url            string    `xml:"Url"`
	CreateDate     time.Time `xml:"CreateDate"`
	ThumbprintList []string  `xml:"ThumbprintList>member,omitempty"`
	ClientIDList   []string  `xml:"ClientIDList>member,omitempty"`
	Tags           []XMLTag  `xml:"Tags>member,omitempty"`
}

// Note: OpenIDConnectProviderListEntry is defined in smithy_types.go

// ============================================================================
// OIDC Provider Response Types
// ============================================================================

// CreateOpenIDConnectProviderResult wraps the ARN for CreateOpenIDConnectProvider response
type CreateOpenIDConnectProviderResult struct {
	XMLName                   xml.Name `xml:"CreateOpenIDConnectProviderResult"`
	OpenIDConnectProviderArn  string   `xml:"OpenIDConnectProviderArn"`
	Tags                      []XMLTag `xml:"Tags>member,omitempty"`
}

// GetOpenIDConnectProviderResult wraps the OIDC provider details
type GetOpenIDConnectProviderResult struct {
	XMLName        xml.Name  `xml:"GetOpenIDConnectProviderResult"`
	Url            string    `xml:"Url"`
	CreateDate     time.Time `xml:"CreateDate"`
	ThumbprintList []string  `xml:"ThumbprintList>member"`
	ClientIDList   []string  `xml:"ClientIDList>member,omitempty"`
	Tags           []XMLTag  `xml:"Tags>member,omitempty"`
}

// ListOpenIDConnectProvidersResult wraps the OIDC providers list
type ListOpenIDConnectProvidersResult struct {
	XMLName                     xml.Name                         `xml:"ListOpenIDConnectProvidersResult"`
	OpenIDConnectProviderList   []OpenIDConnectProviderListEntry `xml:"OpenIDConnectProviderList>member"`
}

// ============================================================================
// Identity Provider Storage Types
// ============================================================================

// SAMLProviderData stores the full SAML provider data
type SAMLProviderData struct {
	Name                 string // Extracted from ARN
	Arn                  string
	SAMLMetadataDocument string
	CreateDate           time.Time
	ValidUntil           time.Time
	Tags                 []XMLTag
}

// OIDCProviderData stores the full OIDC provider data
type OIDCProviderData struct {
	Url            string
	Arn            string
	CreateDate     time.Time
	ThumbprintList []string
	ClientIDList   []string
	Tags           []XMLTag
}

// ============================================================================
// MFA Device XML Types
// ============================================================================

// XMLVirtualMFADevice represents a virtual MFA device
type XMLVirtualMFADevice struct {
	XMLName                 xml.Name  `xml:"VirtualMFADevice"`
	SerialNumber            string    `xml:"SerialNumber"`
	Base32StringSeed        string    `xml:"Base32StringSeed,omitempty"` // Only shown on create
	QRCodePNG               string    `xml:"QRCodePNG,omitempty"`        // Base64 encoded PNG, only shown on create
	User                    *XMLUser  `xml:"User,omitempty"`             // Associated user, if any
	EnableDate              time.Time `xml:"EnableDate,omitempty"`
	Tags                    []XMLTag  `xml:"Tags>member,omitempty"`
}

// XMLMFADevice represents an MFA device (hardware or virtual)
type XMLMFADevice struct {
	XMLName      xml.Name  `xml:"MFADevice"`
	SerialNumber string    `xml:"SerialNumber"`
	UserName     string    `xml:"UserName"`
	EnableDate   time.Time `xml:"EnableDate"`
}

// MFADeviceListItem represents an MFA device in a list (no XMLName)
type MFADeviceListItem struct {
	SerialNumber string    `xml:"SerialNumber"`
	UserName     string    `xml:"UserName"`
	EnableDate   time.Time `xml:"EnableDate"`
}

// VirtualMFADeviceListItem represents a virtual MFA device in a list
type VirtualMFADeviceListItem struct {
	SerialNumber string    `xml:"SerialNumber"`
	User         *XMLUser  `xml:"User,omitempty"`
	EnableDate   time.Time `xml:"EnableDate,omitempty"`
	Tags         []XMLTag  `xml:"Tags>member,omitempty"`
}

// ============================================================================
// MFA Device Response Types
// ============================================================================

// CreateVirtualMFADeviceResult wraps the virtual MFA device for CreateVirtualMFADevice response
type CreateVirtualMFADeviceResult struct {
	XMLName          xml.Name            `xml:"CreateVirtualMFADeviceResult"`
	VirtualMFADevice XMLVirtualMFADevice `xml:"VirtualMFADevice"`
}

// ListVirtualMFADevicesResult wraps the virtual MFA devices list
type ListVirtualMFADevicesResult struct {
	XMLName           xml.Name                   `xml:"ListVirtualMFADevicesResult"`
	VirtualMFADevices []VirtualMFADeviceListItem `xml:"VirtualMFADevices>member"`
	IsTruncated       bool                       `xml:"IsTruncated"`
	Marker            string                     `xml:"Marker,omitempty"`
}

// ListMFADevicesResult wraps the MFA devices list
type ListMFADevicesResult struct {
	XMLName     xml.Name            `xml:"ListMFADevicesResult"`
	MFADevices  []MFADeviceListItem `xml:"MFADevices>member"`
	IsTruncated bool                `xml:"IsTruncated"`
	Marker      string              `xml:"Marker,omitempty"`
}

// ============================================================================
// MFA Device Storage Types
// ============================================================================

// VirtualMFADeviceData stores the full virtual MFA device data
type VirtualMFADeviceData struct {
	SerialNumber     string
	Base32StringSeed string
	QRCodePNG        string    // Base64 encoded PNG
	UserName         string    // Associated user, if any
	EnableDate       time.Time // When it was enabled for a user
	Tags             []XMLTag
	CreateDate       time.Time
}

// ============================================================================
// Server Certificate XML Types
// ============================================================================

// XMLServerCertificate represents an IAM server certificate
type XMLServerCertificate struct {
	XMLName                   xml.Name                      `xml:"ServerCertificate"`
	ServerCertificateMetadata XMLServerCertificateMetadata  `xml:"ServerCertificateMetadata"`
	CertificateBody           string                        `xml:"CertificateBody"`
	CertificateChain          string                        `xml:"CertificateChain,omitempty"`
	Tags                      []XMLTag                      `xml:"Tags>member,omitempty"`
}

// XMLServerCertificateMetadata represents server certificate metadata
type XMLServerCertificateMetadata struct {
	XMLName                 xml.Name  `xml:"ServerCertificateMetadata"`
	ServerCertificateName   string    `xml:"ServerCertificateName"`
	ServerCertificateId     string    `xml:"ServerCertificateId"`
	Arn                     string    `xml:"Arn"`
	Path                    string    `xml:"Path"`
	UploadDate              time.Time `xml:"UploadDate"`
	Expiration              time.Time `xml:"Expiration,omitempty"`
}

// ServerCertificateMetadataListItem represents a certificate in a list (no XMLName)
type ServerCertificateMetadataListItem struct {
	ServerCertificateName   string    `xml:"ServerCertificateName"`
	ServerCertificateId     string    `xml:"ServerCertificateId"`
	Arn                     string    `xml:"Arn"`
	Path                    string    `xml:"Path"`
	UploadDate              time.Time `xml:"UploadDate"`
	Expiration              time.Time `xml:"Expiration,omitempty"`
}

// ============================================================================
// Server Certificate Response Types
// ============================================================================

// UploadServerCertificateResult wraps the metadata for UploadServerCertificate response
type UploadServerCertificateResult struct {
	XMLName                   xml.Name                     `xml:"UploadServerCertificateResult"`
	ServerCertificateMetadata XMLServerCertificateMetadata `xml:"ServerCertificateMetadata"`
	Tags                      []XMLTag                     `xml:"Tags>member,omitempty"`
}

// GetServerCertificateResult wraps the certificate for GetServerCertificate response
type GetServerCertificateResult struct {
	XMLName           xml.Name             `xml:"GetServerCertificateResult"`
	ServerCertificate XMLServerCertificate `xml:"ServerCertificate"`
}

// ListServerCertificatesResult wraps the certificates list
type ListServerCertificatesResult struct {
	XMLName                        xml.Name                            `xml:"ListServerCertificatesResult"`
	ServerCertificateMetadataList  []ServerCertificateMetadataListItem `xml:"ServerCertificateMetadataList>member"`
	IsTruncated                    bool                                `xml:"IsTruncated"`
	Marker                         string                              `xml:"Marker,omitempty"`
}

// ============================================================================
// Server Certificate Storage Types
// ============================================================================

// ServerCertificateData stores the full server certificate data
type ServerCertificateData struct {
	ServerCertificateName string
	ServerCertificateId   string
	Arn                   string
	Path                  string
	CertificateBody       string
	CertificateChain      string
	PrivateKey            string // Not returned in responses, stored for emulation
	UploadDate            time.Time
	Expiration            time.Time
	Tags                  []XMLTag
}

// ============================================================================
// SSH Public Key XML Types
// ============================================================================

// XMLSSHPublicKey represents an IAM SSH public key
type XMLSSHPublicKey struct {
	XMLName        xml.Name  `xml:"SSHPublicKey"`
	UserName       string    `xml:"UserName"`
	SSHPublicKeyId string    `xml:"SSHPublicKeyId"`
	Fingerprint    string    `xml:"Fingerprint"`
	SSHPublicKeyBody string  `xml:"SSHPublicKeyBody"`
	Status         string    `xml:"Status"`
	UploadDate     time.Time `xml:"UploadDate"`
}

// SSHPublicKeyMetadataListItem represents SSH key metadata in a list
type SSHPublicKeyMetadataListItem struct {
	UserName       string    `xml:"UserName"`
	SSHPublicKeyId string    `xml:"SSHPublicKeyId"`
	Status         string    `xml:"Status"`
	UploadDate     time.Time `xml:"UploadDate"`
}

// ============================================================================
// SSH Public Key Response Types
// ============================================================================

// UploadSSHPublicKeyResult wraps the SSH key for UploadSSHPublicKey response
type UploadSSHPublicKeyResult struct {
	XMLName      xml.Name        `xml:"UploadSSHPublicKeyResult"`
	SSHPublicKey XMLSSHPublicKey `xml:"SSHPublicKey"`
}

// GetSSHPublicKeyResult wraps the SSH key for GetSSHPublicKey response
type GetSSHPublicKeyResult struct {
	XMLName      xml.Name        `xml:"GetSSHPublicKeyResult"`
	SSHPublicKey XMLSSHPublicKey `xml:"SSHPublicKey"`
}

// ListSSHPublicKeysResult wraps the SSH keys list
type ListSSHPublicKeysResult struct {
	XMLName       xml.Name                       `xml:"ListSSHPublicKeysResult"`
	SSHPublicKeys []SSHPublicKeyMetadataListItem `xml:"SSHPublicKeys>member"`
	IsTruncated   bool                           `xml:"IsTruncated"`
	Marker        string                         `xml:"Marker,omitempty"`
}

// ============================================================================
// SSH Public Key Storage Types
// ============================================================================

// SSHPublicKeyData stores the full SSH public key data
type SSHPublicKeyData struct {
	UserName         string
	SSHPublicKeyId   string
	Fingerprint      string
	SSHPublicKeyBody string
	Status           string // Active or Inactive
	UploadDate       time.Time
}

// ============================================================================
// Account Alias Response Types
// ============================================================================

// ListAccountAliasesResult wraps the account aliases list
type ListAccountAliasesResult struct {
	XMLName        xml.Name `xml:"ListAccountAliasesResult"`
	AccountAliases []string `xml:"AccountAliases>member"`
	IsTruncated    bool     `xml:"IsTruncated"`
	Marker         string   `xml:"Marker,omitempty"`
}

// ============================================================================
// Password Policy XML Types
// ============================================================================

// XMLPasswordPolicy represents an IAM password policy
type XMLPasswordPolicy struct {
	XMLName                      xml.Name `xml:"PasswordPolicy"`
	MinimumPasswordLength        int      `xml:"MinimumPasswordLength,omitempty"`
	RequireSymbols               bool     `xml:"RequireSymbols"`
	RequireNumbers               bool     `xml:"RequireNumbers"`
	RequireUppercaseCharacters   bool     `xml:"RequireUppercaseCharacters"`
	RequireLowercaseCharacters   bool     `xml:"RequireLowercaseCharacters"`
	AllowUsersToChangePassword   bool     `xml:"AllowUsersToChangePassword"`
	ExpirePasswords              bool     `xml:"ExpirePasswords"`
	MaxPasswordAge               int      `xml:"MaxPasswordAge,omitempty"`
	PasswordReusePrevention      int      `xml:"PasswordReusePrevention,omitempty"`
	HardExpiry                   bool     `xml:"HardExpiry"`
}

// GetAccountPasswordPolicyResult wraps the password policy
type GetAccountPasswordPolicyResult struct {
	XMLName        xml.Name          `xml:"GetAccountPasswordPolicyResult"`
	PasswordPolicy XMLPasswordPolicy `xml:"PasswordPolicy"`
}

// ============================================================================
// Password Policy Storage Types
// ============================================================================

// PasswordPolicyData stores the password policy settings
type PasswordPolicyData struct {
	MinimumPasswordLength        int
	RequireSymbols               bool
	RequireNumbers               bool
	RequireUppercaseCharacters   bool
	RequireLowercaseCharacters   bool
	AllowUsersToChangePassword   bool
	ExpirePasswords              bool
	MaxPasswordAge               int
	PasswordReusePrevention      int
	HardExpiry                   bool
}
