# Microgateway access-log plugin

This is a [HTTP Server Plugin](https://www.krakend.io/docs/extending/http-server-plugins/) for Krakend that pushes access log entries to Firehose.  The entries contain:

 - HTTP Method
 - Relative Path
 - Time (Unix Epoch)
 - Subscription Id (taken from the Authorization header JWT)
 - Android Id (taken from the Authorization header JWT)
 - Product Name (from config, to identify this KL Product and group usage)

The entries are later made available to end users via the Portal in the form of a usage report.

## Using the plugin

Any microgateway that wants to leverage a plugin should initialize a git-submodule in their `/plugins` directory with the respective plugin

### Configuring

In your krakend.json, in the top level `extra_config`, under `plugin/http-server` add the following:

```
"name": [
   "access-log"
],
"access-log": {
   "product_name": "Test Product",
   "buffer_size": 1000,
   "firehose_batch_size": 500,
   "firehose_send_early_timeout_ms": 60000,
   "use_aws_default_credentials": false,
   "aws_secret_key_id": "123456",
   "aws_secret_key": "123456",
   "aws_region": "eu-west-2",
   "delivery_stream_name": "subscriptions-uk-apifactory-api-usage",
   "ignore_paths": [
      "/__health",
      "/__stats",
      "/ignored/*"
   ]
}
```

You can omit buffer_size, firehose_batch_size and firehouse_send_early_timeout_ms to use default values.  However, if you want to override these using environment variables [they must be present in krakend.json](https://www.krakend.io/docs/configuration/environment-vars/).

 - product_name: Your product name (e.g. `Content API`).
 - buffer_size: Size of the channel buffer used to enqueue access log entries to be batched.  This should be set to a value higher than the number of concurrent requests you expect to handle at peak traffic to avoid blocking the goroutine handling the request.
 - firehose_batch_size: Access log entries are read from the channel into a batch, once the batch reaches this size it is pushed to Firehose.  The max batch size for Firehose is 500 with a total size of 4MB, a typical entry will be around 237 bytes encoded as JSON so it is safe to set this to the full 500. 
 - firehose_send_early_timeout_ms: If this number of ms passes before the batch size is reached, the batch will be sent anyway.  This is useful to avoid periods of low usage resulting in sporadic entries not being sent for long periods of time.
 - use_aws_default_credentials: If true then aws_secret_key_id/aws_secret_key are ignored and the default AWS credentials provider is used instead - this will assume the role of the Pod's service account if it has one.
 - aws_region: Region the gateway is deployed to.
 - aws_secret_key_id/aws_secret_key: AWS credentials - override with environment variables in test/production (https://www.krakend.io/docs/configuration/environment-vars/) or omit and provide use_aws_default_credentials if your Pod has a service account with a role.
 - delivery_stream_name: The Firehose delivery stream name, e.g. subscriptions-uk-apifactory-api-usage.
 - ignore_paths: An array of paths to ignore - this should contain any paths that don't require authentication.  You can use a * instead of a path segment to match anything in that segment.

## Working on this repo

### Makefile

The [Makefile](Makefile) has a number of helper commands to get you up and running quickly.

| Command                 | Description                                                                                                                                       |
|-------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------| 
| `make build`            | Builds the plugin in the intended format _<name>.so_                                                                                              |
| `make login`            | Logs you into GitHub Container Registry - must have `GH_PAT` environment variable set in your shell - see [GitHub Personal Access Token](#GH-PAT) |
| `make b`                | Builds a microgateway with the configuration found in `krakend.json` - this is useful for local testing                                           |
| `make r`                | Runs the microgateway built by the step above                                                                                                     |
| `make br`               | Alias to run both the build and run commands in a single step                                                                                     |
| `make test-unit`        | Command to run unit tests                                                                                                                         |
| `make test-integration` | Command to run integration tests                                                                                                                  |

### GH PAT

This refers to the GitHub Personal Access Token. This is a token that is specific to your own GitHub account and grants
access to GitHub resources. The resources we're primarily interested in here are the centrally managed docker images _(used to build krakend plugins & the base gateway image)_

This PAT will need to have permissions that at a minmum can read `repos` and `packages`. You will also need to enable SSO on the GITHUB PAT.

### Authenticating manually

1. Authenticate to the [github container registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry#authenticating-to-the-container-registry)

```sh
export GH_PAT="your personal github access token here - must have read packages scope at a minimum"

echo $GH_PAT | docker login ghcr.io -u USERNAME --password-stdin
```

## Misc

### Generating a JWT

Since it's not the responsibility of this plugin to verify JWTs (this should be done by a plugin earlier in the pipeline in production Krakend instances), 
you can produce arbitrary JWTs to use as the Authorization header with the following JS:

```
const header = '{"alg": "HS256","typ": "JWT"}'

const body = '{"sub": "sub-blabla","name": "Somebody","iat": 123456,"subscription_id": "a9de93fc-2d13-44dd-9272-da7f8c17d155","android_id": "07ff00e4-c1a5-4683-9fcb-613a734d8d3f"}'

console.info(btoa(header).replace("=", "") + "." + btoa(body).replace("=", "") + "." + btoa("invalid signature").replace("=", ""))
```
