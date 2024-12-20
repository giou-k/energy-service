# Energy Service

Energy Service is a personal project to help me document my knowledge and sharpen my skills.

- [Getting started](#GettingStarted)
- [Bring k8 infra up](#Infra)

## Getting Started

Run `make all` to create the docker image.

#### Infra

In order to bring up the kubernetes environment you need `kind` installed.

After installing `kind`, do:

1. Create the cluster:
`make dev-up`

2. Load the docker image:
`make dev-load`

3. Apply the k8 configuration:
`make dev-apply`

#### Run it and more

* `make run` will start the server
* You can check `http://localhost:3011/debug/statsviz/` for nice debug graphs and `http://localhost:3011/debug/pprof/` 
for...not so nice debug stats