#!/bin/sh
set -e
make test
./tests/e2e_bus.sh
