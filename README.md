# Backup in Docker

## Overview

Have you ever looked for a cheap solution to store a database backup somewhere safe and you didn't want to bother with an S3-compatible cloud storage?
Look no further, because there is **Bocker**.  
Bocker is a command line tool which creates a backup from a PostgreSQL database, wraps it in a Docker image, and uploads it to Docker Hub. Of course, Bocker will also do the reverse and restore your database from a backup in Docker Hub.

Is it a good idea? Probably not, but it solved a problem I had!

> **Warning**  
> Do **not** push the image to a publich repository, or everybody in the world will have access to your database backup!

## Usage

```sh
$ bocker --help
Bocker is a command line tool which creates a backup from a PostgreSQL database, 
wraps it in a Docker image, and uploads it to Docker Hub. 
Of course, Bocker will also do the reverse and restore your database from a backup in Docker Hub.

Usage:
  bocker [command]

Available Commands:
  backup      Backup a Postgresql Database
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  restore     Restores a Posgres database

Flags:
  -h, --help   help for bocker

Use "bocker [command] --help" for more information about a command.
```

