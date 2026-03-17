# Plan 1: Fix unzip container SessionId

## File

`docker_executor/try_executor.go`

## Change

Line 124: change `SessionId: ""` to `SessionId: e.Request.SessionId`

```go
// Before
cc := DockerContainerReference{
    CyanId:    e.Request.LocalTemplateId,
    CyanType:  "unzip",
    SessionId: "",
}

// After
cc := DockerContainerReference{
    CyanId:    e.Request.LocalTemplateId,
    CyanType:  "unzip",
    SessionId: e.Request.SessionId,
}
```

## Verification

- `go build ./...` passes
- Container naming in `domain_model.go:48-56` already handles non-empty `SessionId`
- Defer cleanup at line 135-138 uses same `cc` reference — works correctly
