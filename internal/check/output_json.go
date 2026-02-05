package check

import (
	"encoding/json"
	"io"
)

// JSONFormatter outputs check results as JSON.
type JSONFormatter struct{}

// Format writes the check summary as formatted JSON.
func (f *JSONFormatter) Format(w io.Writer, summary *Summary) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(summary)
}
