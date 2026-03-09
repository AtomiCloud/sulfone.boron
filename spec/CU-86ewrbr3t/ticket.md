# Ticket: CU-86ewrbr3t

- **Type**: task
- **Status**: backlog
- **URL**: https://app.clickup.com/t/86ewrbr3t
- **Parent**: 86ewr9nen

## Description

Overview

Boron manages resolver containers and routes resolve requests to them. Resolvers are stateless (like templates) - they have no session and are warmed alongside templates, not via a separate endpoint.

Constraints

Idempotency (CRITICAL):
Resolver warming MUST be idempotent - calling warm multiple times must produce the same result:
Check before pulling: Query existing images via Docker API
Check before creating: Query existing containers via label filter (cyanprint.type=resolver)
Only create if missing: Skip image pull and container creation if already exists

Follow the exact pattern in template_executor.go:
missingTemplateContainer() - checks if container exists
missingTemplateImages() - checks if image exists
Only pulls/starts if missing == true

No Type Constraints:
The boron codebase does NOT constrain container types to a fixed set. CyanType is a string field used in equality comparisons. Adding "resolver" as a 4th type requires no validator changes.

References

[table-embed:1:1 What| 1:2 Reference File| 2:1 Template executor (warm pattern)| 2:2 docker_executor/template_executor.go | 3:1 Container reference| 3:2 docker_executor/domain_model.go ( DockerContainerReference )| 4:1 Proxy pattern for templates| 4:2 server.go ( /proxy/template/:cyanId/... )| 5:1 Warm endpoint| 5:2 server.go ( POST /template/warm )]
Directives

1. Add Resolver Container Type

In docker_executor/domain_model.go, add to existing patterns:
Container type constant: "resolver"
Port: 5553

2. Update TemplateExecutor to Warm Resolvers

In docker_executor/template_executor.go, extend the template warm logic to also warm resolvers:

Add resolver fields to TemplateExecutor:
type TemplateExecutor struct {
Docker DockerClient
Template TemplateVersionPrincipalRes
Resolvers []ResolverReference // NEW - resolvers declared in template
}

Add resolver warming to WarmTemplate():
After template container is running, also:
For each resolver in Resolvers:
Check if resolver container exists (like missingTemplateContainer)
Check if resolver image exists
Pull image if missing
Start resolver container if missing
Health check on port 5553

Resolver container reference (follow template pattern):
DockerContainerReference{
CyanType: "resolver",
CyanId: resolver.ID, // resolver version ID (like template.ID)
SessionId: "", // EMPTY - same as template!
}

3. Update /template/warm Endpoint

In server.go, update the existing /template/warm endpoint:
Parse resolvers from request body (if present)
Pass resolvers to TemplateExecutor

The request body should include resolvers:
type TemplateWarmReq struct {
Template TemplateVersionPrincipalRes `json:"template"`
Resolvers []ResolverReference `json:"resolvers"` // NEW
}

4. Add Proxy Endpoint

In server.go, add proxy route following /proxy/template/:cyanId/... pattern:

POST /proxy/resolver/:cyanId/api/resolve

Should:
Extract cyanId from path
Build endpoint: http://{cyanId}:5553/api/resolve
Forward request body
Copy response back

5. Resolver Model

Create resolver reference struct in model.go:

type ResolverReference struct {
Id string `json:"id"`
Username string `json:"username"`
Name string `json:"name"`
Version int `json:"version"`
DockerReference string `json:"dockerReference"`
DockerTag string `json:"dockerTag"`
}

Container Labels

Follow template pattern (no session):
Labels: map[string]string{
"cyanprint.dev": "true",
"cyanprint.type": "resolver",
"cyanprint.name": fmt.Sprintf("%s/%s", username, name),
// NO session label - resolvers are stateless like templates
}

Flow Summary

POST /template/warm
|
├── TemplateExecutor.WarmTemplate()
| |
| ├── List existing containers (label filter: cyanprint.dev=true)
| ├── List existing images
| ├── List existing volumes
| |
| ├── Check template container missing? -> pull image + start container
| ├── Check template volume missing? -> create volume + unzip
| ├── Health check template (port 5550)
| |
| ├── For each resolver in template:
| | |
| | ├── LIST containers (label filter: cyanprint.type=resolver)
| | ├── LIST images
| | |
| | ├── Check resolver container exists (by CyanId)?
| | | ├── YES -> SKIP (idempotent - already running)
| | | └── NO -> check image, pull if missing, start container
| | |
| | └── Health check resolver (port 5553)
| |
| └── Return
|
└── 200 OK

Key idempotency pattern (copy from missingTemplateContainer()):
List all containers with label filter first
Iterate through list to find matching CyanId
If found -> return false (not missing)
If not found -> return true (missing, needs creation)

Checklist

Add "resolver" container type constant
Create ResolverReference struct in model.go
Update TemplateExecutor to hold resolvers
Add resolver warming logic to WarmTemplate()
Update /template/warm to accept resolvers in request
Add POST /proxy/resolver/:cyanId/api/resolve proxy endpoint
Add integration tests

## Comments

No comments.

---

# Parent: 86ewr9nen (task)

- **Title**: Resolver system
- **Status**: todo
- **URL**: https://app.clickup.com/t/86ewr9nen

## Description

Overview
This spec defines a new artifact type called Resolver for the Sulfone platform. Resolvers solve the problem of file conflicts when multiple templates in a composition need to modify the same file.

Problem Statement
Current Behavior
When templates A, B, C, and D all produce the same file (e.g., package.json), the VFS layerer uses "last-wins" semantics:
Layer 1 (deepest): fileA.MD from Template A
Layer 2: fileA.MD from Template B
Layer 3: fileA.MD from Template C
Layer 4 (highest): fileA.MD from Template D <- This wins, others discarded
Desired Behavior
Templates declare a resolver for files they produce. When conflicts occur, the resolver merges all versions intelligently:
fileA.MD (from Template A) -> ResolverA + ConfigA (Layer 1)
fileA.MD (from Template B) -> ResolverA + ConfigA (Layer 2)
fileA.MD (from Template C) -> ResolverA + ConfigB (Layer 3)
fileA.MD (from Template D) -> ResolverA + ConfigB (Layer 4)

Resolver receives: ConfigB + [(TemplateD, Layer 4), (TemplateC, Layer 3)]
^ Uses highest layer's config

Architecture
Port Assignments
[table-embed:1:1 Artifact Type| 1:2 Port| 1:3 Purpose| 2:1 Template| 2:2 5550| 2:3 Interactive Q&A| 3:1 Processor| 3:2 5551| 3:3 File transformation| 4:1 Plugin| 4:2 5552| 4:3 Post-processing hooks| 5:1 Resolver| 5:2 5553| 5:3 Conflict resolution]
Component Responsibilities
[table-embed:1:1 Component| 1:2 Responsibility| 1:3 Sub-plan| 2:1 Helium| 2:2 Resolver SDK (Node, Python, .NET) - port 5553| 2:3 helium.md| 3:1 Boron| 3:2 Route to resolvers, start resolver containers| 3:3 boron.md| 4:1 Zinc| 4:2 Store resolver artifacts, versions, API endpoints| 4:3 zinc.md| 5:1 Argon| 5:2 Display resolvers in UI| 5:3 argon.md| 6:1 Iridium| 6:2 Invoke resolvers during VFS merge phase| 6:3 iridium.md]

Resolver Configuration
Template Resolver Declaration
Templates declare resolvers in cyan.yaml:
username: atomi
name: nix-init
description: Nix Flake template

resolvers: - resolver: 'atomi/json-merger:1' # username/name:version
config:
strategy: 'deep-merge'
array_strategy: 'append'
files: - 'package.json' - '\*\*/tsconfig.json'

    - resolver: 'atomi/yaml-merger'           # version omitted -> latest at push-time
    config:
      merge_arrays: true
    files:
            - '.github/workflows/*.yaml'

    - resolver: 'atomi/markdown-merger:1'
    config:
      include_source_headers: true
    files:
            - 'README.md'
            - 'CLAUDE.md'

Resolver Config Schema
interface ResolverConfig {
resolver: string; // Format: "username/name" or "username/name:version"
// If version omitted, uses latest at push-time
config: Record<string, any>; // Resolver-specific configuration
files: string[]; // Glob patterns for files to resolve
}

Resolver API (Port 5553)
Endpoint: POST /api/resolve
Request:
{
"config": {
"strategy": "deep-merge",
"array_strategy": "append"
},
"files": [
{
"path": "package.json",
"content": "{ \"name\": \"project\", \"dependencies\": {} }",
"origin": {
"template": "atomi/frontend-template",
"version": 5,
"layer": 4
}
},
{
"path": "package.json",
"content": "{ \"name\": \"project\", \"devDependencies\": {} }",
"origin": {
"template": "atomi/backend-template",
"version": 3,
"layer": 3
}
}
]
}
Response:
{
"path": "package.json",
"content": "{ \"name\": \"project\", \"dependencies\": {}, \"devDependencies\": {} }"
}
Endpoint: GET /health
{ "status": "OK" }

VFS Merge Algorithm
New Flow
┌─────────────────────────────────────────────────────────────┐
│ CompositionOperator │
├─────────────────────────────────────────────────────────────┤
│ 1. Resolve dependencies (post-order traversal) │
│ 2. Execute each template -> VFS (with resolver configs) │
│ 3. Layer all VFS with resolver resolution: │
│ a. Group files by path │
│ b. For each conflicting file group: │
│ - Find resolver from configs │
│ - If resolver exists: call resolver endpoint │
│ - If no resolver: last layer wins (fallback) │
│ 4. 3-way merge with local files │
│ 5. Write to disk │
└─────────────────────────────────────────────────────────────┘

Mathematical Properties
Resolvers SHOULD be commutative and associative:
R(C, [A, B]) = R(C, [B, A]) // Commutativity
R(C, [R(C, [A, B]), C]) = R(C, [A, B, C]) // Associativity

Sub-plans
[table-embed:1:1 File| 1:2 Description| 2:1 helium.md| 2:2 SDK implementation for Node.js, Python, .NET| 3:1 boron.md| 3:2 Routing, container management, proxy endpoints| 4:1 zinc.md| 4:2 Domain models, API endpoints, database migrations| 5:1 argon.md| 5:2 UI components, pages, navigation| 6:1 iridium.md| 6:2 VFS layerer changes, ResolverClient trait]

Migration Path
Phase 1: Infrastructure
Add resolver SDKs to Helium
Add resolver container management to Boron
Add resolver models and migrations to Zinc

Phase 2: Registry
Implement resolver API endpoints in Zinc
Add resolver version management
Update registry client in Iridium

Phase 3: Execution
Update VFS layerer in Iridium
Add resolver client to coordinator
Implement conflict detection and resolution

Phase 4: UI
Add resolver pages to Argon
Update template detail to show resolver configs
Add resolver search and browsing

Open Questions
Resolver discovery: Should templates use resolvers from dependencies, or only declare their own?
Config inheritance: When multiple templates declare same resolver with different configs, use highest layer?
Error handling: If resolver fails - fallback to last-wins, abort, or conflict markers?
Caching: Should resolver outputs be cached for identical inputs?
Binary files: How should resolvers handle binary files? Base64?

Summary
[table-embed:1:1 Component| 1:2 Changes Required| 2:1 Helium| 2:2 New resolver SDK (port 5553)| 3:1 Boron| 3:2 Resolver container management, proxy routing| 4:1 Zinc| 4:2 Resolver/ResolverVersion models, API endpoints| 5:1 Argon| 5:2 Resolver pages, navigation, search| 6:1 Iridium| 6:2 VFS layerer with resolver support]
The resolver system enables intelligent file merging for template compositions, solving "last-wins" problem while maintaining backward compatibility for templates without resolvers.
