#!/usr/bin/env bash

# This script runs the presubmit tests; it is started by prow for each PR.
# For convenience, it can also be executed manually.
# Running the script without parameters, or with the --all-tests
# flag, causes all tests to be executed, in the right order.
# Use the flags --build-tests, --unit-tests and --integration-tests
# to run a specific set of tests.

# Markdown linting failures don't show up properly in Gubernator resulting
# in a net-negative contributor experience.
export DISABLE_MD_LINTING=1
export GO111MODULE=on

source $(dirname $0)/../vendor/knative.dev/hack/presubmit-tests.sh

# TODO(mattmoor): integration tests

# We use the default build, unit and integration test runners.

main "$@"
