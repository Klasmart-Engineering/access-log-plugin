# Microgateway Custom Plugin Template

This repository contains a template project for creating custom KrakenD plugins. _(To be injected into Microgateway deployments)_

Any microgateway that wants to leverage a plugin should initialize a git-submodule in their `/plugins` directory with the respective plugin

## Steps

1. Initialize the go module system

```sh
go mod init yourmodulename
```

2. To build your plugin run - this specifically builds it as a go plugin _(required by krakenD)_

```sh
go build -buildmode=plugin -o yourplugin.so .
```

3. Go into the `Makefile` and update the `build` target to have the same command as step 2. _(namely making the name of
   the outputted `.so` file is correct)_

4. Submit a PR to the [Central Repository](https://github.com/KL-Engineering/central-microgateway-configuration/blob/main/plugins.json) adding your repository to the list of plugin repositories. This will enable us to automatically raise PR's when updates are made centrally

5. Go into `.github/workflows/integration-test.yaml` and update the information accordingly.

## Included Artifacts

### Dockerfile & KrakenD.json

When developing a plugin, it is quite likely that you will want to run it against a KrakenD instance to debug/verify behaviour. The `Dockerfile` and `krakend.json` files give you a barebones gateway that you can very quickly get up and running with. Please feel free to customise this as you like.

### Makefile

The [Makefile](Makefile) has a number of helper commands to get you up and running quickly. There are a number of
**protected** commands _(these are protected as they're used in build pipelines)_. If you think there is a reason to change
them specific to your use-case please contact the API management team to confirm.

For any targets that are not `PROTECTED` please free to edit and add to the `Makefile` as you wish

| Command       | Description                                                                                                                                       | Status      |
| ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- | ----------- |
| `make build`  | Builds the plugin in the intended format _<name>.so_                                                                                              | `PROTECTED` |
| `make login`  | Logs you into GitHub Container Registry - must have `GH_PAT` environment variable set in your shell - see [GitHub Personal Access Token](#GH-PAT) |             |
| `make b`      | Builds a microgateway with the configuration found in `krakend.json` - this is useful for local testing                                           |             |
| `make r`      | Runs the microgateway built by the step above                                                                                                     |             |
| `make br`     | Alias to run both the build and run commands in a single step                                                                                     |             |
| `make run-ci` | Command to run the POSTMAN integration tests in the CI environment                                                                                | `PROTECTED` |

### GitHub Actions

There are two GitHub actions included out of the box

1. `fan-out-updates` is an action that is used to automatically sync consumers of your plugin whenever you create a new release
2. `integration-test` is a basic action that will run integration tests using the artifacts in this repository. It assumes your tests are written as part of your Postman Collection
   - `Dockerfile`
   - `krakend.json`
   - Postman collections

## GH PAT

This refers to the GitHub Personal Access Token. This is a token that is specific to your own GitHub account and grants
access to GitHub resources. The resources we're primarily interested in here are the centrally managed docker images _(used to build krakend plugins & the base gateway image)_

This PAT will need to have permissions that at a minmum can read `repos` and `packages`. You will also
need to enable SSO on the GITHUB PAT.

### Authenticating manually

1. Authenticate to the [github container registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry#authenticating-to-the-container-registry)

```sh
export GH_PAT="your personal github access token here - must have read packages scope at a minimum"

echo $GH_PAT | docker login ghcr.io -u USERNAME --password-stdin
```

## Integration Tests

We opted to use Postman as an integration testing tool, primarily because it's reasonably language agonistic and
well-known in the industry. Similarly it's useful for running both locally and in CI.

Please make sure you commit both the postman `collection` and `environment` files - _environment file should be for
CI/local development_

You can then use the [reuseable workflow](https://github.com/KL-Engineering/central-microgateway-configuration/blob/main/.github/workflows/plugin-integration-test.yaml) to easily set up a CI pipeline for your plugin.

## KrakenD Docs

[Intro to plugins](https://www.krakend.io/docs/extending/introduction/)

There are four different types of plugins you can write:

1. **[HTTP server plugins](https://www.krakend.io/docs/extending/http-server-plugins/)** (or handler plugins): They belong to the router layer and let you modify the request before KrakenD starts processing it, or modify the final response. You can have several plugins at once.
2. **[HTTP client plugins](https://www.krakend.io/docs/extending/http-client-plugins/)** (or proxy client plugins): They belong to the proxy layer and let you change how KrakenD interacts (as a client) with your backend services. You can have one plugin per backend.
3. **[Response Modifier plugins](https://www.krakend.io/docs/extending/plugin-modifiers/)**: They are strictly modifiers and let you change the responses received from your backends.
4. **[Request Modififer plugins](https://www.krakend.io/docs/extending/plugin-modifiers/)**: They are strictly modifiers and let you change the requests sent to your backends.
