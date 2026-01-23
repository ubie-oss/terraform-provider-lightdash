---
name: security-review
description: Comprehensive security review of the Terraform provider codebase using gosec, trunk, osv-scanner, and trivy.
---

# Security Review

This skill provides a systematic workflow for auditing the security of the Terraform Provider for Lightdash. It focuses on Go source code, dependencies, and secrets.

## Prerequisites

Ensure the following tools are installed and available in the PATH:

- `gosec` (Go Security Checker)
- `trunk` (Integrated linting and security platform)
- `osv-scanner` (Open Source Vulnerability Scanner)
- `trivy` (Vulnerability and secret scanner)

## Workflow

### 1. Static Analysis (SAST) for Go

Run `gosec` to identify common security pitfalls in Go code.

```bash
gosec ./internal/...
```

### 2. Integrated Security Check (via Trunk)

Run security-scoped checks using `trunk`. This includes tools like Semgrep and Gitleaks. We explicitly exclude `checkov` as per project preference.

```bash
trunk check --all --scope security --exclude checkov
```

### 3. Open Source Vulnerability Scan

Use `osv-scanner` to check for vulnerabilities in Go dependencies.

```bash
osv-scanner scan source -r .
```

### 4. Filesystem, Secret, and Config Scan

Use `trivy` to scan the filesystem for vulnerabilities, secrets, and misconfigurations (excluding Checkov-style IaC scans if redundant).

```bash
trivy fs . --scanners vuln,secret,config
```

## Reporting

After running these tools, synthesize the findings into a report with the following sections:

### 1. Summary of Findings

- Total issues found per tool.
- Breakdown by severity (Critical, High, Medium, Low).

### 2. Detailed Findings

For each significant finding, provide:

- **Tool**: Which tool reported it.
- **Location**: File path and line number.
- **Severity**: Reported severity.
- **Description**: Brief explanation of the risk.
- **Remediation**: Recommended fix or mitigation.

### 3. False Positives

Document any findings that are determined to be false positives and why.

### 4. Remediation Plan

Prioritize fixes based on severity and impact.

## Usage Note

Always run these checks before major releases or after significant architectural changes.
