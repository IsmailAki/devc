# devc

`devc` turns a GitHub repository into a ready-to-code development container you can open from your editor over SSH.

The primary path is simple: point `devc` at a GitHub repository, let it detect the stack and prepare the environment, then connect with VS Code, JetBrains, or plain SSH. Local projects are supported too, but GitHub is the default experience.

## Why devc

- Start a development environment directly from a GitHub repository
- Keep project toolchains isolated without changing how you work in your editor
- Replace hand-written per-project Dockerfiles with reusable features
- Connect with plain SSH, VS Code Remote - SSH, or JetBrains Gateway
- Rebuild environments when dependencies change while keeping workspace data intact

## Highlights

- GitHub-first onboarding for new environments
- Local project support when you want to containerize an existing repo
- Reproducible feature-based images instead of bespoke setup scripts
- Stable SSH-based access for editors and terminal use
- Non-root workspace handling for writable repos and build outputs
- Docker-in-Docker support through the `docker` feature

## Requirements

You need the following on the host machine:

- Docker
- Git
- An SSH public key in `~/.ssh/` such as `id_ed25519.pub`, `id_rsa.pub`, or `id_ecdsa.pub`

Optional tools:

- VS Code with Remote - SSH
- JetBrains Gateway or a JetBrains IDE with SSH-based remote development
- Go 1.25.5+ if you want to build `devc` from source

## Installation

Install with Go:

```bash
go install github.com/IsmailAki/devc@latest
```

Build from source:

```bash
git clone https://github.com/IsmailAki/devc.git
cd devc
make build
```

The compiled binary is written to `bin/devc`.

To verify the installed version:

```bash
devc --version
```

## Quick Start

### Start from GitHub

Create a ready-to-use environment directly from GitHub:

```bash
devc create https://github.com/owner/repo
devc create github.com/owner/repo
```

The root command also accepts a GitHub repository URL without `https://`:

```bash
devc github.com/owner/repo
devc https://github.com/owner/repo
```

For a specific branch:

```bash
devc create https://github.com/owner/repo --branch develop
```

If automatic language detection is not what you want, override it:

```bash
devc create https://github.com/owner/repo --languages go,node
```

Once the environment is ready, connect with:

```bash
devc ssh
devc connect vscode
devc connect jetbrains
```

### Bring an existing local project

Inside an existing project directory:

```bash
devc init
devc build
devc up
devc ssh
```

This creates `.devc/devc.yml` in your project and lets you pick plugins in a terminal UI.

You can also skip the explicit build step:

```bash
devc up --build
```

### Edit an existing container

Use the interactive editor to update plugins for an existing container:

```bash
devc edit
devc edit <container-name>
```

When you run `devc edit` without a container name, `devc` shows all known containers in a terminal UI, lets you choose one, and then edits its plugins. It checks the selected container's local `.devc/devc.yml` first and falls back to the stored `~/.devc/containers/<name>/devc.yml` if no local config exists.

After editing, you can rebuild immediately from the prompt or later with `devc rebuild [name]`.

## Configuration

For local projects, `devc` reads `.devc/devc.yml` from the repository root. For managed containers created from GitHub, it stores the active config under `~/.devc/containers/<name>/devc.yml`.

The interactive `devc edit` flow edits plugins only in the current version. For a selected container it prefers the local `.devc/devc.yml` when present, and falls back to the stored container config otherwise.

## Built-in features

`devc` currently ships with these built-in features:

- `claude-code`
- `docker`
- `dotnet`
- `go`
- `java`
- `node`
- `opencode`
- `python`
- `ruby`
- `rust`
- `terraform`

To inspect available features and their parameters:

```bash
devc features list
devc features show node
```

You can also define your own features under:

```text
~/.devc/features/<feature-name>/feature.yml
```

## Docker-in-Docker

If the `docker` feature is enabled:

- the container starts in privileged mode
- a dedicated Docker data volume is attached at `/var/lib/docker`
- an internal `dockerd` process is started automatically
- `docker build`, `docker run`, and similar commands work from inside the dev container

Example:

```yaml
features:
  - name: docker
```

## Commands

| Command | Description |
| --- | --- |
| `devc create <repo>` | Create a container from a GitHub repository |
| `devc init` | Create `.devc/devc.yml` for the current project |
| `devc edit [name]` | Open a terminal UI to edit plugins for a container |
| `devc build` | Build the current project's feature image |
| `devc up` | Start the current project's container |
| `devc ssh [name]` | Open an SSH session into a container |
| `devc connect <ide> [name]` | Connect with VS Code or print JetBrains connection info |
| `devc list` | List running containers |
| `devc list --all` | List running and stopped containers |
| `devc rebuild [name]` | Rebuild a container while preserving workspace data |
| `devc stop [name]` | Stop a container |
| `devc destroy [name]` | Remove a container and its managed state |
| `devc features list` | List available features |
| `devc features show <name>` | Show feature details |

Run `devc --help` or `devc <command> --help` for full usage details.

## Editor support

### VS Code

Open the environment in VS Code through Remote - SSH:

```bash
devc connect vscode
```

### JetBrains

Print the SSH or Gateway connection details for a container:

```bash
devc connect jetbrains
```

## Development

Useful development commands:

```bash
make build
make test
make fmt
make lint
make tidy
```

## Current scope

- GitHub repository creation currently targets explicit GitHub URLs, not bare `owner/repo` shorthand
- JetBrains integration currently prints connection details rather than launching the IDE directly
- The project is centered on SSH-based development rather than Docker exec-style interaction
