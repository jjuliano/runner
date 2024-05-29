# kdeps
The missing package manager for Kubernetes

## Features

- Eliminates Kubernetes GLUE code
- Automatically resolve dependencies (services, configs or packages)
- Check if service is running, build, install and check if not

- **Dependency Resolution:** List dependencies of a package and its reverse dependencies.
- **Package Information:** Show detailed information about a package.
- **Package Search:** Search for packages based on keywords.
- **Category Listing:** List packages belonging to specific categories.
- **Dependency Tree:** Visualize the dependency tree of a package.

## Installation

### Prerequisites

- Go (Golang) installed on your system.

### Install from Source

To install the package dependency resolver from source, follow these steps:

1. Clone the repository:

   ```sh
   git clone https://github.com/kdeps/kdeps.git
   ```

2. Navigate to the project directory:

   ```sh
   cd kdeps
   ```

3. Build the project:

   ```sh
   make build
   ```

4. Move the executable to your PATH (optional):

   ```sh
   sudo mv kdeps /usr/local/bin
   ```

Now you can use the `kdeps` command to interact with the package resolver.

## Usage

### Basic Commands

- **List Dependencies:** `kdeps depends [package_names...]`
- **List Reverse Dependencies:** `kdeps rdepends [package_names...]`
- **Show Package Details:** `kdeps show [package_names...]`
- **Search for Packages:** `kdeps search [query]`
- **List Categories:** `kdeps category [categories...]`
- **Show Dependency Tree:** `kdeps tree [package_names...]`
- **Show Dependency Tree (Top-Down):** `kdeps tree-list [package_names...]`
- **List All Package Entries:** `kdeps index`

For more detailed usage and options, run `kdeps --help` or `kdeps [command] --help`.

## License

This project is licensed under the [MIT License](LICENSE).
