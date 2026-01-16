#!/bin/sh

set -e

cd "$(dirname "$0")/.."

go run upgrade/upgrade.go
VERSION=$(grep '#define SQLITE_VERSION_NUMBER' sqlite3-binding.c | grep -o '[0-9]\{7\}')

if [ -z "$VERSION" ]; then
  echo "Error: Could not extract SQLite version"
  exit 1
fi

git branch -d "sqlite-amalgamation-$VERSION" 2>/dev/null || true
git checkout -b "sqlite-amalgamation-$VERSION"
git commit -m "Upgrade SQLite to version $VERSION" sqlite3-binding.c sqlite3-binding.h sqlite3ext.h
git push origin HEAD

gh pr create --title "Upgrade SQLite to version $VERSION" --body "Automated SQLite upgrade to version $VERSION" --web
