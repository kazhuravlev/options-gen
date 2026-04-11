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
			defaults:   Defaults{From: DefaultsFromNone, Param: ""},
			structName: "Options",
			wantTag:    "",
			wantVar:    "",
			wantFunc:   "",
		},
		{
			name:       "tag_default",
			defaults:   Defaults{From: DefaultsFromTag, Param: ""},
			structName: "Options",
			wantTag:    defaultTagName,
			wantVar:    "",
			wantFunc:   "",
		},
		{
			name:       "tag_custom",
			defaults:   Defaults{From: DefaultsFromTag, Param: "cfg"},
			structName: "Options",
			wantTag:    "cfg",
			wantVar:    "",
			wantFunc:   "",
		},
		{
			name:       "var_default",
			defaults:   Defaults{From: DefaultsFromVar, Param: ""},
			structName: "Config",
			wantTag:    "",
			wantVar:    "defaultConfig",
			wantFunc:   "",
		},
		{
			name:       "var_custom",
			defaults:   Defaults{From: DefaultsFromVar, Param: "defaults"},
			structName: "Config",
			wantTag:    "",
			wantVar:    "defaults",
			wantFunc:   "",
		},
		{
			name:       "func_default",
			defaults:   Defaults{From: DefaultsFromFunc, Param: ""},
			structName: "Config",
			wantTag:    "",
			wantVar:    "",
			wantFunc:   "getDefaultConfig",
		},
		{
			name:       "func_custom",
			defaults:   Defaults{From: DefaultsFromFunc, Param: "buildDefaults"},
			structName: "Config",
			wantTag:    "",
			wantVar:    "",
			wantFunc:   "buildDefaults",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			gotTag, gotVar, gotFunc := resolveDefaults(testCase.defaults, testCase.structName)
			require.Equal(t, testCase.wantTag, gotTag)
			require.Equal(t, testCase.wantVar, gotVar)
			require.Equal(t, testCase.wantFunc, gotFunc)
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
