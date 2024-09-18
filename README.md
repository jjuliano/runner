# Runner
A simple graph-based orchestrated runner

## Step 1: Create your workflow
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
      - github-accesss
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

You can download the latest from the release, or use Go.

Runner requires Go `1.22`, to install:

`go install github.com/jjuliano/runner@latest`

Or clone the repo:

```sh
git clone https://github.com/jjuliano/runner.git
cd runner
make build
```

## Author

Runner was created by Joel Bryan Juliano, and under Apache 2.0 License
