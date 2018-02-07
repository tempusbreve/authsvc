# authsvc

Authentication Service


## Developing

This project is written in Go, and follows standard Go development practices.

Vendoring of dependencies is done with the [`dep`](https://golang.github.io/dep/) tool.

There is a git pre-push hook for running some basic sanity tests before pushing.
Please run `.bin/install-pre-push-hook` to install the hook.

Typical deployment of this service will be done with docker.
The Dockerfile and related files are in [`docker/api/`](docker/api/).
