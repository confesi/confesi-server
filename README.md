## Project setup

```sh
cat env-example > .env
```

## Accessing postgres db from container

```sh
docker exec -it confesi-db psql -U postgres
```
