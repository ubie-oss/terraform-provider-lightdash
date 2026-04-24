---
name: security-audit-name
description: Perform security audit for [Project/Component]. Use when checking for vulnerabilities.
---

# Security Audit: [Project]

## Purpose

Systematically scan and identify security vulnerabilities in [Project].

## Audit Workflow

1.  **Dependency Scan**: Run `[scan-tool]` to find vulnerable packages.
2.  **Static Analysis**: Run `[sast-tool]` to find code-level issues.
3.  **Manual Review**: Check critical paths (auth, data access) for common flaws.

## Reporting Format

Use this template for each finding:

- **Severity**: [High/Medium/Low]
- **Description**: [What is the issue?]
- **Impact**: [What happens if exploited?]
- **Recommendation**: [How to fix it?]

## Tooling

- `[tool 1]`: [usage]
- `[tool 2]`: [usage]
