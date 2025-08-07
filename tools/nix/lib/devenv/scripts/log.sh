#!/usr/bin/env bash

function quitsh::_print() {
    local color="$1"
    local header="$2"
    shift 2

    local prefix="• 🐙 "

    local hasColor="0"
    if [ -t 1 ] || [ "${CI:-}" = "true" ]; then
        hasColor="1"
    fi

    if [ "$hasColor" = "0" ]; then
        local msg
        msg=$(printf '%b\n' "$@")
        msg="${msg//$'\n'/$'\n'    }"
        echo -e "$prefix$header$msg"
    else
        local s=$'\033' e='[0m'
        local msg
        msg=$(printf "%b\n" "$@")
        msg="${msg//$'\n'/$'\n'    }"
        echo -e "${s}${color}$prefix$header$msg${s}${e}"
    fi
}

function quitsh::debug() {
    if [ -n "$QUITSH_DEVSHELL_DEBUG" ]; then
        quitsh::_print "[0;94m" "" "DEBUG: " "$@" >&2
    fi
}

function quitsh::trace() {
    if [ -n "$QUITSH_DEVSHELL_DEBUG" ] ||
        [ -n "$QUITSH_DEVSHELL_TRACE" ]; then
        quitsh::_print "[0;94m" "TRACE: " "$@" >&2
    fi
}

function quitsh::info() {
    quitsh::_print "[0;94m" "INFO: " "$@" >&2
}

function quitsh::warn() {
    quitsh::_print "[0;31m" "WARN: " "$@" >&2
}

function quitsh::error() {
    quitsh::_print "[0;31m" "ERROR: " "$@" >&2
}

if [ "$1" = "info" ]; then
    shift 1
    quitsh::info "$@"
elif [ "$1" = "trace" ]; then
    shift 1
    quitsh::trace "$@"
elif [ "$1" = "debug" ]; then
    shift 1
    quitsh::debug "$@"
elif [ "$1" = "warn" ]; then
    shift 1
    quitsh::warn "$@"
elif [ "$1" = "error" ]; then
    shift 1
    quitsh::error "$@"
fi
