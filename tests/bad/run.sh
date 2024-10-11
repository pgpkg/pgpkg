#!/bin/bash

# This script tests that we detect different kinds of error.
# Run this script to apply pgpkg to a bunch of badly written packages.


# Generate a random database name
# Thanks https://stackoverflow.com/questions/32484504/using-random-to-generate-a-random-string-in-bash#32484733

TEMPDB=`LC_ALL=C tr -dc A-Za-z </dev/urandom | head -c 16`
createdb "$TEMPDB"
function cleanup {
  dropdb "$TEMPDB"
}
trap cleanup EXIT
export DSN="postgres://localhost:5432/$TEMPDB?sslmode=disable"

exitStatus=0

for err in `find . -type d -depth 1`
do
  # for these BAD tests, a test that runs successfully is actually a problem
  if pgpkg try $err > /dev/null 2>&1
  then
    # the test passed, which is a problem.
    echo "* PASS: $err (this test should have failed)"
    exitStatus=1  # keep running tests but exit with status when done
  else
    echo "    OK: $err (expected)"
  fi
done

if [ $exitStatus == 1 ]
then
  echo "WARNING: all tests in this directory should 'fail', but at least one did not."
fi

exit $exitStatus
