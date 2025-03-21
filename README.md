<p align="center">
    <img src="docs/assets/images/logo.svg" width="400px">
</p>

---

> [!CAUTION]
>
> This repository is in `beta` and not stable enough since the design space of
> this tool is still explored.

The `quitsh` framework (/kwɪʧ/) is a build-tooling CLI framework designed to
replace loosely-typed scripting languages (e.g., `bash`, `python`, and similar
alternatives) with the statically-typed language `Go`. Its goal is to simplify
tooling tasks while providing robust, extendable solutions for
mono-repositories.

`quitsh` is an opinionated framework born out of frustration with the lack of
simple and extendable tooling for mono-repos. It is language-agnostic and
toolchain-independent, allowing users to focus on their workflows without being
constrained by specific technologies.

### Key Features

#### Code-First Approach

- All tooling logic is implemented in `Go`.
- Tasks are defined primarily in code, avoiding declarative configurations or
  templated non-typed languages, which often add unnecessary complexity despite
  their flexibility.

#### Component Identification

- Components (i.e., buildable units) are identified by placing a configuration
  file (default: `.component.yaml`) in the corresponding subdirectory.

#### Extendability

- `quitsh` serves as a library to build customized CLI tools for your specific
  tasks.
- Users can add custom commands and specialized tooling features using libraries
  like [`cobra`](https://github.com/spf13/cobra).

#### Targets and Steps

- Each component defines **targets**, which consist of multiple **steps**.
- Targets can depend on other targets across the repository.
- Input change sets can be specified for each target to track modifications and
  determine if the target is outdated.

#### Runner System

- Steps within targets are executed by **runners**, which act as reusable
  replacements for traditional build/tooling scripts.
- Runners can have custom YAML configuration options specified per component in
  `.component.yaml`.

#### Toolchain Dispatch

- Runners are associated with specific toolchains.
- By default, `quitsh` includes a
  [Nix development shell dispatch](https://nix.dev/tutorials/first-steps/declarative-shell.html),
  providing stable and reproducible environments.
- While container-based dispatching is not a primary goal, it can be implemented
  by extending the dispatch interface.

- The tool was built to replicate the same procedure one executes during local
  development and also in CI. Having CI align with what you execute locally is
  not a nice thing to have, its a necessity. Nix development shells (or
  containers) help with this. A Nix shell provides a simple and robust
  abstraction to pin a toolchain. The following visualization gives an overview
  about how `quitsh` is used:

  <p align="center">

  ![quitsh-design](docs/assets/images/quitsh-design.drawio.svg)

  </p>

#### Built-in Libraries

The `pkg` folder offers utilities for common development needs, such as:

- Command Execution: [`pkg/exec`](pkg/exec) provides utilities for process
  execution and command chaining.
- Structured Logging: [`pkg/log`](pkg/log) enables consistent and readable
  logging.
- Error Handling: [`pkg/error`](pkg/error) facilitates contextual error
  management.
- Dependency Graphs: Tools for managing and resolving dependency graphs across
  targets.
- Some Go `test` runners (here as an example) for running Go tests (its used
  internally to test `quitsh` it-self).

#### Performance

- Since all tooling is written in `Go`, `quitsh` provides type safety and fast
  performance by default.
- Combined with Nix-based toolchain dispatch and the ability to write tests
  easily, the framework significantly accelerates the "change, test, improve"
  workflow.

#### Nix Integration

- CLI tools built with `quitsh` can be seamlessly packaged into Nix development
  shells, ensuring accessibility for all users of a given repository.

---

# How We Use It?

TODO: unfinished.

Understand what this framework does, is best accomplished by understanding how
we use this framework in our
[components repo (mono-repo)](https://gitlab.com/data-custodian/custodian).

## Components

Our major components are located in
[./components](https://gitlab.com/data-custodian/custodian/-/tree/main/components).
Each component in `quitsh` is defined by a `.component.yaml` (name is
customizable) file which more or less looks like:

```yaml
# The name of the component: Must be a unique.
name: my-component

# A semantic version.
# This is useful for Nix packaging etc.
version: 0.2.1

# A simple annotation (not used internally) what main language this component uses.
language: go

targets:
  # A target called `test` with two steps.
  my-test:
    # The stage to which this target belongs. Does not need to be provided
    # if the CLI is setup to map target names to stages.
    stage: test

    steps:
      # Step 1: Using runner with ID (how it was registered).
      - runner-id: banana-project::my-test-runner
        config: # Your custom runner YAML config, (optional).

      # Step 2: Using a runner with registered key (stage: `test`, name `my-test`)
      - runner: my-test

  # A target called `build-all` with one step.
  build-all:
    stage: build

    # Defining when this target is considered changed:
    # i.e. whenever `self::sources` input change set is changed.
    inputs: ["self::sources"]

    # Defining dependencies on other targets such that this
    # target is executed after target `my-test` above.
    # You can also link to other components (e.g `other-comp::build`).
    depends: ["self::my-test"]

    steps:
      # Step 1: Using a runner with registered key (stage: `build`, name `my-test`)
      - runner: my-build
        config:
          tags: ["doit"]

  lint:
    steps:
      - ... other steps ...

inputs:
  # An input change set with name `sources` which defines
  # patterns to match all source files.
  sources:
    # A regex which matches `*.go` files in `./src` in the components folder.
    patterns:
      - '^./src/.*\.go$'
```

- Quitsh's own [`.component.yaml`](./.component.yaml)

<!---->
<!-- The tool provides entry points for all scripting needs in this monorepo. To -->
<!-- facilitate this and have the notion of components covered for this monorepo, the -->
<!-- `quitsh` streamlines functionality for each component in -->
<!-- [`components`](../../components) by providing: -->
<!---->
<!-- - Executing common steps such as: -->
<!---->
<!--   - `lint` -->
<!--   - `build` -->
<!--   - `test` -->
<!--   - `build-image` -->
<!---->
<!--   for each component in a monorepo backed by **runners**. -->
<!---->
<!-- - Providing other CI scripts and automation for task usually written in `bash` -->
<!--   or `python` backed by additional subcommands on `quitsh`. -->
<!---->

<!---->
<!-- ## Execution of Steps -->
<!---->
<!-- The execution of steps by `quitsh` is done by reading a `.component.yaml` for -->
<!-- each component. The [`.component.yaml`](.component.yaml) file contains -->
<!-- information for each step the `quitsh` provides, e.g. `lint`, `build`, `test`, -->
<!-- `package`, etc. The `.component.yaml` for the `quitsh` itself looks like: -->
<!---->
<!-- ```yaml -->
<!-- name: quitsh -->
<!-- version: 0.0.5 -->
<!-- language: go -->
<!---->
<!-- steps: -->
<!--   lint: -->
<!--     - runner: go -->
<!--   test: -->
<!--     - runner: go -->
<!--   build: -->
<!--     - runner: go -->
<!--   package: -->
<!--     - runner: nix-image -->
<!-- ``` -->
<!---->
<!-- The logic which is executed behind each step is specified by the field `runner`. -->
<!-- A runner is Go code applicable for a certain step which should work for all -->
<!-- components. The runner is registered in -->
<!-- [ `factory` ](./pkg/runner/factory/runner.go). Runners can be written by -->
<!-- implementing the interface [`Runner`](./pkg/runner/runner.go) inside -->
<!-- [`./pkg/runner/runners`](./pkg/runner/runners) and registering them in -->
<!-- [`./pkg/runner/factory/init-runners.go`]. -->
<!---->
<!-- ## Extending Functionality in `quitsh` -->
<!---->
<!-- If you need new functionality for CI and local development which you normally -->
<!-- would write in `bash`/`python` follow the following steps: -->
<!---->
<!-- - If the functionality is **a feature needed in an existing runner and step**: -->
<!--   Extend the runner and make it work with your new test/build/lint feature. -->
<!---->
<!-- - If the functionality is **not related to a runner or the same for each -->
<!--   component with that language**: Extend the `quitsh`s by providing another -->
<!--   subcommand which does what you need, see -->
<!--   [`generate-version`](./cmd/quitsh/cmd/generate-version/generate-version.go). -->
<!---->
<!-- - If the functionality is **for a certain language, e.g. `go` or `python` and -->
<!--   applies to each component which is written in that language**: consider adding -->
<!--   a new runner for an already pre-defined step. If you also need a new destinct -->
<!--   step, discuss the step name with the authors of this tool and -->
<!--   [integrate it here](./pkg/common/step_type.go). -->
<!---->
<!-- ### Additional Runner Config -->
<!---->
<!-- Also there is the possibility to load additional runner specific YAML configs, -->
<!-- e.g. the `go` build runner loads the following -->
<!-- [config](./pkg/runner/runners/go/build-config.go): -->
<!---->
<!-- ```yaml -->
<!-- steps: -->
<!--   build: -->
<!--     - runner: go -->
<!--       config: -->
<!--         version-module: "pkg/myversion-module" # defaults to `pkg/build` -->
<!-- ``` -->
