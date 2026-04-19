# Backup and Restore in Docker (BockeR)

> [!NOTE]
> With the latest pricing and subscription changes by Docker there's no longer unlimited free storage and thus this project is kinda pointless.
> You should probably just upload your backup on an S3(-compatible) bucket :)

## Overview
<img align="right" src="https://github.com/bueti/bocker/assets/383917/98d90d7a-38fa-4df4-90c3-3b9bd345c9af">

Have you ever looked for a cheap solution to store a database backup somewhere safe and you didn't want to bother with an S3-compatible cloud storage?

Look no further, because there is **BockeR**.  

BockeR is a command line tool which creates a backup from a PostgreSQL database, wraps it in a Docker image, and uploads it to Docker Hub. Of course, BockeR will also do the reverse and restore your database from a backup in Docker Hub.


Is it a good idea? Probably not, but it solved a problem I had!


> [!WARNING]
> Do **not** push the image to a public repository, or everybody in the world will have access to your database backup!

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
bocker config set
```

If you want to run bocker on server where there's no keyring tool installed, set the following environment variables:

```sh
export DOCKER_USERNAME=<your docker username>
export DOCKER_PASSWORD=<your docker password>
```

`bocker` will prefer environment variables over the keyring.

To inspect the stored configuration:

```sh
bocker config list                  # shows the username; password is hidden
bocker config list --show-password  # also prints the stored password
```

![bocker config](https://vhs.charm.sh/vhs-6w65TVtSWeJqk5oGv5N9cp.gif)

### List existing backups

To list existing backups you need to tell bocker for which namespace and repository you want to list tags:

```sh
bocker backup list -n <namespace> -r <repository>
```

![bocker backup list](https://vhs.charm.sh/vhs-3LVSVJ42TqACEBIIGcRR4g.gif)

### Restore backup

```sh
bocker restore -r greenlight_backup -o postgres -s greenlight -t greenlight_test --tag 2023-02-14_21-11-43
```

Run `bocker restore -h` for the full list of flags.

![Made with VHS](https://vhs.charm.sh/vhs-3tyELWQdiy2wxPcDn1391H.gif)

### Database passwords

For the host path (no `--container-id`), `pg_dump` / `pg_restore` / `psql` inherit the caller's environment, so setting `PGPASSWORD` (or having a `~/.pgpass`) before running `bocker` works as usual.

When `--container-id` is set, `bocker` runs the Postgres tools inside the container via `docker exec`. If `PGPASSWORD` is exported in your shell, it is forwarded with `docker exec -e PGPASSWORD` so the value stays off argv.

### Cancellation

Ctrl+C cancels the in-flight operation — the Docker push, the image pull, or the running `pg_*` subprocess — instead of letting them finish.

### More
There are some assumptions made:

- The host you are running `bocker` has Docker installed
- `docker login` was run successfully and you must have permission to push images
- You need a Docker Hub Personal Access Token which requires the following permissions: `Read, Write, Delete`

Use `-h` to get help for each subcommand:

```sh
bocker --help
bocker backup --help
bocker restore --help
```
