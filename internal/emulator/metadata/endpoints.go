package metadata

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/robmorgan/infraspec/internal/emulator/core"
)

// MetadataEndpoint represents a metadata endpoint handler
type MetadataEndpoint struct {
	state emulator.StateManager
}

// NewMetadataEndpoint creates a new metadata endpoint handler
func NewMetadataEndpoint(state emulator.StateManager) *MetadataEndpoint {
	return &MetadataEndpoint{state: state}
}

// GetMetadata retrieves metadata for a given path
func (m *MetadataEndpoint) GetMetadata(path string) (string, error) {
	// Remove leading/trailing slashes
	path = strings.Trim(path, "/")

	// Handle root paths - return directory listings
	if path == "" || path == "latest" {
		return "meta-data/\nuser-data\n", nil
	}

	if path == "latest/meta-data" || path == "meta-data" {
		return m.listMetaDataDirectory(), nil
	}

	// Remove "latest/" prefix if present
	path = strings.TrimPrefix(path, "latest/")

	// Route to specific handlers
	switch {
	case path == "user-data":
		return m.getUserData()
	case strings.HasPrefix(path, "meta-data/"):
		return m.handleMetaDataPath(strings.TrimPrefix(path, "meta-data/"))
	default:
		return "", fmt.Errorf("not found: %s", path)
	}
}

// listMetaDataDirectory returns the meta-data directory listing
func (m *MetadataEndpoint) listMetaDataDirectory() string {
	return `ami-id
instance-id
instance-type
local-ipv4
public-ipv4
mac
placement/
network/
iam/
`
}

// handleMetaDataPath handles paths under /meta-data/
func (m *MetadataEndpoint) handleMetaDataPath(path string) (string, error) {
	switch {
	// Core instance info
	case path == "instance-id":
		return m.getStateValue("metadata:instance:instance-id")
	case path == "instance-type":
		return m.getStateValue("metadata:instance:instance-type")
	case path == "ami-id":
		return m.getStateValue("metadata:instance:ami-id")
	case path == "local-ipv4":
		return m.getStateValue("metadata:instance:local-ipv4")
	case path == "public-ipv4":
		return m.getStateValue("metadata:instance:public-ipv4")
	case path == "mac":
		return m.getStateValue("metadata:instance:mac")

	// Placement
	case path == "placement" || path == "placement/":
		return "availability-zone\nregion\n", nil
	case path == "placement/availability-zone":
		return m.getStateValue("metadata:instance:availability-zone")
	case path == "placement/region":
		return m.getStateValue("metadata:instance:region")

	// Network
	case path == "network" || path == "network/":
		return "interfaces/\n", nil
	case path == "network/interfaces" || path == "network/interfaces/":
		return "macs/\n", nil
	case path == "network/interfaces/macs" || path == "network/interfaces/macs/":
		mac, err := m.getStateValue("metadata:instance:mac")
		if err != nil {
			return "", err
		}
		return mac + "/\n", nil
	case strings.HasPrefix(path, "network/interfaces/macs/"):
		return m.handleNetworkInterface(strings.TrimPrefix(path, "network/interfaces/macs/"))

	// IAM
	case path == "iam" || path == "iam/":
		return "security-credentials/\n", nil
	case path == "iam/security-credentials" || path == "iam/security-credentials/":
		roleName, err := m.getStateValue("metadata:instance:iam-role")
		if err != nil {
			return "", err
		}
		if roleName == "" {
			return "", nil
		}
		return roleName + "\n", nil
	case strings.HasPrefix(path, "iam/security-credentials/"):
		roleName := strings.TrimPrefix(path, "iam/security-credentials/")
		return m.getIAMCredentials(roleName)

	default:
		return "", fmt.Errorf("not found: %s", path)
	}
}

// handleNetworkInterface handles network interface metadata paths
func (m *MetadataEndpoint) handleNetworkInterface(path string) (string, error) {
	// Extract MAC address from path
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid network interface path")
	}

	mac := strings.TrimSuffix(parts[0], "/")
	expectedMAC, err := m.getStateValue("metadata:instance:mac")
	if err != nil {
		return "", err
	}

	if mac != expectedMAC {
		return "", fmt.Errorf("MAC address not found: %s", mac)
	}

	// Handle directory listing
	if len(parts) == 1 || parts[1] == "" {
		return "subnet-id\nvpc-id\nlocal-ipv4s\npublic-ipv4s\ndevice-number\n", nil
	}

	// Handle specific fields
	switch parts[1] {
	case "subnet-id":
		return m.getStateValue("metadata:instance:subnet-id")
	case "vpc-id":
		return m.getStateValue("metadata:instance:vpc-id")
	case "local-ipv4s":
		ip, err := m.getStateValue("metadata:instance:local-ipv4")
		if err != nil {
			return "", err
		}
		return ip + "\n", nil
	case "public-ipv4s":
		ip, err := m.getStateValue("metadata:instance:public-ipv4")
		if err != nil {
			return "", err
		}
		return ip + "\n", nil
	case "device-number":
		return "0", nil
	default:
		return "", fmt.Errorf("not found: %s", parts[1])
	}
}

// getUserData retrieves user data
func (m *MetadataEndpoint) getUserData() (string, error) {
	userData, err := m.getStateValue("metadata:instance:user-data")
	if err != nil {
		return "", err
	}
	// Return 404 if user data is empty (AWS behavior)
	if userData == "" {
		return "", fmt.Errorf("not found: user-data")
	}
	return userData, nil
}

// getIAMCredentials retrieves IAM role credentials as JSON
func (m *MetadataEndpoint) getIAMCredentials(roleName string) (string, error) {
	key := fmt.Sprintf("metadata:iam:credentials:%s", roleName)
	var credentials IAMCredentials
	if err := m.state.Get(key, &credentials); err != nil {
		return "", fmt.Errorf("IAM role not found: %s", roleName)
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(credentials)
	if err != nil {
		return "", fmt.Errorf("failed to marshal credentials: %w", err)
	}

	return string(jsonBytes), nil
}

// getStateValue retrieves a string value from state
func (m *MetadataEndpoint) getStateValue(key string) (string, error) {
	var value string
	if err := m.state.Get(key, &value); err != nil {
		return "", fmt.Errorf("metadata not found for key: %s", key)
	}
	return value, nil
}
