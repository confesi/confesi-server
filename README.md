## Project setup

```sh
cat env-example > .env
```

## PostgreSQL

- Accessing PostgreSQL from Docker container.

```sh
docker exec -it confesi-db psql -U postgres confesi
# or use script _db_ (see below)
```

- Migration script
  - Install [Golang Migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate#installation).
  - Bash env.
  - Run from root directory.

```sh
# accessing postgres
./db psql

# new migrations
./db migrate new "<version-name>"

# deploy migration
./db migrate up "<step>" # arg $step can be omitted to deploy just the next one

# deploy rollback
./db migrate down "<step>" # arg $step can be omitted to rollback just the prev one

# fix version
./db migrate fix "<version-number>" # omit leading 0's
```
