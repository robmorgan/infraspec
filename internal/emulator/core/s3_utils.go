package emulator

import "strings"

// S3HostInfo contains information extracted from an S3 virtual-hosted style request.
type S3HostInfo struct {
	IsVirtualHosted bool   // True if this is a virtual-hosted style request
	BucketName      string // Bucket name extracted from the host (empty for path-style)
}

// ParseS3Host parses an HTTP host header to extract S3 virtual-hosted style information.
// It handles the following patterns:
//   - bucket-name.s3.infraspec.sh (virtual-hosted, bucket = "bucket-name")
//   - bucket-name.s3.localhost (virtual-hosted, bucket = "bucket-name")
//   - bucket-name.s3.127.0.0.1.nip.io (virtual-hosted with nip.io, bucket = "bucket-name")
//   - bucket-name.localhost (legacy virtual-hosted, bucket = "bucket-name")
//   - s3.infraspec.sh (path-style base S3 endpoint)
//   - s3.localhost (path-style base S3 endpoint)
//   - s3.127.0.0.1.nip.io (path-style with nip.io)
//
// The host parameter should be the value of the Host header (with or without port).
func ParseS3Host(host string) S3HostInfo {
	if host == "" {
		return S3HostInfo{}
	}

	// Remove port from host if present
	hostWithoutPort := strings.Split(host, ":")[0]
	parts := strings.Split(hostWithoutPort, ".")

	if len(parts) < 2 {
		return S3HostInfo{}
	}

	// Check for nip.io/sslip.io DNS service patterns first
	// Pattern: bucket-name.s3.127.0.0.1.nip.io (virtual-hosted with nip.io)
	// Pattern: s3.127.0.0.1.nip.io (path-style with nip.io)
	if isNipIOHost(parts) {
		// Virtual-hosted: bucket-name.s3.IP.nip.io (parts[0]=bucket, parts[1]=s3)
		if len(parts) >= 5 && parts[1] == "s3" {
			return S3HostInfo{
				IsVirtualHosted: true,
				BucketName:      parts[0],
			}
		}
		// Path-style: s3.IP.nip.io (parts[0]=s3)
		if parts[0] == "s3" {
			return S3HostInfo{
				IsVirtualHosted: false,
				BucketName:      "",
			}
		}
	}

	// Pattern: bucket-name.s3.infraspec.sh or bucket-name.s3.localhost
	// This is virtual-hosted style where bucket name is the first subdomain
	if len(parts) >= 3 && parts[1] == "s3" {
		return S3HostInfo{
			IsVirtualHosted: true,
			BucketName:      parts[0],
		}
	}

	// Pattern: bucket-name.localhost (legacy support)
	// Exclude "s3.localhost" and "localhost.localhost"
	if parts[1] == "localhost" && parts[0] != "localhost" && parts[0] != "s3" {
		return S3HostInfo{
			IsVirtualHosted: true,
			BucketName:      parts[0],
		}
	}

	// Pattern: s3.infraspec.sh or s3.localhost (base S3 endpoint, path-style)
	// Not virtual-hosted, but still an S3 request
	if parts[0] == "s3" {
		return S3HostInfo{
			IsVirtualHosted: false,
			BucketName:      "",
		}
	}

	return S3HostInfo{}
}

// isNipIOHost checks if the host parts indicate a nip.io or sslip.io DNS service.
// These services provide wildcard DNS for IP addresses (e.g., bucket.s3.127.0.0.1.nip.io → 127.0.0.1)
func isNipIOHost(parts []string) bool {
	if len(parts) < 2 {
		return false
	}
	lastPart := parts[len(parts)-1]
	secondLast := parts[len(parts)-2]
	return lastPart == "io" && (secondLast == "nip" || secondLast == "sslip")
}

// IsS3VirtualHostedRequest checks if the given host represents an S3 virtual-hosted style request.
// This includes patterns like:
//   - bucket-name.s3.infraspec.sh
//   - bucket-name.s3.localhost
//   - bucket-name.localhost (legacy)
//
// The host parameter should be the value of the Host header (with or without port).
func IsS3VirtualHostedRequest(host string) bool {
	return ParseS3Host(host).IsVirtualHosted
}

// IsS3Request checks if the given host represents any S3 request (virtual-hosted or path-style).
// This includes patterns like:
//   - bucket-name.s3.infraspec.sh (virtual-hosted)
//   - bucket-name.s3.localhost (virtual-hosted)
//   - bucket-name.s3.127.0.0.1.nip.io (virtual-hosted with nip.io)
//   - bucket-name.localhost (legacy virtual-hosted)
//   - s3.infraspec.sh (path-style)
//   - s3.localhost (path-style)
//   - s3.127.0.0.1.nip.io (path-style with nip.io)
//
// The host parameter should be the value of the Host header (with or without port).
func IsS3Request(host string) bool {
	if host == "" {
		return false
	}

	// Remove port from host if present
	hostWithoutPort := strings.Split(host, ":")[0]
	parts := strings.Split(hostWithoutPort, ".")

	if len(parts) < 2 {
		return false
	}

	// Check for nip.io/sslip.io patterns: *.s3.IP.nip.io or s3.IP.nip.io
	if isNipIOHost(parts) {
		// Look for "s3" in the first or second position
		if parts[0] == "s3" || (len(parts) >= 2 && parts[1] == "s3") {
			return true
		}
	}

	// Pattern: bucket-name.s3.something or s3.something
	if len(parts) >= 2 && (parts[0] == "s3" || (len(parts) >= 3 && parts[1] == "s3")) {
		return true
	}

	// Pattern: bucket-name.localhost (legacy) - but not s3.localhost (handled above)
	if parts[1] == "localhost" && parts[0] != "localhost" {
		return true
	}

	return false
}

// ExtractBucketNameFromHost extracts the bucket name from an S3 virtual-hosted style host.
// Returns an empty string if the host is not a virtual-hosted style request.
//
// The host parameter should be the value of the Host header (with or without port).
func ExtractBucketNameFromHost(host string) string {
	return ParseS3Host(host).BucketName
}
