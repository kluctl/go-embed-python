set -e

DIR=$(cd $(dirname $0) && pwd)
cd $DIR/..

PYTHON_STANDALONE_VERSION=$1
PYTHON_VERSION=$2

if [ "$PYTHON_STANDALONE_VERSION" = "" ]; then
  echo "missing python-standalone version"
  exit 1
fi

if [ "$PYTHON_VERSION" = "" ]; then
  echo "missing python version"
  exit 1
fi

REMOTE_TAGS=$(git ls-remote)
LOCAL_TAGS=$(git tag)
#echo REMOTE_TAGS=$REMOTE_TAGS
#echo LOCAL_TAGS=$LOCAL_TAGS

BUILD_NUM=1

while true; do
  TAG=v0.0.0-$PYTHON_VERSION-$PYTHON_STANDALONE_VERSION-$BUILD_NUM
  if [ "$(echo $REMOTE_TAGS | grep "refs/tags/$TAG")" != "" -o "$(echo $LOCAL_TAGS | grep "$TAG")" != "" ] ; then
      BUILD_NUM=$(($BUILD_NUM+1))
  else
      break
  fi
done

echo $BUILD_NUM
