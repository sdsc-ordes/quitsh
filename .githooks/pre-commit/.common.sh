#!/usr/bin/env bash
# Some common functions for Githooks.

function _print() {
    local flags="$1"
    local header="$2"
    shift 2

    local msg
    msg=$(printf '%b\n' "$@")
    msg="${msg//$'\n'/$'\n'   }"

    # shellcheck disable=SC2086
    echo $flags -e "⚙️ $header$msg"
}

function print_info() {
    _print "" "" "$@"
}

function print_warning() {
    _print "" "WARN: " "$@" >&2
}

function print_prompt() {
    _print "-n" "" "$@" >&2
}

function print_error() {
    _print "" "ERROR: " "$@" >&2
}

function die() {
    print_error "$@"
    exit 1
}

function assert_staged() {
    # Export if run without githooks...
    if [ -z "${STAGED_FILES:-}" ]; then
        CHANGED_FILES=$(git diff --cached --diff-filter=ACMR --name-only)

        # shellcheck disable=SC2181
        if [ $? -eq 0 ]; then
            STAGED_FILES="$CHANGED_FILES"
        fi
    fi
}
