# ecsctl

## *WORK IN PROGRESS*

## Commands

It is organized by subcommands / categories:
```
  clusters         commands to manage clusters
  services         commands to manage services
  task-definitions commands to manage Task Definitions
```

### `clusters` commands
```
  create         Create empty clusters
  add-instance   Add a new EC2 instance to informed cluster
  add-spot-fleet Add a new Spot Fleet to informed cluster
```

### `services` commands
```
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
  - [x] add-instance
  - [x] add-spot-fleet
  - [ ] delete

services
  - [x] copy
  - [ ] deploy
  - [ ] create

task-definitions
  - [x] edit
  - [ ] deregister
  - [ ] env
    - [ ] list
    - [ ] set
    - [ ] delete

scheduled-tasks
  - [ ] create
  - [ ] update

repositories
  - [ ] create
  - [ ] edit
  - [ ] delete
