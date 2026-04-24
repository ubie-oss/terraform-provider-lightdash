# CLAUDE.md

@AGENTS.md

## Claude Code only

The imported file above is the **primary** project configuration. Add here only what Claude Code needs beyond that (for example plan-mode habits, billing-area cautions, or org policy). Do not restate overview, commands, gotchas, or local testing—those belong in `AGENTS.md`.

**Settings precedence (Anthropic):** managed policy and CLI flags override `settings.local.json` and project `.claude/settings.json` for tool behavior.

**Markdown context:** `CLAUDE.md` and `CLAUDE.local.md` are walked upward from the working directory; optional path-scoped rules live in `.claude/rules/` when present.

**One wrapper:** use either this `./CLAUDE.md` or `./.claude/CLAUDE.md`, not two divergent bodies.

**References:** [How CLAUDE.md loads](https://docs.anthropic.com/en/docs/claude-code/claude-md) · [Explore the `.claude` directory](https://code.claude.com/docs/en/claude-directory)
