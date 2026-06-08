# aso-cli

> ## ⏸️ Paused (June 2026)
>
> ASO Maniac and its tooling are currently **paused** — new development, signups, and billing are on hold while we focus elsewhere. This repo stays up for reference; existing installs keep working but are unmaintained for now.


[![Release](https://img.shields.io/github/v/release/ASOManiac/aso-cli)](https://github.com/ASOManiac/aso-cli/releases/latest)
[![Homebrew](https://img.shields.io/badge/homebrew-asomaniac%2Ftap%2Faso-orange)](https://github.com/ASOManiac/homebrew-tap)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

Keyword intelligence CLI for App Store Optimization. Built for AI agents and developers.

A lightweight CLI client for [asomaniac.com](https://asomaniac.com) — keyword analysis, recommendations, and storefront intelligence for iOS, macOS, and other App Store markets.

## Quick start

```bash
brew install asomaniac/tap/aso
aso auth login
aso keywords analyze "vpn" --storefront us
```

## Installation

**Homebrew (recommended):**

```bash
brew install asomaniac/tap/aso
```

**Install script (macOS / Linux):**

```bash
curl -fsSL https://raw.githubusercontent.com/ASOManiac/aso-cli/main/install.sh | bash
```

**Go install:**

```bash
go install github.com/ASOManiac/aso-cli@latest
```

**Binary:** download from [Releases](https://github.com/ASOManiac/aso-cli/releases/latest).

## Authentication

Authenticate with your asomaniac.com API key. Three methods:

```bash
# 1. Browser OAuth (interactive)
aso auth login

# 2. Direct API key (from asomaniac.com/settings)
aso auth login --api-key asm_k_abc123

# 3. Environment variable (CI / scripts / automation)
export ASO_MANIAC_API_KEY=asm_k_abc123
```

Resolution order: env var > `~/.asomaniac/config.json`.

```bash
aso auth status
aso auth whoami
```

## Commands

```bash
# Analyze a keyword — popularity, difficulty, top-ranking apps
aso keywords analyze "photo editor" --storefront us

# Keyword recommendations from a seed
aso keywords recommend "fitness tracker" --storefront us --limit 20

# Batch analyze multiple keywords across storefronts
aso keywords batch "vpn,proxy,privacy" --storefronts us,gb,de

# List all supported storefronts
aso storefronts
```

Run `aso --help` for the full command tree.

## Output formats

JSON by default for piping to `jq` and AI agents:

```bash
aso keywords analyze "vpn" --storefront us | jq '.popularity'
aso keywords analyze "vpn" --output table
```

## For AI agents

Designed for integration with AI agent workflows. See [aso-skills](https://github.com/ASOManiac/aso-skills) for prebuilt agent skills and recipes. JSON is the default output format in non-interactive environments.

## For App Store Connect operations

For ASC operations (builds, TestFlight, submissions, metadata), use:

- **Primary:** [fastlane](https://fastlane.tools)
- **Fallback:** [App Store Connect CLI](https://github.com/rorkai/App-Store-Connect-CLI)

These are not bundled with aso-cli. Install them separately as needed.

## License

MIT — see [LICENSE](LICENSE).

## Contributing

```bash
git clone https://github.com/ASOManiac/aso-cli.git
cd aso-cli
make build
ASC_BYPASS_KEYCHAIN=1 make test
```

See [CONTRIBUTING.md](CONTRIBUTING.md).
