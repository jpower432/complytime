// SPDX-License-Identifier: Apache-2.0

package terminal

import (
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"github.com/stretchr/testify/require"
)

func TestShowDefinitionTable(t *testing.T) {
	tests := []struct {
		name     string
		compDefs []oscalTypes.ComponentDefinition
		wantView string
		wantErr  string
	}{
		{
			name: "Valid/HappyPath",
			compDefs: []oscalTypes.ComponentDefinition{
				{
					Components: &[]oscalTypes.DefinedComponent{
						{
							Title: "MySoftware",
							Type:  "service",
							ControlImplementations: &[]oscalTypes.ControlImplementationSet{
								{
									Source:      "profiles/example/profile.json",
									Description: "My implementation.",
								},
								{
									Props: &[]oscalTypes.Property{
										{
											Name:  extensions.FrameworkProp,
											Value: "anotherexample",
											Ns:    extensions.TrestleNameSpace,
										},
									},
									Description: "my other implementation",
								},
							},
						},
					},
				},
			},
			wantView: populatedTable,
		},
		{
			name: "Valid/NoImplementations",
			compDefs: []oscalTypes.ComponentDefinition{
				{
					Components: &[]oscalTypes.DefinedComponent{
						{
							Title: "MySoftware",
							Type:  "service",
						},
					},
				},
			},
			wantView: emptyTable,
		},
		{
			name:     "Invalid/NoComponentDefinitions",
			compDefs: nil,
			wantErr:  "component definitions inputs cannot be empty",
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			initialModel, err := ShowDefinitionTable(c.compDefs)
			if c.wantErr != "" {
				require.EqualError(t, err, c.wantErr)
			} else {
				require.NoError(t, err)
				gotView := initialModel.View()
				require.Equal(t, c.wantView, gotView)
			}

		})
	}
}

var (
	emptyTable = `┌────────────────────────────────────────────────────────────────────────────────────┐
│ Framework ID                    Supported Components                               │
│────────────────────────────────────────────────────────────────────────────────────│
│                                                                                    │
│                                                                                    │
│                                                                                    │
│                                                                                    │
│                                                                                    │
│                                                                                    │
└────────────────────────────────────────────────────────────────────────────────────┘
`
	populatedTable = `┌────────────────────────────────────────────────────────────────────────────────────┐
│ Framework ID                    Supported Components                               │
│────────────────────────────────────────────────────────────────────────────────────│
│ anotherexample                  MySoftware                                         │
│ example                         MySoftware                                         │
│                                                                                    │
│                                                                                    │
│                                                                                    │
│                                                                                    │
└────────────────────────────────────────────────────────────────────────────────────┘
`
)
