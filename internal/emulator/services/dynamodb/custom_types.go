package dynamodb

// Custom types that CloudMirror doesn't generate from Smithy models.
// DynamoDB uses complex map/union types that require special handling.

// AttributeValue represents a DynamoDB attribute value.
// In DynamoDB, attribute values can be of various types (string, number, binary, etc.)
// This is a simplified representation using interface{} for flexibility.
type AttributeValue map[string]interface{}

// AttributeMap represents a map of attribute names to attribute values.
// This is used for items in DynamoDB (each item is a collection of attributes).
type AttributeMap map[string]AttributeValue
