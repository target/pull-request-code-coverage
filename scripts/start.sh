#!/usr/bin/env bash

set -Eeuox pipefail

env

PLUGIN_MODULE="${PLUGIN_MODULE:-}"
PLUGIN_RUN_DIR="${PLUGIN_RUN_DIR:-}"

git config --global user.name "drone"
git config --global user.email "drone@drone.shipt.com"

git fetch --no-tags origin $DRONE_BRANCH

git --no-pager diff --unified=0 origin/$DRONE_BRANCH $PLUGIN_MODULE | $PLUGIN_RUN_DIR/plugin