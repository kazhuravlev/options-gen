package version

import (
	"runtime/debug"
)

const versionUnknown = "unknown-local"

var version = versionUnknown

func GetVersion() string {
	return resolveVersion(version, debug.ReadBuildInfo)
}

// resolveVersion prefers the explicitly set version (ldflags, task examples:update), then the module
// version from the build info, and finally falls back to versionUnknown.
func resolveVersion(explicit string, readBuildInfo func() (*debug.BuildInfo, bool)) string {
	// In case if not - someone (task examples:update) explicitly set the value of version.
	// An empty value (for example `-X ...version.version=`) is treated as "not set".
	if explicit != "" && explicit != versionUnknown {
		return explicit
	}

	if bi, ok := readBuildInfo(); ok && bi != nil {
		if bi.Main.Version != "" {
			return bi.Main.Version
		}
	}

	return versionUnknown
}
