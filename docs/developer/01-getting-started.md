# Getting Started

## Prerequisites

- **Docker** - Container runtime for isolated execution
- **Go 1.23+** - Build and run Boron locally
- **Zinc Registry** - Access to template/processor/plugin registry

## Installation

```bash
# Clone the repository
git clone https://github.com/AtomiCloud/sulfone.boron.git
cd sulfone.boron

# Install dependencies
go mod download

# Build the binary
go build -o boron ./main.go
```

### Using Nix (Recommended)

```bash
# Enter development environment
nix-shell

# Or use direnv
direnv allow
```

## Setup

### 1. Initialize Docker Network

Boron requires a bridge network named `cyanprint` for inter-container communication.

```bash
docker network create cyanprint
```

**Key File**: `docker.go:391` → `EnforceNetwork()`

### 2. Set Registry Endpoint

Configure the Zinc registry endpoint (defaults to internal cluster):

```bash
# Start server with custom registry
./boron start --registry https://your-registry.example.com
```

**Key File**: `server.go:28` → `server()` function

## Configuration

| Option | Default | Description |
|--------|---------|-------------|
| `--registry` | `https://api.zinc.sulfone.raichu.cluster.atomi.cloud` | Zinc registry endpoint |
| Port | `9000` | HTTP server port |
| Network | `cyanprint` | Docker bridge network name |
| Parallelism | `NumCPU()` | Max concurrent operations |

## Common Issues

### Issue: Network Already Exists

**Symptom**: `Error creating network` or `network already exists`

**Solution**: Boron auto-creates the network on first run. If you see this error, manually remove and recreate:

```bash
docker network rm cyanprint
docker network create cyanprint
```

**Key File**: `docker.go:391` → `EnforceNetwork()`

### Issue: Container Not Ready

**Symptom**: `reached maximum attempts of 60` during health check

**Solution**: Check container logs for startup errors:

```bash
docker logs cyan-processor-<id>-<session>
```

**Key File**: `executor.go:266` → `statusCheck()`

### Issue: Version Resolution Fails

**Symptom**: `processor X does not have a matching version defined in the template`

**Solution**: Verify the template includes all processor/plugin versions being requested. See [Version Resolution](./features/02-version-resolution.md).

**Key File**: `registry.go:147` → `convertProcessor()`

## Next Steps

- [Architecture](./02-architecture.md) - Understand the system design
- [Features](./features/) - Learn about specific capabilities
- [API Reference](./surfaces/api/) - HTTP endpoint documentation
