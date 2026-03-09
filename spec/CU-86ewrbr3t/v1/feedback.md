# Feedback: v1 Task Spec

## FB1: ResolverRes Missing Config and Files Fields

**Status**: ✅ Addressed
**Severity**: High
**Source**: Zinc Registry Swagger (`localhost:9001`)

### Problem

The `ResolverRes` struct in the spec does not match Zinc's `TemplateVersionResolverResp` schema. Zinc returns resolvers with `config` and `files` fields that are missing from the Boron model.

### Zinc Schema (`TemplateVersionResolverResp`)

```json
{
  "id": "uuid",
  "version": "int64",
  "createdAt": "date-time",
  "description": "string (nullable)",
  "dockerReference": "string (nullable)",
  "dockerTag": "string (nullable)",
  "config": {}, // ← MISSING in spec
  "files": ["string"] // ← MISSING in spec (note: []string, NOT []CyanGlobReq)
}
```

### Current Spec (AC2)

```go
type ResolverRes struct {
    ID              string `json:"id"`
    Version         int64  `json:"version"`
    CreatedAt       string `json:"createdAt"`
    Description     string `json:"description"`
    DockerReference string `json:"dockerReference"`
    DockerTag       string `json:"dockerTag"`
    // Missing: Config and Files!
}
```

### Required Fix

Update `ResolverRes` in `docker_executor/models.go`:

```go
type ResolverRes struct {
    ID              string      `json:"id"`
    Version         int64       `json:"version"`
    CreatedAt       string      `json:"createdAt"`
    Description     string      `json:"description"`
    DockerReference string      `json:"dockerReference"`
    DockerTag       string      `json:"dockerTag"`
    Config          interface{} `json:"config"`      // NEW
    Files           []string    `json:"files"`       // NEW - array of strings (glob patterns)
}
```

### Impact Analysis

| Area                   | Impact             | Notes                                                 |
| ---------------------- | ------------------ | ----------------------------------------------------- |
| AC2: ResolverRes Model | ✅ **Only change** | Model must have fields for JSON deserialization       |
| AC3: TemplateExecutor  | ❌ None            | Warming ignores config/files, just starts containers  |
| AC5: /template/warm    | ❌ None            | Works automatically if model is correct               |
| AC6: Proxy Endpoint    | ❌ None            | Transparent forwarding - Iridium handles config/files |

### Design Decisions

1. **Config/files used at resolve time** - Iridium sends them with each resolve request
2. **No storage in Boron** - Transparent proxying, keeps Boron simple
3. **Fields for API compliance only** - Boron deserializes but doesn't actively use them

### Key Difference from Processors

- `CyanProcessorReq.Files` is `[]CyanGlobReq` (complex objects with root/glob/exclude/type)
- `ResolverRes.Files` is `[]string` (simple glob patterns)

This suggests resolvers use a simpler file matching pattern than processors.

### Resolution

✅ **Update AC2** to include `Config` and `Files` fields in `ResolverRes` struct definition.
