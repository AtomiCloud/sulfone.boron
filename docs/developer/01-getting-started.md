# Getting Started

## Prerequisites

- **Docker** - Container runtime for isolated execution
- **Go 1.24+** - Build and run Boron locally
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
nix develop

# Or use direnv
direnv allow
```

## Setup

### 1. Docker Network (Automatic)

Boron requires a bridge network named `cyanprint` for inter-container communication. This is created automatically on first run via `EnforceNetwork()` - no manual setup required.

**Key File**: `docker.go:391` → `EnforceNetwork()`

### 2. Set Registry Endpoint

Configure the Zinc registry endpoint (defaults to internal cluster):

```bash
# Start server with custom registry
./boron start --registry https://your-registry.example.com
```

**Key File**: `server.go:28` → `server()` function

## Configuration

| Option       | Default                                               | Description                |
| ------------ | ----------------------------------------------------- | -------------------------- |
| `--registry` | `https://api.zinc.sulfone.raichu.cluster.atomi.cloud` | Zinc registry endpoint     |
| Port         | `9000`                                                | HTTP server port           |
| Network      | `cyanprint`                                           | Docker bridge network name |
| Parallelism  | `NumCPU()`                                            | Max concurrent operations  |

## Common Issues

### Issue: Network Creation Fails

**Symptom**: `Error creating network` during startup

**Solution**: Boron's `EnforceNetwork()` first checks if the network exists and only creates it if needed - this is idempotent. If creation still fails, it may be a Docker daemon issue. Check Docker is running:

```bash
docker info
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
