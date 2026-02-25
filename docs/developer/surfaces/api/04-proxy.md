# Template Proxy API

Proxy endpoints forward requests to template containers. These enable external systems to communicate with template containers running in the `cyanprint` network.

## POST /proxy/template/:cyanId/api/template/init

Proxy to template container's init endpoint.

**Key File**: `server.go:371`

### Parameters

| Name     | In   | Type     | Required | Description                                   |
| -------- | ---- | -------- | -------- | --------------------------------------------- |
| `cyanId` | path | `string` | Yes      | Template container ID (UUID, dashes optional) |

### Request Body

Forwarded as-is to the template container.

### Response

Returns the template container's response with original status code and headers.

### Response 502 Bad Gateway

```json
{
  "title": "Upstream failed",
  "status": 502,
  "detail": "Failed to forward request to upstream template",
  "type": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/502",
  "trace_id": null,
  "data": ["upstream connection refused"]
}
```

## POST /proxy/template/:cyanId/api/template/validate

Proxy to template container's validate endpoint.

**Key File**: `server.go:437`

### Parameters

| Name     | In   | Type     | Required | Description                                   |
| -------- | ---- | -------- | -------- | --------------------------------------------- |
| `cyanId` | path | `string` | Yes      | Template container ID (UUID, dashes optional) |

### Request Body

Forwarded as-is to the template container.

### Response

Returns the template container's response with original status code and headers.

### Response 502 Bad Gateway

```json
{
  "title": "Upstream failed",
  "status": 502,
  "detail": "Failed to forward request to upstream template",
  "type": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/502",
  "trace_id": null,
  "data": ["upstream connection refused"]
}
```

## Container Addressing

**Key File**: `domain_model.go:48` → `DockerContainerToString()`

Template containers are addressed via the `cyanprint` network (the UUID has its dashes stripped when constructing the name):

- Container name: `cyan-template-<stripped-uuid>` (e.g., `123e4567-e89b-...` → `123e4567e89b...`)
- HTTP endpoint: `http://cyan-template-<stripped-uuid>:5550`

The proxy constructs the endpoint and forwards the request with the original body and headers.

## Related

- [Network Architecture Feature](../../features/10-network-architecture.md) - Container networking details
- [Template Executor Module](../../modules/02-docker-executor.md) - Template container management
