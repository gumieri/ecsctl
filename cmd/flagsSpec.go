package cmd

var requiredSpec = "[REQUIRED] "

var revision string
var revisionSpec = `AWS ECS cluster`

var follow bool
var followSpec = `keep process logging from CloudWatch Logs`

var image string
var imageSpec = `AWS ECR image`

var editorCommand string
var editorCommandSpec = `Override default text editor`

var containerName string
var containerNameSpec = `Container name from Task Definition`

var repository string
var repositorySpec = `AWS ECR repository name`

var tag string
var tagSpec = `AWS ECR image tag`

var cfgFile string
var cfgFileSpec = `config file (default is $HOME/.ecsctl.yaml)`

var profile string
var profileSpec = `AWS Profile`

var region string
var regionSpec = `AWS Region`

var cluster string
var clusterSpec = `AWS ECS cluster`

var toCluster string
var toClusterSpec = `AWS ECS cluster target where the copy will be created`

var spotPrice string
var spotPriceSpec = `Top price to pay for the spot instances`

var spotFleetRole string
var spotFleetRoleSpec = `IAM fleet role grants the Spot fleet permission launch and terminate instances on your behalf`

var instanceRole string
var instanceRoleSpec = `An instance profile is a container for an IAM role and enables you to pass role information to Amazon EC2 Instance when the instance starts`

var targetCapacity int64
var targetCapacitySpec = `The capacity amout defined for the cluster`

var allocationStrategy string
var allocationStrategySpec = `
Valid values: 'diversified' or 'lowestPrice'`

var subnet string
var subnetSpec = `The Subnet ID to launch the instance`

var subnets string
var subnetsSpec = `The Subnets IDs to launch the instances (separeted by comma ',')`

var kernelID string
var kernelIDSpec = `The ID of the Kernel`

var monitoring bool
var monitoringSpec = `Enables monitoring for the instances`

var key string
var keySpec = `Key name to access the instances`

var ebs bool
var ebsSpec = `EBS optimized`

var securityGroups string
var securityGroupsSpec = `Security Groups for the instances (separeted by comma ',')`

var instanceType string
var instanceTypeSpec = `Type of instance to be launched`

var instanceTypes string
var instanceTypesSpec = `Types of instance to be used by the Spot Fleet (separeted by comma ',')`

var credit string
var creditSpec = `The credit option for CPU usage of a T2 (default 'standard') or T3 (default 'unlimited') instance
Valid values: 'standard' or 'unlimited'`

var minimum int64
var minimumSpec = `The minimum number of instances to launch
If you specify a minimum that is more instances than Amazon EC2 can launch in the target Availability Zone, Amazon EC2 launches no instances`

var maximum int64
var maximumSpec = `The maximum number of instances to launch
If you specify more instances than Amazon EC2 can launch in the target Availability Zone, Amazon EC2 launches the largest possible number of instances above MinCount`

var tags string
var tagsSpec = `Tags to Spot Fleet instances ('key=value' separeted by comma ',')
Ex: Name=sample,Project=sample,Lorem=Ipsum`
