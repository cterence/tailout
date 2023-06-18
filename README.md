# xit

`xit` is a command-line tool for quickly creating a cloud-based exit node in your tailnet.

## Installation

To install `xit`, you can use the `go get` command:

```
$ go install github.com/cterence/xit
```

## Prerequisites

To use `xit`, you'll need to have the following installed:

- [Tailscale](https://tailscale.com/)
- [AWS CLI](https://aws.amazon.com/cli/)

## Setup

Before creating exit nodes, you'll need to initialize your tailnet using the `xit init` command.

## Usage

To use `xit`, you can run the `xit create` command to create a new exit node:

```
$ xit create
```

This will create a new exit node in your tailnet using the default configuration.

You can also specify a configuration file using the `--config` flag:

```
$ xit create --config /path/to/config.yaml
```

This will create a new exit node using the configuration specified in the `config.yaml` file.

To stop an exit node, you can run the `xit stop` command:

```
$ xit stop
```

This will stop the exit node running on the current machine.

You can also specify one or more exit nodes to stop by specifying their hostnames:

```
$ xit stop hostname1 hostname2 ...
```

This will stop the exit nodes with the specified hostnames.

## Configuration

`xit` uses a YAML configuration file to specify the settings for the exit node. Here's an example configuration file:

```yaml
tailscale:
  authkey: <your tailscale authkey>
  subnet: 10.0.0.0/24
  hostname: my-exit-node

cloud:
  provider: aws
  region: us-west-2
  instance_type: t2.micro
  ami: ami-0c55b159cbfafe1f0
```

This configuration file specifies the Tailscale authentication key, the subnet to use for the exit node, the hostname of the exit node, and the cloud provider, region, instance type, and AMI to use for the exit node.

You can specify a different configuration file using the `--config` flag when running the `xit create` command.

## License

`xit` is licensed under the MIT License. See the LICENSE file for more information.
