#!/bin/bash

# This script tests that different packages are applied.
# This is just the test runner; the actual tests need to be
# part of each package.


# Generate a random database name
# Thanks https://stackoverflow.com/questions/32484504/using-random-to-generate-a-random-string-in-bash#32484733

TEMPDB=`LC_ALL=C tr -dc A-Za-z </dev/urandom | head -c 16`
createdb "$TEMPDB"


# In case the user hits ^C in the middle of a test.
function cleanup {
  dropdb "$TEMPDB"
}

trap cleanup EXIT

exitStatus=0

for err in `find . -type d -depth 1`
do

  TEMPDB=`LC_ALL=C tr -dc A-Za-z </dev/urandom | head -c 16`
  createdb "$TEMPDB"
  export DSN="postgres://localhost:5432/$TEMPDB?sslmode=disable"

  if ! pgpkg $err > /dev/null 2>&1
  then
    echo "* FAIL: $err"
    exitStatus=1  # keep running tests but exit with status when done
  else
    echo "  pass: $err"
  fi

  dropdb "$TEMPDB"
done

trap "" EXIT

if [ $exitStatus == 1 ]
then
  echo "WARNING: at least one test failed." 1>&2
fi

exit $exitStatus
