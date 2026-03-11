air:
    cd hoarydash/server && IS_DEV=true air

run:
    cd hoarydash/server && IS_DEV=true go run .

release version:
    #!/usr/bin/env bash
    set -euo pipefail
    if git diff --quiet hoarydash/CHANGELOG.md && git diff --cached --quiet hoarydash/CHANGELOG.md; then
        echo "Error: CHANGELOG.md has no changes. Document your changes before releasing."
        exit 1
    fi
    sed -i 's/^version: ".*"/version: "{{version}}"/' hoarydash/config.yaml
    git add hoarydash/config.yaml hoarydash/CHANGELOG.md
    git commit -m "chore: release v{{version}}"
    git tag v{{version}}
    echo "Done. Run 'git push && git push --tags' to publish."

docker:
    docker compose up --build
