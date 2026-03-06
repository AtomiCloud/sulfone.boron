# Plan 1: Update ResolverRes model

**Goal:** Update `ResolverRes` struct to match Zinc's `TemplateVersionResolverResp` schema

**Scope:** Single struct change - add `Config` and `Files` fields to existing `ResolverRes` struct

**Files to modify:**
| File | Changes |
| ---- | ------- |
| `docker_executor/models.go` | Add `Config` and `Files` fields to `ResolverRes` struct |

**Implementation:**

```go
type ResolverRes struct {
    ID              string      `json:"id"`
    Version         int64       `json:"version"`
    CreatedAt       string      `json:"createdAt"`
    Description     string      `json:"description"`
    DockerReference string      `json:"dockerReference"`
    DockerTag       string      `json:"dockerTag"`
    Config          interface{} `json:"config"`      // NEW - matches Zinc
    Files           []string    `json:"files"`        // NEW - simple glob patterns
}
```

**Testing:**

- Verify JSON deserialization works with the new fields
- Run `go build ./...` to confirm compilation

**Checklist:**

- [ ] Add `Config interface{}` field to `ResolverRes` in `docker_executor/models.go`
- [ ] Add `Files []string` field to `ResolverRes` in `docker_executor/models.go`
- [ ] Run `go build ./...`
- [ ] Verify existing code still compiles
