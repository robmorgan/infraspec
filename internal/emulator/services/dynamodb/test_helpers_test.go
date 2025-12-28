package dynamodb

// Shared test helper functions

// strPtr returns a pointer to the string value
func strPtr(s string) *string {
	return &s
}

// boolPtr returns a pointer to the bool value
func boolPtr(b bool) *bool {
	return &b
}

// int64Ptr returns a pointer to the int64 value
func int64Ptr(i int64) *int64 {
	return &i
}
