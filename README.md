# Lade

![Lade logo](lade-logo.png)

[![Build Status](https://img.shields.io/github/actions/workflow/status/lade-io/lade/release.yml)](https://github.com/lade-io/lade/actions/workflows/release.yml)
[![Release](https://img.shields.io/github/v/release/lade-io/lade.svg)](https://github.com/lade-io/lade/releases/latest)

Lade is a developer tool for deploying and managing your apps.

You can learn more about Lade at https://www.lade.io.

## Installation

Lade is supported on MacOS, Linux, and Windows as a standalone binary.
You can download the latest binary from the [releases page](https://github.com/lade-io/lade/releases) on GitHub.

### MacOS

You can install with [Homebrew](https://brew.sh):

```sh
brew install lade-io/tap/lade
```

### Linux

You can download the latest tarball, extract it, and move it to your `$PATH`:

```sh
curl -L https://github.com/lade-io/lade/releases/latest/download/lade-linux-amd64.tar.gz | tar xz
sudo mv lade /usr/local/bin
```

## Build From Source

You can build from source with [Go](https://golang.org):

```sh
go get github.com/lade-io/lade
```

## Getting Started

Create an app:

```sh
$ lade apps create myapp
```

Create an addon:

```sh
$ lade addons create postgres --name mydb
```

Attach an addon to an app:

```sh
$ lade addons attach mydb --app myapp
```

Deploy an app:

```sh
$ lade deploy --app myapp
```

## Command Help

```
Name:
  lade - Manage your Lade resources

Usage:
  lade [command]

Commands:
  addons      Manage addons
  apps        Manage apps
  deploy      Deploy an app
  domains     Manage domains
  env         Manage app environment
  help        Help about any command
  login       Login to your Lade account
  logout      Logout of your Lade account
  logs        Show logs from an app
  plans       List available plans
  ps          Display running tasks
  regions     List available regions
  run         Run a command on an app
  scale       Scale an app

Options:
  -h, --help      Print help message
  -v, --version   Print version and exit

Use "lade [command] --help" for more information about a command.
```
