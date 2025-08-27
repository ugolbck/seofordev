package version

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is set via ldflags during build
var Version = "dev"

// MinimumRequiredVersion is the minimum version required to run the app
const MinimumRequiredVersion = "2.0.0"

// GetVersion returns the current version
func GetVersion() string {
	if Version == "" {
		return "dev"
	}
	return Version
}

// IsVersionAtLeast checks if the current version meets the minimum requirement
func IsVersionAtLeast(minVersion string) (bool, error) {
	currentVersion := GetVersion()

	// Development builds always pass version check
	if currentVersion == "dev" {
		return true, nil
	}

	return compareVersions(currentVersion, minVersion)
}

// CheckMinimumVersion validates that the current version meets the minimum requirement
func CheckMinimumVersion() error {
	isValid, err := IsVersionAtLeast(MinimumRequiredVersion)
	if err != nil {
		return fmt.Errorf("failed to validate version: %w", err)
	}

	if !isValid {
		return fmt.Errorf("version too old: current version %s, minimum required version %s", GetVersion(), MinimumRequiredVersion)
	}

	return nil
}

// compareVersions returns true if version1 >= version2
// Supports semantic versioning (e.g., "1.2.3")
func compareVersions(version1, version2 string) (bool, error) {
	// Remove 'v' prefix if present
	v1 := strings.TrimPrefix(version1, "v")
	v2 := strings.TrimPrefix(version2, "v")

	// Split versions into parts
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Ensure we have at least 3 parts for both versions
	for len(parts1) < 3 {
		parts1 = append(parts1, "0")
	}
	for len(parts2) < 3 {
		parts2 = append(parts2, "0")
	}

	// Compare each part
	for i := 0; i < 3; i++ {
		num1, err := strconv.Atoi(parts1[i])
		if err != nil {
			return false, fmt.Errorf("invalid version format: %s", version1)
		}

		num2, err := strconv.Atoi(parts2[i])
		if err != nil {
			return false, fmt.Errorf("invalid version format: %s", version2)
		}

		if num1 > num2 {
			return true, nil
		} else if num1 < num2 {
			return false, nil
		}
		// If equal, continue to next part
	}

	// All parts are equal
	return true, nil
}
