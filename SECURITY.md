# Security Policy

## Supported Versions

We actively provide security updates for the following versions of the Terraform Provider for Lightdash:

| Version | Supported          |
| ------- | ------------------ |
| 0.5.x   | :white_check_mark: |
| < 0.5.0 | :x:                |

## CI dependency checks

The SBOM workflow runs:

- [govulncheck](https://go.dev/doc/security/vuln/) against the Go module (`./...`)
- [Trivy](https://trivy.dev/docs/latest/guide/supply-chain/sbom/) CycloneDX SBOM generation from the repository, then `trivy sbom` scanning for `HIGH`/`CRITICAL` vulnerabilities (fail-closed)

It triggers on:

1. Pull requests that change Go-related files (`**/*.go`, `**/go.mod`, `**/go.sum`) or the SBOM workflow itself
2. Pushes to `main` (including merges)
3. A weekly schedule (Monday 03:17 UTC)

Locally, run the same SBOM gate with `make sbom` or `mise run sbom`. The Trivy version is resolved from [`mise.lock`](mise.lock) (fuzzy `latest` in `mise.toml`); bump with `mise lock --bump` then reinstall—do not hardcode the version in GitHub Actions.

Tagged releases regenerate and scan the same SBOM before GoReleaser runs, then attach `provider.cdx.json` to the GitHub Release.

## Reporting a Vulnerability

If you discover a security vulnerability in this project, please report it to us as follows:

**Preferred Method**: Use [GitHub's Private Vulnerability Reporting](https://github.com/ubie-oss/terraform-provider-lightdash/security/advisories/new) to report the vulnerability. This allows you to privately disclose the issue to repository maintainers.

Please include the following information in your report:

- A clear description of the vulnerability
- Steps to reproduce the issue
- Potential impact and severity assessment
- Any suggested remediation steps

## Response Process

After submitting a vulnerability report:

1. **Acknowledgment**: We will acknowledge receipt of your report within 48 hours and provide a preliminary assessment.

2. **Investigation**: Our team will investigate the issue and determine its validity and severity.

3. **Updates**: We will keep you informed about our progress throughout the investigation and remediation process.

4. **Disclosure**: Once the issue is resolved, we will work with you on a coordinated disclosure timeline. We aim to release security fixes as soon as possible while giving users time to update.

5. **Credit**: We will credit you in the security advisory and changelog for responsible disclosure (unless you prefer to remain anonymous).

Thank you for helping keep the Terraform Provider for Lightdash and its users secure!
