#!/usr/bin/env bash
# Merges QUITSH_TOOLCHAINS with `$1` for layering DevShells.

old="${QUITSH_TOOLCHAINS:-}"
new="$1"

function clean() {
    local s="$1"
    echo "$s" |
        tr "," "\n" |                                                     # split to lines
        sed -e 's/^[[:space:]]*//' -e 's/ [[:space:]]*$//' -e '/^ *$/d' | # trim and delete empty lines.
        sort -u |
        paste -sd, -
}

if [ -n "$old" ]; then
    new="$old,$new"
fi

clean "$new"
