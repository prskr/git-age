#!/usr/bin/env bash

cwd=$(pwd)

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

cd "${SCRIPT_DIR}/.."
dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient --wd "${cwd}" -- $@