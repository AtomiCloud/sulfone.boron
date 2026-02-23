# Executor API

## POST /executor

Start a new execution session. Creates and starts processor, plugin, and merger containers.

**Key File**: `server.go:183`

### Request Body

```json
{
  "session_id": "my-session",
  "template": {
    "principal": { ... },
    "plugins": [ ... ],
    "processors": [ ... ]
  },
  "write_vol_reference": {
    "cyan_id": "template-uuid",
    "session_id": "my-session"
  },
  "merger": {
    "merger_id": "merger-uuid"
  }
}
```

| Field                 | Type                    | Required | Description                                     |
| --------------------- | ----------------------- | -------- | ----------------------------------------------- |
| `session_id`          | `string`                | Yes      | Unique identifier for this execution            |
| `template`            | `TemplateVersionRes`    | Yes      | Template definition with processors and plugins |
| `write_vol_reference` | `DockerVolumeReference` | Yes      | Session volume reference                        |
| `merger_id`           | `string`                | Yes      | Merger container ID                             |

### Response 200 OK

```json
{
  "status": "OK"
}
```

### Response 503 Service Unavailable

```json
{
  "title": "Failed to start executor",
  "status": 503,
  "detail": "Failed to start cyanprint executor",
  "type": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/503",
  "trace_id": null,
  "data": ["error1", "error2"]
}
```

## POST /executor/:sessionId

Execute the merge pipeline and stream zipped output.

**Key File**: `server.go:68`

### Request Body

```json
{
  "template": {
    "principal": { ... },
    "plugins": [ ... ],
    "processors": [ ... ]
  },
  "merger_id": "merger-uuid"
}
```

| Field       | Type                 | Required | Description         |
| ----------- | -------------------- | -------- | ------------------- |
| `template`  | `TemplateVersionRes` | Yes      | Template definition |
| `merger_id` | `string`             | Yes      | Merger container ID |

### Response 200 OK

Returns `application/x-gzip` stream with header `Content-Disposition: attachment; filename=cyan-output.tar.gz`

### Response 400 Bad Request

```json
{
  "title": "Failed to clean",
  "status": 400,
  "detail": "Failed to session session-id",
  "type": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
  "trace_id": null,
  "data": ["error1", "error2"]
}
```

### Response 503 Service Unavailable

```json
{
  "title": "Failed to contract upstream server",
  "status": 503,
  "detail": "Error contacting upstream (merger) server for zipping",
  "type": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/503",
  "trace_id": null,
  "data": ["error1"]
}
```

## DELETE /executor/:sessionId

Clean up session resources (containers and volumes).

**Key File**: `server.go:34`

### Response 200 OK

```json
{
  "status": "OK"
}
```

### Response 400 Bad Request

```json
{
  "title": "Failed to clean",
  "status": 400,
  "detail": "Failed to session session-id",
  "type": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
  "trace_id": null,
  "data": ["error1", "error2"]
}
```

## POST /executor/:sessionId/warm

Warm a session by pulling images and creating the session volume.

**Key File**: `server.go:248`

### Request Body

```json
{
  "principal": { ... },
  "plugins": [ ... ],
  "processors": [ ... ]
}
```

| Field      | Type                 | Required | Description         |
| ---------- | -------------------- | -------- | ------------------- |
| `template` | `TemplateVersionRes` | Yes      | Template definition |

### Response 200 OK

```json
{
  "session_id": "my-session",
  "vol_ref": {
    "cyan_id": "template-uuid",
    "session_id": "my-session"
  }
}
```

### Response 400 Bad Request

```json
{
  "title": "Failed to warm executor",
  "status": 400,
  "detail": "Failed to warn executor image, templates, and volumes",
  "type": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
  "trace_id": null,
  "data": ["error1", "error2"]
}
```

## Related

- [Session Management Feature](../../features/01-session-management.md) - Session lifecycle details
- [Warming System Feature](../../features/07-warming-system.md) - Warm operation details
