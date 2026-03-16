#!/bin/sh

set -e

cd "$(dirname "$0")/.."

CURRENT_VERSION=$(grep '#define SQLITE_VERSION_NUMBER' sqlite3-binding.c | grep -o '[0-9]\{7\}')

go run upgrade/upgrade.go
VERSION=$(grep '#define SQLITE_VERSION_NUMBER' sqlite3-binding.c | grep -o '[0-9]\{7\}')

if [ -z "$VERSION" ]; then
  echo "Error: Could not extract SQLite version"
  exit 1
fi

if [ "$VERSION" = "$CURRENT_VERSION" ]; then
  echo "Already up to date: version $VERSION"
  git checkout -- sqlite3-binding.c sqlite3-binding.h sqlite3ext.h
  exit 0
fi

git branch -d "sqlite-amalgamation-$VERSION" 2>/dev/null || true
git checkout -b "sqlite-amalgamation-$VERSION"
git commit -m "Upgrade SQLite to version $VERSION" sqlite3-binding.c sqlite3-binding.h sqlite3ext.h
git push origin HEAD

MAJOR=$(echo $VERSION | cut -c1)
MINOR=$(echo $VERSION | cut -c2-4 | sed 's/^0*//')
PATCH=$(echo $VERSION | cut -c5-7 | sed 's/^0*//')
CHANGELOG_URL="https://www.sqlite.org/releaselog/${MAJOR}_${MINOR}_${PATCH}.html"

gh pr create --title "Upgrade SQLite to version $VERSION" --body "Automated SQLite upgrade to version $VERSION

## Changelog
See: $CHANGELOG_URL" --web
