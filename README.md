# Microgateway access-log plugin

This is a Krakend plugin that pushes to Firehose when the gateway is used.

Any microgateway that wants to leverage a plugin should initialize a git-submodule in their `/plugins` directory with the respective plugin

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

## Generating a JWT

Since it's not the responsibility of this plugin to verify JWTs (this should be done by a plugin earlier in the pipeline in production Krakend instances), 
you can produce arbitrary JWTs to use as the Authorization header with the following JS:

```
const header = '{"alg": "HS256","typ": "JWT"}'

const body = '{"sub": "sub-blabla","name": "Somebody","iat": 123456,"subscription_id": "a9de93fc-2d13-44dd-9272-da7f8c17d155","android_id": "07ff00e4-c1a5-4683-9fcb-613a734d8d3f"}'

console.info(btoa(header).replace("=", "") + "." + btoa(body).replace("=", "") + "." + btoa("invalid signature").replace("=", ""))
```