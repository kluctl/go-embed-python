#!/bin/sh

set -e

DIR=$(cd $(dirname $0) && pwd)
cd $DIR/..

PYTHON_STANDALONE_VERSION=$1
PYTHON_VERSION=$2
BUILD_NUM=$3

if [ "$PYTHON_STANDALONE_VERSION" = "" ]; then
  echo "missing python-standalone version"
  exit 1
fi

if [ "$PYTHON_VERSION" = "" ]; then
  echo "missing python version"
  exit 1
fi

if [ "$BUILD_NUM" = "" ]; then
  echo "missing build num"
  exit 1
fi

if [ ! -z "$(git status --porcelain)" ]; then
  echo "working directory is dirty!"
  exit 1
fi

go run ./python/generate --python-standalone-version=$PYTHON_STANDALONE_VERSION --python-version $PYTHON_VERSION
go run ./pip/generate

TAG=v0.0.0-$PYTHON_VERSION-$PYTHON_STANDALONE_VERSION-$BUILD_NUM

echo "checking out temporary branch"
git checkout --detach
git add -f python/internal/data
git add -f pip/internal/data
git commit -m "added python $PYTHON_VERSION from python-standalone $PYTHON_STANDALONE_VERSION"
git tag -f $TAG
git checkout -
