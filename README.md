## Requirements

- Go 1.20.
- Docker/Docker compose.
- [pg-to-dbml](https://github.com/papandreou/pg-to-dbml) CLI.
- [Golang Migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate#installation).
- Bash env.

## Initial project setup

**Generate example `.env` file:**

```sh
cat env-example > .env
```

**Add the app check token:**

Open the `.env` file and follow the [link](https://generate-random.org/api-token-generator) to create the `APPCHECK_TOKEN` env variable.

**Add the `firebase-secrets.json` file to the root of the project:**

This can be obtained via: [Online Firebase console](https://console.firebase.google.com/) > confesi-server-dev > Project settings > Service accounts > Generate new private key. _Ensure this file is not checked into version control_.

**Install Node Modules for Cloud Functions:**

```sh
cd functions ; npm i ; cd ..
```

**Install `firebase-tools`:**

```sh
npm install -g firebase-tools
```

**NOTE**: all scripts are run from the _root_ directory, (ie, `./scripts/database migrate up`.)

## Running the project

**Start the Docker container (with the Docker daemon running):**

```sh
docker compose up --build app
```

## PostgreSQL

- [DB Diagram](https://dbdiagram.io/d/64727d587764f72fcff5bc9a).

- Scripts:

```sh
# accessing postgres
docker exec -it confesi-db psql -U postgres confesi
# OR
./scripts/database psql

# new migrations
./scripts/database migrate new "<version-name>"

# deploy migration
./scripts/database migrate up "<step>" # arg $step can be omitted to deploy just the next one

# deploy rollback
./scripts/database migrate down "<step>" # arg $step can be omitted to rollback just the prev one

# fix version
./scripts/database migrate fix "<version-number>" # omit leading 0's

# generate a new `confesi.dbml`
./scripts/database dbml

# seed data
export POSTGRES_DSN="" # TODO: make a new bash env scripts that exports all of this
go run ./scripts/main.go --seed-schools
```

## Testing Firebase functionality locally

**Start the emulator suite:**

```sh
firebase emulators:start
```

This should open the Emulator Suite UI, usually at [http://127.0.0.1:4000/](http://127.0.0.1:4000/) (address specified after running command).

## Test runner

**Run all tests:**

```sh
./scripts/test ./...
```

**Running tests to a specific package:**

```sh
./scripts/test <./path/to/package>
```

For example, to run tests on the cipher package:

```sh
./scripts/test ./lib/cipher
```
