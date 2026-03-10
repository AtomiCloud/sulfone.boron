# Task Spec: Allow MERGER to use resolver in Boron

**Ticket**: CU-86ewueggx
**Version**: 1
**Parent**: none

## Overview

Currently, MERGER in Boron uses a simple Last-Write-Wins (LWW) strategy when multiple processors output the same file. This spec defines changes to enable MERGER to call resolver containers to intelligently merge conflicting files.

## Context

### Current Behavior

When multiple processors output the same file path (e.g., `package.json`), MERGER uses Last-Write-Wins: the last processor's output overwrites all previous versions. This can cause data loss when processors contribute different sections to the same file.

### Desired Behavior

The MERGER should:
1. Detect all file conflicts across all processor outputs
2. For each unique conflicted file path, determine if a resolver applies
3. Call the applicable resolver ONCE per conflicted file with ALL versions
4. Apply the resolved result to merge output
5. Continue with plugin execution on the resolved output

### Single-Template Scope

**IMPORTANT:** This implementation handles conflicts within a single template where one template author configures all resolvers. Cross-template conflicts (different template authors configuring different resolvers) are Iridium's responsibility.

**Implication:** If multiple resolvers could match the same conflicted file, this indicates a template configuration mistake → ERROR (not a fallback scenario like Iridium's cross-template LWW).

## Acceptance Criteria

### AC1: Detect All Conflicts

The MERGER must identify all files that appear in multiple processor outputs before merging.

- For each unique file path across all processor outputs, track which processors produced it
- A file is "conflicting" if it appears in 2+ processor output directories
- Track processor order for layering information (first processor = layer 0, bottom layer)

### AC2: Match Resolver to Conflicts

For each conflicted file, determine which resolver (if any) should handle it.

- Match file paths against resolver `Files` glob patterns
- Track which resolver(s) match each conflicted file

**Resolver matching rules:**

| Matchers | Action |
|----------|--------|
| Exactly 1 resolver | Include this file in resolver call |
| 0 resolvers | Exclude from resolver call; use LWW |
| 2+ resolvers | ERROR (template misconfiguration) |

**Rationale for ERROR on 2+ resolvers:** Single-template scope means one author with perfect information configured overlapping patterns → mistake, not ambiguity.

### AC3: Single Resolver Call Per Conflicted File

For each conflicted file, make ONE resolver call containing ALL processor layers for that file.

- Collect all versions of ONE conflicted file from all processors that produced it
- Include layer information for each version (processor order determines layer)
- Call applicable resolver ONCE per conflicted file with all versions
- Apply resolved result to merge output
- Repeat for each unique conflicted file path

**Example:** If `package.json` appears in 3 processors, send 1 request with 3 file versions. If `config.json` also conflicts, send a second 1-request with all its versions.

### AC4: Resolver Request Format

For each conflicted file, the resolver receives a request containing ALL versions of that file:

```json
{
  "config": { <resolver configuration from template> },
  "files": [
    {
      "path": "package.json",
      "content": "{ file content from processor 0 }",
      "origin": {
        "template": "{ processor reference }",
        "layer": 0
      }
    },
    {
      "path": "package.json",
      "content": "{ file content from processor 1 }",
      "origin": {
        "template": "{ processor reference }",
        "layer": 1
      }
    },
    {
      "path": "package.json",
      "content": "{ file content from processor 2 }",
      "origin": {
        "template": "{ processor reference }",
        "layer": 2
      }
    }
  ]
}
```

All versions of ONE conflicted file per request.

### AC5: Resolver Response Format

The resolver returns resolved content for the conflicted file.

**Expected:** Resolved content for the input file path.

The MERGER must apply the resolved result to the merge output directory.

### AC6: Handle Non-Matched Conflicts (LWW)

Conflicted files that match ZERO resolvers use LWW.

- For each conflicted file with no resolver match, use the last processor's version
- Log info-level message when LWW is used
- Maintain backward compatibility for templates without resolvers

### AC7: Non-Conflict Files

Files that appear in only ONE processor output are not involved in resolution.

- Copy directly to merge output
- No resolver involvement
- Maintains current behavior

### AC8: Error Handling

Resolver failures fail the entire merge operation.

| Error Scenario | Action |
|----------------|--------|
| Multiple resolvers match same file | ERROR: fail merge with descriptive message |
| Resolver container not running | ERROR: fail merge |
| Resolver call fails (non-200, timeout) | ERROR: fail merge |
| Resolver returns invalid/unexpected data | ERROR: fail merge |

**No fallback to LWW on resolver errors** - fail fast to surface configuration issues.

### AC9: Plugin Integration

Plugins execute AFTER conflict resolution completes.

- Plugins receive fully-resolved merge output (no conflicts remain)
- Plugins transform resolved output as before
- No changes to plugin behavior or API

### AC10: Backward Compatibility

Templates without resolvers work exactly as before.

- Templates with `resolvers: []` → LWW for all conflicts (current behavior)
- Templates with resolvers but no conflicts → no resolver calls (current behavior)
- Only templates with resolvers AND conflicts → new behavior

## Constraints

### Deterministic Layering

Processor order determines layer order:
- First processor in `req.Cyan.Processors` = layer 0 (bottom)
- Last processor = layer N (top)
- Layer information is included in resolver request for merge strategy

### Glob Pattern Matching

Resolver `Files` field uses glob patterns.

**Use `path/filepath.Match()` from Go standard library.**

- This provides POSIX-compliant glob matching
- Supported patterns: `*` (matches any sequence), `?` (matches single character), `[...]` (character class)
- Pattern matching is applied to full file paths (not just filenames)
- Examples:
  - `**/*.json` - Note: `**` has no special meaning in `filepath.Match()`, matches literal `**/` followed by any `.json` file
  - `package.json` - matches only `package.json`
  - `*.json` - matches any `.json` file in current directory
  - `config/*.yaml` - matches YAML files in `config/` directory
  - `src/**/*.go` - matches `.go` files in `src/**/` directory (literal `**`)

**IMPORTANT:** `filepath.Match()` does NOT support recursive `**` globstar. Users must use explicit patterns like `*.json`, `src/*.json`, or specify full paths.

### Single Resolver Scope

This spec assumes one resolver handles all matched conflicts in a template.

- Template authors configure one resolver (or multiple non-overlapping resolvers)
- Overlapping resolver patterns are configuration errors (AC8)
- Cross-template resolver selection is out of scope (Iridium's responsibility)

### Stateless Resolvers

Resolvers are stateless containers warmed alongside templates.

- No file system access within resolver containers
- All file data passed as strings in request/response
- Resolver containers are on the `cyanprint` network

## Out of Scope

- Cross-template conflict resolution (Iridium)
- Resolver SDK implementation (Helium)
- Resolver container orchestration/warming (already implemented)
- Plugin behavior changes
- Testing infrastructure

## Dependencies

- Resolvers must be warmed before merge (implemented in CU-86ewrbr3t)
- Resolver containers listen on port 5553 with `/api/resolve` endpoint
- Resolvers accept conflict requests (all file versions for one file)
- Resolvers return resolved content for each conflicted file

## Examples

### Example 1: Successful Resolution

**Template:** Has resolver with `files: ["**/*.json"]`

**Processor outputs:**
- Processor 0: `package.json`, `app.js`, `utils.js`
- Processor 1: `package.json`, `styles.css`, `index.html`
- Processor 2: `config.json`

**Conflicts detected:** `package.json` (Processor 0, 1)

**Resolver call includes:**
- Both versions of `package.json` (from Processor 0 and 1)
- Layer info: Processor 0 = layer 0, Processor 1 = layer 1

**Resolver returns:** Merged `package.json` with both dependencies and devDependencies

**Non-conflicts:** `app.js`, `utils.js`, `styles.css`, `index.html`, `config.json` → direct copy

### Example 2: No Resolver Match (LWW)

**Template:** Has resolver with `files: ["**/*.yaml"]`

**Processor outputs:**
- Processor 0: `config.json`
- Processor 1: `config.json`

**Conflict:** `config.json` doesn't match `**/*.yaml` pattern

**Result:** LWW → Processor 1's `config.json` overwrites Processor 0's

### Example 3: Multiple Resolvers Match (ERROR)

**Template:**
- Resolver A: `files: ["**/*.json"]`
- Resolver B: `files: ["package.json"]`  // Overlaps!

**Conflict:** `package.json` from multiple processors

**Result:** ERROR - both resolvers match, template configuration is invalid

### Example 4: Multiple Conflicts, Separate Requests

**Template:** Has resolver with `files: ["**/*.json"]`

**Processor outputs:**
- Processor 0: `package.json`, `tsconfig.json`, `app.js`
- Processor 1: `package.json`, `tsconfig.json`, `styles.css`

**Conflicts:** `package.json`, `tsconfig.json` (both appear in Processor 0 and 1)

**Resolver calls:**
- Request 1: `package.json` with versions from Processor 0, 1
- Request 2: `tsconfig.json` with versions from Processor 0, 1

**Resolver returns:** Resolved content for each file in separate responses
