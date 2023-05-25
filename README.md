## Project setup

Generate example `.env` file:
```sh
cat env-example > .env
```

Start the Docker container (with the Docker daemon running):
```sh
docker compose up --build app
```

Add the `firebase-secrets.json` file to the root of the project. This can be obtained via: [Online Firebase console](https://console.firebase.google.com/) > confesi-server-dev > Project settings > Service accounts > Generate new private key.

**Ensure this file is not checked into version control**.

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

## Testing Firebase functionality locally

Install `firebase-tools`:

```sh
npm install -g firebase-tools
```

Start the local emulators. Running this command should open the Emulator Suite UI, usually at [http://127.0.0.1:4000/](http://127.0.0.1:4000/) (address specified after running command):

```sh
firebase emulators:start
```

For example, try adding a user via auth in the Emulator Suite and you'll see the Cloud Function trigger in response.