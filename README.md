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

## Accessing postgres db from container

```sh
docker exec -it confesi-db psql -U postgres
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