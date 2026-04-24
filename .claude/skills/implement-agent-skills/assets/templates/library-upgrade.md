---
name: lib-upgrade-name
description: Safe workflow for upgrading [Library/Package Name]. Use when upgrading [Library Name].
---

# Library Upgrade: [Library Name]

## Purpose

Standardized workflow to safely upgrade `[library-name]` with minimal risk of regression.

## 1. Pre-Upgrade Analysis

1.  **Check Current Version**: `[command to check version]`
2.  **Read Changelog**: Identify breaking changes or deprecated APIs.
3.  **Audit Usages**: Find all locations where the library is used in the project.

## 2. Upgrade Workflow

1.  **Update Dependency**: `[command to update package]`
2.  **Synchronize Lockfile**: `[command to sync lockfile]`
3.  **Fix Breaking Changes**: Refactor code according to the changelog.
4.  **Run Tests**: Ensure all tests pass.

## 3. Validation Checklist

- [ ] No compilation/build errors.
- [ ] All unit tests pass.
- [ ] Manual smoke test of affected features.

## 4. Rollback Plan

- `[command to revert changes]`
- `[command to restore previous version]`
