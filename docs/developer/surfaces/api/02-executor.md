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
| `merger`              | `MergerReq`             | Yes      | Merger container reference                      |

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
  "detail": "Failed to clean <session-id>",
  "type": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
  "trace_id": null,
  "data": ["error1", "error2"]
}
```

### Response 503 Service Unavailable

```json
{
  "title": "Failed to contact upstream server",
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
  "detail": "Failed to clean <session-id>",
  "type": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
  "trace_id": null,
  "data": ["error1", "error2"]
}
```

## POST /executor/:sessionId/warm

Warm a session by pulling images and creating the session volume.

**Key File**: `server.go:248`

### Request Body

> **Note:** The JSON below is illustrative. See `TemplateVersionRes` type definition for full field schema, which includes `id`, `version`, `properties`, `principal`, `plugins`, and `processors`.

```json
{
  "template": { "id": "...", "version": 1, "properties": {} }
}
```

| Field      | Type                 | Required | Description                                                                   |
| ---------- | -------------------- | -------- | ----------------------------------------------------------------------------- |
| `template` | `TemplateVersionRes` | Yes      | Template definition (id, version, properties, principal, plugins, processors) |

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
  "detail": "Failed to warm executor image, templates, and volumes",
  "type": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
  "trace_id": null,
  "data": ["error1", "error2"]
}
```

### Response 503 Service Unavailable

```json
{
  "title": "Failed to configure network",
  "status": 503,
  "detail": "Failed to start cyanprint Docker bridge network",
  "type": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/503",
  "trace_id": null,
  "data": ["error1"]
}
```

## POST /executor/try

Setup a try/test session for local testing. Creates blob volume from local image or path, session volume, pulls missing images, and warms resolvers. The merger is started later via `POST /executor/:sessionId`.

**Key File**: `server.go` (try_executor.go for implementation)

### Request Body

```json
{
  "session_id": "my-session",
  "local_template_id": "local-abc123",
  "source": "image",
  "image_ref": {
    "reference": "my-template",
    "tag": "latest"
  },
  "template": {
    "principal": { ... },
    "plugins": [ ... ],
    "processors": [ ... ],
    "resolvers": [ ... ]
  },
  "merger_id": "merger-uuid"
}
```

| Field               | Type                   | Required    | Description                                                   |
| ------------------- | ---------------------- | ----------- | ------------------------------------------------------------- |
| `session_id`        | `string`               | Yes         | Unique session identifier (generated by Iridium)              |
| `local_template_id` | `string`               | Yes         | Synthetic local template ID (e.g., `local-{uuid}`)            |
| `source`            | `string`               | No          | Blob source: `"image"` (default) or `"path"` (for --dev mode) |
| `image_ref`         | `DockerImageReference` | Conditional | Required when `source="image"`. Local Docker image reference  |
| `path`              | `string`               | Conditional | Required when `source="path"`. Host path to copy files from   |
| `template`          | `TemplateVersionRes`   | Yes         | Template definition with processors, plugins, and resolvers   |
| `merger_id`         | `string`               | Yes         | Merger container ID                                           |

### Response 200 OK

```json
{
  "session_id": "my-session",
  "blob_volume": {
    "cyan_id": "local-abc123",
    "session_id": ""
  },
  "session_volume": {
    "cyan_id": "local-abc123",
    "session_id": "my-session"
  }
}
```

### Response 400 Bad Request

**Invalid request body:**

```json
{
  "title": "Failed to bind request",
  "status": 400,
  "detail": "Request body does not match TryExecutorReq",
  "type": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
  "trace_id": null,
  "data": ["error message"]
}
```

**Invalid source type:**

```json
{
  "title": "Invalid source type",
  "status": 400,
  "detail": "source must be 'image' or 'path', got 'invalid'",
  "type": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
  "trace_id": null,
  "data": null
}
```

**Missing image_ref when source is "image":**

```json
{
  "title": "Missing image_ref",
  "status": 400,
  "detail": "image_ref is required when source is 'image'",
  "type": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
  "trace_id": null,
  "data": null
}
```

### Response 500 Internal Server Error

```json
{
  "title": "Failed to setup try session",
  "status": 500,
  "detail": "Try setup failed",
  "type": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/500",
  "trace_id": null,
  "data": ["error1", "error2"]
}
```

### Response 503 Service Unavailable

```json
{
  "title": "Failed to configure network",
  "status": 503,
  "detail": "Failed to start cyanprint Docker bridge network",
  "type": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/503",
  "trace_id": null,
  "data": ["error1"]
}
```

### Behavior Notes

- **Blob volume** (`cyan-{local_template_id}`) is shared across sessions and created idempotently
- **Session volume** (`cyan-{local_template_id}-{session_id}`) is unique per session
- **Session collision**: Returns error if session volume already exists
- **Source modes**:
  - `image`: Extracts blob from local Docker image
  - `path`: Copies files from host path (for `--dev` mode)
- Cleanup via `DELETE /executor/:sessionId` cleans session volume, preserves blob volume for reuse

## Related

- [Session Management Feature](../../features/01-session-management.md) - Session lifecycle details
- [Warming System Feature](../../features/07-warming-system.md) - Warm operation details
