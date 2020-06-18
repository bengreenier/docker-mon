# docker-mon

![CD](https://github.com/bengreenier/docker-mon/workflows/CD/badge.svg)
![GitHub go.mod Go version (branch)](https://img.shields.io/github/go-mod/go-version/bengreenier/docker-mon/master)
![Docker Pulls](https://img.shields.io/docker/pulls/bengreenier/mon)

Docker container monitor and micro-orchestrator 🩺📦

![Header image](./.github/header.png)

Docker-mon (or `mon` for short) was built to solve two problems, when working with containers outside of an orchestration platform (e.g. [kubernetes](https://kubernetes.io/) or [swarm](https://docs.docker.com/engine/swarm/)):

- Restart containers that are failing their [`HEALTHCHECK`](https://docs.docker.com/engine/reference/builder/#healthcheck)
- Remove un-needed containers and their resources when they're safely stopped

To do this, `mon` runs in it's own docker container and queries the docker daemon API to inspect the state of other containers. As such, `mon` requires the docker control socket be mounted with `-v /var/run/docker.sock:/var/run/docker.sock`. There's also some [additional metadata](#metadata) that can be attached to containers, to control what `mon` will do to them. 

## Getting Started 🚀

- ⏬ Get `mon`
```
# Grab the latest version from dockerhub
docker pull bengreenier/mon:latest
```
- 📝 Configure containers
```
# Start nginx, configuring it for cleanup
# mon will remove the container completely, if it stops with exit code '0'
# mon will restart the container, if it has a healthcheck, and it's failing
docker run --label mon.observe=1 mon.checks.cleanup=1 mon.checks.health=1 nginx:latest
```
- ✨ Start `mon`
```
# Run mon, forwarding the docker control socket
docker run -v /var/run/docker.sock:/var/run/docker.sock bengreenier/mon:latest
```

## Modes 📖

Detailed descriptions of the `mon` operation modes.

### Health Monitoring 👩‍⚕️

Health monitoring is an extension to [`HEALTHCHECK`](https://docs.docker.com/engine/reference/builder/#healthcheck) functionality, to restart containers that are failing. This was originally planned, but never landed in docker itself. There are some other great containers (like [autoheal](https://github.com/willfarrell/docker-autoheal)) that provide this functionality as well.

`mon` observes the container metadata, and if `State.Health.Status` is `Unhealthy`, it will restart the container.

### Cleanup Monitoring 🧼

Cleanup monitoring helps keep the host os from becoming cluttered with content from stopped containers. It will remove containers, links, and volumes that are no longer needed.

`mon` does this by observing the container metadata, and if `State.Running`, `state.Restarting`, are false, and `state.ExitCode` matches the expected value (default is `0`), it will remove the container. 

## Arguments 🙋‍♀️

`mon` supports some command-line arguments to control it's behavior. Here they are:


- `control` - Docker control socket. Default is `unix:///var/run/docker.sock`.
- `prefix` - Docker container prefix to limit observation to. Default is empty, meaning no prefix is required, all containers will be observed.
- `interval` - Interval to poll at (in ms). Default is `10000` (10s).
- `retries` - Max retry count for failed docker commands. Default is `10`.
- `quiet` - Only log when action is taken. Default is `false`.

### Environment Variables 🌍

The same [Arguments](#arguments) that are supported above, can be used as environment variables, prefixed with `MON_`. Here they are:

- `MON_CONTROL` - Docker control socket. Default is `unix:///var/run/docker.sock`.
- `MON_PREFIX` - Docker container prefix to limit observation to. Default is empty, meaning no prefix is required, all containers will be observed.
- `MON_INTERVAL` - Interval to poll at (in ms). Default is `10000` (10s).
- `MON_RETRIES` - Max retry count for failed docker commands. Default is `10`.
- `MON_QUIET` - Only log when action is taken. Default is `false`.

## Metadata 🧬

`mon` supports some additional metadata on containers, that inform it's actions. Here they are:

- `mon.observe` includes the container in mon observations, when set to `1`.
- `mon.checks.health` includes the container in [`HEALTHCHECK`](https://docs.docker.com/engine/reference/builder/#healthcheck) observations, when set to `1`.
- `mon.checks.health.timeout` overrides the expected restart interval (in ms), that a container has to restart. Default is `10000` (10ms).
- `mon.checks.cleanup` includes the container in cleanup observations, when set to `1`.
- `mon.checks.cleanup.code` overrides the expected exit code for the container, which if returned will lead to cleanup. Default is `0`.

## Contributing 👩‍💻

Thanks for your interest! To participate, you'll need [VSCode](https://code.visualstudio.com/), as development occurs in a [DevContainer](https://code.visualstudio.com/docs/remote/containers). Other than that, I don't have much advice at this point. We'll update this section as needed. 

This project follows the same guidelines as the [Microsoft Code Of Conduct](https://opensource.microsoft.com/codeofconduct/), but to escalate issues, please use GitHub Issues, as this project isn't affiliated directly with Microsoft, and issues shouldn't be raised through Microsoft's line of support.

### Cutting a Release

To create a release, create a tag locally, and push it to GitHub. Actions and Dockerhub will do the rest! ✨

```
git tag vx.x.x
git push --tags
```

### Generating Mocks

We use [mockgen](https://github.com/golang/mock) to generate mocks, that are then checked in. To re-generate them:

```
cd internal/app/mon && mockgen -destination ./mocks/dockerd.go . DockerAPI
```
