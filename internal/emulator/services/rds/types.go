package rds

// Result wrapper types for operations where Smithy output shape name differs from OperationNameResult.
// For Query protocol, the XML response must use <OperationNameResult> as the element name.
// Most Result types are now generated in smithy_types.go; only keep types here that aren't generated.

// DescribeDBInstancesResult wraps the DBInstances list for DescribeDBInstances response.
// Smithy defines the output as DBInstanceMessage, but Query protocol needs DescribeDBInstancesResult.
type DescribeDBInstancesResult struct {
	DBInstances []DBInstance `xml:"DBInstances>DBInstance"`
	Marker      *string      `xml:"Marker,omitempty"`
}

// ListTagsForResourceResult wraps the TagList for ListTagsForResource response.
// Smithy defines the output as TagListMessage, but Query protocol needs ListTagsForResourceResult.
type ListTagsForResourceResult struct {
	TagList []Tag `xml:"TagList>Tag"`
}

// AddTagsToResourceResult is empty for AddTagsToResource response.
// Smithy defines no output for this operation.
type AddTagsToResourceResult struct{}
