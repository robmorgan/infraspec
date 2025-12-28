package s3

import "encoding/xml"

// ============================================================================
// XML Response Types for S3 REST-XML Protocol
// These types use XMLName and xmlns attributes for proper S3 XML responses
// Prefixed with "XML" to avoid conflict with Smithy-generated types
// ============================================================================

// ListAllMyBucketsResult represents the response for ListBuckets
type ListAllMyBucketsResult struct {
	XMLName xml.Name   `xml:"ListAllMyBucketsResult"`
	Xmlns   string     `xml:"xmlns,attr"`
	Owner   XMLOwner   `xml:"Owner"`
	Buckets XMLBuckets `xml:"Buckets"`
}

// XMLOwner represents the bucket owner in XML responses
type XMLOwner struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}

// XMLBuckets is a container for XMLBucket elements
type XMLBuckets struct {
	Bucket []XMLBucket `xml:"Bucket"`
}

// XMLBucket represents an S3 bucket in list responses
type XMLBucket struct {
	Name         string `xml:"Name"`
	CreationDate string `xml:"CreationDate"`
}

// XMLGetBucketTaggingOutput represents the response for GetBucketTagging / GetResourceTagging
type XMLGetBucketTaggingOutput struct {
	XMLName xml.Name  `xml:"Tagging"`
	Xmlns   string    `xml:"xmlns,attr"`
	TagSet  XMLTagSet `xml:"TagSet"`
}

// XMLTagSet is a container for XMLTag elements
type XMLTagSet struct {
	Tags []XMLTag `xml:"Tag"`
}

// XMLTag represents a key-value tag in XML responses
type XMLTag struct {
	Key   string `xml:"Key"`
	Value string `xml:"Value"`
}

// VersioningConfiguration represents the response for GetBucketVersioning
type VersioningConfiguration struct {
	XMLName xml.Name `xml:"VersioningConfiguration"`
	Xmlns   string   `xml:"xmlns,attr"`
	Status  string   `xml:"Status,omitempty"`
}

// ListBucketResult represents the response for ListObjectsV2
type ListBucketResult struct {
	XMLName     xml.Name    `xml:"ListBucketResult"`
	Xmlns       string      `xml:"xmlns,attr"`
	Name        string      `xml:"Name"`
	Prefix      string      `xml:"Prefix"`
	KeyCount    int         `xml:"KeyCount"`
	MaxKeys     int         `xml:"MaxKeys"`
	IsTruncated bool        `xml:"IsTruncated"`
	Contents    []XMLObject `xml:"Contents,omitempty"`
}

// XMLObject represents an S3 object in list responses
type XMLObject struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
}

// BucketLoggingStatus represents the response for GetBucketLogging
type BucketLoggingStatus struct {
	XMLName        xml.Name           `xml:"BucketLoggingStatus"`
	Xmlns          string             `xml:"xmlns,attr"`
	LoggingEnabled *XMLLoggingEnabled `xml:"LoggingEnabled,omitempty"`
}

// XMLLoggingEnabled contains the logging configuration when enabled
type XMLLoggingEnabled struct {
	TargetBucket string `xml:"TargetBucket"`
	TargetPrefix string `xml:"TargetPrefix"`
}

// XMLServerSideEncryptionConfiguration represents the response for GetBucketEncryption
// Also used as input type for PutBucketEncryption
type XMLServerSideEncryptionConfiguration struct {
	XMLName xml.Name                      `xml:"ServerSideEncryptionConfiguration"`
	Xmlns   string                        `xml:"xmlns,attr"`
	Rules   []XMLServerSideEncryptionRule `xml:"Rule"`
}

// XMLServerSideEncryptionRule represents an encryption rule
type XMLServerSideEncryptionRule struct {
	ApplyServerSideEncryptionByDefault XMLApplyServerSideEncryptionByDefault `xml:"ApplyServerSideEncryptionByDefault"`
	BucketKeyEnabled                   bool                                  `xml:"BucketKeyEnabled,omitempty"`
}

// XMLApplyServerSideEncryptionByDefault represents the default encryption settings
type XMLApplyServerSideEncryptionByDefault struct {
	SSEAlgorithm   string `xml:"SSEAlgorithm"`
	KMSMasterKeyID string `xml:"KMSMasterKeyID,omitempty"`
}

// XMLPublicAccessBlockConfiguration represents the response for GetPublicAccessBlock
type XMLPublicAccessBlockConfiguration struct {
	XMLName               xml.Name `xml:"PublicAccessBlockConfiguration"`
	Xmlns                 string   `xml:"xmlns,attr"`
	BlockPublicAcls       bool     `xml:"BlockPublicAcls"`
	BlockPublicPolicy     bool     `xml:"BlockPublicPolicy"`
	IgnorePublicAcls      bool     `xml:"IgnorePublicAcls"`
	RestrictPublicBuckets bool     `xml:"RestrictPublicBuckets"`
}
