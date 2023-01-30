# Backup in Docker

This is command line tool to create Postgres database backups, put them in a Docker image and push them to a Docker registry.

Why  would you do that? The Docker registry is a very cheap data storage. Of course you should only do this with private repositories ;)
Is it a good idea? Probably not, but it solved a problem I had!

## TODO

- [x] Create application object
- [x] use MkdirTemp
- [ ] Create Makefile
- [ ] Create a better README
- [ ] Make it work with psql in a Docker container
- [ ] I really should learn how to write tests..
- [ ] How about using the Docker SDK instead of exec.command?
