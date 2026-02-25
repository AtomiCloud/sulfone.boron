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

> **Note:** Port 9000 here refers to the Boron coordinator/server process, not a Merger container (which also uses port 9000 for its own health endpoint).

```bash
curl http://localhost:9000/
```

Response:

```json
{ "status": "OK" }
```
