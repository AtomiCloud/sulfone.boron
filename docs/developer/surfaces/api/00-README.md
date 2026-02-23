# HTTP API Overview

Boron exposes a REST API on port 9000 for build orchestration, template warming, session management, and template proxying.

## Base URL

```text
http://localhost:9000
```

**Key File**: `server.go:28` → `server()` function

## All Endpoints

| Method | Path                                            | Description                                    | Key File        |
| ------ | ----------------------------------------------- | ---------------------------------------------- | --------------- |
| GET    | `/`                                             | Health check                                   | `server.go:30`  |
| POST   | `/executor`                                     | Start a new execution session                  | `server.go:183` |
| POST   | `/executor/:sessionId`                          | Execute merge and get results                  | `server.go:68`  |
| DELETE | `/executor/:sessionId`                          | Clean up session resources                     | `server.go:34`  |
| POST   | `/executor/:sessionId/warm`                     | Warm session with images and volumes           | `server.go:248` |
| POST   | `/template/warm`                                | Warm template (pre-pull images, create volume) | `server.go:312` |
| POST   | `/proxy/template/:cyanId/api/template/init`     | Proxy to template init endpoint                | `server.go:371` |
| POST   | `/proxy/template/:cyanId/api/template/validate` | Proxy to template validate endpoint            | `server.go:437` |
| POST   | `/merge/:sessionId`                             | Internal merge endpoint                        | `server.go:503` |
| POST   | `/zip`                                          | Create tar.gz from directory                   | `server.go:531` |

## Common Response Formats

### Success Response

```json
{
  "status": "OK"
}
```

### Error Response (Problem Details)

```json
{
  "title": "Error title",
  "status": 400,
  "detail": "Detailed error message",
  "type": "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/400",
  "trace_id": null,
  "data": ["error1", "error2"]
}
```

**Key File**: `model.go:12` → `ProblemDetails`

## Authentication

No authentication is currently implemented. The API assumes trusted network access.

## Related

- [Getting Started](../../01-getting-started.md) - Usage examples
- [Server Module](../../modules/01-server.md) - Implementation details
