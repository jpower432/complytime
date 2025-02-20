// SPDX-License-Identifier: Apache-2.0

package complytime

import (
	"testing"

	"github.com/oscal-compass/compliance-to-policy-go/v2/plugin"
	"github.com/stretchr/testify/require"
)

func TestGetSelections(t *testing.T) {
	tests := []struct {
		name       string
		manifests  plugin.Manifests
		selections PluginOptions
		wantMap    map[string]map[string]string
		wantErr    string
	}{
		{
			name: "Valid/Selections",
			manifests: map[string]plugin.Manifest{
				"plugin1": {},
				"plugin2": {},
			},
			selections: PluginOptions{
				Workspace: "testworkspace",
				Profile:   "testprofile",
			},
			wantMap: map[string]map[string]string{
				"plugin1": {
					"workspace": "testworkspace",
					"profile":   "testprofile",
				},
				"plugin2": {
					"workspace": "testworkspace",
					"profile":   "testprofile",
				},
			},
		},
		{
			name: "Invalid/MissingOptions",
			manifests: map[string]plugin.Manifest{
				"plugin1": {},
				"plugin2": {},
			},
			selections: PluginOptions{},
			wantErr:    "failed plugin config validation: workspace must be set",
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			gotMap, err := getSelections(c.manifests, c.selections)
			if c.wantErr != "" {
				require.EqualError(t, err, c.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.wantMap, gotMap)
			}

		})
	}
}
