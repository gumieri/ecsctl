package cmd

var requiredSpec = "REQUIRED - "

var revision string
var revisionSpec = `AWS ECS cluster`

var follow bool
var followSpec = `keep process logging from CloudWatch Logs`

var exit bool
var exitSpec = `Send signal to process exit when dettached. Use only with --follow (-f) option`

var image string
var imageSpec = `AWS ECR image`

var yes bool
var yesSpec = `Answer yes`

var force bool
var forceSpec = `Force the command despite the errors`

var quiet bool
var quietSpec = `Omit messages to the standard output`

var editorCommand string
var editorCommandSpec = `Override default text editor`

var containerName string
var containerNameSpec = `Container name from Task Definition`

var tag string
var tagSpec = `AWS ECR image tag`

var cfgFile string
var cfgFileSpec = `config file (default is $HOME/.ecsctl.yaml)`

var profile string
var profileSpec = `AWS Profile`

var region string
var regionSpec = `AWS Region`

var clusterSpec = `AWS ECS cluster`

var toCluster string
var toClusterSpec = `AWS ECS cluster target where the copy will be created`

var spotPrice string
var spotPriceSpec = `Maximum price to pay for the spot instances (default is On-Demand price)`

var spotFleetRole string
var spotFleetRoleSpec = `IAM fleet role grants the Spot fleet permission launch and terminate instances on your behalf`

var instanceProfile string
var instanceProfileSpec = `An instance profile is a container for an IAM role and enables you to pass role information to Amazon EC2 Instance when the instance starts`

var targetCapacity int64
var targetCapacitySpec = `The capacity amout defined for the cluster`

var allocationStrategy string
var allocationStrategySpec = `The strategy for requesting instances across different Availability Zones.
Valid values:
'lowestPrice': Automatically select the cheapest Availability Zone and instance type (default)
'diversified': Balance Spot instances across selected Availability Zones and instance types`

var subnet string
var subnetSpec = `The Subnet (ID or tag 'Name') to launch the instance
E.g. subnet-123abcd`

var subnets []string
var subnetsSpec = `The Subnet (ID or tag 'Name') to launch the instances. Can be passed multiple times
E.g. --subnet subnet-123abcd -n subnet-456efgh -n lorem-ipsum`

var kernelID string
var kernelIDSpec = `The ID of the Kernel`

var amiID string
var amiIDSpec = `The ID of the AMI`

var monitoring bool
var monitoringSpec = `Enables monitoring for the instances`

var key string
var keySpec = `Key name to access the instances`

var ebs bool
var ebsSpec = `Use EBS optimized`

var securityGroups []string
var securityGroupsSpec = `Security Group (ID, name, or tag 'Name') for the instances. Can be passed multiple times
E.g. --security-group sg-123abcd -g sg-456efgh -g lorem-ipsum`

var instanceType string
var instanceTypeSpec = `Type of instance to be launched
E.g. m4.large`

var instanceTypes []string
var instanceTypesSpec = `Type of instance to be used by the Spot Fleet. Can be passed multiple times
It's possible to change the units provided (target capacity) by a specific instance type adding a number after a colon (:) (default 1)
E.g. --instance-type m4.large:2 -i c4.large:2 -i t3.medium`

var credit string
var creditSpec = `The credit option for CPU usage of a T2 (default 'standard') or T3 (default 'unlimited') instance
Valid values: 'standard', 'unlimited'`

var minimum int64
var minimumSpec = `The minimum number of instances to launch
If you specify a minimum that is more instances than Amazon EC2 can launch in the target Availability Zone, Amazon EC2 launches no instances`

var maximum int64
var maximumSpec = `The maximum number of instances to launch
If you specify more instances than Amazon EC2 can launch in the target Availability Zone, Amazon EC2 launches the largest possible number of instances above MinCount`

var tags []string
var tagsSpec = `Tag to Spot Fleet instances as 'key=value'. Can be passed multiple times
E.g. --tag Name=sample -t Project=sample -t Lorem=Ipsum`

var platform string
var platformSpec = `The base operating system for the cluster
Valid values: 'linux', 'windows', 'windows-2016', 'windows-2019'`

var status string
var statusSpec = `Task definition Status
Valid values: 'all', 'active', 'inactive'`

var listARN bool
var listARNSpec = `List resources ARNs`
