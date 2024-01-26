#!/usr/bin/env bash

dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient -- clean -- %f