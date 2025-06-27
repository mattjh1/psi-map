#!/bin/bash

set -euo pipefail

BUMP_TYPE=${1:-patch}
DRY_RUN=${2:-}

LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
IFS='.' read -r MAJOR MINOR PATCH <<< "${LATEST_TAG#v}"

case "$BUMP_TYPE" in
  patch) PATCH=$((PATCH + 1)) ;;
  minor) MINOR=$((MINOR + 1)); PATCH=0 ;;
  major) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
  *)
    echo "❌ Unknown bump type: $BUMP_TYPE"
    echo "Usage: $0 [patch|minor|major] [--dry-run]"
    exit 1
    ;;
esac

NEW_TAG="v$MAJOR.$MINOR.$PATCH"

echo "🔖 Would release: $NEW_TAG (latest was $LATEST_TAG)"

if [[ "$DRY_RUN" == "--dry-run" ]]; then
  echo "🧪 Dry run: not tagging or pushing anything."
  exit 0
fi

# Check clean working directory
if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "❌ Your working tree has uncommitted changes. Please commit or stash first."
  exit 1
fi

# Tag and push
git tag -a "$NEW_TAG" -m "Release $NEW_TAG"
git push origin "$NEW_TAG"

echo "🚀 Tag $NEW_TAG pushed. GitHub Actions will now handle the release."
