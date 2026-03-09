# Task Spec: Implement Coordinator for Resolver

**Ticket**: CU-86ewrbr3t
**Parent**: 86ewr9nen (Resolver System)
**Version**: 2

## Changelog from v1

- Updated `ResolverRes` model to match Zinc's `TemplateVersionResolverResp` schema
- Added `Config` (interface{}) and `Files` ([]string) fields to `ResolverRes`

## Overview

Boron manages resolver containers and routes resolve requests to them. Resolvers are stateless containers (like templates) that are warmed alongside templates when a composition is processed. This spec defines the changes needed in Boron to support the new resolver container type.

## Acceptance Criteria

### AC1: Add Resolver Container Type Constants

- [ ] Add `CyanTypeResolver = "resolver"` constant for container type
- [ ] Add `ResolverPort = 5553` constant for resolver port
- [ ] No validator changes needed - CyanType is a free-form string

### AC2: Create ResolverRes Model

- [ ] Add `ResolverRes` struct in `docker_executor/models.go` following `PluginRes`/`ProcessorRes` pattern
- [ ] Fields: `ID` (string), `Version` (int64), `CreatedAt` (string), `Description` (string), `DockerReference` (string), `DockerTag` (string), `Config` (interface{}), `Files` ([]string)
- [ ] Add `Resolvers []ResolverRes` field to `TemplateVersionRes` struct

### AC3: Extend TemplateExecutor for Resolver Warming

- [ ] Add `Resolvers []ResolverRes` field to `TemplateExecutor` struct
- [ ] Implement `missingResolverContainer()` function following `missingTemplateContainer()` pattern
- [ ] Implement `missingResolverImages()` function following `missingTemplateImages()` pattern
- [ ] Add resolver warming logic to `WarmTemplate()`:
  - After template container is running
  - For each resolver in Resolvers slice
  - Check if container exists (idempotent)
  - Check if image exists, pull if missing
  - Start container if missing
  - Health check on port 5553 using `/` endpoint
- [ ] Resolver container uses `DockerContainerReference{CyanType: "resolver", CyanId: resolver.ID, SessionId: ""}`

### AC4: Use Existing Container Label Pattern

- [ ] Use only `cyanprint.dev: "true"` label (same as template/processor/plugin)
- [ ] No additional labels needed - follow existing pattern

### AC5: Update /template/warm Endpoint

- [ ] Pass `template.Resolvers` to TemplateExecutor when creating it
- [ ] No new request struct needed - `TemplateVersionRes` already used

### AC6: Add Resolver Proxy Endpoint

- [ ] Add `POST /proxy/resolver/:cyanId/api/resolve` route
- [ ] Extract cyanId from path
- [ ] Build container reference: `DockerContainerReference{CyanType: "resolver", CyanId: cyanId, SessionId: ""}`
- [ ] Build endpoint: `http://{container-name}:5553/api/resolve`
- [ ] Forward request body
- [ ] Copy response back

## Constraints

### Idempotency (CRITICAL)

Resolver warming MUST be idempotent - calling warm multiple times must produce the same result:

1. Check before pulling: Query existing images via Docker API
2. Check before creating: Query existing containers via label filter
3. Only create if missing: Skip image pull and container creation if already exists

Follow the exact pattern in `template_executor.go`:

- `missingTemplateContainer()` - checks if container exists
- `missingTemplateImages()` - checks if image exists
- Only pulls/starts if missing == true

### Error Handling

- If resolver warming fails (image pull, container start, health check), the entire template warm request fails
- Return errors to caller - no partial success

### Stateless Design

- Resolvers have no session (like templates)
- `SessionId` is always empty string
- Container name format: `cyan-resolver-{resolverId}`

## Technical Details

### Container Reference Pattern

```go
DockerContainerReference{
    CyanType:  "resolver",
    CyanId:    resolver.ID,    // resolver version ID (UUID)
    SessionId: "",              // EMPTY - stateless like templates
}
```

### Container Labels

```go
Labels: map[string]string{
    "cyanprint.dev": "true",
}
```

### Health Check

- Endpoint: `GET http://{container-name}:5553/`
- Same pattern as templates (use `/` not `/health`)
- Max 60 attempts, 1 second between attempts

### Model Changes

**ResolverRes (new in models.go):**

```go
type ResolverRes struct {
    ID              string      `json:"id"`
    Version         int64       `json:"version"`
    CreatedAt       string      `json:"createdAt"`
    Description     string      `json:"description"`
    DockerReference string      `json:"dockerReference"`
    DockerTag       string      `json:"dockerTag"`
    Config          interface{} `json:"config"`      // From Zinc TemplateVersionResolverResp
    Files           []string    `json:"files"`       // From Zinc TemplateVersionResolverResp (simple glob patterns)
}
```

**TemplateVersionRes (updated in models.go):**

```go
type TemplateVersionRes struct {
    Principal  TemplateVersionPrincipalRes   `json:"principal"`
    Template   TemplatePrincipalRes          `json:"template"`
    Plugins    []PluginRes                   `json:"plugins"`
    Processors []ProcessorRes                `json:"processors"`
    Templates  []TemplateVersionPrincipalRes `json:"templates"`
    Resolvers  []ResolverRes                 `json:"resolvers"` // NEW
}
```

## Files to Modify

| File                                   | Changes                                                                                                      |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| `docker_executor/domain_model.go`      | Add CyanTypeResolver constant, ResolverPort constant                                                         |
| `docker_executor/models.go`            | Update ResolverRes struct with Config/Files, add Resolvers field to TemplateVersionRes                       |
| `docker_executor/template_executor.go` | Add Resolvers field, missingResolverContainer(), missingResolverImages(), resolver warming in WarmTemplate() |
| `server.go`                            | Pass template.Resolvers to executor, add /proxy/resolver/:cyanId/api/resolve route                           |

## Out of Scope

- Integration tests (no existing test infrastructure)
- Resolver SDK implementation (Helium responsibility)
- Resolver registry API (Zinc responsibility)
- VFS merge logic (Iridium responsibility)

## Dependencies

- Zinc registry must provide TemplateVersionResolverResp with dockerReference, dockerTag, config, and files
- Helium must provide resolver containers listening on port 5553 with `/` health endpoint and `/api/resolve` endpoint
