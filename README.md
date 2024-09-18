# runner
runner is a graph-based orchestrator runner.

It uses graphs to calculate the run-order of your workflows.
So you can define your workflows and dependencies, and let runner orchestrate it for you.

With runner you can:

  * Eliminate overly complex and unmaintainable shell scripts.
  * Git-OPS read: Have a single source-of-truth for your scripts and workflows.
  * CI/CD Ready: runner binary can be deployed to any Docker image and/or include in your CI/CD pipeline.
  * zero Python, zero Javascript. 99% YAML configurations. (And shell-scripts)

## Step 1: Define your workflow
`myService/runner.yml`

```yaml
workflows:
  - resources/github.yaml
  - resources/auth.yaml
  - resources/database.yaml
  - resources/backend.yaml
  - resources/redis.yaml
  - resources/kafka.yaml
  - resources/docker.yaml
  - https://example.com/runner.yaml
```

## Step 2: Create your resources
`myService/resources/backend1.yaml`

```yaml
resources:
  - id: backend1
    name: "Backend1 is responsible for setting up Auth"
    desc: "Backend1 handles authentication between API calls"
    category: "auth"
    requires:
      - github-access
      - helm-charts
      - helm-postgresql
    run:
      - name: "Clone repo"
        check:
          - "FILE:data/files.txt"
          - "ENV:GH_TOKEN"
          - "CMD:git"
        exec: |
            git clone https://github.com/example/backend1.git backend1
      - name: "Compile project"
        check:
          - "ENV:SOME_TOKEN"
          - "ENV:RUNNER_PARAMS1"
          - "CMD:make"
          - "CMD:go"
        exec: "make build"
```

## Step 3: Run

`$ runner run backend1`

## Step 4: (Optional) See direct or reverse dependencies

`$ runner rdepends git`

```text
git
git -> jdberry-tag
git -> jdberry-tag -> ai-tag
git -> jdberry-tag -> ai-tag -> ai-organize-file
```

## Usage

### Preflight/Postflight Checks and Skipping Steps

You can define rules to check conditions before and after executing a command. You can also specify conditions to skip a step.

* Preflight checks are defined in the `check:` array.
* Postflight checks are defined in the `expect:` array.
* Skip conditions are set using the `skip:` array.

You can negate any check by prefixing it with `!`.

```yaml
  check:
    - "ENV:SHOULD_EXISTS"
    - "!ENV:SHOULD_NOT_EXISTS"
  expect:
    - "CMD:command_should_exists_in_path"
  skip:
    - "FILE:/skip/if/this/file/exists.txt"
```

### Supported Prefixes

The following prefixes are available for condition checks:

* `ENV:`  - Check if an environment variable exists.
* `FILE:` - Check if a file exists.
* `DIR:`  - Check if a directory exists.
* `URL:`  - Check if a fully qualified domain name (FQDN) URL exists.
* `CMD:`  - Check if a command exists in the `$PATH`.
* `EXEC:` - Check if a command runs successfully (returns a 0 exit code).

### Environment Variables

You can set environment variables using the `env:` array. Values can be sourced from a `file:`, the result of an `exec:` command, a static `value:`, or by requesting user input with `input:`.

```yaml
  env:
    - name: "FILE_CONTENTS"
      file: "$RUNNER_PARAMS1"
    - name: "FILE_TYPE"
      exec: "file $KDEPS_PARAMS1"
    - name: "HELLO"
      value: "WORLD"
    - name: "GH_TOKEN"
      input: "Please enter the GH_TOKEN:"
```

Additionally, you can directly append variables to `$RUNNER_ENV`.

### Optional Parameters

Optional parameters can be passed using the `--params` flag. The syntax is `--params "Hello;World"`, which translates to `$RUNNER_PARAMS1` ("Hello") and `$RUNNER_PARAMS2` ("World") in the context of the resources.

## Available commands

```
A graph-based orchestrated runner

Usage:
  runner [command]

Available Commands:
  category    List categories of the given resources
  completion  Generate the autocompletion script for the specified shell
  depends     List dependencies of the given resources
  help        Help about any command
  index       List all resource entries
  rdepends    List reverse dependencies of the given resources
  run         Run the commands for the given resources
  search      Search for the given resources
  show        Show details of the given resources
  tree        Show dependency tree of the given resources
  tree-list   Show dependency tree list of the given resources

Flags:
      --config string   config file (default is runner.yaml)
  -h, --help            help for kdeps
      --params string   extra parameters, semi-colon separated

Use "runner [command] --help" for more information about a command.
```

## Installation

You can download the latest release, or install using Go.

#### Using Go (requires Go 1.22 or later):
```bash
go install github.com/jjuliano/runner@latest
```

#### From source:
```bash
git clone https://github.com/jjuliano/runner.git
cd runner
make build
```

---

## Development

To contribute or modify the project, follow these steps:

1. Clone the repository:
   ```bash
   git clone https://github.com/jjuliano/runner.git
   ```

2. Navigate to the project directory:
   ```bash
   cd runner
   ```

3. Install dependencies and build the project:
   ```bash
   make build
   ```

4. Run tests to ensure everything works correctly:
   ```bash
   make test
   ```

Feel free to submit pull requests or report any issues.

---

## Author

**Runner** was created by Joel Bryan Juliano and is licensed under the [Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0).
