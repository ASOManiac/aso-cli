# aso cli reference

Unofficial CLI for the App Store Connect API. AI-friendly command catalog and
workflow notes for the aso cli. Use this alongside the aso cli readme
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

- Explicit flags (prefer `--app` over short flags)
- TTY-aware output defaults (`table` in terminals, `json` when piped/non-interactive)
- No interactive prompts (use `--confirm` for destructive actions)
- Pagination via `--paginate` on list commands

## Common Patterns

- IDs are App Store Connect API resource IDs (use list commands to find them).
- `--app "APP_ID"` is often required (or set `ASC_APP_ID`).
- `--paginate` fetches all pages; use `--limit` and `--next` for manual pagination.
- Output formats: `--output json|table|markdown` and `--pretty` for readable JSON.
- `ASC_DEFAULT_OUTPUT` can pin the default output mode across contexts.
- Destructive operations require `--confirm`.
- Profiles: `--profile "NAME"` and `--strict-auth` for auth resolution safety.
- Debugging: `--debug`, `--api-debug`, `--retry-log`.

## Quick Lookup

| Task | Command |
|------|---------|
| Check auth status | `aso auth status` |
| Run auth doctor | `aso doctor --output json` |
| Check account health | `aso account status` |
| Generate ASC.md | `aso init` |
| Create an app (unofficial web flow) | `aso web apps create --name "My App" --bundle-id "com.example.app" --sku "SKU123"` |
| List apps | `aso apps` |
| List builds | `aso builds list --app "APP_ID"` |
| List TestFlight groups | `aso testflight groups list --app "APP_ID"` |
| List internal TestFlight groups | `aso testflight groups list --app "APP_ID" --internal` |
| Stage a release (pre-submit) | `aso release stage --app "APP_ID" --version "VERSION" --build "BUILD_ID" --copy-metadata-from "PREVIOUS_VERSION" --dry-run` |
| Release (full pipeline) | `aso release run --app "APP_ID" --version "VERSION" --build "BUILD_ID" --metadata-dir "./metadata/version/VERSION" --dry-run` |
| Submit for review (low-level) | `aso submit create --app "APP_ID" --version "VERSION" --build "BUILD_ID" --confirm` |
| Weekly insights summary | `aso insights weekly --app "APP_ID" --source analytics --week "YYYY-MM-DD"` |
| Download localizations | `aso localizations download --version "VERSION_ID" --path "./localizations"` |

## Common Workflows

### Find an App ID and Recent Builds

```bash
aso apps
aso builds list --app "APP_ID" --sort -uploadedDate --limit 5
```

### Stage for Review (high-level: ensure version + copy/apply metadata + attach + validate)

```bash
# Dry-run the staging plan using metadata carry-forward
aso release stage --app "APP_ID" --version "1.0.0" --build "BUILD_ID" --copy-metadata-from "0.9.0" --dry-run

# Stage the version without submitting it for review yet
aso release stage --app "APP_ID" --version "1.0.0" --build "BUILD_ID" --copy-metadata-from "0.9.0" --confirm
```

### Release (high-level: ensure version + apply metadata + attach + validate + submit)

```bash
# Dry-run first to preview all planned steps
aso release run --app "APP_ID" --version "1.0.0" --build "BUILD_ID" --metadata-dir "./metadata/version/1.0.0" --dry-run

# Run the full pipeline
aso release run --app "APP_ID" --version "1.0.0" --build "BUILD_ID" --metadata-dir "./metadata/version/1.0.0" --confirm

# Monitor status after submission
aso status --app "APP_ID"
```

Lower-level alternatives for scripting or partial workflows:

```bash
aso versions list --app "APP_ID"
aso versions attach-build --version-id "VERSION_ID" --build "BUILD_ID"
aso validate --app "APP_ID" --version "1.0.0"
aso submit create --app "APP_ID" --version "1.0.0" --build "BUILD_ID" --confirm
```

### Distribute to TestFlight Group

```bash
aso testflight groups list --app "APP_ID"
aso testflight groups list --app "APP_ID" --internal
aso builds add-groups --build "BUILD_ID" --group "GROUP_ID"
aso builds add-groups --build "BUILD_ID" --group "GROUP_ID" --submit --confirm
```

### Migrate Metadata (Fastlane)

```bash
aso migrate validate --fastlane-dir ./metadata
aso migrate import --app "APP_ID" --fastlane-dir ./metadata
aso migrate export --app "APP_ID" --output ./exported-metadata
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

- `ASC_APP_ID` - Default app ID
- `ASC_PROFILE` - Default auth profile
- `ASC_TIMEOUT`, `ASC_TIMEOUT_SECONDS` - Request timeout
- `ASC_UPLOAD_TIMEOUT`, `ASC_UPLOAD_TIMEOUT_SECONDS` - Upload timeout
- `ASC_DEBUG` - Debug output (`api` enables HTTP logs)
- Web password environment variable (`ASC_WEB` + `_PASSWORD`) - Password source for `aso web auth login` and `aso web apps create`
- `ASC_WEB_SESSION_CACHE`, `ASC_WEB_SESSION_CACHE_DIR`, `ASC_WEB_SESSION_CACHE_BACKEND` - Web-session cache controls for unofficial web flows
- `ASC_IRIS_SESSION_CACHE`, `ASC_IRIS_SESSION_CACHE_DIR` - Deprecated legacy app-create cache settings; imported into the web session cache during the transition window
- `ASC_SPINNER_DISABLED` - Disable interactive stderr spinner
- `ASC_SKILLS_AUTO_CHECK` - Automatic skills update checks (`true`/`1`/`yes`/`y`/`on` enables, `false`/`0`/`no`/`n`/`off` disables; default enabled)

## API References (Offline)

In the aso cli repo, see:
- `docs/openapi/latest.json`
- `docs/openapi/paths.txt`
