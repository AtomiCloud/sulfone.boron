# Boron Documentation

Boron is the **Execution Coordinator** for CyanPrint - a Go-based service that orchestrates Docker containers for template processing.

## Quick Links

- [Developer Documentation](developer/) - Complete technical documentation
- [Getting Started](developer/01-getting-started.md) - Setup and quickstart
- [Features](developer/features/) - All features in one place
- [API Reference](developer/surfaces/api/) - HTTP API endpoints
- [Architecture](developer/02-architecture.md) - User flows and design decisions

## About Boron

Boron sits in the **Execution Cluster** of the CyanPrint platform, orchestrating Docker containers (templates, processors, plugins, and mergers) to transform project templates into customized output.

**Key Responsibilities**:
- Session Management - Isolated execution environments with unique naming
- Container Orchestration - Launch and coordinate template, processor, plugin, and merger containers
- File Merging - Combine outputs from multiple processors into unified output
- Output Packaging - Stream compressed archives (tar.gz) to clients
- Resource Cleanup - Automatic removal of containers and volumes

## Technology Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Language | Go | 1.24 |
| HTTP Framework | Gin | v1.9.1 |
| Docker Client | docker/docker | v28.5.2+incompatible |

## Quick Start

```bash
cd boron
nix develop        # or: direnv allow
pls build
pls run           # starts on port 9000
curl http://localhost:9000/
```

For broader CyanPrint ecosystem documentation (TPPE concepts, Iridium CLI, Helium SDKs), see the project-level documentation in the parent repository.
