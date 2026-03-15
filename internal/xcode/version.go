package xcode

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// BumpType represents the version component to increment.
type BumpType string

const (
	BumpMajor BumpType = "major"
	BumpMinor BumpType = "minor"
	BumpPatch BumpType = "patch"
	BumpBuild BumpType = "build"
)

// ParseBumpType validates and normalizes a bump type string.
func ParseBumpType(s string) (BumpType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "major":
		return BumpMajor, nil
	case "minor":
		return BumpMinor, nil
	case "patch":
		return BumpPatch, nil
	case "build":
		return BumpBuild, nil
	default:
		return "", fmt.Errorf("--type must be one of: major, minor, patch, build")
	}
}

// VersionInfo holds the current version and build number from an Xcode project.
type VersionInfo struct {
	Version     string `json:"version"`
	BuildNumber string `json:"buildNumber"`
	ProjectDir  string `json:"projectDir"`
	Modern      bool   `json:"modern"` // true if project uses MARKETING_VERSION build setting
}

// SetVersionOptions configures what to set.
type SetVersionOptions struct {
	ProjectDir  string
	Version     string
	BuildNumber string
}

// SetVersionResult holds the result of a set operation.
type SetVersionResult struct {
	Version     string `json:"version,omitempty"`
	BuildNumber string `json:"buildNumber,omitempty"`
	ProjectDir  string `json:"projectDir"`
}

// BumpVersionOptions configures the bump operation.
type BumpVersionOptions struct {
	ProjectDir string
	BumpType   BumpType
}

// BumpVersionResult holds the result of a bump operation.
type BumpVersionResult struct {
	BumpType   string `json:"bumpType"`
	OldVersion string `json:"oldVersion,omitempty"`
	NewVersion string `json:"newVersion,omitempty"`
	OldBuild   string `json:"oldBuild,omitempty"`
	NewBuild   string `json:"newBuild,omitempty"`
	ProjectDir string `json:"projectDir"`
}

// GetVersion reads the current marketing version and build number.
func GetVersion(ctx context.Context, projectDir string) (*VersionInfo, error) {
	if err := requireMacOS(); err != nil {
		return nil, err
	}
	if err := requireAgvtool(); err != nil {
		return nil, err
	}

	version, err := runAgvtool(ctx, projectDir, "what-marketing-version", "-terse1")
	if err != nil {
		return nil, fmt.Errorf("failed to read marketing version: %w", err)
	}

	buildNumber, err := runAgvtool(ctx, projectDir, "what-version", "-terse")
	if err != nil {
		return nil, fmt.Errorf("failed to read build number: %w", err)
	}

	parsedVersion := parseAgvtoolVersionOutput(version)
	parsedBuild := parseAgvtoolBuildOutput(buildNumber)
	modern := isVariableReference(parsedVersion)

	// Modern project: agvtool returns $(MARKETING_VERSION). Resolve via xcodebuild.
	if modern {
		resolved, err := readBuildSettings(ctx, projectDir)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve build settings: %w", err)
		}
		if v := resolved["MARKETING_VERSION"]; v != "" {
			parsedVersion = v
		}
		if b := resolved["CURRENT_PROJECT_VERSION"]; b != "" {
			parsedBuild = b
		}
	}

	return &VersionInfo{
		Version:     parsedVersion,
		BuildNumber: parsedBuild,
		ProjectDir:  projectDir,
		Modern:      modern,
	}, nil
}

// SetVersion sets the marketing version and/or build number.
func SetVersion(ctx context.Context, opts SetVersionOptions) (*SetVersionResult, error) {
	if err := requireMacOS(); err != nil {
		return nil, err
	}
	if err := requireAgvtool(); err != nil {
		return nil, err
	}

	result := &SetVersionResult{ProjectDir: opts.ProjectDir}
	modern := isModernProject(ctx, opts.ProjectDir)

	if v := strings.TrimSpace(opts.Version); v != "" {
		if modern {
			if err := updatePbxprojSetting(opts.ProjectDir, "MARKETING_VERSION", v); err != nil {
				return nil, fmt.Errorf("failed to set marketing version: %w", err)
			}
		} else {
			if _, err := runAgvtool(ctx, opts.ProjectDir, "new-marketing-version", v); err != nil {
				return nil, fmt.Errorf("failed to set marketing version: %w", err)
			}
		}
		result.Version = v
	}

	if b := strings.TrimSpace(opts.BuildNumber); b != "" {
		if modern {
			if err := updatePbxprojSetting(opts.ProjectDir, "CURRENT_PROJECT_VERSION", b); err != nil {
				return nil, fmt.Errorf("failed to set build number: %w", err)
			}
		} else {
			if _, err := runAgvtool(ctx, opts.ProjectDir, "new-version", "-all", b); err != nil {
				return nil, fmt.Errorf("failed to set build number: %w", err)
			}
		}
		result.BuildNumber = b
	}

	return result, nil
}

// BumpVersion increments the version or build number.
func BumpVersion(ctx context.Context, opts BumpVersionOptions) (*BumpVersionResult, error) {
	if err := requireMacOS(); err != nil {
		return nil, err
	}
	if err := requireAgvtool(); err != nil {
		return nil, err
	}

	current, err := GetVersion(ctx, opts.ProjectDir)
	if err != nil {
		return nil, err
	}

	result := &BumpVersionResult{
		BumpType:   string(opts.BumpType),
		ProjectDir: opts.ProjectDir,
	}

	if opts.BumpType == BumpBuild {
		result.OldBuild = current.BuildNumber
		if current.Modern {
			newBuild, err := incrementBuildString(current.BuildNumber)
			if err != nil {
				return nil, fmt.Errorf("failed to increment build number: %w", err)
			}
			if err := updatePbxprojSetting(opts.ProjectDir, "CURRENT_PROJECT_VERSION", newBuild); err != nil {
				return nil, fmt.Errorf("failed to set build number: %w", err)
			}
			result.NewBuild = newBuild
		} else {
			if _, err := runAgvtool(ctx, opts.ProjectDir, "next-version", "-all"); err != nil {
				return nil, fmt.Errorf("failed to increment build number: %w", err)
			}
			updated, err := GetVersion(ctx, opts.ProjectDir)
			if err != nil {
				return nil, fmt.Errorf("failed to read updated build number: %w", err)
			}
			result.NewBuild = updated.BuildNumber
		}
		return result, nil
	}

	// Version bump (major/minor/patch).
	result.OldVersion = current.Version
	newVersion, err := bumpVersionString(current.Version, opts.BumpType)
	if err != nil {
		return nil, err
	}

	if current.Modern {
		if err := updatePbxprojSetting(opts.ProjectDir, "MARKETING_VERSION", newVersion); err != nil {
			return nil, fmt.Errorf("failed to set marketing version: %w", err)
		}
	} else {
		if _, err := runAgvtool(ctx, opts.ProjectDir, "new-marketing-version", newVersion); err != nil {
			return nil, fmt.Errorf("failed to set marketing version: %w", err)
		}
	}
	result.NewVersion = newVersion

	return result, nil
}

func requireMacOS() error {
	if runtimeGOOS != "darwin" {
		return fmt.Errorf("xcode version commands require macOS")
	}
	return nil
}

func requireAgvtool() error {
	_, err := lookPathFn("agvtool")
	if err != nil {
		return fmt.Errorf("agvtool not found: install Xcode command-line tools")
	}
	return nil
}

func runAgvtool(ctx context.Context, projectDir string, args ...string) (string, error) {
	cmd := commandContextFn(ctx, "agvtool", args...)
	cmd.Dir = projectDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrText := strings.TrimSpace(stderr.String())
		if stderrText != "" {
			return "", fmt.Errorf("%w: %s", err, stderrText)
		}
		return "", err
	}

	return stdout.String(), nil
}

// readBuildSettings runs xcodebuild -showBuildSettings and extracts key=value pairs.
func readBuildSettings(ctx context.Context, projectDir string) (map[string]string, error) {
	xcodeproj, err := findXcodeproj(projectDir)
	if err != nil {
		return nil, err
	}

	cmd := commandContextFn(ctx, "xcodebuild", "-showBuildSettings", "-project", xcodeproj)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrText := strings.TrimSpace(stderr.String())
		if stderrText != "" {
			return nil, fmt.Errorf("%w: %s", err, stderrText)
		}
		return nil, err
	}

	settings := make(map[string]string)
	for _, line := range strings.Split(stdout.String(), "\n") {
		trimmed := strings.TrimSpace(line)
		if idx := strings.Index(trimmed, " = "); idx > 0 {
			key := strings.TrimSpace(trimmed[:idx])
			value := strings.TrimSpace(trimmed[idx+3:])
			settings[key] = value
		}
	}
	return settings, nil
}

// findXcodeproj finds the .xcodeproj directory in a project dir.
func findXcodeproj(projectDir string) (string, error) {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return "", fmt.Errorf("failed to read project directory: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() && strings.HasSuffix(entry.Name(), ".xcodeproj") {
			return filepath.Join(projectDir, entry.Name()), nil
		}
	}
	return "", fmt.Errorf("no .xcodeproj found in %s", projectDir)
}

// findPbxprojPath finds the project.pbxproj inside the .xcodeproj.
func findPbxprojPath(projectDir string) (string, error) {
	xcodeproj, err := findXcodeproj(projectDir)
	if err != nil {
		return "", err
	}
	pbxproj := filepath.Join(xcodeproj, "project.pbxproj")
	if _, err := os.Stat(pbxproj); err != nil {
		return "", fmt.Errorf("project.pbxproj not found in %s", xcodeproj)
	}
	return pbxproj, nil
}

// updatePbxprojSetting replaces all occurrences of a build setting in project.pbxproj.
// Matches lines like: MARKETING_VERSION = 1.2.3;
func updatePbxprojSetting(projectDir, setting, newValue string) error {
	pbxprojPath, err := findPbxprojPath(projectDir)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(pbxprojPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", pbxprojPath, err)
	}

	oldContent := string(data)
	var replaced int
	var result strings.Builder
	result.Grow(len(oldContent))

	for _, line := range strings.Split(oldContent, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, setting+" = ") && strings.HasSuffix(trimmed, ";") {
			// Preserve original indentation.
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			result.WriteString(indent)
			result.WriteString(setting)
			result.WriteString(" = ")
			result.WriteString(newValue)
			result.WriteString(";\n")
			replaced++
		} else {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	// Remove trailing extra newline from the loop.
	output := strings.TrimSuffix(result.String(), "\n")

	if replaced == 0 {
		return fmt.Errorf("%s not found in %s", setting, pbxprojPath)
	}

	return os.WriteFile(pbxprojPath, []byte(output), 0o644)
}

// isVariableReference checks if a value is an Xcode variable like $(MARKETING_VERSION).
func isVariableReference(value string) bool {
	return strings.Contains(value, "$(")
}

// isModernProject detects if the project uses build-setting-based versioning.
func isModernProject(ctx context.Context, projectDir string) bool {
	version, err := runAgvtool(ctx, projectDir, "what-marketing-version", "-terse1")
	if err != nil {
		return false
	}
	return isVariableReference(parseAgvtoolVersionOutput(version))
}

// parseAgvtoolVersionOutput extracts the version from agvtool output.
// `agvtool what-marketing-version -terse1` outputs lines like "=1.2.3" or "target=1.2.3".
func parseAgvtoolVersionOutput(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if idx := strings.LastIndex(line, "="); idx >= 0 {
			return strings.TrimSpace(line[idx+1:])
		}
		return line
	}
	return strings.TrimSpace(output)
}

// parseAgvtoolBuildOutput extracts the build number from agvtool output.
// `agvtool what-version -terse` outputs just the number.
func parseAgvtoolBuildOutput(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return ""
	}
	return strings.TrimSpace(lines[len(lines)-1])
}

// bumpVersionString increments a semver-style version string.
func bumpVersionString(current string, bumpType BumpType) (string, error) {
	current = strings.TrimSpace(current)
	if current == "" {
		return "", fmt.Errorf("current version is empty")
	}

	parts := strings.Split(current, ".")
	components := make([]int, len(parts))
	for i, p := range parts {
		val, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return "", fmt.Errorf("version %q is not a valid numeric version", current)
		}
		components[i] = val
	}

	switch bumpType {
	case BumpMajor:
		if len(components) < 1 {
			return "", fmt.Errorf("version %q has no major component", current)
		}
		components[0]++
		for i := 1; i < len(components); i++ {
			components[i] = 0
		}
	case BumpMinor:
		if len(components) < 2 {
			return "", fmt.Errorf("version %q needs at least major.minor format for minor bump", current)
		}
		components[1]++
		for i := 2; i < len(components); i++ {
			components[i] = 0
		}
	case BumpPatch:
		if len(components) < 3 {
			return "", fmt.Errorf("version %q needs major.minor.patch format for patch bump", current)
		}
		components[2]++
	default:
		return "", fmt.Errorf("unsupported bump type %q for version bump", bumpType)
	}

	result := make([]string, len(components))
	for i, v := range components {
		result[i] = strconv.Itoa(v)
	}
	return strings.Join(result, "."), nil
}

// incrementBuildString increments a numeric build string by 1.
func incrementBuildString(current string) (string, error) {
	current = strings.TrimSpace(current)
	if current == "" {
		return "", fmt.Errorf("build number is empty")
	}

	// Support dotted build numbers (e.g. 1.2.3 → 1.2.4).
	parts := strings.Split(current, ".")
	last := parts[len(parts)-1]
	val, err := strconv.Atoi(last)
	if err != nil {
		return "", fmt.Errorf("build number %q is not numeric", current)
	}
	parts[len(parts)-1] = strconv.Itoa(val + 1)
	return strings.Join(parts, "."), nil
}
