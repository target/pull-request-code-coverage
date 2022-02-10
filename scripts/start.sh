#!/usr/bin/env bash

set -Eeuo pipefail



if [[ ! -f ~/.netrc ]]
then
    echo "~/.netrc does not exist, creating..."

    cat >~/.netrc <<EOF
machine $VELA_NETRC_MACHINE
login $VELA_NETRC_USERNAME
password $VELA_NETRC_PASSWORD
EOF
fi


set -x

module="${PARAMETER_MODULE:-}"
PARAMETER_RUN_DIR="${PARAMETER_RUN_DIR:-}"
branch="${VELA_PULL_REQUEST_TARGET:-}"

git config --global user.name "vela"
git config --global user.email "vela@xyz.com"


git fetch --no-tags origin  "$branch"
git --no-pager diff --unified=0 origin/"$branch" $module | $PARAMETER_RUN_DIR/plugin
