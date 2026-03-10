air:
    cd hoarydash/server && IS_DEV=true air

run:
    cd hoarydash/server && IS_DEV=true go run .

release version:
    sed -i 's/^version: ".*"/version: "{{version}}"/' hoarydash/config.yaml
    git add hoarydash/config.yaml
    git commit -m "chore: release v{{version}}"
    git tag v{{version}}
    @echo "Done. Run 'git push && git push --tags' to publish."

docker:
    docker compose up --build
