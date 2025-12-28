package metadata

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/robmorgan/infraspec/internal/emulator/core"
)

// InitializeDefaults sets up default metadata in the state manager
func InitializeDefaults(state emulator.StateManager) error {
	// Generate unique IDs
	instanceID := fmt.Sprintf("i-%s", uuid.New().String()[:17])
	amiID := fmt.Sprintf("ami-%s", uuid.New().String()[:17])

	// Create default instance metadata
	metadata := &InstanceMetadata{
		InstanceID:       instanceID,
		InstanceType:     "t3.micro",
		AMIID:            amiID,
		AvailabilityZone: "us-east-1a",
		Region:           "us-east-1",
		LocalIPv4:        "172.31.32.10",
		PublicIPv4:       "54.123.45.67",
		MAC:              "0e:12:34:56:78:9a",
		UserData:         "", // Empty by default
		IAMRole:          "test-role",
		VPCID:            fmt.Sprintf("vpc-%s", uuid.New().String()[:17]),
		SubnetID:         fmt.Sprintf("subnet-%s", uuid.New().String()[:17]),
	}

	// Store each metadata field
	if err := state.Set("metadata:instance:instance-id", metadata.InstanceID); err != nil {
		return fmt.Errorf("failed to set instance-id: %w", err)
	}
	if err := state.Set("metadata:instance:instance-type", metadata.InstanceType); err != nil {
		return fmt.Errorf("failed to set instance-type: %w", err)
	}
	if err := state.Set("metadata:instance:ami-id", metadata.AMIID); err != nil {
		return fmt.Errorf("failed to set ami-id: %w", err)
	}
	if err := state.Set("metadata:instance:availability-zone", metadata.AvailabilityZone); err != nil {
		return fmt.Errorf("failed to set availability-zone: %w", err)
	}
	if err := state.Set("metadata:instance:region", metadata.Region); err != nil {
		return fmt.Errorf("failed to set region: %w", err)
	}
	if err := state.Set("metadata:instance:local-ipv4", metadata.LocalIPv4); err != nil {
		return fmt.Errorf("failed to set local-ipv4: %w", err)
	}
	if err := state.Set("metadata:instance:public-ipv4", metadata.PublicIPv4); err != nil {
		return fmt.Errorf("failed to set public-ipv4: %w", err)
	}
	if err := state.Set("metadata:instance:mac", metadata.MAC); err != nil {
		return fmt.Errorf("failed to set mac: %w", err)
	}
	if err := state.Set("metadata:instance:user-data", metadata.UserData); err != nil {
		return fmt.Errorf("failed to set user-data: %w", err)
	}
	if err := state.Set("metadata:instance:iam-role", metadata.IAMRole); err != nil {
		return fmt.Errorf("failed to set iam-role: %w", err)
	}
	if err := state.Set("metadata:instance:vpc-id", metadata.VPCID); err != nil {
		return fmt.Errorf("failed to set vpc-id: %w", err)
	}
	if err := state.Set("metadata:instance:subnet-id", metadata.SubnetID); err != nil {
		return fmt.Errorf("failed to set subnet-id: %w", err)
	}

	// Initialize default IAM role credentials
	if err := initializeIAMCredentials(state, metadata.IAMRole); err != nil {
		return fmt.Errorf("failed to initialize IAM credentials: %w", err)
	}

	return nil
}

// initializeIAMCredentials creates default IAM role credentials
func initializeIAMCredentials(state emulator.StateManager, roleName string) error {
	credentials := &IAMCredentials{
		AccessKeyID:     fmt.Sprintf("ASIA%s", generateRandomString(16)),
		SecretAccessKey: generateRandomString(40),
		Token:           generateRandomString(356),
		Expiration:      time.Now().Add(6 * time.Hour),
		Code:            "Success",
		LastUpdated:     time.Now(),
		Type:            "AWS-HMAC",
	}

	key := fmt.Sprintf("metadata:iam:credentials:%s", roleName)
	if err := state.Set(key, credentials); err != nil {
		return fmt.Errorf("failed to set IAM credentials: %w", err)
	}

	return nil
}

// generateRandomString generates a random alphanumeric string of specified length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		// Use UUID as source of randomness
		id := uuid.New()
		result[i] = charset[int(id[i%16])%len(charset)]
	}
	return string(result)
}
