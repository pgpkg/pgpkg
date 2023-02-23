#!/bin/bash

# run this script to apply pgpkg to a bunch of badly written packages.

for err in entity-syntax func-syntax sql-syntax table-ref test-exception func-duplicates passing-tests failing-tests function-args
do
  echo ""
  echo $err
  pgpkg -verbose $err
done
