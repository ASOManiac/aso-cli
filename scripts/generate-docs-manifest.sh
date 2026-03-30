#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
OUTPUT="$PROJECT_DIR/docs/manifest.json"

VERSION=$(git -C "$PROJECT_DIR" describe --tags --always 2>/dev/null || echo "dev")

echo "Generating CLI docs manifest (version: $VERSION)..."

mkdir -p "$(dirname "$OUTPUT")"

# For now, output the static manifest. In the future, this could parse
# command help output to auto-generate.
cat > "$OUTPUT" << MANIFEST
{
  "version": "$VERSION",
  "install": {
    "brew": "brew install asomaniac/tap/aso",
    "go": "go install github.com/ASOManiac/aso-cli@latest",
    "curl": "curl -fsSL https://raw.githubusercontent.com/ASOManiac/aso-cli/main/install.sh | bash"
  },
  "auth": {
    "methods": [
      { "name": "Browser OAuth", "command": "aso maniac login" },
      { "name": "Direct API key", "command": "aso maniac login --api-key <KEY>" },
      { "name": "Environment variable", "command": "export ASO_MANIAC_API_KEY=<KEY>" }
    ]
  },
  "commands": {
    "maniac": [
      { "name": "keywords analyze", "description": "Score keyword popularity, difficulty, and top apps" },
      { "name": "keywords recommend", "description": "AI-powered keyword suggestions from a seed" },
      { "name": "keywords batch", "description": "Analyze multiple keywords across storefronts" },
      { "name": "competitors", "description": "Competitor keyword overlap analysis" },
      { "name": "trends", "description": "Historical keyword popularity trends" },
      { "name": "rank track", "description": "Start tracking keyword rankings for an app" },
      { "name": "rank history", "description": "View historical rank positions" },
      { "name": "dashboard", "description": "Open the web dashboard for your portfolio" },
      { "name": "export", "description": "Export rankings and keyword data (CSV/JSON)" },
      { "name": "config", "description": "Manage CLI configuration and defaults" }
    ],
    "asc": [
      { "name": "apps list", "description": "List all apps in App Store Connect" },
      { "name": "apps info", "description": "Get app details and metadata" },
      { "name": "builds list", "description": "List builds for an app" },
      { "name": "testflight submit", "description": "Submit a build for TestFlight review" },
      { "name": "metadata update", "description": "Update app metadata (title, description, keywords)" }
    ],
    "ascCommandCount": 64
  }
}
MANIFEST

echo "Generated $OUTPUT"
