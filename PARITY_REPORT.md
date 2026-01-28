# CloudMirror Parity Report

## Summary

| Metric | Value |
|--------|-------|
| **Services Analyzed** | 9 |
| **Total Operations** | 1384 |
| **Supported** | 257 |
| **Missing** | 1126 |
| **Overall Coverage** | 18.6% |

---

## Service Coverage

| Service | Coverage | Supported | Missing | Deprecated |
|---------|----------|-----------|---------|------------|
| applicationautoscaling | 100.0% | 14 | 0 | 0 |
| dynamodb | 57.9% | 33 | 24 | 0 |
| ec2 | 6.3% | 47 | 702 | 0 |
| iam | 59.7% | 105 | 71 | 0 |
| lambda | 0.0% | 0 | 84 | 1 |
| rds | 5.5% | 9 | 154 | 0 |
| s3 | 21.7% | 23 | 83 | 0 |
| sqs | 69.6% | 16 | 7 | 0 |
| sts | 90.9% | 10 | 1 | 0 |


---


## Application Auto Scaling (applicationautoscaling)

**Coverage:** 100.0% (14/14)


### Supported (14)
- DeleteScalingPolicy
- DeleteScheduledAction
- DescribeScalableTargets
- DescribeScalingActivities
- DescribeScalingPolicies
- DescribeScheduledActions
- GetPredictiveScalingForecast
- ListTagsForResource
- PutScalingPolicy
- PutScheduledAction
- DeregisterScalableTarget
- RegisterScalableTarget
- TagResource
- UntagResource





---


## DynamoDB (dynamodb)

**Coverage:** 57.9% (33/57)


### Supported (33)
- CreateBackup
- CreateGlobalTable
- CreateTable
- DeleteBackup
- DeleteItem
- DeleteResourcePolicy
- DeleteTable
- DescribeBackup
- DescribeContinuousBackups
- DescribeContributorInsights
- DescribeEndpoints
- DescribeExport
- DescribeGlobalTable
- DescribeGlobalTableSettings
- DescribeImport
- DescribeKinesisStreamingDestination
- DescribeLimits
- DescribeTable
- DescribeTableReplicaAutoScaling
- DescribeTimeToLive
- GetItem
- GetResourcePolicy
- ListBackups
- ListTables
- ListTagsOfResource
- PutItem
- UpdateContinuousBackups
- UpdateTable
- UpdateTimeToLive
- Query
- Scan
- TagResource
- UntagResource




### Missing - High Priority
- ListContributorInsights
- ListExports
- ListGlobalTables
- ListImports
- PutResourcePolicy



---


## EC2 (ec2)

**Coverage:** 6.3% (47/749)


### Supported (47)
- CreateInternetGateway
- CreateKeyPair
- CreateLaunchTemplate
- CreateSecurityGroup
- CreateSubnet
- CreateTags
- CreateVolume
- CreateVpc
- DeleteInternetGateway
- DeleteKeyPair
- DeleteLaunchTemplate
- DeleteSecurityGroup
- DeleteSubnet
- DeleteTags
- DeleteVolume
- DeleteVpc
- DescribeImages
- DescribeInstanceAttribute
- DescribeInstanceCreditSpecifications
- DescribeInstanceTypes
- DescribeInstances
- DescribeInternetGateways
- DescribeKeyPairs
- DescribeLaunchTemplates
- DescribeNetworkAcls
- DescribeNetworkInterfaces
- DescribeRouteTables
- DescribeSecurityGroups
- DescribeSubnets
- DescribeTags
- DescribeVolumes
- DescribeVpcAttribute
- DescribeVpcs
- AttachInternetGateway
- AttachVolume
- DetachInternetGateway
- DetachVolume
- ModifyVpcAttribute
- StartInstances
- StopInstances
- AuthorizeSecurityGroupEgress
- AuthorizeSecurityGroupIngress
- ImportKeyPair
- RevokeSecurityGroupEgress
- RevokeSecurityGroupIngress
- RunInstances
- TerminateInstances




### Missing - High Priority
- CreateCapacityManagerDataExport
- CreateCapacityReservation
- CreateCapacityReservationBySplitting
- CreateCapacityReservationFleet
- CreateCarrierGateway
- CreateClientVpnEndpoint
- CreateClientVpnRoute
- CreateCoipCidr
- CreateCoipPool
- CreateCustomerGateway
- CreateDefaultSubnet
- CreateDefaultVpc
- CreateDelegateMacVolumeOwnershipTask
- CreateDhcpOptions
- CreateEgressOnlyInternetGateway
- CreateFleet
- CreateFlowLogs
- CreateFpgaImage
- CreateImage
- CreateImageUsageReport
- CreateInstanceConnectEndpoint
- CreateInstanceEventWindow
- CreateInstanceExportTask
- CreateInterruptibleCapacityReservationAllocation
- CreateIpam
- CreateIpamExternalResourceVerificationToken
- CreateIpamPolicy
- CreateIpamPool
- CreateIpamPrefixListResolver
- CreateIpamPrefixListResolverTarget
- CreateIpamResourceDiscovery
- CreateIpamScope
- CreateLaunchTemplateVersion
- CreateLocalGatewayRoute
- CreateLocalGatewayRouteTable
- CreateLocalGatewayRouteTableVirtualInterfaceGroupAssociation
- CreateLocalGatewayRouteTableVpcAssociation
- CreateLocalGatewayVirtualInterface
- CreateLocalGatewayVirtualInterfaceGroup
- CreateMacSystemIntegrityProtectionModificationTask
- CreateManagedPrefixList
- CreateNatGateway
- CreateNetworkAcl
- CreateNetworkAclEntry
- CreateNetworkInsightsAccessScope
- CreateNetworkInsightsPath
- CreateNetworkInterface
- CreateNetworkInterfacePermission
- CreatePlacementGroup
- CreatePublicIpv4Pool
- CreateReplaceRootVolumeTask
- CreateReservedInstancesListing
- CreateRestoreImageTask
- CreateRoute
- CreateRouteServer
- CreateRouteServerEndpoint
- CreateRouteServerPeer
- CreateRouteTable
- CreateSnapshot
- CreateSnapshots
- CreateSpotDatafeedSubscription
- CreateStoreImageTask
- CreateSubnetCidrReservation
- CreateTrafficMirrorFilter
- CreateTrafficMirrorFilterRule
- CreateTrafficMirrorSession
- CreateTrafficMirrorTarget
- CreateTransitGateway
- CreateTransitGatewayConnect
- CreateTransitGatewayConnectPeer
- CreateTransitGatewayMeteringPolicy
- CreateTransitGatewayMeteringPolicyEntry
- CreateTransitGatewayMulticastDomain
- CreateTransitGatewayPeeringAttachment
- CreateTransitGatewayPolicyTable
- CreateTransitGatewayPrefixListReference
- CreateTransitGatewayRoute
- CreateTransitGatewayRouteTable
- CreateTransitGatewayRouteTableAnnouncement
- CreateTransitGatewayVpcAttachment
- CreateVerifiedAccessEndpoint
- CreateVerifiedAccessGroup
- CreateVerifiedAccessInstance
- CreateVerifiedAccessTrustProvider
- CreateVpcBlockPublicAccessExclusion
- CreateVpcEncryptionControl
- CreateVpcEndpoint
- CreateVpcEndpointConnectionNotification
- CreateVpcEndpointServiceConfiguration
- CreateVpcPeeringConnection
- CreateVpnConcentrator
- CreateVpnConnection
- CreateVpnConnectionRoute
- CreateVpnGateway
- DeleteCapacityManagerDataExport
- DeleteCarrierGateway
- DeleteClientVpnEndpoint
- DeleteClientVpnRoute
- DeleteCoipCidr
- DeleteCoipPool
- DeleteCustomerGateway
- DeleteDhcpOptions
- DeleteEgressOnlyInternetGateway
- DeleteFleets
- DeleteFlowLogs
- DeleteFpgaImage
- DeleteImageUsageReport
- DeleteInstanceConnectEndpoint
- DeleteInstanceEventWindow
- DeleteIpam
- DeleteIpamExternalResourceVerificationToken
- DeleteIpamPolicy
- DeleteIpamPool
- DeleteIpamPrefixListResolver
- DeleteIpamPrefixListResolverTarget
- DeleteIpamResourceDiscovery
- DeleteIpamScope
- DeleteLaunchTemplateVersions
- DeleteLocalGatewayRoute
- DeleteLocalGatewayRouteTable
- DeleteLocalGatewayRouteTableVirtualInterfaceGroupAssociation
- DeleteLocalGatewayRouteTableVpcAssociation
- DeleteLocalGatewayVirtualInterface
- DeleteLocalGatewayVirtualInterfaceGroup
- DeleteManagedPrefixList
- DeleteNatGateway
- DeleteNetworkAcl
- DeleteNetworkAclEntry
- DeleteNetworkInsightsAccessScope
- DeleteNetworkInsightsAccessScopeAnalysis
- DeleteNetworkInsightsAnalysis
- DeleteNetworkInsightsPath
- DeleteNetworkInterface
- DeleteNetworkInterfacePermission
- DeletePlacementGroup
- DeletePublicIpv4Pool
- DeleteQueuedReservedInstances
- DeleteRoute
- DeleteRouteServer
- DeleteRouteServerEndpoint
- DeleteRouteServerPeer
- DeleteRouteTable
- DeleteSnapshot
- DeleteSpotDatafeedSubscription
- DeleteSubnetCidrReservation
- DeleteTrafficMirrorFilter
- DeleteTrafficMirrorFilterRule
- DeleteTrafficMirrorSession
- DeleteTrafficMirrorTarget
- DeleteTransitGateway
- DeleteTransitGatewayConnect
- DeleteTransitGatewayConnectPeer
- DeleteTransitGatewayMeteringPolicy
- DeleteTransitGatewayMeteringPolicyEntry
- DeleteTransitGatewayMulticastDomain
- DeleteTransitGatewayPeeringAttachment
- DeleteTransitGatewayPolicyTable
- DeleteTransitGatewayPrefixListReference
- DeleteTransitGatewayRoute
- DeleteTransitGatewayRouteTable
- DeleteTransitGatewayRouteTableAnnouncement
- DeleteTransitGatewayVpcAttachment
- DeleteVerifiedAccessEndpoint
- DeleteVerifiedAccessGroup
- DeleteVerifiedAccessInstance
- DeleteVerifiedAccessTrustProvider
- DeleteVpcBlockPublicAccessExclusion
- DeleteVpcEncryptionControl
- DeleteVpcEndpointConnectionNotifications
- DeleteVpcEndpointServiceConfigurations
- DeleteVpcEndpoints
- DeleteVpcPeeringConnection
- DeleteVpnConcentrator
- DeleteVpnConnection
- DeleteVpnConnectionRoute
- DeleteVpnGateway
- DescribeAccountAttributes
- DescribeAddressTransfers
- DescribeAddresses
- DescribeAddressesAttribute
- DescribeAggregateIdFormat
- DescribeAvailabilityZones
- DescribeAwsNetworkPerformanceMetricSubscriptions
- DescribeBundleTasks
- DescribeByoipCidrs
- DescribeCapacityBlockExtensionHistory
- DescribeCapacityBlockExtensionOfferings
- DescribeCapacityBlockOfferings
- DescribeCapacityBlockStatus
- DescribeCapacityBlocks
- DescribeCapacityManagerDataExports
- DescribeCapacityReservationBillingRequests
- DescribeCapacityReservationFleets
- DescribeCapacityReservationTopology
- DescribeCapacityReservations
- DescribeCarrierGateways
- DescribeClassicLinkInstances
- DescribeClientVpnAuthorizationRules
- DescribeClientVpnConnections
- DescribeClientVpnEndpoints
- DescribeClientVpnRoutes
- DescribeClientVpnTargetNetworks
- DescribeCoipPools
- DescribeConversionTasks
- DescribeCustomerGateways
- DescribeDeclarativePoliciesReports
- DescribeDhcpOptions
- DescribeEgressOnlyInternetGateways
- DescribeElasticGpus
- DescribeExportImageTasks
- DescribeExportTasks
- DescribeFastLaunchImages
- DescribeFastSnapshotRestores
- DescribeFleetHistory
- DescribeFleetInstances
- DescribeFleets
- DescribeFlowLogs
- DescribeFpgaImageAttribute
- DescribeFpgaImages
- DescribeHostReservationOfferings
- DescribeHostReservations
- DescribeHosts
- DescribeIamInstanceProfileAssociations
- DescribeIdFormat
- DescribeIdentityIdFormat
- DescribeImageAttribute
- DescribeImageReferences
- DescribeImageUsageReportEntries
- DescribeImageUsageReports
- DescribeImportImageTasks
- DescribeImportSnapshotTasks
- DescribeInstanceConnectEndpoints
- DescribeInstanceEventNotificationAttributes
- DescribeInstanceEventWindows
- DescribeInstanceImageMetadata
- DescribeInstanceSqlHaHistoryStates
- DescribeInstanceSqlHaStates
- DescribeInstanceStatus
- DescribeInstanceTopology
- DescribeInstanceTypeOfferings
- DescribeIpamByoasn
- DescribeIpamExternalResourceVerificationTokens
- DescribeIpamPolicies
- DescribeIpamPools
- DescribeIpamPrefixListResolverTargets
- DescribeIpamPrefixListResolvers
- DescribeIpamResourceDiscoveries
- DescribeIpamResourceDiscoveryAssociations
- DescribeIpamScopes
- DescribeIpams
- DescribeIpv6Pools
- DescribeLaunchTemplateVersions
- DescribeLocalGatewayRouteTableVirtualInterfaceGroupAssociations
- DescribeLocalGatewayRouteTableVpcAssociations
- DescribeLocalGatewayRouteTables
- DescribeLocalGatewayVirtualInterfaceGroups
- DescribeLocalGatewayVirtualInterfaces
- DescribeLocalGateways
- DescribeLockedSnapshots
- DescribeMacHosts
- DescribeMacModificationTasks
- DescribeManagedPrefixLists
- DescribeMovingAddresses
- DescribeNatGateways
- DescribeNetworkInsightsAccessScopeAnalyses
- DescribeNetworkInsightsAccessScopes
- DescribeNetworkInsightsAnalyses
- DescribeNetworkInsightsPaths
- DescribeNetworkInterfaceAttribute
- DescribeNetworkInterfacePermissions
- DescribeOutpostLags
- DescribePlacementGroups
- DescribePrefixLists
- DescribePrincipalIdFormat
- DescribePublicIpv4Pools
- DescribeRegions
- DescribeReplaceRootVolumeTasks
- DescribeReservedInstances
- DescribeReservedInstancesListings
- DescribeReservedInstancesModifications
- DescribeReservedInstancesOfferings
- DescribeRouteServerEndpoints
- DescribeRouteServerPeers
- DescribeRouteServers
- DescribeScheduledInstanceAvailability
- DescribeScheduledInstances
- DescribeSecurityGroupReferences
- DescribeSecurityGroupRules
- DescribeSecurityGroupVpcAssociations
- DescribeServiceLinkVirtualInterfaces
- DescribeSnapshotAttribute
- DescribeSnapshotTierStatus
- DescribeSnapshots
- DescribeSpotDatafeedSubscription
- DescribeSpotFleetInstances
- DescribeSpotFleetRequestHistory
- DescribeSpotFleetRequests
- DescribeSpotInstanceRequests
- DescribeSpotPriceHistory
- DescribeStaleSecurityGroups
- DescribeStoreImageTasks
- DescribeTrafficMirrorFilterRules
- DescribeTrafficMirrorFilters
- DescribeTrafficMirrorSessions
- DescribeTrafficMirrorTargets
- DescribeTransitGatewayAttachments
- DescribeTransitGatewayConnectPeers
- DescribeTransitGatewayConnects
- DescribeTransitGatewayMeteringPolicies
- DescribeTransitGatewayMulticastDomains
- DescribeTransitGatewayPeeringAttachments
- DescribeTransitGatewayPolicyTables
- DescribeTransitGatewayRouteTableAnnouncements
- DescribeTransitGatewayRouteTables
- DescribeTransitGatewayVpcAttachments
- DescribeTransitGateways
- DescribeTrunkInterfaceAssociations
- DescribeVerifiedAccessEndpoints
- DescribeVerifiedAccessGroups
- DescribeVerifiedAccessInstanceLoggingConfigurations
- DescribeVerifiedAccessInstances
- DescribeVerifiedAccessTrustProviders
- DescribeVolumeAttribute
- DescribeVolumeStatus
- DescribeVolumesModifications
- DescribeVpcBlockPublicAccessExclusions
- DescribeVpcBlockPublicAccessOptions
- DescribeVpcClassicLink
- DescribeVpcClassicLinkDnsSupport
- DescribeVpcEncryptionControls
- DescribeVpcEndpointAssociations
- DescribeVpcEndpointConnectionNotifications
- DescribeVpcEndpointConnections
- DescribeVpcEndpointServiceConfigurations
- DescribeVpcEndpointServicePermissions
- DescribeVpcEndpointServices
- DescribeVpcEndpoints
- DescribeVpcPeeringConnections
- DescribeVpnConcentrators
- DescribeVpnConnections
- DescribeVpnGateways
- GetActiveVpnTunnelStatus
- GetAllowedImagesSettings
- GetAssociatedEnclaveCertificateIamRoles
- GetAssociatedIpv6PoolCidrs
- GetAwsNetworkPerformanceData
- GetCapacityManagerAttributes
- GetCapacityManagerMetricData
- GetCapacityManagerMetricDimensions
- GetCapacityReservationUsage
- GetCoipPoolUsage
- GetConsoleOutput
- GetConsoleScreenshot
- GetDeclarativePoliciesReportSummary
- GetDefaultCreditSpecification
- GetEbsDefaultKmsKeyId
- GetEbsEncryptionByDefault
- GetEnabledIpamPolicy
- GetFlowLogsIntegrationTemplate
- GetGroupsForCapacityReservation
- GetHostReservationPurchasePreview
- GetImageAncestry
- GetImageBlockPublicAccessState
- GetInstanceMetadataDefaults
- GetInstanceTpmEkPub
- GetInstanceTypesFromInstanceRequirements
- GetInstanceUefiData
- GetIpamAddressHistory
- GetIpamDiscoveredAccounts
- GetIpamDiscoveredPublicAddresses
- GetIpamDiscoveredResourceCidrs
- GetIpamPolicyAllocationRules
- GetIpamPolicyOrganizationTargets
- GetIpamPoolAllocations
- GetIpamPoolCidrs
- GetIpamPrefixListResolverRules
- GetIpamPrefixListResolverVersionEntries
- GetIpamPrefixListResolverVersions
- GetIpamResourceCidrs
- GetLaunchTemplateData
- GetManagedPrefixListAssociations
- GetManagedPrefixListEntries
- GetNetworkInsightsAccessScopeAnalysisFindings
- GetNetworkInsightsAccessScopeContent
- GetPasswordData
- GetReservedInstancesExchangeQuote
- GetRouteServerAssociations
- GetRouteServerPropagations
- GetRouteServerRoutingDatabase
- GetSecurityGroupsForVpc
- GetSerialConsoleAccessStatus
- GetSnapshotBlockPublicAccessState
- GetSpotPlacementScores
- GetSubnetCidrReservations
- GetTransitGatewayAttachmentPropagations
- GetTransitGatewayMeteringPolicyEntries
- GetTransitGatewayMulticastDomainAssociations
- GetTransitGatewayPolicyTableAssociations
- GetTransitGatewayPolicyTableEntries
- GetTransitGatewayPrefixListReferences
- GetTransitGatewayRouteTableAssociations
- GetTransitGatewayRouteTablePropagations
- GetVerifiedAccessEndpointPolicy
- GetVerifiedAccessEndpointTargets
- GetVerifiedAccessGroupPolicy
- GetVpcResourcesBlockingEncryptionEnforcement
- GetVpnConnectionDeviceSampleConfiguration
- GetVpnConnectionDeviceTypes
- GetVpnTunnelReplacementStatus
- ListImagesInRecycleBin
- ListSnapshotsInRecycleBin
- ListVolumesInRecycleBin



---


## IAM (iam)

**Coverage:** 59.7% (105/176)


### Supported (105)
- CreateAccessKey
- CreateAccountAlias
- CreateGroup
- CreateInstanceProfile
- CreateLoginProfile
- CreateOpenIDConnectProvider
- CreatePolicy
- CreateRole
- CreateSAMLProvider
- CreateServiceLinkedRole
- CreateUser
- CreateVirtualMFADevice
- DeleteAccessKey
- DeleteAccountAlias
- DeleteAccountPasswordPolicy
- DeleteGroup
- DeleteGroupPolicy
- DeleteInstanceProfile
- DeleteLoginProfile
- DeleteOpenIDConnectProvider
- DeletePolicy
- DeleteRole
- DeleteRolePolicy
- DeleteSAMLProvider
- DeleteSSHPublicKey
- DeleteServerCertificate
- DeleteServiceLinkedRole
- DeleteUser
- DeleteUserPolicy
- DeleteVirtualMFADevice
- GetAccessKeyLastUsed
- GetAccountPasswordPolicy
- GetGroup
- GetGroupPolicy
- GetInstanceProfile
- GetLoginProfile
- GetOpenIDConnectProvider
- GetPolicy
- GetPolicyVersion
- GetRole
- GetRolePolicy
- GetSAMLProvider
- GetSSHPublicKey
- GetServerCertificate
- GetServiceLinkedRoleDeletionStatus
- GetUser
- GetUserPolicy
- ListAccessKeys
- ListAccountAliases
- ListAttachedGroupPolicies
- ListAttachedRolePolicies
- ListAttachedUserPolicies
- ListGroupPolicies
- ListGroups
- ListGroupsForUser
- ListInstanceProfiles
- ListInstanceProfilesForRole
- ListMFADevices
- ListOpenIDConnectProviders
- ListPolicies
- ListPolicyVersions
- ListRolePolicies
- ListRoleTags
- ListRoles
- ListSAMLProviders
- ListSSHPublicKeys
- ListServerCertificates
- ListUserPolicies
- ListUserTags
- ListUsers
- ListVirtualMFADevices
- PutGroupPolicy
- PutRolePolicy
- PutUserPolicy
- AddRoleToInstanceProfile
- AddUserToGroup
- AttachGroupPolicy
- AttachRolePolicy
- AttachUserPolicy
- DetachGroupPolicy
- DetachRolePolicy
- DetachUserPolicy
- EnableMFADevice
- RemoveRoleFromInstanceProfile
- RemoveUserFromGroup
- UpdateAccessKey
- UpdateAccountPasswordPolicy
- UpdateAssumeRolePolicy
- UpdateGroup
- UpdateLoginProfile
- UpdateOpenIDConnectProviderThumbprint
- UpdateRole
- UpdateRoleDescription
- UpdateSAMLProvider
- UpdateSSHPublicKey
- UpdateServerCertificate
- UpdateUser
- DeactivateMFADevice
- ResyncMFADevice
- TagRole
- TagUser
- UntagRole
- UntagUser
- UploadSSHPublicKey
- UploadServerCertificate




### Missing - High Priority
- CreateDelegationRequest
- CreatePolicyVersion
- CreateServiceSpecificCredential
- DeletePolicyVersion
- DeleteRolePermissionsBoundary
- DeleteServiceSpecificCredential
- DeleteSigningCertificate
- DeleteUserPermissionsBoundary
- GetAccountAuthorizationDetails
- GetAccountSummary
- GetContextKeysForCustomPolicy
- GetContextKeysForPrincipalPolicy
- GetCredentialReport
- GetDelegationRequest
- GetHumanReadableSummary
- GetMFADevice
- GetOrganizationsAccessReport
- GetOutboundWebIdentityFederationInfo
- GetServiceLastAccessedDetails
- GetServiceLastAccessedDetailsWithEntities
- ListDelegationRequests
- ListEntitiesForPolicy
- ListInstanceProfileTags
- ListMFADeviceTags
- ListOpenIDConnectProviderTags
- ListOrganizationsFeatures
- ListPoliciesGrantingServiceAccess
- ListPolicyTags
- ListSAMLProviderTags
- ListServerCertificateTags
- ListServiceSpecificCredentials
- ListSigningCertificates
- PutRolePermissionsBoundary
- PutUserPermissionsBoundary



---


## Lambda (lambda)

**Coverage:** 0.0% (0/84)




### Missing - High Priority
- CreateAlias
- CreateCapacityProvider
- CreateCodeSigningConfig
- CreateEventSourceMapping
- CreateFunction
- CreateFunctionUrlConfig
- DeleteAlias
- DeleteCapacityProvider
- DeleteCodeSigningConfig
- DeleteEventSourceMapping
- DeleteFunction
- DeleteFunctionCodeSigningConfig
- DeleteFunctionConcurrency
- DeleteFunctionEventInvokeConfig
- DeleteFunctionUrlConfig
- DeleteLayerVersion
- DeleteProvisionedConcurrencyConfig
- GetAccountSettings
- GetAlias
- GetCapacityProvider
- GetCodeSigningConfig
- GetDurableExecution
- GetDurableExecutionHistory
- GetDurableExecutionState
- GetEventSourceMapping
- GetFunction
- GetFunctionCodeSigningConfig
- GetFunctionConcurrency
- GetFunctionConfiguration
- GetFunctionEventInvokeConfig
- GetFunctionRecursionConfig
- GetFunctionScalingConfig
- GetFunctionUrlConfig
- GetLayerVersion
- GetLayerVersionByArn
- GetLayerVersionPolicy
- GetPolicy
- GetProvisionedConcurrencyConfig
- GetRuntimeManagementConfig
- ListAliases
- ListCapacityProviders
- ListCodeSigningConfigs
- ListDurableExecutionsByFunction
- ListEventSourceMappings
- ListFunctionEventInvokeConfigs
- ListFunctionUrlConfigs
- ListFunctionVersionsByCapacityProvider
- ListFunctions
- ListFunctionsByCodeSigningConfig
- ListLayerVersions
- ListLayers
- ListProvisionedConcurrencyConfigs
- ListTags
- ListVersionsByFunction
- PutFunctionCodeSigningConfig
- PutFunctionConcurrency
- PutFunctionEventInvokeConfig
- PutFunctionRecursionConfig
- PutFunctionScalingConfig
- PutProvisionedConcurrencyConfig
- PutRuntimeManagementConfig



---


## RDS (rds)

**Coverage:** 5.5% (9/163)


### Supported (9)
- CreateDBInstance
- DeleteDBInstance
- DescribeDBInstances
- ListTagsForResource
- AddTagsToResource
- ModifyDBInstance
- RebootDBInstance
- StartDBInstance
- StopDBInstance




### Missing - High Priority
- CreateBlueGreenDeployment
- CreateCustomDBEngineVersion
- CreateDBCluster
- CreateDBClusterEndpoint
- CreateDBClusterParameterGroup
- CreateDBClusterSnapshot
- CreateDBInstanceReadReplica
- CreateDBParameterGroup
- CreateDBProxy
- CreateDBProxyEndpoint
- CreateDBSecurityGroup
- CreateDBShardGroup
- CreateDBSnapshot
- CreateDBSubnetGroup
- CreateEventSubscription
- CreateGlobalCluster
- CreateIntegration
- CreateOptionGroup
- CreateTenantDatabase
- DeleteBlueGreenDeployment
- DeleteCustomDBEngineVersion
- DeleteDBCluster
- DeleteDBClusterAutomatedBackup
- DeleteDBClusterEndpoint
- DeleteDBClusterParameterGroup
- DeleteDBClusterSnapshot
- DeleteDBInstanceAutomatedBackup
- DeleteDBParameterGroup
- DeleteDBProxy
- DeleteDBProxyEndpoint
- DeleteDBSecurityGroup
- DeleteDBShardGroup
- DeleteDBSnapshot
- DeleteDBSubnetGroup
- DeleteEventSubscription
- DeleteGlobalCluster
- DeleteIntegration
- DeleteOptionGroup
- DeleteTenantDatabase
- DescribeAccountAttributes
- DescribeBlueGreenDeployments
- DescribeCertificates
- DescribeDBClusterAutomatedBackups
- DescribeDBClusterBacktracks
- DescribeDBClusterEndpoints
- DescribeDBClusterParameterGroups
- DescribeDBClusterParameters
- DescribeDBClusterSnapshotAttributes
- DescribeDBClusterSnapshots
- DescribeDBClusters
- DescribeDBEngineVersions
- DescribeDBInstanceAutomatedBackups
- DescribeDBLogFiles
- DescribeDBMajorEngineVersions
- DescribeDBParameterGroups
- DescribeDBParameters
- DescribeDBProxies
- DescribeDBProxyEndpoints
- DescribeDBProxyTargetGroups
- DescribeDBProxyTargets
- DescribeDBRecommendations
- DescribeDBSecurityGroups
- DescribeDBShardGroups
- DescribeDBSnapshotAttributes
- DescribeDBSnapshotTenantDatabases
- DescribeDBSnapshots
- DescribeDBSubnetGroups
- DescribeEngineDefaultClusterParameters
- DescribeEngineDefaultParameters
- DescribeEventCategories
- DescribeEventSubscriptions
- DescribeEvents
- DescribeExportTasks
- DescribeGlobalClusters
- DescribeIntegrations
- DescribeOptionGroupOptions
- DescribeOptionGroups
- DescribeOrderableDBInstanceOptions
- DescribePendingMaintenanceActions
- DescribeReservedDBInstances
- DescribeReservedDBInstancesOfferings
- DescribeSourceRegions
- DescribeTenantDatabases
- DescribeValidDBInstanceModifications



---


## S3 (s3)

**Coverage:** 21.7% (23/106)


### Supported (23)
- CreateBucket
- CreateBucketMetadataConfiguration
- CreateBucketMetadataTableConfiguration
- CreateMultipartUpload
- CreateSession
- DeleteBucket
- DeleteBucketPolicy
- DeletePublicAccessBlock
- GetBucketEncryption
- GetBucketLogging
- GetBucketPolicy
- GetBucketVersioning
- GetObject
- GetPublicAccessBlock
- ListBuckets
- ListObjectsV2
- PutBucketEncryption
- PutBucketLogging
- PutBucketPolicy
- PutBucketVersioning
- PutObject
- PutPublicAccessBlock
- HeadBucket




### Missing - High Priority
- DeleteBucketAnalyticsConfiguration
- DeleteBucketCors
- DeleteBucketEncryption
- DeleteBucketIntelligentTieringConfiguration
- DeleteBucketInventoryConfiguration
- DeleteBucketLifecycle
- DeleteBucketMetadataConfiguration
- DeleteBucketMetadataTableConfiguration
- DeleteBucketMetricsConfiguration
- DeleteBucketOwnershipControls
- DeleteBucketReplication
- DeleteBucketTagging
- DeleteBucketWebsite
- DeleteObject
- DeleteObjectTagging
- DeleteObjects
- GetBucketAbac
- GetBucketAccelerateConfiguration
- GetBucketAcl
- GetBucketAnalyticsConfiguration
- GetBucketCors
- GetBucketIntelligentTieringConfiguration
- GetBucketInventoryConfiguration
- GetBucketLifecycleConfiguration
- GetBucketLocation
- GetBucketMetadataConfiguration
- GetBucketMetadataTableConfiguration
- GetBucketMetricsConfiguration
- GetBucketNotificationConfiguration
- GetBucketOwnershipControls
- GetBucketPolicyStatus
- GetBucketReplication
- GetBucketRequestPayment
- GetBucketTagging
- GetBucketWebsite
- GetObjectAcl
- GetObjectAttributes
- GetObjectLegalHold
- GetObjectLockConfiguration
- GetObjectRetention
- GetObjectTagging
- GetObjectTorrent
- ListBucketAnalyticsConfigurations
- ListBucketIntelligentTieringConfigurations
- ListBucketInventoryConfigurations
- ListBucketMetricsConfigurations
- ListDirectoryBuckets
- ListMultipartUploads
- ListObjectVersions
- ListObjects
- ListParts
- PutBucketAbac
- PutBucketAccelerateConfiguration
- PutBucketAcl
- PutBucketAnalyticsConfiguration
- PutBucketCors
- PutBucketIntelligentTieringConfiguration
- PutBucketInventoryConfiguration
- PutBucketLifecycleConfiguration
- PutBucketMetricsConfiguration
- PutBucketNotificationConfiguration
- PutBucketOwnershipControls
- PutBucketReplication
- PutBucketRequestPayment
- PutBucketTagging
- PutBucketWebsite
- PutObjectAcl
- PutObjectLegalHold
- PutObjectLockConfiguration
- PutObjectRetention
- PutObjectTagging



---


## SQS (sqs)

**Coverage:** 69.6% (16/23)


### Supported (16)
- CreateQueue
- DeleteMessage
- DeleteMessageBatch
- DeleteQueue
- GetQueueAttributes
- GetQueueUrl
- ListQueueTags
- ListQueues
- ChangeMessageVisibility
- PurgeQueue
- ReceiveMessage
- SendMessage
- SendMessageBatch
- SetQueueAttributes
- TagQueue
- UntagQueue




### Missing - High Priority
- ListDeadLetterSourceQueues
- ListMessageMoveTasks



---


## STS (sts)

**Coverage:** 90.9% (10/11)


### Supported (10)
- GetAccessKeyInfo
- GetCallerIdentity
- GetDelegatedAccessToken
- GetFederationToken
- GetSessionToken
- AssumeRole
- AssumeRoleWithSAML
- AssumeRoleWithWebIdentity
- AssumeRoot
- DecodeAuthorizationMessage




### Missing - High Priority
- GetWebIdentityToken



---



_Generated by CloudMirror on 0001-01-01 00:00:00 UTC_
