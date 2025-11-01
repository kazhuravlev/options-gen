package version

import (
	"runtime/debug"
)

const versionUnknown = "unknown-local"

var version = versionUnknown

func GetVersion() string {
	// In case if not - someone (task examples:update) explicitly set the value of version.
	if version != versionUnknown {
		return version
	}

	if bi, ok := debug.ReadBuildInfo(); ok {
		if bi.Main.Version != "" {
			return bi.Main.Version
		}
	}

	return versionUnknown
}
