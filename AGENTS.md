# AGENTS.md

Lightweight Go CLI for keyword intelligence and ASO. Built with [ffcli](https://github.com/peterbourgon/ff).

## Command surface

| Command | Auth | Purpose |
|---|---|---|
| `aso keywords analyze` | yes | Keyword popularity / difficulty / top-ranking apps |
| `aso keywords recommend` | yes | Seed-based keyword recommendations |
| `aso keywords batch` | yes | Multi-keyword / multi-storefront analysis |
| `aso storefronts` | no | List supported storefront codes |
| `aso auth maniac login` | no | Authenticate (OAuth or `--api-key`) |
| `aso auth maniac logout` | no | Clear stored credentials |
| `aso auth maniac status` | yes | Show authenticated account state |
| `aso auth maniac whoami` | yes | Account ID, plan, API usage |

## Authentication

Three methods. Resolution order: env var > config file.

```bash
aso auth maniac login                                  # browser OAuth
aso auth maniac login --api-key asm_k_abc123           # direct API key
export ASO_MANIAC_API_KEY=asm_k_abc123                 # env var
```

Config file: `~/.asomaniac/config.json`.

## JSON output

Default in non-TTY environments. Examples:

```jsonc
// aso keywords analyze vpn --storefront us
{
  "keyword": "vpn",
  "storefront": "us",
  "popularity": 87,
  "difficulty": 92,
  "top_apps": [{"name": "ExpressVPN", "app_id": "...", "rating": 4.2}]
}

// aso auth maniac whoami
{
  "account_id": "user_123",
  "plan": "pro",
  "api_calls_this_month": 42,
  "api_limit": 10000
}
```

## Common errors

| Error | Cause | Fix |
|---|---|---|
| `auth: no credentials found` | API key not set | `aso auth maniac login` or set `ASO_MANIAC_API_KEY` |
| `401 unauthorized` | Invalid or expired key | Verify at asomaniac.com/settings; regenerate |
| `429 too many requests` | Rate limit | Check plan with `aso auth maniac whoami`; wait or upgrade |
| `invalid storefront code` | Typo | `aso storefronts` to list valid codes |

## Discovering commands at runtime

```bash
aso --help
aso keywords --help
aso keywords analyze --help
```

Always check `--help` for current flags.

## Agent integration tips

1. JSON is default in non-TTY. Parse with `jq` or a JSON parser, never regex.
2. Run `aso auth maniac status` before workflows to validate setup.
3. Respect 429 with exponential backoff.
4. For full workflows, use [aso-skills](https://github.com/ASOManiac/aso-skills).

## Environment variables

| Variable | Purpose |
|---|---|
| `ASO_MANIAC_API_KEY` | Override API key (precedence over config file) |
| `ASO_DEFAULT_OUTPUT` | Default output format: `json`, `table`, `markdown` |

## References

- Full command reference: `docs/COMMANDS.md` (auto-generated)
- aso-skills: https://github.com/ASOManiac/aso-skills
- Development setup: `CONTRIBUTING.md`
