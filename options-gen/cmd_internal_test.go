package optionsgen

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveDefaults(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		defaults   Defaults
		structName string
		wantTag    string
		wantVar    string
		wantFunc   string
	}{
		{
			name:       "none",
			defaults:   Defaults{From: DefaultsFromNone},
			structName: "Options",
		},
		{
			name:       "tag_default",
			defaults:   Defaults{From: DefaultsFromTag},
			structName: "Options",
			wantTag:    defaultTagName,
		},
		{
			name:       "tag_custom",
			defaults:   Defaults{From: DefaultsFromTag, Param: "cfg"},
			structName: "Options",
			wantTag:    "cfg",
		},
		{
			name:       "var_default",
			defaults:   Defaults{From: DefaultsFromVar},
			structName: "Config",
			wantVar:    "defaultConfig",
		},
		{
			name:       "var_custom",
			defaults:   Defaults{From: DefaultsFromVar, Param: "defaults"},
			structName: "Config",
			wantVar:    "defaults",
		},
		{
			name:       "func_default",
			defaults:   Defaults{From: DefaultsFromFunc},
			structName: "Config",
			wantFunc:   "getDefaultConfig",
		},
		{
			name:       "func_custom",
			defaults:   Defaults{From: DefaultsFromFunc, Param: "buildDefaults"},
			structName: "Config",
			wantFunc:   "buildDefaults",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotTag, gotVar, gotFunc := resolveDefaults(tc.defaults, tc.structName)
			require.Equal(t, tc.wantTag, gotTag)
			require.Equal(t, tc.wantVar, gotVar)
			require.Equal(t, tc.wantFunc, gotFunc)
		})
	}
}

func TestResolveOutOptionTypeName(t *testing.T) {
	t.Parallel()

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		got, err := resolveOutOptionTypeName("Config", "")
		require.NoError(t, err)
		require.Equal(t, "OptConfigSetter", got)
	})

	t.Run("custom", func(t *testing.T) {
		t.Parallel()

		got, err := resolveOutOptionTypeName("Config", "CustomSetter")
		require.NoError(t, err)
		require.Equal(t, "CustomSetter", got)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()

		_, err := resolveOutOptionTypeName("Config", "custom_setter")
		require.EqualError(t, err, "outOptionTypeName must be a valid type name, contains only letters a-z or A-Z")
	})
}
