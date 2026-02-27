# kubepose

A terminal UI for exploring and managing Kubernetes clusters. Browse resources, inspect details, and perform common kubectl operations through an interactive keyboard-driven interface.

## Features

- **Multi-tab resource browsing** - View Pods, Deployments, Services, and other resources in configurable tabs
- **Fuzzy search** - Filter resources in real-time
- **Multiple output formats** - View resources as table, YAML, or JSON
- **Resource actions** - Describe, logs, delete, edit, exec, port-forward, scale, and rollout restart
- **Context & namespace switching** - Quickly switch between clusters and namespaces
- **Multi-select** - Select multiple resources for bulk operations
- **Fully configurable** - Customize keybindings, tabs, and commands via YAML

## Installation

```bash
go install github.com/clobrano/kubepose@latest
```

### From source

```bash
git clone https://github.com/clobrano/kubepose.git
cd kubepose
make install
```

### Requirements

- Go 1.21+
- kubectl configured with cluster access

## Usage

```bash
kubepose
```

On first run, a default configuration file is created at `~/.config/kubepose/config.yaml`.

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `q` | Quit |
| `/` | Search/filter resources |
| `Enter` | View resource details (table) |
| `Y` | View as YAML |
| `J` | View as JSON |
| `d` | Describe resource |
| `l` / `L` | Logs / Follow logs |
| `D` | Delete resource |
| `e` | Edit resource |
| `x` | Exec into pod |
| `R` | Rollout restart |
| `c` | Switch context |
| `n` | Switch namespace |
| `Tab` | Next tab |
| `Shift+Tab` | Previous tab |
| `1-9` | Jump to tab |
| `j/k` | Move down/up |
| `Space` | Toggle selection |
| `r` | Refresh |
| `?` | Help |

## Configuration

Edit `~/.config/kubepose/config.yaml` to customize:

```yaml
# Path to kubectl binary
kubectl_bin: "kubectl"

# Pager for long output
pager: "less"

# Custom keybindings
keybindings:
  quit: "q"
  describe: "d"
  logs: "l"
  # ...

# Configure tabs
tabs:
  - name: "Pods"
    resource: "pods"
  - name: "Deployments"
    resource: "deployments"
  - name: "Services"
    resource: "services"
```

## Development

```bash
# Build
make build

# Run tests
make test

# Install locally
make install
```

## License

MIT
