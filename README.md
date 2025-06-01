# CENTRAL REPO
This repository contains the logic needed to execute a "central repository" node Ermes. It contains:
1) Node config file
2) HTTP gateway
3) Redis connectivity

## Requirements
- Golang
- WasmEdge
- Redis

For this repository to work properly, you must also clone
the [WASM pool library](https://github.com/ermes-thesis/ermes-wasm-test) with this structure.

```
projects
└───central-repo
│   │   ...
│   
└───ermes-wasm-test
    │   ...
```

It is recommended to use the provided DevContainer, which already satisfies all the requirements

## How to use the DevContainer
The DevContainer can be opened either using GoLand or VSCode with the DevContainers extension

### How to use it with GoLand
1) Start docker
2) Clone this repo
3) Open the project. A popup should ask to reopen the project in the container. Click `reopen`.

Alternatively you can also start the container by navigating to the `devcontainer.json` file and use the `Services` tab to manage existing containers.

> This DevContainer starts a Docker Compose. Apparently GoLand is not able yet to stop it when the IDE gets closed since the DevContainer support is still in Beta. You can stop the containers in the Docker Compose from the `Services` tab or directly from Docker Desktop

### How to use it with VSCode
1) Start docker
2) Clone this repo
3) Open the project. A popup should ask to reopen the project in the container. Click `reopen`

Alternatively you can also start the container navigating to the VSCode Command Palette by pressing `ctrl + shift + P` (Or by selecting `View -> Command Palette`). At this point search and select the option: `Dev Containers: Reopen in Container`.

## Node config file
This repo contains a `.env` file in the `.devcontainer` directory. This file contains several environment variables that will be read by the node once it starts and use them to configure JWT token encryption and the connection to the Redis DB

## HTTP Gateway interface
The HTTP gateway exposes the following interface:
- `http://<ip_address>:3000/group-access`
  - Can be called by the Ermes middleware to request or renew the group token associated to the user performing the HTTP request
- `http://<ip_address>:3000/add-to-group?group=<group_name>`
  - Can be called by a wasm function, executed within an Ermes node to add the user executing the function and performing the HTTP request to a group specified as query parameter
- `http://<ip_address>:3000/remove-from-group?group=<group_name>`
  - Can be called by a wasm function, executed within an Ermes node to remove the user executing the function and performing the HTTP request from a group specified as query parameter
- `http://<ip_address>:3000/create-group?group=<group_name>`
  - Can be called by a wasm function (or externally by an Admin), executed within an Ermes node to create a new existing group with the name specified as query parameter 
- `http://<ip_address>:3000/delete-group`
  - Can be called by a wasm function (or externally by an Admin), executed within an Ermes node to delete an existing group with the name specified as query parameter

## Redis
This node uses Redis as main DB storage. It is used to store the existing groups and the association between each registered user and their groups.

### Data format
The node stores the existing groups in a REDIS SET, with key: `existing_groups` and as value the list of all the existing groups at that moment in the Ermes architecture

It stores the associations user-groups in other REDIS SETs, that have as key the user's ID and as values all the groups associated to that specific user 

## How to run
Either in GoLand or VSCode, to start the node you can run the main function in `main.go` file in any way allowed by the IDE, or from the terminal by typing `go run main`.

