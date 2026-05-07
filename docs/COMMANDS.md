# Command Reference Guide

This file is generated from live CLI help output.
For authoritative command behavior, also use:

```bash
aso --help
aso <command> --help
aso <command> <subcommand> --help
```

To regenerate:

```bash
make generate-command-docs
```

## Usage Pattern

```bash
aso <subcommand> [flags]
```

## Global Flags

- `--profile` - Use named authentication profile
- `--report` - Report format for CI output (e.g., junit)
- `--report-file` - Path to write CI report file
- `--version` - Print version and exit (default: false)

## Command Families

### Subcommands

- `auth` - Manage authentication for the ASO Maniac API.
- `keywords` - Analyze popularity, difficulty, and get AI-powered suggestions.
- `storefronts` - List all 60+ supported App Store country codes.
- `docs` - Access embedded documentation guides and reference helpers.
- `completion` - Print shell completion scripts.

## Scripting Tips

- JSON is the default output format in non-interactive environments — pipe to `jq` or any JSON parser.
- Use `--output table` for human-readable output in scripts that need it.
- Authenticate once with `aso auth maniac login` (or `ASO_MANIAC_API_KEY`) and reuse the session in CI.

## High-Signal Examples

```bash
# Authenticate
aso auth maniac login
aso auth maniac whoami

# Analyze a single keyword
aso keywords analyze "vpn" --storefront us

# Get recommendations from a seed
aso keywords recommend "fitness tracker" --storefront us --limit 20

# Batch analyze across storefronts
aso keywords batch "vpn,proxy,privacy" --storefronts us,gb,de

# List supported storefronts
aso storefronts
```

## Related Documentation

- [../README.md](../README.md) - onboarding and common workflows
- [../CONTRIBUTING.md](../CONTRIBUTING.md) - contribution and dev workflow
