package iam

import (
	"strings"
	"time"
)

// AWSManagedPolicy represents a pre-defined AWS managed policy
type AWSManagedPolicy struct {
	PolicyName       string
	PolicyId         string
	Arn              string
	Path             string
	Description      string
	DefaultVersionId string
	Document         string
	CreateDate       time.Time
	UpdateDate       time.Time
}

// permissiveStubDocument is a permissive policy document used as a fallback
// for AWS managed policies that are not explicitly defined.
const permissiveStubDocument = `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "*",
            "Resource": "*"
        }
    ]
}`

// awsManagedPolicies contains the top 25 most commonly used AWS managed policies.
// These are pre-defined and available without explicit creation.
var awsManagedPolicies = map[string]AWSManagedPolicy{
	// Lambda policies
	"arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole": {
		PolicyName:       "AWSLambdaBasicExecutionRole",
		PolicyId:         "ANPAJNCQGXC42545SKXIK",
		Arn:              "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
		Path:             "/service-role/",
		Description:      "Provides write permissions to CloudWatch Logs.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Resource": "*"
        }
    ]
}`,
	},
	"arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole": {
		PolicyName:       "AWSLambdaVPCAccessExecutionRole",
		PolicyId:         "ANPAJVTME3YLVNL72YR2K",
		Arn:              "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole",
		Path:             "/service-role/",
		Description:      "Provides minimum permissions for a Lambda function to execute while accessing a resource within a VPC.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents",
                "ec2:CreateNetworkInterface",
                "ec2:DescribeNetworkInterfaces",
                "ec2:DeleteNetworkInterface",
                "ec2:AssignPrivateIpAddresses",
                "ec2:UnassignPrivateIpAddresses"
            ],
            "Resource": "*"
        }
    ]
}`,
	},
	"arn:aws:iam::aws:policy/service-role/AWSLambdaDynamoDBExecutionRole": {
		PolicyName:       "AWSLambdaDynamoDBExecutionRole",
		PolicyId:         "ANPAIP7WNAGMIPYNW4WQG",
		Arn:              "arn:aws:iam::aws:policy/service-role/AWSLambdaDynamoDBExecutionRole",
		Path:             "/service-role/",
		Description:      "Provides list and read access to DynamoDB streams and write permissions to CloudWatch logs.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "dynamodb:DescribeStream",
                "dynamodb:GetRecords",
                "dynamodb:GetShardIterator",
                "dynamodb:ListStreams",
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Resource": "*"
        }
    ]
}`,
	},
	"arn:aws:iam::aws:policy/service-role/AWSLambdaSQSQueueExecutionRole": {
		PolicyName:       "AWSLambdaSQSQueueExecutionRole",
		PolicyId:         "ANPAJFWJZI6LQQTROCBEY",
		Arn:              "arn:aws:iam::aws:policy/service-role/AWSLambdaSQSQueueExecutionRole",
		Path:             "/service-role/",
		Description:      "Provides receive message, delete message, and read attribute access to SQS queues, and write permissions to CloudWatch logs.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "sqs:ReceiveMessage",
                "sqs:DeleteMessage",
                "sqs:GetQueueAttributes",
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents"
            ],
            "Resource": "*"
        }
    ]
}`,
	},
	"arn:aws:iam::aws:policy/AWSLambda_FullAccess": {
		PolicyName:       "AWSLambda_FullAccess",
		PolicyId:         "ANPAZKAPJZG4ONJPM5YFF",
		Arn:              "arn:aws:iam::aws:policy/AWSLambda_FullAccess",
		Path:             "/",
		Description:      "Grants full access to AWS Lambda service, AWS Lambda console features, and other related AWS services.",
		DefaultVersionId: "v1",
		Document:         permissiveStubDocument,
	},

	// Administrator and PowerUser policies
	"arn:aws:iam::aws:policy/AdministratorAccess": {
		PolicyName:       "AdministratorAccess",
		PolicyId:         "ANPAIWMBCKSKIEE64ZLYK",
		Arn:              "arn:aws:iam::aws:policy/AdministratorAccess",
		Path:             "/",
		Description:      "Provides full access to AWS services and resources.",
		DefaultVersionId: "v1",
		Document:         permissiveStubDocument,
	},
	"arn:aws:iam::aws:policy/PowerUserAccess": {
		PolicyName:       "PowerUserAccess",
		PolicyId:         "ANPAJYRXTHIB4FOVS3ZXS",
		Arn:              "arn:aws:iam::aws:policy/PowerUserAccess",
		Path:             "/",
		Description:      "Provides full access to AWS services and resources, but does not allow management of Users and groups.",
		DefaultVersionId: "v1",
		Document:         permissiveStubDocument,
	},
	"arn:aws:iam::aws:policy/ReadOnlyAccess": {
		PolicyName:       "ReadOnlyAccess",
		PolicyId:         "ANPAILL3HVNFSB6DCOWYQ",
		Arn:              "arn:aws:iam::aws:policy/ReadOnlyAccess",
		Path:             "/",
		Description:      "Provides read-only access to AWS services and resources.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "*:Describe*",
                "*:Get*",
                "*:List*"
            ],
            "Resource": "*"
        }
    ]
}`,
	},

	// S3 policies
	"arn:aws:iam::aws:policy/AmazonS3FullAccess": {
		PolicyName:       "AmazonS3FullAccess",
		PolicyId:         "ANPAIFIR6V6BVTRAHWINE",
		Arn:              "arn:aws:iam::aws:policy/AmazonS3FullAccess",
		Path:             "/",
		Description:      "Provides full access to all buckets via the AWS Management Console.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "s3:*",
            "Resource": "*"
        }
    ]
}`,
	},
	"arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess": {
		PolicyName:       "AmazonS3ReadOnlyAccess",
		PolicyId:         "ANPAIZTJ4DXE7G6AGAE6M",
		Arn:              "arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess",
		Path:             "/",
		Description:      "Provides read only access to all buckets via the AWS Management Console.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:Get*",
                "s3:List*"
            ],
            "Resource": "*"
        }
    ]
}`,
	},

	// DynamoDB policies
	"arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess": {
		PolicyName:       "AmazonDynamoDBFullAccess",
		PolicyId:         "ANPAIY2XFNA232OKQCJWC",
		Arn:              "arn:aws:iam::aws:policy/AmazonDynamoDBFullAccess",
		Path:             "/",
		Description:      "Provides full access to Amazon DynamoDB via the AWS Management Console.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "dynamodb:*",
            "Resource": "*"
        }
    ]
}`,
	},
	"arn:aws:iam::aws:policy/AmazonDynamoDBReadOnlyAccess": {
		PolicyName:       "AmazonDynamoDBReadOnlyAccess",
		PolicyId:         "ANPAINUGF2JSOSUY76KYA",
		Arn:              "arn:aws:iam::aws:policy/AmazonDynamoDBReadOnlyAccess",
		Path:             "/",
		Description:      "Provides read only access to Amazon DynamoDB via the AWS Management Console.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "dynamodb:BatchGetItem",
                "dynamodb:DescribeTable",
                "dynamodb:GetItem",
                "dynamodb:ListTables",
                "dynamodb:Query",
                "dynamodb:Scan"
            ],
            "Resource": "*"
        }
    ]
}`,
	},

	// EC2 policies
	"arn:aws:iam::aws:policy/AmazonEC2FullAccess": {
		PolicyName:       "AmazonEC2FullAccess",
		PolicyId:         "ANPAI3VAJF5ZCRZ7MCQE6",
		Arn:              "arn:aws:iam::aws:policy/AmazonEC2FullAccess",
		Path:             "/",
		Description:      "Provides full access to Amazon EC2 via the AWS Management Console.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "ec2:*",
            "Resource": "*"
        }
    ]
}`,
	},
	"arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess": {
		PolicyName:       "AmazonEC2ReadOnlyAccess",
		PolicyId:         "ANPAIGDT4S5U7O6UNL6C2",
		Arn:              "arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess",
		Path:             "/",
		Description:      "Provides read only access to Amazon EC2 via the AWS Management Console.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "ec2:Describe*",
            "Resource": "*"
        }
    ]
}`,
	},

	// RDS policies
	"arn:aws:iam::aws:policy/AmazonRDSFullAccess": {
		PolicyName:       "AmazonRDSFullAccess",
		PolicyId:         "ANPAI2D4VEWVHYVK6PTUM",
		Arn:              "arn:aws:iam::aws:policy/AmazonRDSFullAccess",
		Path:             "/",
		Description:      "Provides full access to Amazon RDS via the AWS Management Console.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "rds:*",
            "Resource": "*"
        }
    ]
}`,
	},
	"arn:aws:iam::aws:policy/AmazonRDSReadOnlyAccess": {
		PolicyName:       "AmazonRDSReadOnlyAccess",
		PolicyId:         "ANPAJKTTTYV2IIHKLZ346",
		Arn:              "arn:aws:iam::aws:policy/AmazonRDSReadOnlyAccess",
		Path:             "/",
		Description:      "Provides read only access to Amazon RDS via the AWS Management Console.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "rds:Describe*",
                "rds:ListTagsForResource"
            ],
            "Resource": "*"
        }
    ]
}`,
	},

	// IAM policies
	"arn:aws:iam::aws:policy/IAMFullAccess": {
		PolicyName:       "IAMFullAccess",
		PolicyId:         "ANPAI7XKCFMBPM3QQRRVQ",
		Arn:              "arn:aws:iam::aws:policy/IAMFullAccess",
		Path:             "/",
		Description:      "Provides full access to IAM via the AWS Management Console.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "iam:*",
            "Resource": "*"
        }
    ]
}`,
	},
	"arn:aws:iam::aws:policy/IAMReadOnlyAccess": {
		PolicyName:       "IAMReadOnlyAccess",
		PolicyId:         "ANPAJKSO7NDY4T57MWDSQ",
		Arn:              "arn:aws:iam::aws:policy/IAMReadOnlyAccess",
		Path:             "/",
		Description:      "Provides read only access to IAM via the AWS Management Console.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "iam:Get*",
                "iam:List*"
            ],
            "Resource": "*"
        }
    ]
}`,
	},

	// SQS policies
	"arn:aws:iam::aws:policy/AmazonSQSFullAccess": {
		PolicyName:       "AmazonSQSFullAccess",
		PolicyId:         "ANPAI4UIINUVGB5SEC57G",
		Arn:              "arn:aws:iam::aws:policy/AmazonSQSFullAccess",
		Path:             "/",
		Description:      "Provides full access to Amazon SQS via the AWS Management Console.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "sqs:*",
            "Resource": "*"
        }
    ]
}`,
	},
	"arn:aws:iam::aws:policy/AmazonSQSReadOnlyAccess": {
		PolicyName:       "AmazonSQSReadOnlyAccess",
		PolicyId:         "ANPAJFWMCWF2ZYOGKRXZS",
		Arn:              "arn:aws:iam::aws:policy/AmazonSQSReadOnlyAccess",
		Path:             "/",
		Description:      "Provides read only access to Amazon SQS via the AWS Management Console.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "sqs:GetQueueAttributes",
                "sqs:GetQueueUrl",
                "sqs:ListDeadLetterSourceQueues",
                "sqs:ListQueues"
            ],
            "Resource": "*"
        }
    ]
}`,
	},

	// CloudWatch policies
	"arn:aws:iam::aws:policy/CloudWatchFullAccess": {
		PolicyName:       "CloudWatchFullAccess",
		PolicyId:         "ANPAIKEABORKUXN6DEAZU",
		Arn:              "arn:aws:iam::aws:policy/CloudWatchFullAccess",
		Path:             "/",
		Description:      "Provides full access to CloudWatch.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "cloudwatch:*",
                "logs:*"
            ],
            "Resource": "*"
        }
    ]
}`,
	},
	"arn:aws:iam::aws:policy/CloudWatchReadOnlyAccess": {
		PolicyName:       "CloudWatchReadOnlyAccess",
		PolicyId:         "ANPAJN23PDQP7SZQAE3QE",
		Arn:              "arn:aws:iam::aws:policy/CloudWatchReadOnlyAccess",
		Path:             "/",
		Description:      "Provides read only access to CloudWatch.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "cloudwatch:Describe*",
                "cloudwatch:Get*",
                "cloudwatch:List*",
                "logs:Describe*",
                "logs:Get*",
                "logs:FilterLogEvents"
            ],
            "Resource": "*"
        }
    ]
}`,
	},

	// SSM policies
	"arn:aws:iam::aws:policy/AmazonSSMFullAccess": {
		PolicyName:       "AmazonSSMFullAccess",
		PolicyId:         "ANPAJA7V6HI7WXOQIQXNU",
		Arn:              "arn:aws:iam::aws:policy/AmazonSSMFullAccess",
		Path:             "/",
		Description:      "Provides full access to Amazon SSM.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "ssm:*",
            "Resource": "*"
        }
    ]
}`,
	},
	"arn:aws:iam::aws:policy/AmazonSSMReadOnlyAccess": {
		PolicyName:       "AmazonSSMReadOnlyAccess",
		PolicyId:         "ANPAJODSKQGGJTHRYZ6TM",
		Arn:              "arn:aws:iam::aws:policy/AmazonSSMReadOnlyAccess",
		Path:             "/",
		Description:      "Provides read only access to Amazon SSM.",
		DefaultVersionId: "v1",
		Document: `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ssm:Describe*",
                "ssm:Get*",
                "ssm:List*"
            ],
            "Resource": "*"
        }
    ]
}`,
	},
}

// isAWSManagedPolicyArn checks if the given ARN is an AWS managed policy ARN.
// AWS managed policies have the format: arn:aws:iam::aws:policy/...
func isAWSManagedPolicyArn(arn string) bool {
	return strings.HasPrefix(arn, "arn:aws:iam::aws:policy/")
}

// getAWSManagedPolicy returns the AWS managed policy for the given ARN.
// If the policy is not in our predefined list, it returns a stub policy.
// Returns nil if the ARN is not an AWS managed policy ARN.
func getAWSManagedPolicy(arn string) *AWSManagedPolicy {
	if !isAWSManagedPolicyArn(arn) {
		return nil
	}

	// Check if we have an explicit definition
	if policy, ok := awsManagedPolicies[arn]; ok {
		// Set timestamps if not already set
		if policy.CreateDate.IsZero() {
			policy.CreateDate = time.Date(2015, 2, 6, 18, 40, 58, 0, time.UTC)
			policy.UpdateDate = time.Date(2015, 2, 6, 18, 40, 58, 0, time.UTC)
		}
		return &policy
	}

	// Return a stub policy for unrecognized AWS managed policies
	policyName := extractPolicyNameFromArn(arn)
	path := extractPolicyPathFromArn(arn)

	return &AWSManagedPolicy{
		PolicyName:       policyName,
		PolicyId:         "ANPA" + generateStubPolicyId(arn),
		Arn:              arn,
		Path:             path,
		Description:      "AWS managed policy (stub)",
		DefaultVersionId: "v1",
		Document:         permissiveStubDocument,
		CreateDate:       time.Date(2015, 2, 6, 18, 40, 58, 0, time.UTC),
		UpdateDate:       time.Date(2015, 2, 6, 18, 40, 58, 0, time.UTC),
	}
}

// extractPolicyPathFromArn extracts the path from an AWS managed policy ARN.
// For example: arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
// returns "/service-role/"
func extractPolicyPathFromArn(arn string) string {
	parts := strings.Split(arn, ":policy")
	if len(parts) < 2 {
		return "/"
	}
	pathAndName := parts[1]
	// Get everything up to and including the last slash
	lastSlash := strings.LastIndex(pathAndName, "/")
	if lastSlash <= 0 {
		return "/"
	}
	return pathAndName[:lastSlash+1]
}

// generateStubPolicyId generates a deterministic policy ID suffix based on the ARN.
// This ensures the same ARN always gets the same ID.
func generateStubPolicyId(arn string) string {
	// Simple hash-like function to generate a 17-char alphanumeric ID
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 17)
	hash := uint64(0)
	for i, c := range arn {
		hash = hash*31 + uint64(c) + uint64(i)
	}
	for i := range result {
		result[i] = chars[hash%uint64(len(chars))]
		hash /= uint64(len(chars))
		if hash == 0 {
			hash = uint64(i + 1)
		}
	}
	return string(result)
}

// toXMLPolicy converts an AWSManagedPolicy to XMLPolicy for response serialization.
func (p *AWSManagedPolicy) toXMLPolicy() XMLPolicy {
	return XMLPolicy{
		PolicyName:       p.PolicyName,
		PolicyId:         p.PolicyId,
		Arn:              p.Arn,
		Path:             p.Path,
		Description:      p.Description,
		DefaultVersionId: p.DefaultVersionId,
		CreateDate:       p.CreateDate,
		UpdateDate:       p.UpdateDate,
		AttachmentCount:  0, // AWS managed policies don't track this per-account
		IsAttachable:     true,
	}
}

// toXMLPolicyVersion converts an AWSManagedPolicy to XMLPolicyVersion for response serialization.
func (p *AWSManagedPolicy) toXMLPolicyVersion() XMLPolicyVersion {
	return XMLPolicyVersion{
		VersionId:        p.DefaultVersionId,
		Document:         p.Document,
		IsDefaultVersion: true,
		CreateDate:       p.CreateDate,
	}
}
