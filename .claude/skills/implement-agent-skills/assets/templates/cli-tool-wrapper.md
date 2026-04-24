---
name: cli-wrapper-name
description: Wrap [CLI Tool Name] for agent use. Use when [trigger scenarios].
---

# [CLI Tool Name] Wrapper

## Purpose

This skill provides a standardized way for the agent to interact with `[cli-command]`.

## 1. Safety & Verification

Before executing any commands, the agent must:

1.  **Check Help**: Run `[cli-command] --help` to verify available flags and syntax.
2.  **Verify Version**: Run `[cli-command] --version` to ensure compatibility.

## 2. Common Workflows

### Workflow: [Workflow Name]

1.  [Step 1]
2.  [Step 2]

## 3. Error Handling

- If [error pattern] occurs: [mitigation step]
- If [error pattern] occurs: [mitigation step]

## 4. Examples

### Example: [Scenario]

**Command**: `[cli-command] [args]`
**Expected Output**: [Description of success]
