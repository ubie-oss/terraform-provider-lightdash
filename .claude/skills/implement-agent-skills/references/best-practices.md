# Best Practices for Implementation

When creating a new Agent Skill, adhere to the following standards from the official [Agent Skills specification](https://agentskills.io/specification):

## Directory & Naming

- **Structure**: Each skill must be in its own directory: `skills/<skill-name>/`.
- **File**: The main file must be named `SKILL.md`.
- **Naming Convention**:
  - Use **unicode lowercase alphanumeric characters and hyphens** (`a-z`, `0-9`, and `-`).
  - Must match the parent directory name exactly.
  - Length: 1-64 characters.
  - Must not start or end with a hyphen.
  - Must not contain consecutive hyphens (`--`).
  - Example: `pdf-processing` (Good), `PDF_Helper` (Bad), `pdf--processing` (Bad).

## Frontmatter & Metadata

Every `SKILL.md` must start with YAML frontmatter containing required and optional fields.

```yaml
---
name: skill-name
description: Brief summary of what this skill does and when to use it.
license: Apache-2.0 (Optional)
compatibility: Requires git, docker, jq (Optional, max 500 chars)
metadata: (Optional)
  author: example-org
  version: "1.0"
allowed-tools: Bash(git:*) Read (Optional, space-delimited list)
---
```

### Required Fields

- **name**: Must match the directory name and follow naming conventions.
- **description**: 1-1024 characters. Should describe both **what** the skill does and **when** to use it. Include specific keywords for agent discovery.

### Optional Fields

- **license**: License name or reference to a bundled license file.
- **compatibility**: Environment requirements (e.g., specific OS, tools, or internet access).
- **metadata**: Arbitrary key-value mapping for custom properties.
- **allowed-tools**: Space-delimited list of pre-approved tools.

## Standard Directory Structure

Organize additional resources into these optional subdirectories:

- `scripts/`: Contains executable code (Python, Bash, JS). Scripts should be self-contained and handle errors gracefully.
- `references/`: Detailed technical documentation, templates, or domain-specific data.
- `assets/`: Static resources like configuration templates, images, or lookup tables.

## Progressive Disclosure

Structure the skill for efficient context usage:

1.  **Metadata** (~100 tokens): `name` and `description` (loaded at startup).
2.  **Instructions** (< 5000 tokens recommended): The full `SKILL.md` body (loaded when activated).
3.  **Resources** (as needed): External files in `scripts/`, `references/`, or `assets/` (loaded only when required).

## File References

- **Paths**: Use relative paths from the skill root (e.g., `[details](references/DETAIL.md)`).
- **Depth**: Keep file references one level deep from `SKILL.md`. Avoid deeply nested chains.

## Validation

Before finalizing a skill, validate it using the `skills-ref` CLI:

```bash
skills-ref validate ./my-skill
```

This ensures the frontmatter and naming conventions strictly follow the specification.
