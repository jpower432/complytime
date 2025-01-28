// SPDX-License-Identifier: Apache-2.0

package complytime

import (
	"path/filepath"
	"testing"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/stretchr/testify/require"
)

func TestPlan(t *testing.T) {
	tmpDir := t.TempDir()
	testPlanPath := filepath.Join(tmpDir, "assessment-plan.json")

	// Testing reading and writing Assessment plan workflows

	testPlan := oscalTypes.AssessmentPlan{
		UUID: "228ff6d0-0d67-4c15-9c16-ece9a554c4de",
		Metadata: oscalTypes.Metadata{
			Title:        "example",
			OscalVersion: "1.1.2",
			Version:      "1.0.0",
		},
	}

	err := WritePlan(&testPlan, testPlanPath)
	require.NoError(t, err)

	_, err = PlanSettings(testPlanPath)
	require.ErrorIs(t, err, ErrNoActivities)

	localDefs := oscalTypes.LocalDefinitions{
		Activities: &[]oscalTypes.Activity{
			{
				Description: "activity",
				Title:       "my-activity",
				UUID:        "228ff6d0-0d67-4c15-9c16-ece9a554c4df",
			},
		},
	}
	testPlan.LocalDefinitions = &localDefs

	err = WritePlan(&testPlan, testPlanPath)
	require.NoError(t, err)

	_, err = PlanSettings(testPlanPath)
	require.NoError(t, err)
}
