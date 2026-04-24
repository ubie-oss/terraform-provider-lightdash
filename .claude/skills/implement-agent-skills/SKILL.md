---
name: implement-agent-skills
description: Comprehensive workflow for creating, implementing, and validating Agent Skills. Use when asked to "create a new skill", "author a skill", "add a capability", or when standardizing project-specific workflows. Support for platform detection (Cursor, Claude Code, Gemini CLI, Codex) and template selection.
---

# Implement Agent Skills

This skill provides a structured process for creating effective Agent Skills that extend agent capabilities with specialized knowledge, workflows, and tools.

## Platform Setup

Before starting, identify the target AI coding tool environment to ensure compatibility.

- See [references/platforms.md](references/platforms.md) for detection logic and target paths.

## Skill Creation Process

Follow these steps to develop a new skill from scratch or improve an existing one.

### 1. Understand the Skill with Concrete Examples

- Clarify the functionality the skill should support.
- Identify what a user would say to trigger this skill.
- _Examples_: "How many users logged in today?", "Rotate this PDF", "Build me a dashboard".

### 2. Brainstorm and Evaluate Approaches

- Broadly and deeply think of **five** different approaches to implement the skill.
- Score each approach (0 to 100) based on Feasibility, Performance, Maintainability, and Complexity.
- Recommend the best approach and show clear reasons why it was chosen.
- Proceed with the recommended approach unless the user explicitly wants to select a different one.
- Specifically, the PLAN mode MUST include the decision process.

### 3. Plan Reusable Skill Contents

Analyze the examples to identify reusable resources:

- **Scripts**: For tasks requiring deterministic reliability.
- **References**: For documentation, schemas, or domain knowledge.
- **Assets**: For templates, boilerplate code, or static files.

### 4. Initialize the Skill

- Create the skill directory following the [naming standards](references/best-practices.md).
- Create the `SKILL.md` file with required YAML frontmatter.
- Use an appropriate template from the [template catalog](references/templates.md).

### 5. Edit the Skill

- **Implement Resources**: Build out `scripts/`, `references/`, and `assets/`.
- **Write Instructions**: Fill in the `SKILL.md` body using imperative language.
- **Apply Progressive Disclosure**: Keep `SKILL.md` lean by linking to references.

### 6. Package the Skill

- Ensure all file references are relative and only one level deep.
- Validate the YAML frontmatter and directory structure.
- (Optional) Use a packaging script if available in the project.

### 7. Iterate

- Test the skill on real-world tasks.
- Refine descriptions to improve discovery.
- Update instructions based on edge cases discovered during testing.

## Resources

- **Best Practices**: [references/best-practices.md](references/best-practices.md)
- **Templates**: [references/templates.md](references/templates.md)
- **Platforms**: [references/platforms.md](references/platforms.md)
