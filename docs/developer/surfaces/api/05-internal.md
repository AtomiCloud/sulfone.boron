# Internal API

These endpoints are used internally by Boron during the merge pipeline.

## POST /merge/:sessionId

Merge output directories from multiple processor containers into a single directory. Called by the merger during the 3-stage pipeline.

**Key File**: `server.go:503`

### Parameters

| Name        | In   | Type     | Required | Description        |
| ----------- | ---- | -------- | -------- | ------------------ |
| `sessionId` | path | `string` | Yes      | Session identifier |

### Request Body

```json
{
  "from_dirs": ["/workspace/area/uuid1", "/workspace/area/uuid2"],
  "to_dir": "/workspace/area/merge-uuid",
  "template": { ... }
}
```

| Field       | Type                 | Required | Description                  |
| ----------- | -------------------- | -------- | ---------------------------- |
| `from_dirs` | `[]string`           | Yes      | Processor output directories |
| `to_dir`    | `string`             | Yes      | Merge destination directory  |
| `template`  | `TemplateVersionRes` | Yes      | Template definition          |

### Response 200 OK

```json
{
  "status": "OK"
}
```

### Response 400 Bad Request

```json
{
  "error": ["error message"]
}
```

## POST /zip

Create a tar.gz archive from a directory and stream it to the client. Called after merge to deliver final results.

**Key File**: `server.go:531`

### Request Body

```json
{
  "target_dir": "/workspace/area/merge-uuid"
}
```

| Field        | Type     | Required | Description      |
| ------------ | -------- | -------- | ---------------- |
| `target_dir` | `string` | Yes      | Directory to zip |

### Response 200 OK

Returns `application/x-gzip` stream with tar.gz archive.

Headers:

- `Content-Disposition: attachment; filename=cyan-output.tar.gz`
- `Content-Type: application/x-gzip`

### Response 400 Bad Request

```json
{
  "error": ["target directory not found or unreadable"]
}
```

### Response 500 Internal Server Error

```json
{
  "error": ["failed to create archive: permission denied"]
}
```

## Related

- [Merger System Feature](../../features/03-merger-system.md) - 3-stage pipeline details
- [File Merging Algorithm](../../algorithms/03-file-merging.md) - Merge implementation
