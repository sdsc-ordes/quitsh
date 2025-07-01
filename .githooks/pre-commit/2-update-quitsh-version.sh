#!/usr/bin/env bash
# shellcheck disable=SC1091
#
# Update the versions in quitsh etc.
# The Nix package build is anyway correct, but its better to keep the
# Go version variable aligned.
#
set -e
set -u
set -o pipefail

DIR=$(cd "$(dirname "$0")" && pwd)
. "$DIR/.common.sh"

function updateVersion() {
    local comp="$1"
    local moduleFile="$2"

    version=$(grep 'version:' "$comp/.component.yaml" | sed -E 's/version: "?(.*)"?/\1/g') ||
        die "Could not get version in '$comp'."

    if ! grep -q "$version" "$moduleFile"; then
        print_info "Updating version '$version' in $moduleFile"
        sed -i -E "s@buildVersion = \".*\"@buildVersion = \"$version\"@" "$moduleFile" ||
            die "Could not update version in '$moduleFile'."
    fi
}

updateVersion "." "pkg/build/version.go"
