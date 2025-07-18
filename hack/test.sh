#!/bin/bash
set -eux -o pipefail

which go-junit-report || go install github.com/jstemmer/go-junit-report@latest

TEST_RESULTS=${TEST_RESULTS:-test-results}
TEST_FLAGS=${TEST_FLAGS:-}

if test "${ARGOCD_TEST_PARALLELISM:-}" != ""; then
	TEST_FLAGS="$TEST_FLAGS -p $ARGOCD_TEST_PARALLELISM"
fi
if test "${ARGOCD_TEST_VERBOSE:-}" != ""; then
	TEST_FLAGS="$TEST_FLAGS -v"
fi

mkdir -p "$TEST_RESULTS"

# `TEST_FLAGS` cannot be quoted as empty needs to evaluate to 0 arguments.
# `TEST_FLAGS` cannot be turned into array without backward incompatible change of script input
# shellcheck disable=SC2086
GODEBUG="tarinsecurepath=0,zipinsecurepath=0" \
    gotestsum --rerun-fails-report=rerunreport.txt --junitfile="$TEST_RESULTS/junit.xml" --format=testname \
    --rerun-fails="$RERUN_FAILS" --packages="$PACKAGES" -- -cover $TEST_FLAGS "$@"
