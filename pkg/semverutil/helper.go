/*
Copyright 2019 The KubeOne Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package semverutil

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

func MustParseConstraint(constraint string) *semver.Constraints {
	result, err := semver.NewConstraint(constraint)
	if err != nil {
		panic(err)
	}

	return result
}

// GetMinorRange generates a list of versions (without patch) in the range [minVersionStr, maxVersionStr].
func GetMinorRange(minVersionStr, maxVersionStr string) ([]string, error) {
	// Parse the minimum and maximum versions
	minVersion, err := semver.NewVersion(minVersionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse minimum version: %w", err)
	}

	maxVersion, err := semver.NewVersion(maxVersionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse maximum version: %w", err)
	}

	var versions []string

	currentVersion := minVersion

	for currentVersion.Compare(maxVersion) <= 0 {
		versionWithoutPatch, err := GetVersionWithoutPatch(currentVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to format version: %w", err)
		}
		versions = append(versions, versionWithoutPatch)

		nextVersion, err := IncrementMinor(currentVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to increment minor version: %w", err)
		}

		currentVersion = nextVersion
	}

	return versions, nil
}

// IncrementMinor increments the minor version of the given semantic version.
func IncrementMinor(version *semver.Version) (*semver.Version, error) {
	if version == nil {
		return nil, fmt.Errorf("input version cannot be nil")
	}

	newVersionStr := fmt.Sprintf("%d.%d.%d", version.Major(), version.Minor()+1, 0)
	newVersion, err := semver.NewVersion(newVersionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to increment minor version: %w", err)
	}

	return newVersion, nil
}

// GetVersionWithoutPatch returns the version without the patch release (e.g., "1.30.5" -> "1.30").
func GetVersionWithoutPatch(version *semver.Version) (string, error) {
	if version == nil {
		return "", fmt.Errorf("input version cannot be nil")
	}

	return fmt.Sprintf("%d.%d", version.Major(), version.Minor()), nil
}

// GetPatchRange generates a list of versions (with patch) in the range [minVersionStr, maxVersionStr], incrementing the patch version.
func GetPatchRange(minVersionStr, maxVersionStr string) ([]string, error) {
	// Parse the minimum and maximum versions
	minVersion, err := semver.NewVersion(minVersionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse minimum version: %w", err)
	}

	maxVersion, err := semver.NewVersion(maxVersionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse maximum version: %w", err)
	}

	var versions []string

	currentVersion := minVersion

	for currentVersion.Compare(maxVersion) <= 0 {
		versions = append(versions, currentVersion.String())

		nextVersion, err := IncrementPatch(currentVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to increment patch version: %w", err)
		}

		currentVersion = nextVersion
	}

	return versions, nil
}

// IncrementPatch increments the patch version of the given semantic version.
func IncrementPatch(version *semver.Version) (*semver.Version, error) {
	if version == nil {
		return nil, fmt.Errorf("input version cannot be nil")
	}

	newVersionStr := fmt.Sprintf("%d.%d.%d", version.Major(), version.Minor(), version.Patch()+1)
	newVersion, err := semver.NewVersion(newVersionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to increment patch version: %w", err)
	}

	return newVersion, nil
}
