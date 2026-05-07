# Support

## Start Here

- Quick start and troubleshooting: [README.md](README.md)
- Questions and workflow help: [GitHub Discussions](https://github.com/ASOManiac/aso-cli/discussions)
- Bugs and feature requests: [GitHub Issues](https://github.com/ASOManiac/aso-cli/issues)

## Use Discussions For

- Install or upgrade help
- Authentication and keychain setup questions
- "How do I...?" workflow questions
- Automation, CI, or scripting advice
- Sharing examples, tips, and patterns with other users

## Use Issues For

- Reproducible bugs or regressions
- Incorrect help text, broken docs, or misleading output
- Clear feature requests for missing commands, flags, or workflows

## Useful Bug Report Checklist

Include as many of these as you can:

- `aso version`
- Your OS and shell
- How you installed `aso` (`brew`, install script, source build, other)
- The exact command you ran
- Redacted stdout/stderr output
- Whether the issue still reproduces with `ASC_BYPASS_KEYCHAIN=1`

## Common Gotchas

- Authentication resolves from `ASO_MANIAC_API_KEY` first, then `~/.asomaniac/config.json`. Use `aso auth status` to see the resolved source.
- Output defaults are TTY-aware: interactive terminals default to `table`, pipes and CI default to `json`. Override with `--output` or `ASC_DEFAULT_OUTPUT`.
- Storefronts are 2-letter country codes (e.g. `us`, `gb`, `de`). Run `aso storefronts` to list valid codes.
