package helpers

import "time"

// StringPtr returns a pointer to the given string value.
func StringPtr(v string) *string { return &v }

// BoolPtr returns a pointer to the given bool value.
func BoolPtr(v bool) *bool { return &v }

// Int32Ptr returns a pointer to the given int32 value.
func Int32Ptr(v int32) *int32 { return &v }

// Int64Ptr returns a pointer to the given int64 value.
func Int64Ptr(v int64) *int64 { return &v }

// TimePtr returns a pointer to the given time.Time value.
func TimePtr(t time.Time) *time.Time { return &t }
