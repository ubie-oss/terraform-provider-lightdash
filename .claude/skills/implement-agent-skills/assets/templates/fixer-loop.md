---
name: fixer-loop-name
description: Automatic fix-and-verify loop for [Error Type/Lint]. Use when fixing recurring errors.
---

# Fixer Loop: [Error Type]

## Purpose

An autonomous loop for the agent to identify, fix, and verify [Error Type] issues.

## Loop Logic

1.  **Identify**: Run `[test/lint command]` to list current failures.
2.  **Analyze**: Examine the error message and the failing code.
3.  **Fix**: Apply the minimum necessary change to resolve the error.
4.  **Verify**: Re-run `[test/lint command]`.
    - If passed: Move to next issue.
    - If failed: Analyze new failure and repeat loop.

## Termination Criteria

- No more errors reported by `[command]`.
- Reached max iteration limit (default: 5).

## Examples

### Scenario: [Example Case]

1.  `npm test` fails with [error].
2.  Agent edits `src/logic.js`.
3.  `npm test` passes.
