package version

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func buildInfoWithMainVersion(mainVersion string) func() (*debug.BuildInfo, bool) {
	return func() (*debug.BuildInfo, bool) {
		info := new(debug.BuildInfo)
		info.Main.Version = mainVersion

		return info, true
	}
}

func noBuildInfo() (*debug.BuildInfo, bool) {
	return nil, false
}

func nilBuildInfoButOK() (*debug.BuildInfo, bool) {
	return nil, true
}

func TestResolveVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		explicit      string
		readBuildInfo func() (*debug.BuildInfo, bool)
		want          string
	}{
		{
			name:          "explicit version wins over build info",
			explicit:      "v1.2.3",
			readBuildInfo: buildInfoWithMainVersion("v9.9.9"),
			want:          "v1.2.3",
		},
		{
			name:          "module version from build info",
			explicit:      versionUnknown,
			readBuildInfo: buildInfoWithMainVersion("v0.55.0"),
			want:          "v0.55.0",
		},
		{
			name:          "devel build info is returned as is",
			explicit:      versionUnknown,
			readBuildInfo: buildInfoWithMainVersion("(devel)"),
			want:          "(devel)",
		},
		{
			name:          "empty module version falls back to unknown",
			explicit:      versionUnknown,
			readBuildInfo: buildInfoWithMainVersion(""),
			want:          versionUnknown,
		},
		{
			name:          "missing build info falls back to unknown",
			explicit:      versionUnknown,
			readBuildInfo: noBuildInfo,
			want:          versionUnknown,
		},
		{
			name:          "nil build info with ok flag falls back to unknown",
			explicit:      versionUnknown,
			readBuildInfo: nilBuildInfoButOK,
			want:          versionUnknown,
		},
		{
			name:          "empty explicit version is treated as not set",
			explicit:      "",
			readBuildInfo: buildInfoWithMainVersion("v0.1.0"),
			want:          "v0.1.0",
		},
		{
			name:          "empty explicit version without build info falls back to unknown",
			explicit:      "",
			readBuildInfo: noBuildInfo,
			want:          versionUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, resolveVersion(tt.explicit, tt.readBuildInfo))
		})
	}
}
