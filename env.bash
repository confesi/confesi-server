# this won't be exported if you have the DATABASE_URL
# pointing towards production sql server.
[[ -z "${DATABASE_URL}" ]] && export DATABASE_URL="postgres://postgres:postgres@localhost:5432/confesi?sslmode=disable"
export APPCHECK_TOKEN="test_token"
export POSTGRES_DSN="postgres://postgres:postgres@localhost:5432/confesi?sslmode=disable"
export CIPHER_KEY="thisis32bitlongpassphraseimusing"
export CIPHER_NONCE="asdfasdfasdfasdfasdfasdf"
export CONFESI_ROOT=$PWD

function db() {
    migration_dir="./migrations"
    case $1 in
        "up")
            step=$2
            [[ -z "${2}" ]] && step="1"
            migrate -verbose -path $migration_dir -database $DATABASE_URL up $step
            ;;
        "down")
            step=$2
            [[ -z "${2}" ]] && step="1"
            migrate -verbose -path $migration_dir -database $DATABASE_URL down $step 
            ;;
        "new")
            migrate -verbose create -ext sql -dir $migration_dir -seq $2
            ;;
        "fix")
            migrate -verbose -path $migration_dir -database $DATABASE_URL -verbose force $2
            ;;
        "psql")
            docker exec -it confesi-db psql -U postgres confesi
            ;;
        "dbml")
            pg-to-dbml --c=${DATABASE_URL} -o="./"
            ;;
        "seed")
            go run ./scripts/main.go $2
            ;;
        *)
            echo "usage: db [up|down|new|fix|psql|dbml]"
            echo "for more usage details: https://github.com/golang-migrate/migrate/blob/master/GETTING_STARTED.md"
            ;;
    esac
}

function request() {
    sh $CONFESI_ROOT/scripts/requests
}

function gotest() {
    if [ -z "${1}" ]; then
        echo "Testing al packages"
        go test ./... -coverprofile cover.out -v
    else
        echo "Testing ${1}"
        go test $1 -coverprofile cover.out -v
    fi

    [[ -f cover.out ]] && rm cover.out
}

function token() {
    sh $CONFESI_ROOT/scripts/token
}
