#!/bin/sh

set -e

cd "$(dirname "$0")/.."

CURRENT_VERSION=$(grep '#define SQLITE_VERSION ' sqlite3-binding.c | grep -o '[0-9]*\.[0-9]*\.[0-9]*')

if [ -z "$CURRENT_VERSION" ]; then
  echo "Error: Could not extract current SQLite version from sqlite3-binding.c"
  exit 1
fi

LATEST_VERSION=$(curl -fsSL https://www.sqlite.org/download.html \
  | grep -o 'sqlite-amalgamation-[0-9]*\.zip' \
  | head -n 1 \
  | sed 's/sqlite-amalgamation-\([0-9]\)\([0-9][0-9]\)\([0-9][0-9]\)[0-9][0-9]\.zip/\1.\2.\3/' \
  | sed 's/\.0*\([0-9]\)/.\1/g')

if [ -z "$LATEST_VERSION" ]; then
  echo "Error: Could not extract latest SQLite version from sqlite.org"
  exit 1
fi

echo "Current version: $CURRENT_VERSION"
echo "Latest version:  $LATEST_VERSION"

if [ "$CURRENT_VERSION" = "$LATEST_VERSION" ]; then
  echo "Already up to date."
  exit 0
fi

echo "Upgrade available: $CURRENT_VERSION -> $LATEST_VERSION"
echo "Run upgrade/upgrade.sh to upgrade."
exit 1
