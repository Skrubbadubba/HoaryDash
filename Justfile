air:
    cd hoarydash/server && IS_DEV=true air

run:
    cd hoarydash/server && IS_DEV=true go run .

feature name:
    git checkout main
    git pull
    git checkout -b {{name}}
    echo "Branch '{{name}}' created from main."

# Merge current branch to main, tag, and push
# Requires CHANGELOG.md to have uncommitted changes
ship version:
    #!/usr/bin/env bash
    set -euo pipefail

    BRANCH=$(git rev-parse --abbrev-ref HEAD)

    if [ "$BRANCH" = "main" ]; then
        echo "Error: already on main. Run this from your feature branch."
        exit 1
    fi

    if git diff --quiet hoarydash/CHANGELOG.md && git diff --cached --quiet hoarydash/CHANGELOG.md; then
        echo "Error: CHANGELOG.md has no changes. Document your changes before shipping."
        exit 1
    fi

    sed -i 's/^version: ".*"/version: "{{version}}"/' hoarydash/config.yaml
    git add hoarydash/config.yaml hoarydash/CHANGELOG.md
    git commit -m "chore: release v{{version}}"

    git checkout main
    git pull
    git merge --no-ff "$BRANCH" -m "chore: merge $BRANCH for v{{version}}"

    git tag v{{version}}

    echo "Done. Run 'git push && git push --tags' to publish."
    echo "Then delete your branch with: git branch -D $BRANCH"

hotfix version:
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

tags:
    git push && git push --tags

docker:
    docker compose up --build
