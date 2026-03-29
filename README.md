# ASO CLI

[![Release](https://img.shields.io/github/v/release/ASOManiac/aso-cli)](https://github.com/ASOManiac/aso-cli/releases/latest)
[![Homebrew](https://img.shields.io/badge/homebrew-asomaniac%2Ftap%2Faso-orange)](https://github.com/ASOManiac/homebrew-tap)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

App Store Optimization + App Store Connect CLI.

A fork of [App-Store-Connect-CLI](https://github.com/rudrankriyam/App-Store-Connect-CLI) with premium ASO commands powered by [asomaniac.com](https://asomaniac.com).

All original App Store Connect commands work out of the box, plus premium ASO commands for AI-powered keyword research, competitor analysis, rank tracking, and more.

## Quick Start

```bash
# Install
brew install asomaniac/tap/aso

# Authenticate with ASO Maniac (free: 100 API calls/month)
aso login

# Try a command
aso keywords suggest --app-id 1234567890 --storefront us
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

**GitHub releases**: download the binary for your platform from [Releases](https://github.com/ASOManiac/aso-cli/releases/latest).

## Free Commands (App Store Connect)

All commands from the upstream App Store Connect CLI are included and free to use. Authenticate with your App Store Connect API key:

```bash
aso auth login \
  --name "MyApp" \
  --key-id "ABC123" \
  --issuer-id "DEF456" \
  --private-key /path/to/AuthKey.p8 \
  --network
```

Then use any command: `aso apps list`, `aso builds list`, `aso testflight`, `aso submit`, `aso release run`, etc.

For full documentation on all inherited commands, see the [upstream CLI docs](https://github.com/rudrankriyam/App-Store-Connect-CLI).

## Premium ASO Commands

Premium ASO commands are available at the top level. Authenticate once with `aso login`.

| Command | Description |
|---------|-------------|
| `aso login` | Authenticate with your ASO Maniac account via browser |
| `aso logout` | Remove stored ASO Maniac credentials |
| `aso whoami` | Display current user, plan, and API usage |
| `aso keywords` | Analyze, suggest, and batch-process keywords |
| `aso competitors` | Analyze competitor keyword overlap |
| `aso rank` | Track and view keyword ranking history |
| `aso trends` | View keyword popularity trends over time |
| `aso export` | Export keyword data in CSV, JSON, or TSV format |
| `aso dashboard` | View your ASO portfolio dashboard overview |
| `aso storefronts` | List all supported App Store storefronts |

## Authentication

ASO Maniac commands use a separate auth flow from App Store Connect:

```bash
# Opens a browser window to authenticate with asomaniac.com
aso login

# Verify your session
aso whoami

# Remove credentials
aso logout
```

Credentials are stored in `~/.asomaniac/config.json`. Free plan includes 100 API calls per month.

## Output Formats

Default output is JSON, designed for piping to `jq` and consumption by AI agents:

```bash
# Pipe to jq for filtering
aso keywords suggest --app-id 1234567890 --storefront us | jq '.keywords[:5]'

# Export to CSV for spreadsheets
aso export --app-id 1234567890 --format csv > keywords.csv
```

In an interactive terminal, `aso` auto-detects TTY and defaults to `table` for App Store Connect commands. Override with `--output json` or `--output table` on any command.

## Attribution

Built on [App-Store-Connect-CLI](https://github.com/rudrankriyam/App-Store-Connect-CLI) by [Rudrank Riyam](https://github.com/rudrankriyam). All original commands are preserved and fully functional.

## License

MIT License -- see [LICENSE](LICENSE) for details.
