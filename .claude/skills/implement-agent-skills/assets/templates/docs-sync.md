---
name: docs-sync-name
description: Keep [Documentation Path] in sync with [Source Path]. Use when code changes affect documentation.
---

# Documentation Sync: [Target]

## Purpose

Ensure that `[target-docs]` accurately reflect the current state of `[source-code]`.

## Workflow

1.  **Detect Changes**: Identify files modified in `[source-path]`.
2.  **Analyze Impact**: Determine which documentation pages are affected.
3.  **Update Docs**: Revise content, examples, and signatures in `[target-path]`.
4.  **Verify Links**: Ensure all internal and external links are still valid.

## Sync Checklist

- [ ] Code examples match the latest implementation.
- [ ] Function/API signatures are up to date.
- [ ] No broken links or outdated references.

## Examples

### Scenario: [Example Sync]

- **Source**: `src/api/auth.js` (added new field `role`)
- **Doc**: `docs/api/authentication.md` (updated field table and example JSON)
