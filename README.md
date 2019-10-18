# ecsctl

## *WORK IN PROGRESS*

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
  - [ ] env
    - [ ] list
    - [ ] set
    - [ ] delete

scheduled-tasks
  - [ ] create
  - [ ] edit
  - [ ] delete
  - [ ] update

repositories
  - [x] create
  - [ ] edit
  - [x] delete
