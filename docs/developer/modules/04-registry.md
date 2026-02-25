# Registry Module

**What**: HTTP client for querying the Zinc registry to resolve processor and plugin versions.

**Why**: Translates user-friendly references (e.g., `username/processor`) into concrete version IDs and Docker image references.

**Key Files**:

- `docker_executor/registry.go:11` → `RegistryClient` struct
- `docker_executor/registry.go:147` → `convertProcessor()`
- `docker_executor/registry.go:205` → `convertPlugin()`

## Responsibilities

What this module is responsible for:

- Query Zinc registry for version metadata
- Parse Cyan references (username/name:version)
- Resolve version references to concrete IDs
- Match resolved versions against template definitions
- Handle latest version fallback

## Structure

```text
docker_executor/registry.go
├── RegistryClient struct        # HTTP client
├── getProcessorVersion()        # Get specific version
├── getProcessorVersionLatest()  # Get latest version
├── getPluginVersion()           # Get specific version
├── getPluginVersionLatest()     # Get latest version
├── convertProcessor()           # Resolve processor reference
└── convertPlugin()              # Resolve plugin reference
```

| File                          | Purpose                   |
| ----------------------------- | ------------------------- |
| `docker_executor/registry.go` | Zinc registry HTTP client |

## Dependencies

```mermaid
flowchart LR
    A[Registry] --> B[Zinc API]
    C[Merger] --> A
```

| Dependency | Why                        |
| ---------- | -------------------------- |
| Zinc API   | Source of version metadata |

## Used By

| Module | Why                                  |
| ------ | ------------------------------------ |
| Merger | Uses registry for version resolution |

## Key Internal Functions

> **Note:** The methods below are unexported (internal) helpers on `RegistryClient`. They are not part of the public API.

### Registry Client

**Key File**: `docker_executor/registry.go:11` → `RegistryClient` struct

```go
type RegistryClient struct {
    Endpoint string
}
```

### Version Query Methods

**Key File**: `docker_executor/registry.go:15` → Version queries

```go
func (rc RegistryClient) getProcessorVersion(username, name, version string) (RegistryProcessorVersionRes, error)
func (rc RegistryClient) getProcessorVersionLatest(username, name string) (RegistryProcessorVersionRes, error)
func (rc RegistryClient) getPluginVersion(username, name, version string) (RegistryPluginVersionRes, error)
func (rc RegistryClient) getPluginVersionLatest(username, name string) (RegistryPluginVersionRes, error)
```

### Conversion Methods

**Key Files**: `docker_executor/registry.go:147` → `convertProcessor()`, `docker_executor/registry.go:205` → `convertPlugin()`

```go
func (rc RegistryClient) convertProcessor(cp CyanProcessorReq, processors []ProcessorRes) (CyanProcessor, error)
func (rc RegistryClient) convertPlugin(cp CyanPluginReq, plugins []PluginRes) (CyanPlugin, error)
```

## API Endpoints Used

| Resource          | Endpoint                                               | Purpose              |
| ----------------- | ------------------------------------------------------ | -------------------- |
| Processor version | `/api/v1/Processor/slug/:user/:name/versions/:version` | Get specific version |
| Processor latest  | `/api/v1/Processor/slug/:user/:name/versions/latest`   | Get latest version   |
| Plugin version    | `/api/v1/Plugin/slug/:user/:name/versions/:version`    | Get specific version |
| Plugin latest     | `/api/v1/Plugin/slug/:user/:name/versions/latest`      | Get latest version   |

## Related

- [Version Resolution Feature](../features/02-version-resolution.md) - Feature documentation
- [Version Resolution Algorithm](../algorithms/01-version-resolution.md) - Implementation details
- [Merger Module](./03-merger.md) - Uses registry for resolution
