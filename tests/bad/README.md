# "Bad" tests

Each subdirectory in this directory contains a package that is expected to fail on installation.
The failure might be caused by any issue, including bad packaging, a migration problem, or
a pgpkg test failure.

These tests are listed in pkg_test.go and should be flagged with "expectFailure" set to true.

Run tests from the pgpkg directory using "go test ."
