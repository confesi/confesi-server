# [Confesi](https://confesi.com)

![unit tests](https://github.com/mattrltrent/confesi-server/actions/workflows/unit_tests.yml/badge.svg)
![linting](https://github.com/mattrltrent/confesi-server/actions/workflows/linting.yml/badge.svg)

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

**Add your AWS access tokens to the `.env` file.**

Specifically what IAM roles are needed will be determined in the future. Currently, a general admin user suffices.

**Add the Firebase API key:**

This can be obtained via: [Online Firebase console](https://console.firebase.google.com/) > confesi-server-dev > Project settings > Scroll down till you see "your apps" > Confesi web app. The key should be listed under the `apiKey` field. Add it as `FB_API_KEY` in the `.env` file.

**Add the `firebase-secrets.json` file to the root of the project:**

This can be obtained via: [Online Firebase console](https://console.firebase.google.com/) > confesi-server-dev > Project settings > Service accounts > Generate new private key. _Ensure this file is not checked into version control_.

**Install Node Modules for Cloud Functions:**

```sh
cd functions ; npm i ; cd ..
```

**Install `firebase-tools`:**

```sh
sudo npm install -g firebase-tools
```

**Install the Redis UI to view the cache in real time:**

```sh
sudo npm install -g redis-commander
```

**Install `swag` for API doc website:**

```sh
go install github.com/swaggo/swag/cmd/swag@latest
```

If you encounter problems with installation, check out the official installation docs [here](https://github.com/swaggo/swag) or [this](https://stackoverflow.com/questions/73387155/swag-the-term-swag-is-not-recognized-as-the-name-of-a-cmdlet-function-scri) helpful Stack Overflow question for troubleshooting.


## Running the project

For both steps below, ensure the Docker daemon is running.

**Run/build the Docker container (first time running the project):**

```sh
docker compose up --build app
```

**Run the Docker container (after you've built it the first time):**

```sh
docker-compose up
```

## Scripts

**The `env.bash` file has functions needed for development:**

```sh
source env.bash
```

**The following scripts are now available:**

Replaces all instances of bearer tokens in `requests.http` files with a new token. Useful for testing API routes since Firebase's tokens refresh every hour:

```sh
requests <my_new_token>
```

Get an access token for a user:

```sh
token <email> <password>
```

Fetch new token for user and update it for all `requests.http` files at once:

```sh
token <email> <password> | requests
```

## PostgreSQL

- [DB Diagram](https://dbdiagram.io/d/64727d587764f72fcff5bc9a).

- Scripts (these are made available through the `env.bash`):

```sh
# accessing postgres
docker exec -it confesi-db psql -U postgres confesi
# OR
db psql

# new migrations
db migrate new "<version-name>"

# deploy migration
# arg $step can be omitted to deploy just the next one
db migrate up "<step>"

# deploy rollback
# arg $step can be omitted to rollback just the prev one
db migrate down "<step>"

# fix version
# omit leading 0's
db migrate fix "<version-number>"

# seed/mock data
# this will call `go run ./scripts/main.go`
db seed "<seed-action>"

# generate a new `confesi.dbml`
db dbml
```

## Redis cache

**Start the web UI:**

```sh
redis-commander
```

... this should open the viewer, usually at [http://127.0.0.1:8081/](http://127.0.0.1:8081/) (address specified after running the command).

## Testing Firebase functionality locally

**Start the emulator suite:**

```sh
firebase emulators:start
```

... this should open the Emulator Suite UI, usually at [http://127.0.0.1:4000/](http://127.0.0.1:4000/) (address specified after running command).

## Test runner

- These will also be available in `env.bash`
- Note that since `test` is a UNix command, to invoke the testing function, call `gotest`

**Run all tests:**

```sh
gotest ./...
```

**Running tests to a specific package:**

```sh
gotest <./path/to/package>
```

... for example, to run tests on the cipher package:

```sh
gotest ./lib/cipher
```

## Documentation

**Viewing the docs:**

[Start the API server via Docker](https://github.com/mattrltrent/confesi-server#running-the-project), then open [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) in a browser. The docs are automatically built every time the API server starts or hot reloads.

**Editing the docs:**

[Here](https://github.com/swaggo/swag#declarative-comments-format) is a good resource to learn how to properly annotate handler functions so they appear in the Swagger UI. [Here](https://mholt.github.io/json-to-go/) is another good resource to help with speedy struct-to-JSON conversions.

**Rebuilding the docs:**

Once the docs have been edited, run:

```sh
swag init
```

... this should rebuild the doc files under `/docs` to reflect the changes.

**Formatting the docs:**

You can also run to format the Swagger "code".

```sh
swag fmt
```