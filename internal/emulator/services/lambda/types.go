package lambda

import (
	"time"
)

// StoredFunction represents a Lambda function stored in state
type StoredFunction struct {
	FunctionName        string                     `json:"FunctionName"`
	FunctionArn         string                     `json:"FunctionArn"`
	Runtime             string                     `json:"Runtime,omitempty"`
	Role                string                     `json:"Role"`
	Handler             string                     `json:"Handler,omitempty"`
	Description         string                     `json:"Description,omitempty"`
	Timeout             int32                      `json:"Timeout"`
	MemorySize          int32                      `json:"MemorySize"`
	CodeSha256          string                     `json:"CodeSha256"`
	CodeSize            int64                      `json:"CodeSize"`
	Version             string                     `json:"Version"`
	State               string                     `json:"State"`
	StateReason         string                     `json:"StateReason,omitempty"`
	StateReasonCode     string                     `json:"StateReasonCode,omitempty"`
	LastModified        string                     `json:"LastModified"`
	LastUpdateStatus    string                     `json:"LastUpdateStatus,omitempty"`
	PackageType         string                     `json:"PackageType"`
	Architectures       []string                   `json:"Architectures,omitempty"`
	RevisionId          string                     `json:"RevisionId"`
	Tags                map[string]string          `json:"Tags,omitempty"`
	Environment         *StoredEnvironment         `json:"Environment,omitempty"`
	VpcConfig           *StoredVpcConfig           `json:"VpcConfig,omitempty"`
	DeadLetterConfig    *StoredDeadLetterConfig    `json:"DeadLetterConfig,omitempty"`
	TracingConfig       *StoredTracingConfig       `json:"TracingConfig,omitempty"`
	EphemeralStorage    *StoredEphemeralStorage    `json:"EphemeralStorage,omitempty"`
	Layers              []string                   `json:"Layers,omitempty"`
	KMSKeyArn           string                     `json:"KMSKeyArn,omitempty"`
	FileSystemConfigs   []StoredFileSystemConfig   `json:"FileSystemConfigs,omitempty"`
	ImageUri            string                     `json:"ImageUri,omitempty"`
	ImageConfigResponse *StoredImageConfigResponse `json:"ImageConfigResponse,omitempty"`
	LoggingConfig       *StoredLoggingConfig       `json:"LoggingConfig,omitempty"`

	// Code storage (in-memory, base64-encoded for mock)
	Code *FunctionCode `json:"Code,omitempty"`

	// Published versions
	PublishedVersions map[string]*StoredVersion `json:"PublishedVersions,omitempty"`
	NextVersionNumber int                       `json:"NextVersionNumber"`

	// Concurrency settings
	ReservedConcurrentExecutions *int32 `json:"ReservedConcurrentExecutions,omitempty"`

	// Resource policy
	Policy string `json:"Policy,omitempty"`
}

// StoredVersion represents a published function version
type StoredVersion struct {
	Version      string `json:"Version"`
	Description  string `json:"Description,omitempty"`
	CodeSha256   string `json:"CodeSha256"`
	CodeSize     int64  `json:"CodeSize"`
	RevisionId   string `json:"RevisionId"`
	FunctionArn  string `json:"FunctionArn"`
	LastModified string `json:"LastModified"`
}

// StoredAlias represents a function alias
type StoredAlias struct {
	Name            string              `json:"Name"`
	FunctionName    string              `json:"FunctionName"`
	FunctionVersion string              `json:"FunctionVersion"`
	Description     string              `json:"Description,omitempty"`
	AliasArn        string              `json:"AliasArn"`
	RevisionId      string              `json:"RevisionId"`
	RoutingConfig   *AliasRoutingConfig `json:"RoutingConfig,omitempty"`
}

// AliasRoutingConfig represents alias routing configuration
type AliasRoutingConfig struct {
	AdditionalVersionWeights map[string]float64 `json:"AdditionalVersionWeights,omitempty"`
}

// StoredFunctionUrl represents a function URL configuration
type StoredFunctionUrl struct {
	FunctionName     string      `json:"FunctionName"`
	FunctionArn      string      `json:"FunctionArn"`
	FunctionUrl      string      `json:"FunctionUrl"`
	AuthType         string      `json:"AuthType"`
	Cors             *CorsConfig `json:"Cors,omitempty"`
	InvokeMode       string      `json:"InvokeMode,omitempty"`
	CreationTime     string      `json:"CreationTime"`
	LastModifiedTime string      `json:"LastModifiedTime"`
}

// CorsConfig represents CORS configuration for function URL
type CorsConfig struct {
	AllowCredentials bool     `json:"AllowCredentials,omitempty"`
	AllowHeaders     []string `json:"AllowHeaders,omitempty"`
	AllowMethods     []string `json:"AllowMethods,omitempty"`
	AllowOrigins     []string `json:"AllowOrigins,omitempty"`
	ExposeHeaders    []string `json:"ExposeHeaders,omitempty"`
	MaxAge           int32    `json:"MaxAge,omitempty"`
}

// StoredEnvironment represents function environment variables for storage
type StoredEnvironment struct {
	Variables map[string]string `json:"Variables,omitempty"`
}

// StoredVpcConfig represents VPC configuration for storage
type StoredVpcConfig struct {
	SubnetIds        []string `json:"SubnetIds,omitempty"`
	SecurityGroupIds []string `json:"SecurityGroupIds,omitempty"`
	VpcId            string   `json:"VpcId,omitempty"`
}

// StoredDeadLetterConfig represents dead letter queue configuration for storage
type StoredDeadLetterConfig struct {
	TargetArn string `json:"TargetArn,omitempty"`
}

// StoredTracingConfig represents X-Ray tracing configuration for storage
type StoredTracingConfig struct {
	Mode string `json:"Mode,omitempty"`
}

// StoredEphemeralStorage represents ephemeral storage configuration for storage
type StoredEphemeralStorage struct {
	Size int32 `json:"Size"`
}

// StoredFileSystemConfig represents EFS file system configuration for storage
type StoredFileSystemConfig struct {
	Arn            string `json:"Arn"`
	LocalMountPath string `json:"LocalMountPath"`
}

// StoredImageConfigResponse represents container image configuration for storage
type StoredImageConfigResponse struct {
	ImageConfig *StoredImageConfig      `json:"ImageConfig,omitempty"`
	Error       *StoredImageConfigError `json:"Error,omitempty"`
}

// StoredImageConfig represents container image settings for storage
type StoredImageConfig struct {
	Command          []string `json:"Command,omitempty"`
	EntryPoint       []string `json:"EntryPoint,omitempty"`
	WorkingDirectory string   `json:"WorkingDirectory,omitempty"`
}

// StoredImageConfigError represents an error in image configuration for storage
type StoredImageConfigError struct {
	ErrorCode string `json:"ErrorCode,omitempty"`
	Message   string `json:"Message,omitempty"`
}

// StoredLoggingConfig represents CloudWatch logging configuration for storage
type StoredLoggingConfig struct {
	LogFormat           string `json:"LogFormat,omitempty"`
	ApplicationLogLevel string `json:"ApplicationLogLevel,omitempty"`
	SystemLogLevel      string `json:"SystemLogLevel,omitempty"`
	LogGroup            string `json:"LogGroup,omitempty"`
}

// FunctionCode represents the deployment package
type FunctionCode struct {
	S3Bucket        string `json:"S3Bucket,omitempty"`
	S3Key           string `json:"S3Key,omitempty"`
	S3ObjectVersion string `json:"S3ObjectVersion,omitempty"`
	ZipFile         string `json:"ZipFile,omitempty"` // Base64-encoded
	ImageUri        string `json:"ImageUri,omitempty"`
}

// CreateFunctionInput represents the input to CreateFunction
type CreateFunctionInput struct {
	FunctionName      string                   `json:"FunctionName"`
	Role              string                   `json:"Role"`
	Runtime           string                   `json:"Runtime,omitempty"`
	Handler           string                   `json:"Handler,omitempty"`
	Code              *FunctionCode            `json:"Code"`
	Description       string                   `json:"Description,omitempty"`
	Timeout           *int32                   `json:"Timeout,omitempty"`
	MemorySize        *int32                   `json:"MemorySize,omitempty"`
	Publish           bool                     `json:"Publish,omitempty"`
	VpcConfig         *StoredVpcConfig         `json:"VpcConfig,omitempty"`
	PackageType       string                   `json:"PackageType,omitempty"`
	DeadLetterConfig  *StoredDeadLetterConfig  `json:"DeadLetterConfig,omitempty"`
	Environment       *StoredEnvironment       `json:"Environment,omitempty"`
	KMSKeyArn         string                   `json:"KMSKeyArn,omitempty"`
	TracingConfig     *StoredTracingConfig     `json:"TracingConfig,omitempty"`
	Tags              map[string]string        `json:"Tags,omitempty"`
	Layers            []string                 `json:"Layers,omitempty"`
	FileSystemConfigs []StoredFileSystemConfig `json:"FileSystemConfigs,omitempty"`
	ImageConfig       *StoredImageConfig       `json:"ImageConfig,omitempty"`
	Architectures     []string                 `json:"Architectures,omitempty"`
	EphemeralStorage  *StoredEphemeralStorage  `json:"EphemeralStorage,omitempty"`
	LoggingConfig     *StoredLoggingConfig     `json:"LoggingConfig,omitempty"`
}

// UpdateFunctionCodeInput represents the input to UpdateFunctionCode
type UpdateFunctionCodeInput struct {
	S3Bucket        string   `json:"S3Bucket,omitempty"`
	S3Key           string   `json:"S3Key,omitempty"`
	S3ObjectVersion string   `json:"S3ObjectVersion,omitempty"`
	ZipFile         string   `json:"ZipFile,omitempty"` // Base64-encoded
	ImageUri        string   `json:"ImageUri,omitempty"`
	Publish         bool     `json:"Publish,omitempty"`
	DryRun          bool     `json:"DryRun,omitempty"`
	RevisionId      string   `json:"RevisionId,omitempty"`
	Architectures   []string `json:"Architectures,omitempty"`
}

// UpdateFunctionConfigurationInput represents the input to UpdateFunctionConfiguration
type UpdateFunctionConfigurationInput struct {
	Description       string                   `json:"Description,omitempty"`
	Handler           string                   `json:"Handler,omitempty"`
	MemorySize        *int32                   `json:"MemorySize,omitempty"`
	Role              string                   `json:"Role,omitempty"`
	Runtime           string                   `json:"Runtime,omitempty"`
	Timeout           *int32                   `json:"Timeout,omitempty"`
	VpcConfig         *StoredVpcConfig         `json:"VpcConfig,omitempty"`
	DeadLetterConfig  *StoredDeadLetterConfig  `json:"DeadLetterConfig,omitempty"`
	Environment       *StoredEnvironment       `json:"Environment,omitempty"`
	KMSKeyArn         string                   `json:"KMSKeyArn,omitempty"`
	TracingConfig     *StoredTracingConfig     `json:"TracingConfig,omitempty"`
	RevisionId        string                   `json:"RevisionId,omitempty"`
	Layers            []string                 `json:"Layers,omitempty"`
	FileSystemConfigs []StoredFileSystemConfig `json:"FileSystemConfigs,omitempty"`
	ImageConfig       *StoredImageConfig       `json:"ImageConfig,omitempty"`
	EphemeralStorage  *StoredEphemeralStorage  `json:"EphemeralStorage,omitempty"`
	LoggingConfig     *StoredLoggingConfig     `json:"LoggingConfig,omitempty"`
}

// InvokeInput represents the input to Invoke
type InvokeInput struct {
	InvocationType string `json:"InvocationType,omitempty"` // RequestResponse, Event, DryRun
	LogType        string `json:"LogType,omitempty"`        // None, Tail
	ClientContext  string `json:"ClientContext,omitempty"`  // Base64-encoded
	Qualifier      string `json:"Qualifier,omitempty"`
	Payload        []byte `json:"Payload,omitempty"`
}

// TagResourceInput represents the input to TagResource
type TagResourceInput struct {
	Tags map[string]string `json:"Tags"`
}

// CreateAliasInput represents the input to CreateAlias
type CreateAliasInput struct {
	Name            string              `json:"Name"`
	FunctionVersion string              `json:"FunctionVersion"`
	Description     string              `json:"Description,omitempty"`
	RoutingConfig   *AliasRoutingConfig `json:"RoutingConfig,omitempty"`
}

// UpdateAliasInput represents the input to UpdateAlias
type UpdateAliasInput struct {
	FunctionVersion string              `json:"FunctionVersion,omitempty"`
	Description     string              `json:"Description,omitempty"`
	RoutingConfig   *AliasRoutingConfig `json:"RoutingConfig,omitempty"`
	RevisionId      string              `json:"RevisionId,omitempty"`
}

// PublishVersionInput represents the input to PublishVersion
type PublishVersionInput struct {
	CodeSha256  string `json:"CodeSha256,omitempty"`
	Description string `json:"Description,omitempty"`
	RevisionId  string `json:"RevisionId,omitempty"`
}

// CreateFunctionUrlConfigInput represents the input to CreateFunctionUrlConfig
type CreateFunctionUrlConfigInput struct {
	AuthType   string      `json:"AuthType"`
	Cors       *CorsConfig `json:"Cors,omitempty"`
	InvokeMode string      `json:"InvokeMode,omitempty"`
	Qualifier  string      `json:"Qualifier,omitempty"`
}

// UpdateFunctionUrlConfigInput represents the input to UpdateFunctionUrlConfig
type UpdateFunctionUrlConfigInput struct {
	AuthType   string      `json:"AuthType,omitempty"`
	Cors       *CorsConfig `json:"Cors,omitempty"`
	InvokeMode string      `json:"InvokeMode,omitempty"`
}

// PutFunctionConcurrencyInput represents the input to PutFunctionConcurrency
type PutFunctionConcurrencyInput struct {
	ReservedConcurrentExecutions int32 `json:"ReservedConcurrentExecutions"`
}

// AddPermissionInput represents the input to AddPermission
type AddPermissionInput struct {
	StatementId         string `json:"StatementId"`
	Action              string `json:"Action"`
	Principal           string `json:"Principal"`
	SourceArn           string `json:"SourceArn,omitempty"`
	SourceAccount       string `json:"SourceAccount,omitempty"`
	EventSourceToken    string `json:"EventSourceToken,omitempty"`
	Qualifier           string `json:"Qualifier,omitempty"`
	RevisionId          string `json:"RevisionId,omitempty"`
	PrincipalOrgID      string `json:"PrincipalOrgID,omitempty"`
	FunctionUrlAuthType string `json:"FunctionUrlAuthType,omitempty"`
}

// StoredEventSourceMapping represents an event source mapping
type StoredEventSourceMapping struct {
	UUID                                string                               `json:"UUID"`
	EventSourceArn                      string                               `json:"EventSourceArn,omitempty"`
	FunctionArn                         string                               `json:"FunctionArn"`
	FunctionName                        string                               `json:"FunctionName"`
	State                               string                               `json:"State"`
	StateTransitionReason               string                               `json:"StateTransitionReason,omitempty"`
	LastModified                        string                               `json:"LastModified"`
	LastProcessingResult                string                               `json:"LastProcessingResult,omitempty"`
	BatchSize                           *int32                               `json:"BatchSize,omitempty"`
	MaximumBatchingWindowInSeconds      *int32                               `json:"MaximumBatchingWindowInSeconds,omitempty"`
	ParallelizationFactor               *int32                               `json:"ParallelizationFactor,omitempty"`
	StartingPosition                    string                               `json:"StartingPosition,omitempty"`
	StartingPositionTimestamp           string                               `json:"StartingPositionTimestamp,omitempty"`
	MaximumRecordAgeInSeconds           *int32                               `json:"MaximumRecordAgeInSeconds,omitempty"`
	BisectBatchOnFunctionError          *bool                                `json:"BisectBatchOnFunctionError,omitempty"`
	MaximumRetryAttempts                *int32                               `json:"MaximumRetryAttempts,omitempty"`
	TumblingWindowInSeconds             *int32                               `json:"TumblingWindowInSeconds,omitempty"`
	Enabled                             *bool                                `json:"Enabled,omitempty"`
	FilterCriteria                      *FilterCriteria                      `json:"FilterCriteria,omitempty"`
	DestinationConfig                   *DestinationConfig                   `json:"DestinationConfig,omitempty"`
	Queues                              []string                             `json:"Queues,omitempty"`
	SourceAccessConfigurations          []SourceAccessConfiguration          `json:"SourceAccessConfigurations,omitempty"`
	SelfManagedEventSource              *SelfManagedEventSource              `json:"SelfManagedEventSource,omitempty"`
	FunctionResponseTypes               []string                             `json:"FunctionResponseTypes,omitempty"`
	AmazonManagedKafkaEventSourceConfig *AmazonManagedKafkaEventSourceConfig `json:"AmazonManagedKafkaEventSourceConfig,omitempty"`
	SelfManagedKafkaEventSourceConfig   *SelfManagedKafkaEventSourceConfig   `json:"SelfManagedKafkaEventSourceConfig,omitempty"`
	ScalingConfig                       *ScalingConfig                       `json:"ScalingConfig,omitempty"`
	DocumentDBEventSourceConfig         *DocumentDBEventSourceConfig         `json:"DocumentDBEventSourceConfig,omitempty"`
}

// Note: FilterCriteria, Filter, DestinationConfig, OnSuccess, OnFailure,
// SourceAccessConfiguration, SelfManagedEventSource, ScalingConfig,
// AmazonManagedKafkaEventSourceConfig, SelfManagedKafkaEventSourceConfig,
// DocumentDBEventSourceConfig are defined in smithy_types.go

// CreateEventSourceMappingInput represents input to CreateEventSourceMapping
type CreateEventSourceMappingInput struct {
	EventSourceArn                      string                               `json:"EventSourceArn,omitempty"`
	FunctionName                        string                               `json:"FunctionName"`
	Enabled                             *bool                                `json:"Enabled,omitempty"`
	BatchSize                           *int32                               `json:"BatchSize,omitempty"`
	MaximumBatchingWindowInSeconds      *int32                               `json:"MaximumBatchingWindowInSeconds,omitempty"`
	ParallelizationFactor               *int32                               `json:"ParallelizationFactor,omitempty"`
	StartingPosition                    string                               `json:"StartingPosition,omitempty"`
	StartingPositionTimestamp           string                               `json:"StartingPositionTimestamp,omitempty"`
	MaximumRecordAgeInSeconds           *int32                               `json:"MaximumRecordAgeInSeconds,omitempty"`
	BisectBatchOnFunctionError          *bool                                `json:"BisectBatchOnFunctionError,omitempty"`
	MaximumRetryAttempts                *int32                               `json:"MaximumRetryAttempts,omitempty"`
	TumblingWindowInSeconds             *int32                               `json:"TumblingWindowInSeconds,omitempty"`
	FilterCriteria                      *FilterCriteria                      `json:"FilterCriteria,omitempty"`
	DestinationConfig                   *DestinationConfig                   `json:"DestinationConfig,omitempty"`
	Queues                              []string                             `json:"Queues,omitempty"`
	SourceAccessConfigurations          []SourceAccessConfiguration          `json:"SourceAccessConfigurations,omitempty"`
	SelfManagedEventSource              *SelfManagedEventSource              `json:"SelfManagedEventSource,omitempty"`
	FunctionResponseTypes               []string                             `json:"FunctionResponseTypes,omitempty"`
	AmazonManagedKafkaEventSourceConfig *AmazonManagedKafkaEventSourceConfig `json:"AmazonManagedKafkaEventSourceConfig,omitempty"`
	SelfManagedKafkaEventSourceConfig   *SelfManagedKafkaEventSourceConfig   `json:"SelfManagedKafkaEventSourceConfig,omitempty"`
	ScalingConfig                       *ScalingConfig                       `json:"ScalingConfig,omitempty"`
	DocumentDBEventSourceConfig         *DocumentDBEventSourceConfig         `json:"DocumentDBEventSourceConfig,omitempty"`
}

// UpdateEventSourceMappingInput represents input to UpdateEventSourceMapping
type UpdateEventSourceMappingInput struct {
	Enabled                        *bool                        `json:"Enabled,omitempty"`
	BatchSize                      *int32                       `json:"BatchSize,omitempty"`
	MaximumBatchingWindowInSeconds *int32                       `json:"MaximumBatchingWindowInSeconds,omitempty"`
	ParallelizationFactor          *int32                       `json:"ParallelizationFactor,omitempty"`
	MaximumRecordAgeInSeconds      *int32                       `json:"MaximumRecordAgeInSeconds,omitempty"`
	BisectBatchOnFunctionError     *bool                        `json:"BisectBatchOnFunctionError,omitempty"`
	MaximumRetryAttempts           *int32                       `json:"MaximumRetryAttempts,omitempty"`
	TumblingWindowInSeconds        *int32                       `json:"TumblingWindowInSeconds,omitempty"`
	FilterCriteria                 *FilterCriteria              `json:"FilterCriteria,omitempty"`
	DestinationConfig              *DestinationConfig           `json:"DestinationConfig,omitempty"`
	SourceAccessConfigurations     []SourceAccessConfiguration  `json:"SourceAccessConfigurations,omitempty"`
	FunctionName                   string                       `json:"FunctionName,omitempty"`
	FunctionResponseTypes          []string                     `json:"FunctionResponseTypes,omitempty"`
	ScalingConfig                  *ScalingConfig               `json:"ScalingConfig,omitempty"`
	DocumentDBEventSourceConfig    *DocumentDBEventSourceConfig `json:"DocumentDBEventSourceConfig,omitempty"`
}

// StoredLayer represents a Lambda layer
type StoredLayer struct {
	LayerName           string                        `json:"LayerName"`
	LatestVersionNumber int64                         `json:"LatestVersionNumber"`
	LayerArn            string                        `json:"LayerArn"`
	Versions            map[int64]*StoredLayerVersion `json:"Versions,omitempty"`
}

// StoredLayerVersion represents a specific layer version
type StoredLayerVersion struct {
	LayerVersionArn         string   `json:"LayerVersionArn"`
	Version                 int64    `json:"Version"`
	Description             string   `json:"Description,omitempty"`
	CreatedDate             string   `json:"CreatedDate"`
	CompatibleRuntimes      []string `json:"CompatibleRuntimes,omitempty"`
	CompatibleArchitectures []string `json:"CompatibleArchitectures,omitempty"`
	LicenseInfo             string   `json:"LicenseInfo,omitempty"`
	CodeSha256              string   `json:"CodeSha256"`
	CodeSize                int64    `json:"CodeSize"`
	// Code storage (in-memory, base64-encoded for mock)
	Content *LayerContent `json:"Content,omitempty"`
	// Resource policy
	Policy string `json:"Policy,omitempty"`
}

// LayerContent represents layer code content
type LayerContent struct {
	S3Bucket        string `json:"S3Bucket,omitempty"`
	S3Key           string `json:"S3Key,omitempty"`
	S3ObjectVersion string `json:"S3ObjectVersion,omitempty"`
	ZipFile         string `json:"ZipFile,omitempty"` // Base64-encoded
}

// PublishLayerVersionInput represents input to PublishLayerVersion
type PublishLayerVersionInput struct {
	LayerName               string        `json:"LayerName"`
	Description             string        `json:"Description,omitempty"`
	Content                 *LayerContent `json:"Content"`
	CompatibleRuntimes      []string      `json:"CompatibleRuntimes,omitempty"`
	CompatibleArchitectures []string      `json:"CompatibleArchitectures,omitempty"`
	LicenseInfo             string        `json:"LicenseInfo,omitempty"`
}

// AddLayerVersionPermissionInput represents input to AddLayerVersionPermission
type AddLayerVersionPermissionInput struct {
	StatementId    string `json:"StatementId"`
	Action         string `json:"Action"`
	Principal      string `json:"Principal"`
	OrganizationId string `json:"OrganizationId,omitempty"`
	RevisionId     string `json:"RevisionId,omitempty"`
}

// StoredProvisionedConcurrencyConfig represents a provisioned concurrency configuration
type StoredProvisionedConcurrencyConfig struct {
	FunctionArn                              string `json:"FunctionArn"`
	Qualifier                                string `json:"Qualifier"` // Version or alias name
	RequestedProvisionedConcurrentExecutions int32  `json:"RequestedProvisionedConcurrentExecutions"`
	AvailableProvisionedConcurrentExecutions int32  `json:"AvailableProvisionedConcurrentExecutions"`
	AllocatedProvisionedConcurrentExecutions int32  `json:"AllocatedProvisionedConcurrentExecutions"`
	Status                                   string `json:"Status"` // IN_PROGRESS, READY, FAILED
	StatusReason                             string `json:"StatusReason,omitempty"`
	LastModified                             string `json:"LastModified"`
}

// PutProvisionedConcurrencyConfigInput represents input to PutProvisionedConcurrencyConfig
type PutProvisionedConcurrencyConfigInput struct {
	ProvisionedConcurrentExecutions int32 `json:"ProvisionedConcurrentExecutions"`
}

// StoredEventInvokeConfig represents a function event invoke configuration
type StoredEventInvokeConfig struct {
	FunctionArn              string             `json:"FunctionArn"`
	Qualifier                string             `json:"Qualifier,omitempty"` // $LATEST, version, or alias
	MaximumEventAgeInSeconds *int32             `json:"MaximumEventAgeInSeconds,omitempty"`
	MaximumRetryAttempts     *int32             `json:"MaximumRetryAttempts,omitempty"`
	DestinationConfig        *DestinationConfig `json:"DestinationConfig,omitempty"`
	LastModified             string             `json:"LastModified"`
}

// PutFunctionEventInvokeConfigInput represents input to PutFunctionEventInvokeConfig
type PutFunctionEventInvokeConfigInput struct {
	MaximumEventAgeInSeconds *int32             `json:"MaximumEventAgeInSeconds,omitempty"`
	MaximumRetryAttempts     *int32             `json:"MaximumRetryAttempts,omitempty"`
	DestinationConfig        *DestinationConfig `json:"DestinationConfig,omitempty"`
}

// Note: AccountLimit and AccountUsage are defined in smithy_types.go

// Utility functions

// ptr returns a pointer to the given value
func ptr[T any](v T) *T {
	return &v
}

// now returns the current time in ISO 8601 format
func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}
