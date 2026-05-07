# Contributing

For development setup, integration tests, and PR guidelines, see the root `CONTRIBUTING.md`.

This file covers patterns for AI agents working on the codebase.

## Git Workflow

- Branch from `main` and keep one logical change per branch
- Do not commit directly to `main` unless explicitly instructed; prefer PRs
- Prefer `git worktree add` for parallel tasks; remove with `git worktree remove` when done
- Rebase on `main` before merging; avoid merge commits
- Commit small, coherent changes; no WIP commits on shared branches
- Use concise, present-tense commit messages that match repo style
- Never commit secrets or local config files (keys, `.env`, `~/.asomaniac/config.json`)

## Before Committing

```bash
make format     # Format code
make lint       # Check for issues
make check-command-docs  # Verify command docs stay in sync
ASC_BYPASS_KEYCHAIN=1 make test  # Run all tests without keychain prompts
git diff        # Review changes before staging
```

## CLI Structure

- Command implementations live in `internal/cli/<domain>` packages
- Each domain exposes a top-level `XCommand() *ffcli.Command`
- `cmd/` only contains the root entry point; do not add wrapper files
- Register top-level commands in `cmd/root.go` (order matters)
- Shared CLI helpers go in `internal/cli/shared`

## Adding a New Command

1. Add or extend a domain package in `internal/cli/<domain>`
2. Implement a command factory (e.g., `XCommand() *ffcli.Command`)
3. Register it as a subcommand in `cmd/root.go`
4. Write tests in the domain package
5. Update README.md and `docs/COMMANDS.md` (run `make generate-command-docs`)

## Adding a New API Call

1. Add a method to `internal/asomaniac/client.go`
2. Add types for request/response structs
3. Add table/markdown renderers in the matching `internal/cli/<domain>` package
4. Write HTTP client tests with mocked responses
5. Group repeated request-wiring tests, but keep at least one representative non-empty decode assertion and one representative output-structure assertion where formatting is user-facing

## Releases

Tag releases with plain semver like `2.0.0` (no `v` prefix).

### Pre-Release Checklist

Before tagging a release, verify:

```bash
# 1. All tests pass
ASC_BYPASS_KEYCHAIN=1 make test

# 2. Audit help output for every command group
for cmd in auth keywords storefronts docs completion; do
  echo "=== $cmd ===" && ./aso $cmd --help 2>&1
done

# 3. Check for duplicate sections (should see SUBCOMMANDS only once per command)
# 4. Verify bold formatting renders correctly
```

**Common issues to check:**
- No duplicate "Subcommands:" sections (don't list subcommands in LongHelp; DefaultUsageFunc handles it)
- All flags have descriptions
- Examples are up to date
