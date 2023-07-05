# this won't be exported if you have the DATABASE_URL
# pointing towards production sql server.
[[ -z "${DATABASE_URL}" ]] && export DATABASE_URL="postgres://postgres:postgres@localhost:5432/confesi?sslmode=disable"
export APPCHECK_TOKEN="test_token"
export POSTGRES_DSN="postgres://postgres:postgres@localhost:5432/confesi?sslmode=disable"
export CIPHER_KEY="thisis32bitlongpassphraseimusing"
export CIPHER_NONCE="asdfasdfasdfasdfasdfasdf"

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
            echo "usage: migrate [up|down|new|fix|psql|dbml]"
            echo "for extensive usage details: https://github.com/golang-migrate/migrate/blob/master/GETTING_STARTED.md"
            ;;
    esac
}

function request() {
    # check if token value is provided as a command-line argument
    if [ $# -eq 1 ]; then
        new_token="$1"
    elif [ -p /dev/stdin ]; then
        # read the token value from standard input
        read -r new_token
    else
        echo "Usage: request <token_value> or provide token value through standard input"
        exit 1
    fi

    # check if the first script failed
    if [ $? -eq 0 ]; then
        dirs=$(find . -type f -name "requests.http" -exec dirname {} \; | sort -u)

        # replace token values in files
        for dir in ${dirs[@]}; do
            find "$dir" -name "requests.http" -exec sed -E -i '' "s@(Authorization: Bearer )[^[:space:]]+.*@\1$new_token@" {} +
        done

        echo "Token values updated successfully!"
    else
        echo "Failed to obtain token. Aborting script."
        exit 1
    fi
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
    set -o pipefail

    # handle script termination
    abort() {
        echo "Script aborted. Couldn't get token." >&2
        exit 1
    }
    trap 'abort' ERR

    # get the absolute path of the script's directory
    script_dir="$(dirname "$(readlink -f "$0")")"

    # read environment variables from .env file (comments ignored via grep)
    export $(grep -v '^#' "$script_dir/../.env" | xargs)

    # show how to use if user can't bash properly
    if [ "$#" -ne 2 ]; then
        echo "Usage: $0 <email> <password>"
        exit 1
    fi

    # extract vars
    email="$1"
    password="$2"

    # make the authentication request
    response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d '{
            "email": "'"$email"'",
            "password": "'"$password"'",
            "returnSecureToken": true
        }' "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=${FB_API_KEY}")

    # extract the token from the response
    token=$(echo "$response" | grep -o '"idToken": *"[^"]*' | grep -o '[^"]*$')

    # check if the token exists
    if [ -n "$token" ]; then
        echo "$token"
        exit 0
    else
        echo "Login failed. Check your credentials."
        exit 1
    fi
}
