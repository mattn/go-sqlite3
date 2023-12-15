#!/bin/sh

VERSION=${1:-$(curl -sS https://api.github.com/repos/sqlcipher/sqlcipher/tags | jq -r '.[] | .name' | sort | tail -n1)}
IMAGE=sqlcipher-amalgamation:${VERSION}

docker build -t ${IMAGE} --build-arg version=${VERSION} $(dirname $0)

for ext in c h
do
  file=$(dirname $0)/../../sqlcipher-binding.${ext}
  echo '#ifdef USE_SQLCIPHER' >${file}
  docker run --rm ${IMAGE} sh -c "cat /sqlcipher/sqlite3.${ext}" >>${file}
  docker run --rm ${IMAGE} sh -c "cat /sqlcipher/*userauth.${ext}" >>${file}
  echo '#endif // !USE_SQLCIPHER' >>${file}
done
