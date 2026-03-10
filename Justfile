run:
    export HOARY_BASE=.. && export IS_DEV="true" && cd server && go run .
air:
    cd server && IS_DEV=true air

docker:
    docker compose up --build

release version:
    sed -i 's/^version: ".*"/version: "{{version}}"/' hoarydash/config.yaml
    git add hoarydash/config.yaml
    git commit -m "chore: release v{{version}}"
    git tag v{{version}}
    @echo "Done. Run 'git push && git push --tags' to publish."