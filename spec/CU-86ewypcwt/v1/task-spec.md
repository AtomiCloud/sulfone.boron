# Task Spec: Fix unzip container name collision in parallel try sessions

## Problem

When running `cyan test template` with `--parallel > 1`, all try sessions for the same template generate the same unzip container name because `SessionId` is hardcoded to `""`. Docker requires unique container names, so only the first session succeeds — all others fail with:

```
Conflict. The container name "/cyan-unzip-..." is already in use
```

## Root Cause

In `docker_executor/try_executor.go:124`, the `populateBlobFromImage()` function creates a `DockerContainerReference` with `SessionId: ""`. The naming function in `domain_model.go:48-56` generates the same container name for all sessions sharing the same `LocalTemplateId`:

```go
// try_executor.go:121-125
cc := DockerContainerReference{
    CyanId:    e.Request.LocalTemplateId,
    CyanType:  "unzip",
    SessionId: "",  // <-- bug: should be e.Request.SessionId
}
```

## Fix

Change `SessionId: ""` to `SessionId: e.Request.SessionId` on line 124 of `docker_executor/try_executor.go`.

The naming function already supports session-scoped names — non-empty `SessionId` is appended to the container name.

## Safety

- The unzip container is ephemeral (created, used, removed in a defer at line 135-138)
- Cleanup uses the same `cc` struct reference, so it resolves correctly
- The blob volume intentionally shares `SessionId: ""` across sessions (it checks for existence and reuses)
- No API or interface changes

## Out of Scope

- `populateBlobFromPath()` copy-helper container (line 104) also uses `SessionId: ""` — separate concern, less common collision scenario
