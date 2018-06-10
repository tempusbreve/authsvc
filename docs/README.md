# Developer Notes

## Installing and running

[![GoDoc](https://godoc.org/breve.us/authsvc?status.svg)](http://godoc.org/breve.us/authsvc)

This project is written in Go, and follows standard Go development practices.

Vendoring of dependencies is done with the [`dep`](https://golang.github.io/dep/) tool.
Use `make deps` to ensure the dependencies.

There is a git pre-push hook for running some basic sanity tests before pushing.
Please run `make install-hooks` to install the hook.

Typical deployment of this service will be done with docker.
The Dockerfile and related files are in [`build/docker/authsvc/`](build/docker/authsvc/).
You can build the docker images with `make images`, and push them to the docker repo with `make push-images`


To run locally, use the `make run` command.
You can override default parameters in a file called `.local.config`.
For example you could use:

```sh
  #!/usr/bin/env bash

  export PORT=${PORT:-4884}
  export VERBOSE=${VERBOSE:-true}
  export DATA_HOME="${DATA_HOME:-.local/data}"
  export PUBLIC_HOME="${PUBLIC_HOME:-build/docker/authsvc/public}"
  # The following are randomly generated seed values; you should change them.
  export SEED_BLOCK="${SEED_BLOCK:-yTVbPPsuijznJ0G05+EgXpoBTuT64FwpHS/X2CThfow=}"
  export SEED_HASH="${SEED_HASH:-uB0qbJMdJZn2E0jdjC8gPnaxEa/tNLDKMtzb956BzaAg8XlqEsPLCNGi0jhTsa/TDwIYQxQIm8CyEcnU9E4bWw==}"
  export STORAGE_ENGINE="${STORAGE_ENGINE:-boltdb}"
  export CLIENTS="${CLIENTS:-.local/clients.json}"
```

You can define your registered OAuth2 clients in the file named with the `--clients` parameter or the envirionment variable `CLIENTS`:

```json
  [
    {
      "id": "mattermost-client",
      "name": "Mattermost",
      "endpoints": [
        "https://mattermost.example.com/signup/gitlab/complete"
      ]
    }
  ]
```

Where the `id` is the OAuth2 Client ID, and the `endpoints` are the acceptable redirect endpoints after being authorized.

## Intra Package Dependencies

I try to keep the package dependencies clean; the intra-package dependency graph is one way I keep track:

![dependencies](package_dependencies.svg)
