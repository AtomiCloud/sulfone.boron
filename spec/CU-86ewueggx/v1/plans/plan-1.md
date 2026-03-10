# Plan 1: Implement Resolver Support in MERGER

**Goal:** Enable MERGER to detect conflicts and call resolvers to intelligently merge conflicting files.

**Scope:** Add resolver conflict detection and resolution to the MERGER component.

## Files to Modify

| File | Changes |
|------|---------|
| `docker_executor/model.go` | Add resolver request/response data structures |
| `docker_executor/merger.go` | Rewrite `MergeFiles()` to detect conflicts, match resolvers, and resolve |
| `go.mod` | Add `github.com/bmatcuk/doublestar/v4` dependency |

## Implementation Approach

### Step 0: Add Dependency

**Goal:** Add globstar-compatible glob library.

**In `go.mod`:**

```
require github.com/bmatcuk/doublestar/v4 v4.x.x
```

Run `go mod tidy` to add the dependency.

**Why:** Go's `path/filepath.Match()` does NOT support `**` globstar. Helium uses `glob` v11 (Node.js) and Iridium uses `glob` v0.3 (Rust) - both support `**`. Using `doublestar` ensures compatibility.

### Step 1: Add Resolver Models

**Goal:** Define data structures for resolver communication.

**In `docker_executor/model.go`:**

Add the following structs after existing model definitions:

```go
// ResolverFile represents a file version sent to resolver
type ResolverFile struct {
    Path    string        `json:"path"`
    Content string        `json:"content"`
    Origin  ResolverOrigin `json:"origin"`
}

// ResolverOrigin tracks where a file version came from
type ResolverOrigin struct {
    Template string `json:"template"`  // Processor reference or name
    Layer    int    `json:"layer"`     // Layer order (0 = bottom)
}

// ResolverRequest is sent to resolver container
type ResolverRequest struct {
    Config interface{}    `json:"config"`  // From resolver.Config
    Files  []ResolverFile `json:"files"`   // All versions of ONE conflicting file
}

// ResolverResponse is received from resolver container
type ResolverResponse struct {
    Path    string `json:"path"`
    Content string `json:"content"`
}
```

### Step 2: Rewrite MergeFiles with Conflict Resolution

**Goal:** Detect conflicts, match resolvers, and resolve using LWW or resolver calls.

**In `docker_executor/merger.go`, replace `MergeFiles()` implementation:**

**High-level approach:**

1. Build a map of all files from all processor output directories
2. Identify conflicts (files appearing in 2+ directories)
3. For each conflict:
   - Match against resolver glob patterns using `doublestar.Match()`
   - Determine action: resolve (1 match), LWW (0 match), ERROR (2+ matches)
   - Execute action
4. For non-conflicts: Copy directly to merge directory

**Detailed approach:**

```go
func (m Merger) MergeFiles(fromDirs []string, mergeDir string) error {
    // Step 1: Collect all files from all processor outputs
    fileMap := make(map[string][]processorFile)  // path -> list of versions
    for i, dir := range fromDirs {
        walk all files in dir, add to fileMap with processor index as layer
    }

    // Step 2: Identify conflicts and non-conflicts
    var conflicts []string
    var nonConflicts []string
    for path, versions := range fileMap {
        if len(versions) > 1 {
            conflicts = append(conflicts, path)
        } else {
            nonConflicts = append(nonConflicts, path)
        }
    }

    // Step 3: Handle each conflict
    for _, conflictPath := range conflicts {
        // Match resolvers using doublestar.Match()
        matchingResolver := findMatchingResolver(conflictPath, m.Template.Resolvers)

        if len(matchingResolver) == 0 {
            // LWW: use last version
            lastVersion := fileMap[conflictPath][len(fileMap[conflictPath])-1]
            copy last version to mergeDir
        } else if len(matchingResolver) == 1 {
            // Call resolver with all versions
            resolver := matchingResolver[0]
            files := buildResolverFiles(conflictPath, fileMap[conflictPath])
            request := ResolverRequest{
                Config: resolver.Config,
                Files:  files,
            }
            response, err := callResolver(resolver.ID, request)
            if err != nil { return err }
            write resolved content to mergeDir
        } else {
            // Multiple resolvers match - ERROR
            return fmt.Errorf("multiple resolvers match conflicting file '%s': %+v",
                conflictPath, getResolverIDs(matchingResolver))
        }
    }

    // Step 4: Copy non-conflicts
    for _, path := range nonConflicts {
        copy directly to mergeDir
    }

    return nil
}
```

**Helper functions:**

```go
func findMatchingResolver(path string, resolvers []ResolverRes) []ResolverRes {
    // Use doublestar.Match() for globstar-compatible matching
    // Return all resolvers that match the file path
}

func buildResolverFiles(path string, versions []processorFile) []ResolverFile {
    // Read file contents from each processor output
    // Build ResolverFile structs with layer info
}

func callResolver(resolverID string, req ResolverRequest) (*ResolverResponse, error) {
    // HTTP POST to resolver container
    // Container name: "cyan-resolver-{id-without-dashes}"
    // Endpoint: "http://...:5553/api/resolve"
}
```

### Step 3: Glob Pattern Matching

**Use `doublestar.Match()` for resolver pattern matching.**

```go
import "github.com/bmatcuk/doublestar/v4"

func findMatchingResolver(path string, resolvers []ResolverRes) []ResolverRes {
    var matches []ResolverRes
    for _, resolver := range resolvers {
        for _, pattern := range resolver.Files {
            matched, err := doublestar.Match(pattern, path)
            if err == nil && matched {
                matches = append(matches, resolver)
                break  // One pattern match per resolver is enough
            }
        }
    }
    return matches
}
```

**Globstar support:** `doublestar.Match()` supports `**` for recursive matching, compatible with:
- Helium's `glob` package (Node.js)
- Iridium's `glob` crate (Rust)

**Examples of expected behavior:**
- `**/*.json` - matches any `.json` file at any depth (recursive)
- `*.json` - matches any `.json` file in current directory
- `package.json` - matches exactly `package.json`
- `config/*.yaml` - matches YAML files in `config/` directory
- `src/**/*.go` - matches `.go` files in `src/**/` (recursive)

### Step 4: Error Handling

All errors fail entire merge:

| Error | Message |
|-------|----------|
| Multiple resolvers match | "Multiple resolvers match conflicting file '{path}': [id1, id2]. Template resolver configuration may be misconfigured." |
| Resolver container not running | "Resolver container for '{id}' not found or not running" |
| Resolver call fails (non-200) | "Resolver call failed for '{file-path}': {status} {body}" |
| Resolver returns wrong path | "Resolver returned invalid path: expected '{expected}', got '{actual}'" |
| File read error | "Failed to read file '{path}': {error}" |

## Edge Cases

### Empty Processors
- If `fromDirs` is empty, succeed with no output
- If all directories are empty, create empty merge directory

### Identical File Contents
- Even if file contents are identical across processors, still call resolver
- Let resolver decide how to handle identical versions

### Nested File Paths
- File paths include directory structure (e.g., `src/config.json`)
- Conflicts determined by full path (not just filename)
- Copy maintains directory structure in merge output

### Resolvers with Empty Config
- Resolver with `Config: null` or `Config: {}` is valid
- Pass empty config object in request

### Many Conflicts
- Handle large numbers of conflicts efficiently
- One resolver call per conflict (not parallel - fail-fast on errors)

### Non-Matching Resolver Pattern
- LWW behavior logged at info level
- Maintain backward compatibility

## Integration Points

### No Integration with Other Components

- Merger is self-contained - no coordination needed with other components
- Plugins work on resolved output (no changes needed)
- Template warming already handles resolver containers (CU-86ewrbr3t)

### Existing Code Reuse

- `copyFile()` function can be reused for non-conflict files
- `PostJSON[Req, Res]` generic function already exists for HTTP calls
- Container naming from `DockerContainerToString()` and `DockerContainerReference`

## Testing Strategy

Since no test infrastructure exists, verify manually:

1. **Test LWW (no resolvers):**
   - Create template with `resolvers: []`
   - Create processors with conflicting outputs
   - Run merge
   - Verify last processor's file wins

2. **Test resolver call (1 resolver):**
   - Create template with 1 resolver matching `**/*.json`
   - Create processors with conflicting `package.json`
   - Run merge
   - Verify resolver is called and resolved content is written

3. **Test multiple conflicts:**
   - Create template with 1 resolver matching `**/*.json`
   - Create processors with conflicts on `package.json`, `tsconfig.json`
   - Run merge
   - Verify 2 separate resolver calls are made

4. **Test multiple resolvers (ERROR):**
   - Create template with 2 overlapping resolvers for `*.json`
   - Create processors with conflicting `package.json`
   - Run merge
   - Verify error is returned

5. **Test glob patterns:**
   - Verify `**/*.json` matches recursively (e.g., `src/config.json`)
   - Verify `*.json` matches `.json` files in current directory only
   - Verify nested patterns like `config/*.yaml` work

6. **Test globstar compatibility:**
   - Use same patterns as Helium/Iridium examples
   - Verify `**/tsconfig.json` matches at any depth
   - Verify `**/*.json` matches all `.json` files recursively

## Implementation Checklist

### Code Changes

- [ ] Add `github.com/bmatcuk/doublestar/v4` to `go.mod`
- [ ] Add `ResolverFile` struct to `model.go`
- [ ] Add `ResolverOrigin` struct to `model.go`
- [ ] Add `ResolverRequest` struct to `model.go`
- [ ] Add `ResolverResponse` struct to `model.go`
- [ ] Rewrite `MergeFiles()` in `merger.go`
- [ ] Implement conflict detection (map building)
- [ ] Implement `findMatchingResolver()` with `doublestar.Match()`
- [ ] Implement resolver container name construction
- [ ] Implement `callResolver()` with HTTP POST
- [ ] Implement LWW fallback for non-matched conflicts
- [ ] Implement error for multiple resolver matches
- [ ] Handle and propagate resolver errors
- [ ] Copy non-conflict files directly

### Testing

- [ ] Manual test: LWW behavior (no resolvers)
- [ ] Manual test: Single resolver call
- [ ] Manual test: Multiple conflicts with 1 resolver
- [ ] Manual test: Multiple resolvers match (ERROR case)
- [ ] Manual test: Glob pattern matching behavior
- [ ] Manual test: Globstar recursive matching (`**/*.json`)
- [ ] Manual test: Pattern compatibility with Helium/Iridium
