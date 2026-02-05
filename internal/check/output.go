package check

import "io"

// Formatter defines the interface for output formatters.
type Formatter interface {
	// Format writes the check summary to the given writer.
	Format(w io.Writer, summary *Summary) error
}

// NewFormatter creates a formatter based on the format name.
// Supported formats: "text" (default), "json".
func NewFormatter(format string) Formatter {
	switch format {
	case "json":
		return &JSONFormatter{}
	default:
		return &TextFormatter{}
	}
}
