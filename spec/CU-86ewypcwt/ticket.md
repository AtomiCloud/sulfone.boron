# Ticket: CU-86ewypcwt

- **Type**: Bug
- **Status**: backlog
- **URL**: https://app.clickup.com/t/86ewypcwt
- **Parent**: none
- **Assignee**: Adelphi Liong (adelphi@atomi.cloud)

## Description

Bug

Parallel cyan test sessions collide on unzip container name, causing all but the first session to fail with a Docker container name conflict.

Error (every failing test)

failed to start unzip container: Conflict. The container name "/cyan-unzip-b2f2cd5913934a6abb9130654a50de8f" is already in use

Root Cause

In boron/docker_executor/try_executor.go:121-125, the unzip container is created with SessionId: "":

cc := DockerContainerReference{
CyanId: e.Request.LocalTemplateId,
CyanType: "unzip",
SessionId: "", // <-- should be e.Request.SessionId
}

The naming function in domain_model.go:48-56 already supports session-scoped names — if SessionId is non-empty, it appends it:

if container.SessionId == "" {
return "cyan-" + container.CyanType + "-" + templateVersionId
}
return "cyan-" + container.CyanType + "-" + templateVersionId + "-" + container.SessionId

Since all parallel test cases share the same LocalTemplateId (generated once during warmup) but have unique SessionIds, the container name collides.

Fix

One-line change in boron/docker_executor/try_executor.go:124:

// Before
SessionId: "",
// After
SessionId: e.Request.SessionId,

No breakage risk — the unzip container is ephemeral (created, used, removed in a defer). The naming function already handles both formats. Cleanup uses in-memory DockerContainerReference structs, not name lookups.

Reproduction

cyan test template -c http://localhost:9000 --parallel 10

## Comments

No comments on this ticket.
