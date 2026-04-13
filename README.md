## Devc

`devc` is a CLI for creating isolated development containers with a simple SSH-based workflow.

It is designed for two common cases:

- Spin up a container for an existing local project
- Create a ready-to-use development environment directly from a GitHub repository

Instead of maintaining complex Dockerfiles per project, you describe the tools you need in a small config file, build the image, and connect through your editor or plain SSH.

## Highlights

- Simple CLI for local and GitHub-based development containers
- Reproducible toolchains through reusable features
- SSH-first workflow that works with editors you already use
- VS Code integration out of the box
- JetBrains connection instructions via SSH/Gateway
- Rebuild containers when feature requirements change
- Automatic container state and SSH config management

## How It Works

`devc` builds a base Ubuntu image, installs the features you configure, starts a Docker container, and adds an SSH entry to your local `~/.ssh/config`.

From there you can:

- open the environment in VS Code
- connect from JetBrains Gateway
- use `ssh <container-name>` directly

There are two working modes:

- Local project mode: your current project folder is bind-mounted into the container under `/workspace/<project>`
- GitHub mode: the repository is cloned inside a Docker volume and managed under `~/.devc`

## Requirements

Before using `devc`, make sure you have:

- Docker installed and running
- An SSH public key in `~/.ssh/` such as `id_ed25519.pub`, `id_rsa.pub`, or `id_ecdsa.pub`
- Git installed
- Internet access for downloading base images, features, and repositories

Optional:

- VS Code with the Remote - SSH extension
- JetBrains Gateway or a JetBrains IDE with SSH-based remote workflow
- Go 1.25+ if you want to build `devc` from source

## Installation

Build from source:

```bash
make build
```

The binary will be created at `bin/devc`.

To install it system-wide:

```bash
sudo make install
```

Or run it without installing:

```bash
make run ARGS="list"
```

## Quick Start

### Local Project

Inside your project directory:

```bash
devc init
devc build
devc up
devc ssh
```

To open the same container in VS Code:

```bash
devc connect vscode
```

### GitHub Repository

Create a container directly from GitHub:

```bash
devc create owner/repo
```

You can also use the short form handled by the root command:

```bash
devc owner/repo
devc https://github.com/owner/repo
```

For a specific branch:

```bash
devc create owner/repo --branch develop
```

## Typical Workflow

### 1. Initialize a project

```bash
devc init
```

This creates `.devc/devc.yml` in your project.

### 2. Define your tools

Example config:

```yaml
name: my-project
description: Local development environment

features:
  - name: node
    version: "20"
  - name: python
    version: "3.12"
  - name: docker
```

### 3. Build the image

```bash
devc build
```

### 4. Start the container

```bash
devc up
```

### 5. Connect

```bash
devc ssh
devc connect vscode
devc connect jetbrains
```

### 6. Rebuild after feature changes

After editing `.devc/devc.yml`:

```bash
devc rebuild
```

## Available Commands

| Command | What it does |
| --- | --- |
| `devc init` | Creates a new local project config |
| `devc build` | Builds the Docker image for the current project |
| `devc up` | Starts the development container |
| `devc ssh [name]` | Opens an SSH session into a container |
| `devc connect <ide> [name]` | Connects using `vscode` or shows JetBrains connection info |
| `devc list` | Lists running containers |
| `devc list --all` | Lists running and stopped containers |
| `devc features list` | Lists built-in and user-defined features |
| `devc features show <name>` | Shows feature details and parameters |
| `devc stop [name]` | Stops a running container |
| `devc rebuild [name]` | Rebuilds a container while preserving project data |
| `devc destroy [name]` | Removes a container and its managed state |
| `devc create <repo>` | Creates a container from a GitHub repository |

## Built-In Features

`devc` currently ships with these built-in features:

- `node`
- `python`
- `go`
- `rust`
- `java`
- `ruby`
- `dotnet`
- `docker`
- `terraform`
- `opencode`
- `claude-code`

To inspect the exact parameters for any feature:

```bash
devc features show node
```

## Editor Support

### VS Code

`devc connect vscode` launches VS Code using Remote - SSH and opens the project path inside the container.

### JetBrains

`devc connect jetbrains` prints the SSH/Gateway connection details you can use manually.

## Data and State

`devc` stores container metadata under:

```text
~/.devc/containers/<container-name>/
```

That directory is used for managed state such as:

- `state.json`
- `metadata.json`
- stored `devc.yml`

For local projects, your actual source code stays in your existing folder on the host.

For GitHub-created containers, the repository lives inside a Docker volume attached to the container.

## Custom Features

In addition to built-in features, `devc` can load user-defined features from:

```text
~/.devc/features/<feature-name>/feature.yml
```

This lets you keep team-specific or personal tooling outside the core project.

## Notes and Limitations

- GitHub repository creation currently targets GitHub URLs and `owner/repo` shorthand
- Language detection for `devc create` uses the GitHub API and may fail if rate limits are exceeded
- JetBrains support currently provides connection instructions rather than a full automatic launcher
- `devc` expects SSH-based access and will configure entries in your local `~/.ssh/config`

## Development

Useful project commands:

```bash
make build
make test
make fmt
make lint
```

## License

Add your project license here.
