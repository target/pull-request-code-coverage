#!/usr/bin/env bash

set -Eeuo pipefail

get_env_var () {
    if [ -z "$1" ]
    then
        echo "$2"
    fi
    echo "$1"
}

netrcMachine="$(get_env_var "${DRONE_NETRC_MACHINE:-}" "${VELA_NETRC_MACHINE:-}")"
netrcUsername="$(get_env_var "${DRONE_NETRC_USERNAME:-}" "${VELA_NETRC_USERNAME:-}")"
netrcPassword="$(get_env_var "${DRONE_NETRC_PASSWORD:-}" "${VELA_NETRC_PASSWORD:-}")"



if [[ ! -f ~/.netrc ]]
then
    echo "~/.netrc does not exist, creating..."

    cat >~/.netrc <<EOF
machine $netrcMachine
login $netrcUsername
password $netrcPassword
EOF
fi


set -x

module="$(get_env_var "${PLUGIN_MODULE:-}" "${PARAMETER_MODULE:-}")"
PLUGIN_RUN_DIR="${PLUGIN_RUN_DIR:-}"
branch="$(get_env_var "${DRONE_BRANCH:-}" "${VELA_PULL_REQUEST_TARGET:-}")"

git config --global user.name "drone"
git config --global user.email "drone@drone.shipt.com"


git fetch --no-tags origin  "$branch"
git --no-pager diff --unified=0 origin/"$branch" $module | $PLUGIN_RUN_DIR/plugin
