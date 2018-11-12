#!/usr/bin/env bash

set -Eeuo pipefail

if [[ ! -f ~/.netrc ]]
then
    echo "~/.netrc does not exist, creating..."
    cat >~/.netrc <<EOF
machine $DRONE_NETRC_MACHINE
login $DRONE_NETRC_USERNAME
password $DRONE_NETRC_PASSWORD
EOF
fi

set -x

PLUGIN_MODULE="${PLUGIN_MODULE:-}"
PLUGIN_RUN_DIR="${PLUGIN_RUN_DIR:-}"

git config --global user.name "drone"
git config --global user.email "drone@drone.shipt.com"

git fetch --no-tags origin $DRONE_BRANCH

git --no-pager diff --unified=0 origin/$DRONE_BRANCH $PLUGIN_MODULE | $PLUGIN_RUN_DIR/plugin