package cmd

var linuxUserData = `
#!/bin/bash
echo ECS_CLUSTER={{.Cluster}} >> /etc/ecs/ecs.config;echo ECS_BACKEND_HOST= >> /etc/ecs/ecs.config;
`

var linuxSpotFleetUserData = `
#!/bin/bash
echo ECS_CLUSTER={{.Cluster}} >> /etc/ecs/ecs.config
echo ECS_BACKEND_HOST= >> /etc/ecs/ecs.config
export PATH=/usr/local/bin:$PATH
yum -y install jq
easy_install pip
pip install awscli
aws configure set default.region {{.Region}}
cat <<EOF > /etc/init/spot-instance-termination-notice-handler.conf
description "Start spot instance termination handler monitoring script"
author "Amazon Web Services"
start on started ecs
script
echo \$\$ > /var/run/spot-instance-termination-notice-handler.pid
exec /usr/local/bin/spot-instance-termination-notice-handler.sh
end script
pre-start script
logger "[spot-instance-termination-notice-handler.sh]: spot instance termination
notice handler started"
end script
EOF
cat <<EOF > /usr/local/bin/spot-instance-termination-notice-handler.sh
#!/bin/bash
while sleep 5; do
	if [ -z \$(curl -Isf http://169.254.169.254/latest/meta-data/spot/termination-time)]; then
		/bin/false
	else
		logger "[spot-instance-termination-notice-handler.sh]: spot instance termination notice detected"
		STATUS=DRAINING
		ECS_CLUSTER=\$(curl -s http://localhost:51678/v1/metadata | jq .Cluster | tr -d \")
		CONTAINER_INSTANCE=\$(curl -s http://localhost:51678/v1/metadata | jq .ContainerInstanceArn | tr -d \")
		logger "[spot-instance-termination-notice-handler.sh]: putting instance in state \$STATUS"

		/usr/local/bin/aws  ecs update-container-instances-state --cluster \$ECS_CLUSTER --container-instances \$CONTAINER_INSTANCE --status \$STATUS

		logger "[spot-instance-termination-notice-handler.sh]: putting myself to sleep..."
		sleep 120 # exit loop as instance expires in 120 secs after terminating notification
	fi
done
EOF
chmod +x /usr/local/bin/spot-instance-termination-notice-handler.sh
`

var windowsUserData = `<powershell> Initialize-ECSAgent -Cluster {{.Cluster}} -EnableTaskIAMRole -LoggingDrivers '["json-file","awslogs"]' </powershell>`

var baseWindowsUserData = `
<powershell>
# Set up directories the agent uses
New-Item -Type directory -Path ${env:ProgramFiles}\Amazon\ECS -Force
New-Item -Type directory -Path ${env:ProgramData}\Amazon\ECS -Force
New-Item -Type directory -Path ${env:ProgramData}\Amazon\ECS\data -Force
# Set up configuration
$ecsExeDir = "${env:ProgramFiles}\Amazon\ECS"
[Environment]::SetEnvironmentVariable("ECS_CLUSTER", "{{.Cluster}}", "Machine")
[Environment]::SetEnvironmentVariable("ECS_LOGFILE", "${env:ProgramData}\Amazon\ECS\log\ecs-agent.log", "Machine")
[Environment]::SetEnvironmentVariable("ECS_DATADIR", "${env:ProgramData}\Amazon\ECS\data", "Machine")
# Download the agent
$agentVersion = "latest"
$agentZipUri = "https://s3.amazonaws.com/amazon-ecs-agent/ecs-agent-windows-$agentVersion.zip"
$zipFile = "${env:TEMP}\ecs-agent.zip"
Invoke-RestMethod -OutFile $zipFile -Uri $agentZipUri
# Put the executables in the executable directory.
Expand-Archive -Path $zipFile -DestinationPath $ecsExeDir -Force
Set-Location ${ecsExeDir}
# Set $EnableTaskIAMRoles to $true to enable task IAM roles
# Note that enabling IAM roles will make port 80 unavailable for tasks.
[bool]$EnableTaskIAMRoles = $false
if (${EnableTaskIAMRoles}) {
  $HostSetupScript = Invoke-WebRequest https://raw.githubusercontent.com/aws/amazon-ecs-agent/master/misc/windows-deploy/hostsetup.ps1
  Invoke-Expression $($HostSetupScript.Content)
}
# Install the agent service
New-Service -Name "AmazonECS" ` + "`" + `
        -BinaryPathName "$ecsExeDir\amazon-ecs-agent.exe -windows-service" ` + "`" + `
        -DisplayName "Amazon ECS" ` + "`" + `
        -Description "Amazon ECS service runs the Amazon ECS agent" ` + "`" + `
        -DependsOn Docker ` + "`" + `
        -StartupType Manual
sc.exe failure AmazonECS reset=300 actions=restart/5000/restart/30000/restart/60000
sc.exe failureflag AmazonECS 1
Start-Service AmazonECS
</powershell>
`
