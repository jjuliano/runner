# runner
**runner** is a graph-based orchestrator designed to calculate and execute workflows in the correct order, based on their dependencies. With **runner**, you can simplify complex shell scripts and streamline your workflow automation.

## Features

- **Simplified Automation**: Replace unmaintainable shell scripts with a structured, YAML-based approach.
- **GitOps-Ready**: Store your scripts and workflows in version control for a single source of truth.
- **CI/CD Integration**: Deploy **runner** within Docker images or as part of your CI/CD pipelines.
- **Language Agnostic**: Use 99% YAML configurations, and shell scripts for execution. No Python or JavaScript required.

## Quick Start Guide

### Step 1: Define Your Workflow

Create a `runner.yml` file that lists the necessary resources for your workflow.

`myService/runner.yml`:

```yaml
workflows:
  - resources/github.yaml
  - resources/helmcharts.yaml
  - resources/auth.yaml
  - resources/database.yaml
  - resources/backend.yaml
  - resources/redis.yaml
  - resources/kafka.yaml
  - resources/docker.yaml
  - https://example.com/frontend.yaml
```

### Step 2: Create Your Resources

Define the resources needed for each step of your workflow.

`myService/resources/backend1.yaml`:

```yaml
resources:
  - id: backend1
    name: "Backend1 - Setup Authentication"
    desc: "Handles authentication between API calls"
    category: "auth"
    requires:
      - github-access
      - helm-charts
      - helm-postgresql
    run:
      - name: "Clone repository"
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

### Step 3: Execute the Workflow

Run the workflow by specifying the desired resource.

```bash
$ runner run backend1
```

### Step 4: (Optional) Explore Dependencies

List direct or reverse dependencies.

```bash
$ runner rdepends git
```

Example output:

```text
git
git -> jdberry-tag
git -> jdberry-tag -> ai-tag
git -> jdberry-tag -> ai-tag -> ai-organize-file
```

## Advanced Usage

### Preflight/Postflight Checks and Skipping Steps

Define conditions that need to be met before (`check:`) or after (`expect:`) a step runs. You can also specify skip conditions using the `skip:` array.

Negate any condition by prefixing it with `!`.

```yaml
check:
  - "ENV:THIS_PREFLIGHT_ENV_VAR_SHOULD_EXIST"
  - "!ENV:SHOULD_NOT_EXIST"
expect:
  - "CMD:this_postflight_command_check_should_be_available_in_path"
skip:
  - "FILE:/skip/if/this/file/exists.txt"
```

### Supported Check Prefixes

- `ENV:` – Checks if an environment variable exists.
- `FILE:` – Verifies if a file exists.
- `DIR:` – Checks if a directory exists.
- `URL:` – Confirms if a URL is reachable.
- `CMD:` – Ensures a command is available in the `$PATH`.
- `EXEC:` – Runs a command to check if it completes successfully (exit code 0).

### Setting Environment Variables

You can set environment variables dynamically using `env:` blocks, sourcing values from files, commands, or user input.

```yaml
env:
  - name: "FILE_CONTENTS"
    file: "$RUNNER_PARAMS1"
  - name: "FILE_TYPE"
    exec: "file $RUNNER_PARAMS1"
  - name: "HELLO"
    value: "WORLD"
  - name: "GH_TOKEN"
    input: "Please enter the GH_TOKEN:"
```

Variables can also be appended directly to `$RUNNER_ENV`. i.e. `echo FOO='bar' >> $RUNNER_ENV`

### Passing Optional Parameters

You can pass optional parameters using the `--params` flag. The format is `--params "param1;param2"`, which sets `$RUNNER_PARAMS1` and `$RUNNER_PARAMS2` in the workflow context.

## CLI Commands

```
A graph-based orchestrator for workflows.

Usage:
  runner [command]

Available Commands:
  category    List categories of the given resources
  completion  Generate the autocompletion script for the specified shell
  depends     List dependencies of the given resources
  help        Help for any command
  index       List all resource entries
  rdepends    List reverse dependencies of the given resources
  run         Execute commands for the specified resources
  search      Search for resources
  show        Show details of the specified resources
  tree        Display a dependency tree
  tree-list   List dependencies in a tree-like format

Flags:
      --config string   Config file (default is runner.yaml)
  -h, --help            Display help for runner
      --params string   Extra parameters (semi-colon separated)

Use "runner [command] --help" for more information.
```

## Installation

### Option 1: Using Go (Go 1.22 or later)

Install **runner** via `go install`:

```bash
go install github.com/jjuliano/runner@latest
```

### Option 2: Build from Source

Clone the repository and build **runner**:

```bash
git clone https://github.com/jjuliano/runner.git
cd runner
make build
```

## Contributing

To contribute to **runner**, follow these steps:

1. Clone the repository:
   ```bash
   git clone https://github.com/jjuliano/runner.git
   ```

2. Navigate to the project directory:
   ```bash
   cd runner
   ```

3. Build the project:
   ```bash
   make build
   ```

4. Run tests:
   ```bash
   make test
   ```

Feel free to open pull requests or report issues.

## License

**runner** is developed by Joel Bryan Juliano and is licensed under the [Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0).
