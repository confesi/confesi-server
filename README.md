# [Confesi](https://confesi.com)

![unit tests](https://github.com/mattrltrent/confesi-server/actions/workflows/unit_tests.yml/badge.svg)
<!-- ![linting](https://github.com/mattrltrent/confesi-server/actions/workflows/linting.yml/badge.svg) commented out until linting is added back-->

## Notes

- All scripts are run from the _root_ directory, (ie, `./scripts/database migrate up`.)

- Please be **very** cautious about signing up using the assorted `requests.http` files. The ones that invoke email actions (reseting password, updating email, creating account, etc.) are **live**. So, only try them using addresses you actually own.

## Requirements

- Go 1.20.
- Docker/Docker compose.
- [pg-to-dbml](https://github.com/papandreou/pg-to-dbml) CLI.
- [Golang Migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate#installation).
- Bash env.

## Initial project setup

**Download the file `IP2LOCATION-LITE-DB5.IPV6.BIN` externally:**

Then put it here, like so `~/assets/IP2LOCATION/IP2LOCATION-LITE-DB5.IPV6.BIN`

**Add required env variables to GitHub Secrets Manager for tests to pass:**

Repo > Settings > Secrets and variables > Actions > New repository secret

**Generate example `.env` file:**

```sh
cat env-example > .env
```

**Add the app check token:**

Open the `.env` file and follow the [link](https://generate-random.org/api-token-generator) to create the `APPCHECK_TOKEN` env variable.

**Ensure you have the correct 16-byte `MASK_SECRET` in the `.env` file.**

An example is provided in the `env-example`, but obviously generate your own for prod.

**Add your AWS data to the `.env` file.**

Specifically what IAM roles are needed will be determined in the future. Currently, a general admin user suffices.

This includes your: access key, secret access key, and region.

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

**Replaces all instances of bearer tokens in `requests.http` files with a new token. Useful for testing API routes since Firebase's tokens refresh every hour.**

```sh
./scripts/requests <my_new_token>
```

**Get an access token for a user:**

```sh
./scripts/token <email> <password>
```

**Fetch new token for user and update it for all `requests.http` files at once:**

```sh
./scripts/token <email> <password> | ./scripts/requests
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

# endpoint speed checker
go run ./scripts/main.go --test-endpoints-speed

# seed data (use the POSTGRES_DSN found in `/scripts/test` not `.env`)
export POSTGRES_DSN="" 
export MASK_SECRET="" # Found in `.env`

go run ./scripts/main.go --seed-all # Seed every seedable table

go run ./scripts/main.go --seed-schools # Seed schools
go run ./scripts/main.go --seed-feedback-types # Seed feedback types
go run ./scripts/main.go --seed-report-types # Seed report types
go run ./scripts/main.go --seed-post-categories # Seed post categories
go run ./scripts/main.go --seed-faculties # Seed faculties
go run ./scripts/main.go --seed-years-of-study # Seed years of study

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

**Run all tests:**

```sh
./scripts/test ./...
```

**Running tests to a specific package:**

```sh
./scripts/test <./path/to/package>
```

... for example, to run tests on the cipher package:

```sh
./scripts/test ./lib/cipher
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

## Todo: Prod

- Update Firebase Security Rules to ensure that a user can only access a document in the `rooms` collection if their `uid` is listed in the `user_id` field.

- Firebase indices created (else client (possibly server?) throws exceptions).