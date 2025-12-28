package metadata

import "time"

// InstanceMetadata represents EC2 instance metadata
type InstanceMetadata struct {
	InstanceID       string `json:"instanceId"`
	InstanceType     string `json:"instanceType"`
	AMIID            string `json:"imageId"`
	AvailabilityZone string `json:"availabilityZone"`
	Region           string `json:"region"`
	LocalIPv4        string `json:"privateIp"`
	PublicIPv4       string `json:"publicIp"`
	MAC              string `json:"macAddress"`
	UserData         string `json:"userData"`
	IAMRole          string `json:"iamInstanceProfile,omitempty"`
	VPCID            string `json:"vpcId"`
	SubnetID         string `json:"subnetId"`
}

// IAMCredentials represents temporary IAM role credentials
type IAMCredentials struct {
	AccessKeyID     string    `json:"AccessKeyId"`
	SecretAccessKey string    `json:"SecretAccessKey"`
	Token           string    `json:"Token"`
	Expiration      time.Time `json:"Expiration"`
	Code            string    `json:"Code"`
	LastUpdated     time.Time `json:"LastUpdated"`
	Type            string    `json:"Type"`
}

// IMDSv2Token represents a session token for IMDSv2
type IMDSv2Token struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	TTL       int       `json:"ttl"`
}

// NetworkInterface represents a network interface
type NetworkInterface struct {
	MAC              string   `json:"mac"`
	DeviceNumber     int      `json:"deviceNumber"`
	SubnetID         string   `json:"subnetId"`
	VPCID            string   `json:"vpcId"`
	SecurityGroupIDs []string `json:"securityGroupIds"`
	LocalIPv4s       []string `json:"localIpv4s"`
	PublicIPv4s      []string `json:"publicIpv4s"`
}
