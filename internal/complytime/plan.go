// SPDX-License-Identifier: Apache-2.0

package complytime

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/oscal-compass/oscal-sdk-go/generators"
	"github.com/oscal-compass/oscal-sdk-go/settings"
)

// WritePlan writes an AssessmentPlan to a given path location with consistency.
func WritePlan(plan *oscalTypes.AssessmentPlan, planLocation string) error {
	// To ensure we can easily read the plan once written, include under
	// OSCAL Model type to include the top-level "assessment-plan" key.
	oscalModels := oscalTypes.OscalModels{
		AssessmentPlan: plan,
	}
	assessmentPlanData, err := json.MarshalIndent(oscalModels, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(planLocation, assessmentPlanData, 0600)
}

var ErrNoActivities = errors.New("no local activities detected")

// PlanSettings return a new compliance Settings instance based on the
// given assessment plan path.
func PlanSettings(assessmentPlanPath string) (settings.Settings, error) {
	file, err := os.Open(assessmentPlanPath)
	if err != nil {
		return settings.Settings{}, err
	}
	plan, err := generators.NewAssessmentPlan(file)
	if err != nil {
		return settings.Settings{}, fmt.Errorf("failed to load assessment plan from %s: %w", assessmentPlanPath, err)
	}

	if plan.LocalDefinitions != nil && plan.LocalDefinitions.Activities != nil {
		return settings.NewAssessmentActivitiesSettings(*plan.LocalDefinitions.Activities), nil
	}
	return settings.Settings{}, ErrNoActivities
}
