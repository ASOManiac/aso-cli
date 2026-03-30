# ASO CLI

[![Release](https://img.shields.io/github/v/release/ASOManiac/aso-cli)](https://github.com/ASOManiac/aso-cli/releases/latest)
[![Homebrew](https://img.shields.io/badge/homebrew-asomaniac%2Ftap%2Faso-orange)](https://github.com/ASOManiac/homebrew-tap)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

**App Store Optimization + App Store Connect in one CLI.**

A fork of [App-Store-Connect-CLI](https://github.com/rudrankriyam/App-Store-Connect-CLI) with premium ASO intelligence powered by [asomaniac.com](https://asomaniac.com). All 70+ original App Store Connect commands work out of the box, plus AI-powered keyword research, competitor analysis, and rank tracking as top-level commands (`aso keywords`, `aso competitors`, `aso rank`, etc.).

## Quick Start

```bash
# Install
brew install asomaniac/tap/aso

# Authenticate with ASO Maniac (free: 100 API calls/month)
aso auth maniac login

# Analyze a keyword
aso keywords analyze "vpn" --storefront us

# Check your App Store Connect setup too
aso auth status
```

## Installation

**Homebrew** (recommended):

```bash
brew install asomaniac/tap/aso
```

**Install script** (macOS / Linux):

```bash
curl -fsSL https://raw.githubusercontent.com/ASOManiac/aso-cli/main/install.sh | bash
```

**Go install**:

```bash
go install github.com/ASOManiac/aso-cli@latest
```

**Binary**: grab the latest from [Releases](https://github.com/ASOManiac/aso-cli/releases/latest).

## Two CLIs in One

`aso` ships two command families that use **separate auth**:

| Family | Prefix | Auth | What it does |
|--------|--------|------|-------------|
| App Store Connect | `aso <command>` | ASC API key (.p8) | Manage apps, builds, TestFlight, releases, metadata, analytics |
| ASO Maniac | `aso <command>` | asomaniac.com API key | AI keyword analysis, competitor intel, rank tracking, trends |

Run `aso --help` for the full command tree.

---

## ASO Maniac Commands

Premium keyword intelligence at your fingertips.

### Authentication

Three ways to authenticate — pick what fits:

```bash
# 1. Browser OAuth (interactive)
aso auth maniac login

# 2. Paste an API key from asomaniac.com/settings
aso auth maniac login --api-key asm_k_abc123

# 3. Environment variable (CI/Docker/scripts)
export ASO_MANIAC_API_KEY=asm_k_abc123
```

Key resolution order: `ASO_MANIAC_API_KEY` env var > `~/.asomaniac/config.json`.

```bash
# Check what's active
aso auth maniac status

# See your plan and usage
aso auth maniac whoami
```

### Keyword Intelligence

```bash
# Analyze a keyword — popularity, difficulty, top-ranking apps
aso keywords analyze "photo editor" --storefront us

# AI-powered keyword suggestions from a seed
aso keywords recommend "fitness tracker" --storefront us --limit 20

# Batch analyze across multiple storefronts
aso keywords batch "vpn,proxy,privacy" --storefronts us,gb,de
```

### Competitor Analysis

```bash
# Find competitor apps and compare keyword overlap
aso competitors --app-id 1234567890 --storefront us
```

### Rank Tracking

```bash
# Start tracking an app's keywords
aso rank track 1234567890 --storefront us --keywords "vpn,proxy"

# View ranking history over time
aso rank history 1234567890 --keyword "vpn" --storefront us
```

### Trends & Data

```bash
# Keyword popularity trends
aso trends "vpn" --storefront us --from 2025-01-01

# Portfolio dashboard
aso dashboard

# Export data for spreadsheets or pipelines
aso export --format csv --type keywords

# List all 60+ supported storefronts
aso storefronts
```

### Command Reference

| Command | Description |
|---------|-------------|
| `aso auth maniac login` | Sign in via browser or `--api-key` |
| `aso auth maniac logout` | Remove stored credentials |
| `aso auth maniac whoami` | Show account, plan, and API usage |
| `aso auth maniac status` | Check connection, auth, and key source |
| `aso keywords` | Analyze, recommend, and batch-process keywords |
| `aso competitors` | Competitor keyword overlap analysis |
| `aso trends` | Keyword popularity over time |
| `aso rank` | Track and view keyword ranking history |
| `aso dashboard` | Portfolio overview with alerts |
| `aso export` | Download data as CSV, JSON, or TSV |
| `aso storefronts` | List all supported App Store country codes |

---

## App Store Connect Commands

All 70+ commands from the upstream ASC CLI are included. Authenticate with your App Store Connect API key:

```bash
aso auth login \
  --name "MyApp" \
  --key-id "ABC123" \
  --issuer-id "DEF456" \
  --private-key /path/to/AuthKey.p8 \
  --network
```

Then use any command:

```bash
aso apps                                          # List apps
aso builds list --app "APP_ID"                    # Recent builds
aso testflight groups list --app "APP_ID"         # TestFlight groups
aso release run --app "APP_ID" --version "1.0.0"  # Full release pipeline
aso metadata pull --app "APP_ID"                  # Pull metadata
aso submit create --app "APP_ID" --confirm        # Submit for review
aso schema search "appStoreVersions"              # Inspect 1,208 API endpoints
```

For full documentation, see [upstream CLI docs](https://github.com/rudrankriyam/App-Store-Connect-CLI).

### Schema Inspector

The CLI embeds the full App Store Connect OpenAPI spec (1,208 endpoints) for offline reference:

```bash
aso schema search "screenshots"       # Find endpoints by path
aso schema search "GET.appInfos"      # Filter by method + path
aso schema show "/v1/apps/{id}"       # Full endpoint detail
```

## Output Formats

Default output is JSON, designed for piping to `jq` and AI agents:

```bash
# Pipe to jq
aso keywords analyze "vpn" --storefront us | jq '.popularity'

# Export to CSV
aso export --format csv --type keywords > keywords.csv
```

In interactive terminals, App Store Connect commands auto-detect TTY and default to `table`. Override with `--output json|table|markdown`.

## Attribution

Built on [App-Store-Connect-CLI](https://github.com/rudrankriyam/App-Store-Connect-CLI) by [Rudrank Riyam](https://github.com/rudrankriyam). All original commands are preserved and fully functional.

## License

MIT License — see [LICENSE](LICENSE) for details.
