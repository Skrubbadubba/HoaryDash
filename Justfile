run:
    export HOARY_BASE=.. && export IS_DEV="true" && cd server && go run .
air:
    cd server && IS_DEV=true air

docker:
    docker compose up --build