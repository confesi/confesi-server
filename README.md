## Requirements

- Go 1.20
- Docker/Docker compose.
- [pg-to-dbml](https://github.com/papandreou/pg-to-dbml) CLI.
- [Golang Migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate#installation).
- Bash env.

## Project setup

```sh
cat env-example > .env
```

## PostgreSQL

- [DB Diagram](https://dbdiagram.io/d/64727d587764f72fcff5bc9a).

- Scripts:

```sh
# accessing postgres
docker exec -it confesi-db psql -U postgres confesi
# OR
./db psql

# new migrations
./db migrate new "<version-name>"

# deploy migration
./db migrate up "<step>" # arg $step can be omitted to deploy just the next one

# deploy rollback
./db migrate down "<step>" # arg $step can be omitted to rollback just the prev one

# fix version
./db migrate fix "<version-number>" # omit leading 0's

# generate a new `confesi.dbml`
./db dbml
```
