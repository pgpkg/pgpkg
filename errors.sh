#!/bin/bash

# run this script to apply pgpkg to a bunch of badly written packages.

for err in tests/entity-syntax tests/func-syntax tests/sql-syntax tests/table-ref tests/test-exception tests/func-duplicates
do
  echo ""
  echo $err
  pgpkg $err
done
