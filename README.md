# ecsctl *WORK IN PROGRESS*

## Installation

All builds are available at the [releases](https://github.com/gumieri/ecsctl/releases) page. The program is a single executable and to install it is simply placing it into a directory configured in the `PATH` of the system.

A simple way to install, if you are using Mac OS or Linux, is to run the commands bellow with elevated permissions (like as `root`):
```bash
curl -L https://github.com/gumieri/ecsctl/releases/latest/download/ecsctl-`uname -s`-`uname -m` -o /usr/local/bin/ecsctl
chmod +x /usr/local/bin/ecsctl
```

For modern Linux systems, it would be recomended to place the executable at `~/.local/bin`, the command for it does not require elevated permission:
```bash
mkdir -p ~/.local/bin
curl -L https://github.com/gumieri/ecsctl/releases/latest/download/ecsctl-`uname -s`-`uname -m` -o ~/.local/bin/ecsctl
chmod +x ~/.local/bin/ecsctl
```

To upgrade program to the latest version available, just run:
```bash
ecsctl upgrade
```

## Configurations
`ecsctl` will look for configurations at `$XDG_CONFIG_HOME/ecsctl/config.yaml` or `~/.ecsctl/config.yaml`.
The file can be a JSON, TOML, YAML, HCL or envfile. Any configuration can also be set as environment variable and it also respect the `AWS_*` environment variables or roles from AWS IAM (by extension of the AWS SDK).

A configuration example:
```yaml
region: "us-east-1"
cluster: "awesome"
```

## Commands

It is organized by subcommands / categories:
```
  clusters         Commands to manage clusters
  repositories     Commands to manage repositories (ECR)
  services         Commands to manage services
  task-definitions Commands to manage Task Definitions
```

### `clusters` commands
```
  add-instance   Add a new EC2 instance to informed cluster
  add-spot-fleet Add a new Spot Fleet to informed cluster
  create         Create empty clusters. If not specified a name, create a cluster named default
  delete         Delete clusters
  list           List clusters
```

### `repositories` commands
```
  create      Create repositories
  delete      Delete repositories
```

### `services` commands
```
  list        List services of specified cluster
  copy        Copy a service to another cluster
  deploy      Deploy a service
```

### `task-definitions` commands
```
  list        List Task Definition Families
  edit        Edit a Task Definition
  run         Run a Task Definition
```

## Roadmap

clusters
  - [x] create
  - [ ] edit (container insights and tags)
  - [x] delete
  - [x] list
  - [x] add-instance
  - [x] add-spot-fleet

services
  - [ ] create
  - [ ] edit
  - [ ] delete
  - [x] list
  - [x] copy
  - [ ] deploy

task-definitions
  - [ ] create
  - [x] edit
  - [ ] delete
  - [ ] deregister
  - [x] env
    - [x] list
    - [x] set
    - [x] delete

scheduled-tasks
  - [x] create
  - [x] edit
  - [ ] delete
  - [x] update

repositories
  - [x] create
  - [ ] edit
  - [x] delete
