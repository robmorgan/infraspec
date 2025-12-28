package graph

// NewAWSSchema creates a RelationshipSchema pre-populated with common AWS resource relationships.
// This schema defines the relationships between AWS resources as they exist in the real AWS API.
func NewAWSSchema() *RelationshipSchema {
	schema := NewRelationshipSchema()

	// ==========================================================================
	// EC2 Relationships
	// ==========================================================================

	// Subnet -> VPC (subnets belong to VPCs)
	schema.AddRelationship("ec2", "subnet", "ec2", "vpc", SchemaEntry{
		Type:           RelContains,
		Cardinality:    CardManyToOne,
		DeleteBehavior: DeleteRestrict,
		Required:       true,
		Description:    "Subnets are contained within VPCs",
	})

	// Security Group -> VPC (security groups belong to VPCs)
	schema.AddRelationship("ec2", "security-group", "ec2", "vpc", SchemaEntry{
		Type:           RelContains,
		Cardinality:    CardManyToOne,
		DeleteBehavior: DeleteRestrict,
		Required:       true,
		Description:    "Security groups are contained within VPCs",
	})

	// Internet Gateway -> VPC (IGW attached to VPC)
	schema.AddRelationship("ec2", "internet-gateway", "ec2", "vpc", SchemaEntry{
		Type:           RelAttachedTo,
		Cardinality:    CardOneToOne,
		DeleteBehavior: DeleteRestrict,
		Required:       false,
		Description:    "Internet gateways can be attached to VPCs",
	})

	// NAT Gateway -> Subnet (NAT gateways are launched in subnets)
	schema.AddRelationship("ec2", "nat-gateway", "ec2", "subnet", SchemaEntry{
		Type:           RelContains,
		Cardinality:    CardManyToOne,
		DeleteBehavior: DeleteRestrict,
		Required:       true,
		Description:    "NAT gateways are created in subnets",
	})

	// Route Table -> VPC (route tables belong to VPCs)
	schema.AddRelationship("ec2", "route-table", "ec2", "vpc", SchemaEntry{
		Type:           RelContains,
		Cardinality:    CardManyToOne,
		DeleteBehavior: DeleteRestrict,
		Required:       true,
		Description:    "Route tables are contained within VPCs",
	})

	// Route Table -> Subnet (route table associations)
	schema.AddRelationship("ec2", "route-table", "ec2", "subnet", SchemaEntry{
		Type:           RelAssociatedWith,
		Cardinality:    CardManyToMany,
		DeleteBehavior: DeleteSetNull,
		Required:       false,
		Description:    "Route tables can be associated with subnets",
	})

	// Network ACL -> VPC (network ACLs belong to VPCs)
	schema.AddRelationship("ec2", "network-acl", "ec2", "vpc", SchemaEntry{
		Type:           RelContains,
		Cardinality:    CardManyToOne,
		DeleteBehavior: DeleteRestrict,
		Required:       true,
		Description:    "Network ACLs are contained within VPCs",
	})

	// Instance -> Subnet (instances are launched in subnets)
	schema.AddRelationship("ec2", "instance", "ec2", "subnet", SchemaEntry{
		Type:           RelReferences,
		Cardinality:    CardManyToOne,
		DeleteBehavior: DeleteRestrict,
		Required:       false, // Instances can be in EC2-Classic (legacy)
		Description:    "Instances are launched in subnets",
	})

	// Instance -> Security Group (instances use security groups)
	schema.AddRelationship("ec2", "instance", "ec2", "security-group", SchemaEntry{
		Type:           RelReferences,
		Cardinality:    CardManyToMany,
		DeleteBehavior: DeleteRestrict,
		Required:       false,
		Description:    "Instances reference security groups for network rules",
	})

	// Instance -> Key Pair (instances use key pairs)
	schema.AddRelationship("ec2", "instance", "ec2", "key-pair", SchemaEntry{
		Type:           RelReferences,
		Cardinality:    CardManyToOne,
		DeleteBehavior: DeleteSetNull,
		Required:       false,
		Description:    "Instances can use key pairs for SSH access",
	})

	// EBS Volume -> Instance (volumes attached to instances)
	schema.AddRelationship("ec2", "volume", "ec2", "instance", SchemaEntry{
		Type:           RelAttachedTo,
		Cardinality:    CardManyToOne, // A volume can only be attached to one instance at a time
		DeleteBehavior: DeleteSetNull, // Volume can exist unattached
		Required:       false,
		Description:    "EBS volumes can be attached to instances",
	})

	// Network Interface -> Subnet (ENIs belong to subnets)
	schema.AddRelationship("ec2", "network-interface", "ec2", "subnet", SchemaEntry{
		Type:           RelContains,
		Cardinality:    CardManyToOne,
		DeleteBehavior: DeleteRestrict,
		Required:       true,
		Description:    "Network interfaces are created in subnets",
	})

	// Network Interface -> Security Group
	schema.AddRelationship("ec2", "network-interface", "ec2", "security-group", SchemaEntry{
		Type:           RelReferences,
		Cardinality:    CardManyToMany,
		DeleteBehavior: DeleteRestrict,
		Required:       false,
		Description:    "Network interfaces reference security groups",
	})

	// ==========================================================================
	// IAM Relationships
	// ==========================================================================

	// Policy -> Role (policy attachment blocks role deletion)
	// Edge direction: policy points to role, so deleting a role with attached policies fails
	schema.AddRelationship("iam", "policy", "iam", "role", SchemaEntry{
		Type:           RelAssociatedWith,
		Cardinality:    CardManyToMany,
		DeleteBehavior: DeleteRestrict, // Can't delete role with attached policies
		Required:       false,
		Description:    "IAM policy attachments prevent role deletion until detached",
	})

	// User -> Policy (user-policy attachments)
	schema.AddRelationship("iam", "user", "iam", "policy", SchemaEntry{
		Type:           RelAssociatedWith,
		Cardinality:    CardManyToMany,
		DeleteBehavior: DeleteSetNull,
		Required:       false,
		Description:    "IAM users can have managed policies attached",
	})

	// Group -> Policy (group-policy attachments)
	schema.AddRelationship("iam", "group", "iam", "policy", SchemaEntry{
		Type:           RelAssociatedWith,
		Cardinality:    CardManyToMany,
		DeleteBehavior: DeleteSetNull,
		Required:       false,
		Description:    "IAM groups can have managed policies attached",
	})

	// User -> Group (user group membership)
	schema.AddRelationship("iam", "user", "iam", "group", SchemaEntry{
		Type:           RelAssociatedWith,
		Cardinality:    CardManyToMany,
		DeleteBehavior: DeleteSetNull,
		Required:       false,
		Description:    "IAM users can be members of groups",
	})

	// Instance Profile -> Role
	schema.AddRelationship("iam", "instance-profile", "iam", "role", SchemaEntry{
		Type:           RelContains,
		Cardinality:    CardOneToOne, // Instance profiles contain exactly one role
		DeleteBehavior: DeleteRestrict,
		Required:       false,
		Description:    "Instance profiles contain a single IAM role",
	})

	// ==========================================================================
	// Cross-Service Relationships (EC2 <-> IAM)
	// ==========================================================================

	// Instance -> Instance Profile
	schema.AddRelationship("ec2", "instance", "iam", "instance-profile", SchemaEntry{
		Type:           RelReferences,
		Cardinality:    CardManyToOne,
		DeleteBehavior: DeleteSetNull,
		Required:       false,
		Description:    "EC2 instances can be associated with IAM instance profiles",
	})

	// ==========================================================================
	// RDS Relationships
	// ==========================================================================

	// DB Instance -> DB Subnet Group
	schema.AddRelationship("rds", "db-instance", "rds", "db-subnet-group", SchemaEntry{
		Type:           RelReferences,
		Cardinality:    CardManyToOne,
		DeleteBehavior: DeleteRestrict,
		Required:       false, // Can use default subnet group
		Description:    "RDS instances reference DB subnet groups for networking",
	})

	// DB Instance -> Security Group (VPC security groups)
	schema.AddRelationship("rds", "db-instance", "ec2", "security-group", SchemaEntry{
		Type:           RelReferences,
		Cardinality:    CardManyToMany,
		DeleteBehavior: DeleteRestrict,
		Required:       false,
		Description:    "RDS instances reference VPC security groups",
	})

	// DB Instance -> DB Parameter Group
	schema.AddRelationship("rds", "db-instance", "rds", "db-parameter-group", SchemaEntry{
		Type:           RelReferences,
		Cardinality:    CardManyToOne,
		DeleteBehavior: DeleteRestrict,
		Required:       false, // Uses default if not specified
		Description:    "RDS instances reference parameter groups for configuration",
	})

	// DB Instance -> Option Group
	schema.AddRelationship("rds", "db-instance", "rds", "option-group", SchemaEntry{
		Type:           RelReferences,
		Cardinality:    CardManyToOne,
		DeleteBehavior: DeleteRestrict,
		Required:       false,
		Description:    "RDS instances can reference option groups",
	})

	// DB Subnet Group -> Subnet
	schema.AddRelationship("rds", "db-subnet-group", "ec2", "subnet", SchemaEntry{
		Type:           RelContains,
		Cardinality:    CardManyToMany, // Subnet groups contain multiple subnets
		DeleteBehavior: DeleteRestrict,
		Required:       true,
		Description:    "DB subnet groups contain EC2 subnets",
	})

	// ==========================================================================
	// S3 Relationships
	// ==========================================================================

	// Bucket -> IAM Role (bucket policy references)
	// Note: This is a loose reference via bucket policy, not a hard dependency
	schema.AddRelationship("s3", "bucket", "iam", "role", SchemaEntry{
		Type:           RelReferences,
		Cardinality:    CardManyToMany,
		DeleteBehavior: DeleteSetNull,
		Required:       false,
		Description:    "S3 buckets can reference IAM roles in bucket policies",
	})

	// ==========================================================================
	// DynamoDB Relationships
	// ==========================================================================

	// Table -> IAM Role (for streams, backups, etc.)
	schema.AddRelationship("dynamodb", "table", "iam", "role", SchemaEntry{
		Type:           RelReferences,
		Cardinality:    CardManyToOne,
		DeleteBehavior: DeleteSetNull,
		Required:       false,
		Description:    "DynamoDB tables can reference IAM roles for streams and backups",
	})

	// ==========================================================================
	// Lambda Relationships
	// ==========================================================================

	// Function -> IAM Role (execution role)
	schema.AddRelationship("lambda", "function", "iam", "role", SchemaEntry{
		Type:           RelReferences,
		Cardinality:    CardManyToOne,
		DeleteBehavior: DeleteRestrict,
		Required:       true,
		Description:    "Lambda functions require an execution role",
	})

	// Function -> Security Group (VPC-enabled functions)
	schema.AddRelationship("lambda", "function", "ec2", "security-group", SchemaEntry{
		Type:           RelReferences,
		Cardinality:    CardManyToMany,
		DeleteBehavior: DeleteRestrict,
		Required:       false, // Only for VPC-enabled functions
		Description:    "VPC-enabled Lambda functions reference security groups",
	})

	// Function -> Subnet (VPC-enabled functions)
	schema.AddRelationship("lambda", "function", "ec2", "subnet", SchemaEntry{
		Type:           RelReferences,
		Cardinality:    CardManyToMany,
		DeleteBehavior: DeleteRestrict,
		Required:       false, // Only for VPC-enabled functions
		Description:    "VPC-enabled Lambda functions reference subnets",
	})

	// ==========================================================================
	// SQS Relationships
	// ==========================================================================

	// Queue -> IAM Role/Policy (queue policy references)
	schema.AddRelationship("sqs", "queue", "iam", "role", SchemaEntry{
		Type:           RelReferences,
		Cardinality:    CardManyToMany,
		DeleteBehavior: DeleteSetNull,
		Required:       false,
		Description:    "SQS queues can reference IAM roles in queue policies",
	})

	return schema
}

// AWSResourceTypes returns common AWS resource type definitions.
// This can be used to document or validate resource IDs.
var AWSResourceTypes = map[string]string{
	// EC2
	"ec2:vpc":               "Amazon VPC",
	"ec2:subnet":            "VPC Subnet",
	"ec2:security-group":    "Security Group",
	"ec2:instance":          "EC2 Instance",
	"ec2:volume":            "EBS Volume",
	"ec2:internet-gateway":  "Internet Gateway",
	"ec2:nat-gateway":       "NAT Gateway",
	"ec2:route-table":       "Route Table",
	"ec2:network-acl":       "Network ACL",
	"ec2:network-interface": "Network Interface",
	"ec2:key-pair":          "Key Pair",

	// IAM
	"iam:role":             "IAM Role",
	"iam:policy":           "IAM Policy",
	"iam:user":             "IAM User",
	"iam:group":            "IAM Group",
	"iam:instance-profile": "Instance Profile",

	// RDS
	"rds:db-instance":        "RDS DB Instance",
	"rds:db-subnet-group":    "DB Subnet Group",
	"rds:db-parameter-group": "DB Parameter Group",
	"rds:option-group":       "Option Group",

	// S3
	"s3:bucket": "S3 Bucket",

	// DynamoDB
	"dynamodb:table": "DynamoDB Table",

	// Lambda
	"lambda:function": "Lambda Function",

	// SQS
	"sqs:queue": "SQS Queue",

	// STS
	"sts:assumed-role": "Assumed Role Session",
}
