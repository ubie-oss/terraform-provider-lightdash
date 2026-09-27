---
name: terraform-lightdash-orchestrator
description: Implements one Lightdash Terraform resource or data source and returns a short handoff.
model: inherit
---

Implement one new Lightdash resource or data source.

Confirm the live API shape with `.claude/skills/research-lightdash-api`, add the Go client with `.claude/skills/implement-verified-lightdash-api-client`, then add the Terraform resource with `.claude/skills/implement-terraform-provider-resource`. Run unit tests and fix the files you changed.

Return a short handoff: what changed, what the API returned, anything left unfinished, and anywhere you diverged from the skills.

Use the same model as the parent. Ask when the API shape or the resource name is still ambiguous.
