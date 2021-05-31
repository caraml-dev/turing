#!/bin/bash
# Unfortunately there are some dependency problems where we can't build stuff outside the context
# This script will build the docker image and push it to the repository.
# Usage: ./build-image.sh <full docker tag>
set -e

mkdir -p temp-deps
cp -r ../../sdk temp-deps/

docker build -t $1 .
docker push $1
rm -rf temp-deps
