# tailout

`tailout` is a command-line tool for quickly creating a cloud-based exit node in your tailnet.

## Installation

To install `tailout`, you can download the latest release from the [releases page](https://github.com/cterence/tailout/releases).

You can also use the `go install` command:

```
$ go install github.com/cterence/tailout
```

## Prerequisites

To use `tailout`, you'll need to have the following installed:

- [Tailscale](https://tailscale.com/)
- An AWS account

At the moment, `tailout` only supports AWS as a cloud provider. Support for other cloud providers will be added in the future.

## Setup

`tailout init` will initialize `tailout` and create a configuration file in `~/.tailout/config.yaml`. You can edit this file to change the default configuration for `tailout`.

You will also need to set up your AWS credentials. tailout will look for default credentials, like environment variables for access keys or an AWS profile.

To easily check if your credentials are set up correctly, you can use the `aws sts get-caller-identity` command. If you see an error, you'll need to set up your credentials.

## Usage

Create an exit node in your tailnet:

```
tailout create
```

Connect to your exit node:

```
tailout connect
```

Get the status of your exit node:

```
tailout status
```

Disconnect from your exit node:

```
tailout disconnect
```

## Configuration

`tailout` will look for a configuration file at the following paths:

- `/etc/tailout/config.{yml,yaml,hcl,json,toml}`
- `$HOME/.tailout/config.{yml,yaml,hcl,json,toml}`

For exemple, you could have this content in `/etc/tailout/config.yml`:

```yaml
tailscale:
  api_key: tskey-api-xxx-xxx
  auth_key: tskey-auth-xxx-xxx
  tailnet: <your tailnet name>
region: eu-west-3
create:
  shutdown: 15m
```

You can specify any of the above settings as command-line flags or environment variables prefixed by `TAILOUT_`.

For example, to specify the Tailscale API key, you can use the `--tailscale-api-key` flag or the `TAILOUT_TAILSCALE_API_KEY` environment variable.
