// SPDX-License-Identifier: Apache-2.0

package terminal

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/complytime/complytime/internal/complytime"
)

func TestShowDefinitionTable(t *testing.T) {
	tests := []struct {
		name       string
		frameworks []complytime.Framework
		wantView   string
	}{
		{
			name: "Valid/HappyPath",
			frameworks: []complytime.Framework{
				{
					ID:                  "anotherexample",
					Title:               "Example Profile (moderate)",
					SupportedComponents: []string{"My Software"},
				},
				{
					ID:                  "example",
					Title:               "Example Profile (low)",
					SupportedComponents: []string{"My Software"},
				},
			},
			wantView: populatedTable,
		},
		{
			name:       "Valid/Empty",
			frameworks: []complytime.Framework{},
			wantView:   emptyTable,
		},
		{
			name: "Valid/LongTitle",
			frameworks: []complytime.Framework{
				{
					ID:                  "anotherexample",
					Title:               "This is a very very very long title (moderate)",
					SupportedComponents: []string{"My Software"},
				},
			},
			wantView: longTitle,
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			initialModel := ShowDefinitionTable(c.frameworks)
			gotView := initialModel.View()
			require.Equal(t, c.wantView, gotView)
		})
	}
}

var (
	emptyTable = `┌──────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ Title                           Framework ID          Supported Components                               │
│──────────────────────────────────────────────────────────────────────────────────────────────────────────│
│                                                                                                          │
│                                                                                                          │
│                                                                                                          │
│                                                                                                          │
│                                                                                                          │
│                                                                                                          │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────┘

Choose an option from the Framework ID column to use with complytime plan.
`
	populatedTable = `┌──────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ Title                           Framework ID          Supported Components                               │
│──────────────────────────────────────────────────────────────────────────────────────────────────────────│
│ Example Profile (moderate)      anotherexample        My Software                                        │
│ Example Profile (low)           example               My Software                                        │
│                                                                                                          │
│                                                                                                          │
│                                                                                                          │
│                                                                                                          │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────┘

Choose an option from the Framework ID column to use with complytime plan.
`
	longTitle = `┌──────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ Title                           Framework ID          Supported Components                               │
│──────────────────────────────────────────────────────────────────────────────────────────────────────────│
│ This is a very very very long…  anotherexample        My Software                                        │
│                                                                                                          │
│                                                                                                          │
│                                                                                                          │
│                                                                                                          │
│                                                                                                          │
└──────────────────────────────────────────────────────────────────────────────────────────────────────────┘

Choose an option from the Framework ID column to use with complytime plan.
`
)
