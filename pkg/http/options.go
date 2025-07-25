package http

// Options for running http commands
type Options struct {
	Url         string
	Method      string
	Headers     map[string]string
	File        *FileDetails
	ContentType string
	FormData    map[string]string
	RequestBody string
}

type FileDetails struct {
	path      string
	fieldName string
}
