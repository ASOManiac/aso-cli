# Workflow Patterns

`aso workflow` lets you compose existing `aso` commands and shell commands into
repeatable release pipelines.

## Verified local Xcode -> TestFlight workflow

This pattern was validated against a real app using:

- `aso builds latest --next` to choose the next build number for a version
- `aso xcode archive` to create a deterministic `.xcarchive`
- `aso xcode export` to create a deterministic `.ipa`
- `aso publish testflight --group ... --wait` to upload, wait for processing,
  and add the build to a TestFlight group

Create `.asc/export-options-app-store.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>method</key>
  <string>app-store-connect</string>
  <key>signingStyle</key>
  <string>automatic</string>
  <key>teamID</key>
  <string>YOUR_TEAM_ID</string>
  <key>uploadSymbols</key>
  <true/>
</dict>
</plist>
```

Create `.asc/workflow.json`:

```json
{
  "env": {
    "APP_ID": "1234567890",
    "PROJECT_PATH": "App.xcodeproj",
    "SCHEME": "App",
    "CONFIGURATION": "Release",
    "EXPORT_OPTIONS": ".asc/export-options-app-store.plist",
    "TESTFLIGHT_GROUP": "Beta",
    "VERSION": ""
  },
  "workflows": {
    "testflight_beta": {
      "description": "Archive, export, upload, and distribute an app to a TestFlight group.",
      "steps": [
        {
          "name": "validate_version",
          "run": "if [ -z \"$VERSION\" ]; then echo \"VERSION is required\" >&2; exit 1; fi"
        },
        {
          "name": "beta_resolve_next_build",
          "run": "aso builds latest --app \"$APP_ID\" --version \"$VERSION\" --platform IOS --next --initial-build-number 1 --output json",
          "outputs": {
            "BUILD_NUMBER": "$.nextBuildNumber"
          }
        },
        {
          "name": "beta_archive",
          "run": "aso xcode archive --project \"$PROJECT_PATH\" --scheme \"$SCHEME\" --configuration \"$CONFIGURATION\" --archive-path \".asc/artifacts/App-$VERSION-${steps.beta_resolve_next_build.BUILD_NUMBER}.xcarchive\" --clean --overwrite --xcodebuild-flag=-destination --xcodebuild-flag=generic/platform=iOS --xcodebuild-flag=-allowProvisioningUpdates --xcodebuild-flag=MARKETING_VERSION=$VERSION --xcodebuild-flag=CURRENT_PROJECT_VERSION=${steps.beta_resolve_next_build.BUILD_NUMBER} --output json",
          "outputs": {
            "ARCHIVE_PATH": "$.archive_path",
            "VERSION": "$.version",
            "BUILD_NUMBER": "$.build_number"
          }
        },
        {
          "name": "beta_export",
          "run": "aso xcode export --archive-path ${steps.beta_archive.ARCHIVE_PATH} --export-options \"$EXPORT_OPTIONS\" --ipa-path \".asc/artifacts/App-$VERSION-${steps.beta_archive.BUILD_NUMBER}.ipa\" --overwrite --xcodebuild-flag=-allowProvisioningUpdates --output json",
          "outputs": {
            "IPA_PATH": "$.ipa_path",
            "VERSION": "$.version",
            "BUILD_NUMBER": "$.build_number"
          }
        },
        {
          "name": "beta_publish",
          "run": "aso publish testflight --app \"$APP_ID\" --ipa ${steps.beta_export.IPA_PATH} --group \"$TESTFLIGHT_GROUP\" --wait --poll-interval 10s --output json",
          "outputs": {
            "BUILD_ID": "$.buildId",
            "BUILD_NUMBER": "$.buildNumber"
          }
        }
      ]
    }
  }
}
```

Run it:

```bash
aso workflow validate
aso workflow run --dry-run testflight_beta VERSION:1.2.3
aso workflow run testflight_beta VERSION:1.2.3
```

Notes:

- `VERSION` must be a valid next marketing version for your app. If the latest
  App Store version is already `READY_FOR_DISTRIBUTION`, reusing that same
  version can cause App Store Connect to reject the upload.
- `TESTFLIGHT_GROUP` accepts either a beta group name or group ID.
- Add `"ASC_BYPASS_KEYCHAIN": "1"` to the top-level `env` block if you want the
  workflow to resolve credentials from environment variables or config instead
  of the macOS keychain.
- Output-producing step names should stay unique within the workflow file when
  you define multiple workflows that use `outputs`.
