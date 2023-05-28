## Requirements

- Go 1.20
- Docker/Docker compose.
- [pg-to-dbml](https://github.com/papandreou/pg-to-dbml) CLI.
- [Golang Migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate#installation).
- Bash env.

## Initial project setup

**Generate example `.env` file:**

```sh
cat env-example > .env
```

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
./database psql

# new migrations
./database migrate new "<version-name>"

# deploy migration
./database migrate up "<step>" # arg $step can be omitted to deploy just the next one

# deploy rollback
./database migrate down "<step>" # arg $step can be omitted to rollback just the prev one

# fix version
./database migrate fix "<version-number>" # omit leading 0's

# generate a new `confesi.dbml`
./database dbml
```

## Testing Firebase functionality locally

**Start the emulator suite:**

```sh
firebase emulators:start
```

This should open the Emulator Suite UI, usually at [http://127.0.0.1:4000/](http://127.0.0.1:4000/) (address specified after running command)
