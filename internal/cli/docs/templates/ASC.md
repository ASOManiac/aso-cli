# aso cli reference

Keyword intelligence CLI for the App Store, powered by [asomaniac.com](https://asomaniac.com).
AI-friendly command catalog for the aso cli. Use this alongside the aso cli readme
(examples) and `aso --help` (source of truth). Generate this file in any repo
with `aso init` (or `aso docs init`).

## Command Discovery (Source of Truth)

```bash
aso --help
aso <command> --help
aso <command> <subcommand> --help
```

Do not memorize flags. Always use `--help` for the current interface.

## Core Principles

- JSON output by default in non-TTY environments (pipe-friendly)
- Use `--output table` or `--output markdown` for human-readable output
- Authenticate once with `aso auth login` and reuse the session

## Common Patterns

- Storefronts use 2-letter App Store country codes (e.g. `us`, `gb`, `de`); list them with `aso storefronts`.
- Output formats: `--output json|table|markdown` and `--pretty` for readable JSON.
- `ASC_DEFAULT_OUTPUT` can pin the default output mode across contexts.
- Profiles: `--profile "NAME"` for managing multiple ASO Maniac accounts.

## Quick Lookup

| Task | Command |
|------|---------|
| Authenticate (browser OAuth) | `aso auth login` |
| Authenticate with API key | `aso auth login --api-key asm_k_abc123` |
| Check auth status | `aso auth status` |
| Show account, plan, usage | `aso auth whoami` |
| Log out | `aso auth logout` |
| Analyze a single keyword | `aso keywords analyze "vpn" --storefront us` |
| Get keyword recommendations | `aso keywords recommend "fitness tracker" --storefront us --limit 20` |
| Batch analyze across storefronts | `aso keywords batch "vpn,proxy,privacy" --storefronts us,gb,de` |
| List supported storefronts | `aso storefronts` |
| Generate ASC.md | `aso init` |

## Common Workflows

### Analyze a Keyword

```bash
aso keywords analyze "photo editor" --storefront us
aso keywords analyze "photo editor" --storefront us --output table
aso keywords analyze "photo editor" --storefront us | jq '.popularity'
```

### Discover Keywords from a Seed

```bash
aso keywords recommend "fitness tracker" --storefront us --limit 20
aso keywords recommend "meditation" --storefront us --output markdown
```

### Compare Multiple Keywords or Storefronts

```bash
aso keywords batch "vpn,proxy,privacy" --storefronts us,gb,de
aso keywords batch "vpn,proxy,privacy" --storefronts us --output json | jq '.[] | {keyword, popularity}'
```

## Command Groups

Use `aso <command> --help` for subcommands and flags.

- `auth` - Manage authentication for the ASO Maniac API.
- `keywords` - Analyze popularity, difficulty, and get AI-powered suggestions.
- `storefronts` - List all 60+ supported App Store country codes.
- `docs` - Access embedded documentation guides and reference helpers.
- `completion` - Print shell completion scripts.

## Global Flags

- `--profile` - Use a named authentication profile
- `--report` - Report format for CI output
- `--report-file` - Path to write CI report file
- `--version` - Print version and exit

## Environment Variables (Selected)

- `ASO_MANIAC_API_KEY` - Override the configured API key (precedence over config file)
- `ASC_DEFAULT_OUTPUT` - Default output format (`json`, `table`, `markdown`)
- `ASC_CONFIG_PATH` - Override the config file location (defaults to `~/.asomaniac/config.json`)

## See Also

- Full command reference: `docs/COMMANDS.md` (auto-generated)
- Agent skills and recipes: https://github.com/ASOManiac/aso-skills
- API: https://asomaniac.com
