#!/usr/bin/env bash

set -Eeuo pipefail

PLUGIN_MODULE="${PLUGIN_MODULE:-}"
PLUGIN_RUN_DIR="${PLUGIN_RUN_DIR:-}"

git fetch --no-tags origin $DRONE_BRANCH
git --no-pager diff --unified=0 origin/$DRONE_BRANCH $PLUGIN_MODULE | $PLUGIN_RUN_DIR/plugin