# How agents discover instructions

| Tool                            | Notes                                                                                                           | Official reference                                                                                                                                                                                                                                  |
| ------------------------------- | --------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Cursor**                      | Root and nested `AGENTS.md`; may combine with [`.cursor/rules/`](.cursor/rules/) and other agent config.        | [Rules / AGENTS.md](https://cursor.com/docs/rules)                                                                                                                                                                                                  |
| **OpenAI Codex**                | Instruction chain from `~/.codex` then repo root → cwd; merges `AGENTS.md` / `AGENTS.override.md` (later wins). | [Custom instructions with AGENTS.md](https://developers.openai.com/codex/guides/agents-md/)                                                                                                                                                         |
| **Claude Code**                 | [CLAUDE.md](CLAUDE.md) should start with `@AGENTS.md`; Claude-only notes stay in `CLAUDE.md`.                   | [Import additional files](https://code.claude.com/docs/en/memory#import-additional-files), [CLAUDE.md](https://docs.anthropic.com/en/docs/claude-code/claude-md), [.claude directory](https://code.claude.com/docs/en/claude-directory)             |
| **Gemini CLI**                  | [`.gemini/settings.json`](.gemini/settings.json) sets `context.fileName` to include `AGENTS.md`.                | [Context files](https://geminicli.com/docs/cli/gemini-md/)                                                                                                                                                                                          |
| **GitHub Copilot coding agent** | `AGENTS.md` in repo (nested files: nearest wins).                                                               | [Custom instructions](https://docs.github.com/en/copilot/customizing-copilot/adding-custom-instructions-for-github-copilot), [changelog](https://github.blog/changelog/2025-08-28-copilot-coding-agent-now-supports-agents-md-custom-instructions/) |

### Codex and optional extras

- **Size:** Codex concatenates discovered docs (~32 KiB combined by default). If you outgrow that, add **nested `AGENTS.md`** in subtrees (Codex and Cursor).
- **Optional files:** `.github/copilot-instructions.md` or `.github/instructions/*.instructions.md` only if a Copilot surface does not use `AGENTS.md`. **`.codex/agents/*.toml`** for Codex worker configs ([multi-agent](https://developers.openai.com/codex/multi-agent/))—not a substitute for this file.

### Repo paths (not obvious from the table)

- **`.claude/`** — Claude Code skills, agents, rules, settings. `.claude/agents/*.md` is also a Cursor subagent path; same name in `.cursor/agents/` wins.
- **Skills:** Cursor auto-discovers [`.cursor/skills/`](https://cursor.com/docs/skills) or `.agents/skills/` only. Reference [`.claude/skills/`](.claude/skills/) with `@.claude/skills/...` (or mirror into `.cursor/skills/` if you want auto-discovery).
