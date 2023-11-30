# Energy Service

Energy Service is a personal project to help me document my knowledge and sharpen my skills.

- [Bring k8 infra up](#Infra)

## Infra

In order to bring up the kubernetes environment you need `kind` installed.

After installing `kind`, do:

1. Create the cluster:
`make dev-up`

2. Load the docker image:
`make dev-load`

3. Apply the k8 configuration:
`make dev-apply`