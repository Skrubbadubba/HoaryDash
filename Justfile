run:
    export IS_DEV="true" && cd server && go run .

docker:
    docker compose up --build