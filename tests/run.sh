#!/bin/bash
#
# Run all the tests.
#

pushd good
./run.sh
popd

pushd bad
./run.sh
popd
