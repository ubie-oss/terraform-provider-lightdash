---
name: investigate-deprecated-apis
description: Accountability for identifying deprecated Lightdash API usage in the provider by comparing code patterns against the official OpenAPI specification. Use when tasked to audit the codebase for technical debt or outdated API dependencies.
---

# Investigate Deprecated Lightdash APIs

## Responsibility

To provide a systematic methodology for identifying technical debt related to Lightdash API deprecations. This skill defines the "what" and "why" of the audit process, ensuring the provider remains resilient against breaking API changes.

## Deprecation Criteria

An endpoint is considered a candidate for migration if:

- It is explicitly marked `deprecated: true` in the OpenAPI specification.
- It resides in `/api/v1/` and a corresponding functional equivalent exists in `/api/v2/`.
- Official documentation recommends an alternative workflow.

## Audit Methodology

### 1. Pattern Discovery

Identify potential API usage by searching `internal/lightdash/api/` for versioned path patterns.

- **Tools**: Use `grep` or `rg` to find `"/api/v1/` or `"/api/v2/`.
- **Target**: Build a unique list of (HTTP Method, Path Template).

### 2. Spec Cross-Reference

Compare identified patterns against the official Lightdash OpenAPI source of truth.

- **Source**: `https://raw.githubusercontent.com/lightdash/lightdash/refs/heads/main/packages/backend/src/generated/swagger.json`
- **Action**: Check each used path for the `deprecated` flag.

### 3. Impact Analysis

For each deprecated endpoint, determine:

- Which Terraform resource(s) or data source(s) depend on it.
- The severity of the deprecation (is there a sunset date?).
- The complexity of the migration (e.g., role string to role UUID).

## References

- [Lightdash API Reference](https://docs.lightdash.com/api-reference/v1/introduction)
- [Migration Strategies](references/migration-strategies.md) - Common mapping patterns.
- [Research Lightdash API](../research-lightdash-api/SKILL.md) - For deep-dive verification of individual endpoints.
