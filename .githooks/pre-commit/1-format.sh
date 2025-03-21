#!/usr/bin/env bash
# shellcheck disable=SC1091
#
# Format all files with Go fmt.

set -e
set -u
set -o pipefail

DIR=$(cd "$(dirname "$0")" && pwd)
. "$DIR/.common.sh"

if ! echo "${QUITSH_TOOLCHAINS:-}" | grep -q "general"; then
    echo "! Toolchain not is not loaded." >&2
    exit 0
fi

FILES=()
if [ -n "${STAGED_FILES:-}" ]; then
    readarray -t FILES < <(echo "$STAGED_FILES")
else
    readarray -t FILES < <(cat "$STAGED_FILES_FILE")
fi

# shellcheck disable=SC2128
if [ "${#FILES[@]}" = "0" ]; then
    print_info "No files to format."
    exit 0
fi

print_info "Running 'quitsh format'..."
just quitsh format "${FILES[@]}"
