# GROUP TOKEN PROVIDER

This repository contains the logic needed to execute a _Group Token Provider_ (GTP) node of Ermes. It contains:

1) Node config file
2) HTTP gateway
3) Redis connectivity

## Requirements

- Golang
- WasmEdge
- Redis

For this repository to work properly, you must also clone
the [Ermes common](https://github.com/ermes-thesis/ermes-common) with this structure.

```
projects
└───ermes-central-repo
│   │   ...
│   
└───ermes-common
    │   ...
```

## Deployment

To deploy this node, follow the instructions in the [Ermes deploy](https://github.com/ermes-thesis/ermes-deploy)
repository.

## Local development

To run this node locally, you need to satisfy the requirements listed above.
You can install them on your machine or use the provided DevContainer, which already satisfies all the requirements

### How to use the DevContainer

The DevContainer can be opened either using GoLand or VSCode with the DevContainers extension

#### How to use it with GoLand

1) Start docker
2) Clone this repo
3) Open the project. A popup should ask to reopen the project in the container. Click `reopen`.

Alternatively you can also start the container by navigating to the `devcontainer.json` file and use the `Services` tab
to manage existing containers.

> This DevContainer starts a Docker Compose. Apparently GoLand is not able yet to stop it when the IDE gets closed since
> the DevContainer support is still in Beta. You can stop the containers in the Docker Compose from the `Services` tab
> or
> directly either from terminal or Docker Desktop

#### How to use it with VSCode

1) Start docker
2) Clone this repo
3) Open the project. A popup should ask to reopen the project in the container. Click `reopen`

Alternatively you can also start the container navigating to the VSCode Command Palette by pressing `ctrl + shift + P` (
Or by selecting `View -> Command Palette`). At this point search and select the option:
`Dev Containers: Reopen in Container`.

### How to run

There are some alternatives:

- Either in GoLand or VSCode, to start the node you can open the given devcontainer, and run the main function in
  `main.go` file in any way allowed by the IDE, or from the terminal by typing `go run main`.
- In GoLand, create a new `Go Build` run configuration. Specify for this configuration:
    - Run kind: package
    - Package path: ermes-central-repo
    - Module: ermes-central-repo
    - Run on: Click on `Manage targets...`, and then set:
        - Server: Docker
        - Configuration files: The path to the docker-compose.yml file in the project
        - Service: Go
        - Click on `Additional Settings` and set `Project sources directory` to `/workspaces/ermes-central-repo`
    - Ready to Go!

### How to test

This project uses [mockery](https://vektra.github.io/mockery/latest/) to generate mocks for testing purposes.
To install mockery run

```bash
go install github.com/vektra/mockery/v3@v3.5.5
```

To generate the mocks open a terminal in the root folder of the project and run

```bash
$ mockery
```

You can run tests with

```bash
$ go test ./...
```
