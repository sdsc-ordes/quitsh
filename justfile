set positional-arguments
set fallback := false
set shell := ["bash", "-cue"]
comp_dir := justfile_directory()
root_dir := `git rev-parse --show-toplevel`
out_dir := comp_dir / ".output"
build_dir := out_dir / "build"
flake_dir := comp_dir / "tools" /  "nix"

# Default target if you do not specify a target.
default:
    just --list

# Enter a Nix development shell.
develop *args:
    just nix-develop default

# Use our own `cli` tool (built by Nix, its a `quitsh` framework)
# to build this component.
build *args:
    just go-cli exec-target quitsh::build "$@"

# Use our own `cli` tool (built by Nix, its a `quitsh` framework)
# to test this component.
test *args:
    just go-cli exec-target quitsh::test "$@"

# Use our own `cli` tool (built by Nix, its a `quitsh` framework)
# to lint this component.
lint:
    just go-cli exec-target quitsh::lint "$@"

# Format all files.
format *args:
    nix run --accept-flake-config "{{flake_dir}}#treefmt" -- "$@"

# Build the `cli` tool with Nix.
package-nix:
    nix build -L "{{flake_dir}}#cli" -o "{{out_dir}}/package/cli"

# Clean the output folder.
clean:
    rm -rf "{{out_dir}}"

## CI =========================================================================
ci *args:
    just nix-develop "ci" "$@"
## ============================================================================

## Build over Go ==============================================================
## These commands are only for trouble-shooting.

# Run the CLI tool (`quitsh`).
go-cli *args:
    cd tools/cli && \
    go run ./cmd/cli/main.go -C ../.. "$@"

# Build `quitsh`.
go-build:
    go build ./...

# Test `quitsh`.
go-test *args:
    just go-test-unit-tests "$@"
    just go-test-integration "$@"

# Run the unit-tests.
# To execute specific tests you can use `-run regexp`.
go-test-unit-tests *args:
    go test -tags debug,test \
        -cover \
        -covermode=count \
        -coverpkg ./... \
        "$@" ./...

# Enter the Nix development shell `$1` and execute the command `${@:2}`.
[private]
nix-develop *args:
    #!/usr/bin/env bash
    set -eu
    shell="$1"; shift 1;
    args=("$@") && [ "${#args[@]}" != 0 ] || args="$SHELL"
    mkdir -p .devenv/state && pwd >.devenv/state/pwd
    nix develop \
        --accept-flake-config \
        --override-input devenv-root "path:.devenv/state/pwd" \
        "{{flake_dir}}#$shell" \
        --command "${args[@]}"

# Build the integration test.
[private]
go-build-integration-test:
    cd "{{comp_dir}}" && \
    cd test && \
    go build -tags debug,test,integration,coverage \
        -cover \
        -covermode=count \
        -coverpkg ./... \
        -o "{{out_dir}}/coverage/bin/quitsh-integration-test" \
        ./cmd/quitsh-integration-test/main.go

# Test the integration test.
# To execute specific tests you can use `-run regexp`.
go-test-integration *args: go-build-integration-test
    mkdir -p "{{out_dir}}/coverage"

    cd "{{comp_dir}}" && \
    cd test && \
    QUITSH_BIN_DIR="{{out_dir}}/coverage/bin" \
    QUITSH_COVERAGE_DIR="{{out_dir}}/coverage" \
        go test -tags debug,test,integration "$@" ./...
## ============================================================================
