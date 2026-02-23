# Health Check API

## GET /

Check if the Boron server is running.

**Key File**: `server.go:30`

## Response

### 200 OK

```json
{
  "status": "OK"
}
```

## Example

```bash
curl http://localhost:9000/
```

Response:

```json
{ "status": "OK" }
```
