# Contributing

Thanks for your interest in contributing to aso-cli!

## Development Setup

Requirements:
- Go 1.26+

Clone and build:
```bash
git clone https://github.com/ASOManiac/aso-cli.git
cd aso-cli
make build
```

Run tests:
```bash
ASC_BYPASS_KEYCHAIN=1 make test
```

Optional tooling:
```bash
make tools   # installs gofumpt + golangci-lint
make lint    # uses golangci-lint if installed, else go vet
make format  # gofmt + gofumpt (requires gofumpt; install with make tools)
```

## Local Validation

Run this checklist before opening a PR:

```bash
make tools               # Install gofumpt + golangci-lint
make format              # Format code
make lint                # Lint code
make check-command-docs  # Verify command docs are up to date
ASC_BYPASS_KEYCHAIN=1 make test  # Run tests
make build               # Build binary
./aso --help             # Smoke-test the binary
```

## Pull Request Guidelines

- Keep PRs small and focused.
- Add or update tests for new behavior.
- When pruning repetitive tests, prefer grouped/table-driven suites, but keep representative high-signal assertions for response decoding and user-facing output formatting.
- Update `README.md` if behavior or scope changes.
- Avoid committing any credentials or API keys.

## Support and Questions

- Use GitHub Discussions for install help, auth setup, workflow questions, and "how do I...?" support.
- Use GitHub Issues for reproducible bugs and concrete feature requests.
- When reporting a bug, include `aso --version`, your OS, install method, exact command, and stdout/stderr.

## Issue Triage Labels

Every newly created GitHub issue should leave initial triage with exactly one label from
each of these buckets:

- Type: `bug`, `enhancement`, or `question`
- Priority: `p0`, `p1`, `p2`, or `p3`
- Difficulty: `easy`, `medium`, or `hard`

Label meanings:

- `bug`: broken behavior, regression, incorrect output, or misleading UX
- `enhancement`: new feature, workflow improvement, or behavior expansion
- `question`: clarification or discussion where the work is not yet well-defined
- `p0`: release-blocking, security-sensitive, data-loss, or core workflow outage
- `p1`: high-impact bug or important near-term work
- `p2`: normal roadmap work or a bug with a reasonable workaround or limited blast radius
- `p3`: longer-horizon, convenience, exploratory, or low-urgency work
- `easy`: small, low-risk, localized change
- `medium`: moderate cross-file change or some product/UX/design work
- `hard`: large, high-risk, or architecture-heavy change

External contributors may not have permission to label issues directly. Maintainers and
agents should add any missing labels during first triage, and new issues should not be left
without a type, priority, and difficulty label set.

## Security

If you find a security issue, please report it responsibly by opening a private issue
or contacting the maintainer directly.
