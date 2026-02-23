#!/bin/sh
set -e
make test
./tests/e2e.sh
