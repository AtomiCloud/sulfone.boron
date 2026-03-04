# Plan 1: Resolver Coordinator Implementation

**Goal**: Add resolver container type support to Boron for warming and proxying resolver containers alongside templates.

**Scope**: All changes needed to support the resolver container type in Boron, including documentation.

## Files to Modify

### Code Files

| File                                   | Changes                                                                     |
| -------------------------------------- | --------------------------------------------------------------------------- |
| `docker_executor/domain_model.go`      | Add `CyanTypeResolver` and `ResolverPort` constants                         |
| `docker_executor/models.go`            | Add `ResolverRes` struct, add `Resolvers` field to `TemplateVersionRes`     |
| `docker_executor/template_executor.go` | Add `Resolvers` field, implement resolver warming functions                 |
| `server.go`                            | Pass resolvers to executor, add `/proxy/resolver/:cyanId/api/resolve` route |

### Documentation Files

| File                                                     | Changes                                                |
| -------------------------------------------------------- | ------------------------------------------------------ |
| `docs/developer/surfaces/api/03-template.md`             | Add `resolvers` field to request body documentation    |
| `docs/developer/surfaces/api/04-proxy.md`                | Add resolver proxy endpoint documentation              |
| `docs/developer/surfaces/api/00-README.md`               | Add resolver proxy endpoint to endpoint list           |
| `docs/developer/features/07-warming-system.md`           | Add resolver warming section                           |
| `docs/developer/concepts/template-vs-cyan-processors.md` | Add note that resolvers follow same pattern            |
| `docs/developer/02-architecture.md`                      | Add resolver containers to architecture diagram/tables |

## Implementation Approach

### Step 1: Add Constants (domain_model.go)

Add container type and port constants near existing type definitions:

- `CyanTypeResolver = "resolver"` - the container type string
- `ResolverPort = 5553` - the port resolvers listen on

### Step 2: Add Model Structs (models.go)

1. Add `ResolverRes` struct following the `PluginRes`/`ProcessorRes` pattern:

   - `ID string` - resolver version ID (UUID)
   - `Version int64` - version number
   - `CreatedAt string` - timestamp
   - `Description string` - optional description
   - `DockerReference string` - docker image reference
   - `DockerTag string` - docker image tag

2. Add `Resolvers []ResolverRes` field to `TemplateVersionRes` struct (after `Templates` field)

### Step 3: Extend TemplateExecutor (template_executor.go)

1. Add `Resolvers []ResolverRes` field to `TemplateExecutor` struct

2. Implement `missingResolverContainer()` function:

   - Follow `missingTemplateContainer()` pattern exactly
   - Filter containers by `CyanType == "resolver"`
   - Match by `CyanId == resolver.ID`
   - Return `(bool, DockerContainerReference)`

3. Implement `missingResolverImages()` function:

   - Follow `missingTemplateImages()` pattern
   - Check if resolver image exists using `DockerReference` and `DockerTag`
   - Return `(bool, DockerImageReference)`

4. Add resolver warming to `WarmTemplate()` function:

   - After template container health check succeeds
   - Iterate through `de.Resolvers` slice
   - For each resolver:
     - Call `missingResolverContainer()` to check if exists
     - If missing: call `missingResolverImages()`, pull if needed, start container
     - Health check on `http://{container-name}:5553/`
   - Collect and return any errors (fail entire request on error)

5. Container creation details:
   - `DockerContainerReference{CyanType: "resolver", CyanId: resolver.ID, SessionId: ""}`
   - Labels: `{"cyanprint.dev": "true"}`
   - Container name will be: `cyan-resolver-{id-without-dashes}`

### Step 4: Update Server (server.go)

1. In `/template/warm` endpoint handler (~line 339):

   - Change `exec := docker_executor.TemplateExecutor{...}` to include resolvers:

   ```go
   exec := docker_executor.TemplateExecutor{
       Docker:   d,
       Template: template.Principal,
       Resolvers: template.Resolvers,  // NEW
   }
   ```

2. Add proxy route after existing template proxy routes (~line 400+):
   ```go
   r.POST("/proxy/resolver/:cyanId/api/resolve", func(c *gin.Context) {
       cyanId := c.Param("cyanId")
       d := docker_executor.DockerContainerReference{
           CyanType:  "resolver",
           CyanId:    cyanId,
           SessionId: "",
       }
       endpoint := "http://" + docker_executor.DockerContainerToString(d) + ":5553/api/resolve"
       // Follow existing proxy pattern: read body, forward, copy response
   })
   ```

### Step 5: Update Documentation

1. **Update `docs/developer/surfaces/api/03-template.md`**:

   - Add `resolvers` field to request body table
   - Add `resolvers` to example JSON
   - Type: `[]ResolverRes`, Required: No

2. **Update `docs/developer/surfaces/api/04-proxy.md`**:

   - Rename title from "Template Proxy API" to "Proxy API"
   - Add new section: `## POST /proxy/resolver/:cyanId/api/resolve`
   - Document parameters (cyanId path param)
   - Document request/response format
   - Add container addressing for resolvers (port 5553)

3. **Update `docs/developer/surfaces/api/00-README.md`**:

   - Add resolver proxy endpoint to the endpoint list table

4. **Update `docs/developer/features/07-warming-system.md`**:

   - Add "Resolver Warming" section after "Template Warming"
   - Document resources: resolver image, resolver container
   - Update flow diagram to include resolver warming step
   - Add edge cases for resolver warming

5. **Update `docs/developer/concepts/template-vs-cyan-processors.md`**:

   - Update note to include resolvers: "This applies to processors, plugins, templates, and resolvers"
   - Add brief mention that resolvers follow the same pattern

6. **Update `docs/developer/02-architecture.md`**:
   - Add Resolver Containers to the Runtime in the architecture diagram
   - Add resolver to component tables where applicable
   - Update port assignments if there's a table

## Edge Cases

1. **Empty resolvers slice** - Should work gracefully, no resolver warming needed
2. **Resolver already running** - Idempotent, skip creation
3. **Resolver image pull fails** - Return error, fail entire warm request
4. **Resolver health check times out** - Return error, fail entire warm request
5. **Multiple resolvers** - Warm all of them, fail if any fail

## Testing Strategy

Since there's no existing test infrastructure, verify manually:

1. Start with warm request without resolvers - should work as before
2. Warm request with resolver - check container created with correct labels
3. Call warm again - should be idempotent (no new containers)
4. Test proxy endpoint - forward request to resolver container
5. Test failure cases - invalid image, health check timeout

## Dependencies

- Zinc registry must return `resolvers` array in `TemplateVersionRes`
- Resolver containers must expose port 5553 with `/` health endpoint and `/api/resolve` endpoint

## Implementation Checklist

### Code Changes

- [ ] Add `CyanTypeResolver` constant in `domain_model.go`
- [ ] Add `ResolverPort` constant in `domain_model.go`
- [ ] Add `ResolverRes` struct in `models.go`
- [ ] Add `Resolvers []ResolverRes` to `TemplateVersionRes` in `models.go`
- [ ] Add `Resolvers []ResolverRes` field to `TemplateExecutor` struct
- [ ] Implement `missingResolverContainer()` function
- [ ] Implement `missingResolverImages()` function
- [ ] Add resolver warming loop to `WarmTemplate()`
- [ ] Update `/template/warm` handler to pass resolvers
- [ ] Add `/proxy/resolver/:cyanId/api/resolve` route

### Documentation Changes

- [ ] Update `docs/developer/surfaces/api/03-template.md` - add resolvers field
- [ ] Update `docs/developer/surfaces/api/04-proxy.md` - add resolver proxy endpoint, rename title
- [ ] Update `docs/developer/surfaces/api/00-README.md` - add endpoint to list
- [ ] Update `docs/developer/features/07-warming-system.md` - add resolver warming section
- [ ] Update `docs/developer/concepts/template-vs-cyan-processors.md` - add resolvers note
- [ ] Update `docs/developer/02-architecture.md` - add resolver containers
