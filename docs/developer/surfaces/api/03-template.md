# Template API

## POST /template/warm

Warm a template by pulling images and creating the template volume.

**Key File**: `server.go:312`

### Request Body

```json
{
  "principal": {
    "id": "template-uuid",
    "version": 1,
    "properties": {
      "templateDockerReference": "registry/image",
      "templateDockerTag": "latest",
      "blobDockerReference": "registry/blob-image",
      "blobDockerTag": "latest"
    }
  },
  "template": { ... },
  "plugins": [ ... ],
  "processors": [ ... ]
}
```

| Field      | Type                 | Required | Description                                     |
| ---------- | -------------------- | -------- | ----------------------------------------------- |
| `template` | `TemplateVersionRes` | Yes      | Template definition with processors and plugins |

### Response 200 OK

```json
{
  "status": "OK"
}
```

### Response 400 Bad Request

```json
{
  "title": "Failed to warm template",
  "status": 400,
  "detail": "Failed to warm template image, templates, and volumes",
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

## Related

- [Warming System Feature](../../features/07-warming-system.md) - Warm operation details
