# Backup and Restore in Docker (BockeR)

## Overview

Have you ever looked for a cheap solution to store a database backup somewhere safe and you didn't want to bother with an S3-compatible cloud storage?
Look no further, because there is **BockeR**.  
BockeR is a command line tool which creates a backup from a PostgreSQL database, wraps it in a Docker image, and uploads it to Docker Hub. Of course, BockeR will also do the reverse and restore your database from a backup in Docker Hub.

Is it a good idea? Probably not, but it solved a problem I had!

> **Warning**  
> Do **not** push the image to a publich repository, or everybody in the world will have access to your database backup!

## Installation

### Homebrew

Linux and macOS binaries are available in Homebrew:

```sh
brew install bueti/tap/bocker
```

Or `brew tap bueti/tap` and then `brew install bocker`.

### Manual

Download the appropriate file from the [Releases](https://github.com/bueti/bocker/releases) page, unpack the file and put the binary in your PATH.

## Usage

### Configuration

To configure your username and password run:
```sh
bocker config
```

![bocker config](https://vhs.charm.sh/vhs-QLGa2GLyPLUaxnhl2uKyJ.gif)

### List existing backups

To list existing backups you need to tell bocker for which namespace and repository you want to list tags:

```sh
bocker backup list -n <namespace> -r <repository>
```

![bocker backup list](https://vhs.charm.sh/vhs-3LVSVJ42TqACEBIIGcRR4g.gif)

### More
There are some assumptions made:

- The host you are running `bocker` has Docker installed
- `docker login` was run successfully and you must have permission to push images
- You need a Docker Hub Personal Access Token which requires the following permissions: `Read, Write, Delete`

Use `-h` to get help for each subcommand:

```sh
$ bocker --help
Bocker is a command line tool which creates a backup from a PostgreSQL database, 
wraps it in a Docker image, and uploads it to Docker Hub.  Of course, Bocker will also do the 
reverse and restore your database from a backup in Docker Hub.

Usage:
  bocker [command]

Available Commands:
  backup      Backup a Postgresql Database
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  restore     Restores a Posgres database

Flags:
  -h, --help                help for bocker
  -n, --namespace string    Docker Namespace (default "bueti")
  -r, --repository string   Docker Repository

Use "bocker [command] --help" for more information about a command.
```
