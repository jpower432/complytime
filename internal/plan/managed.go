// SPDX-License-Identifier: Apache-2.0

package plan

import (
	"errors"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/oscal-compass/oscal-sdk-go/extensions"
)

type Managed struct {
	plan         *oscalTypes.AssessmentPlan
	framkeworkId string
}

// NewManaged returns a Managed for a given Assessment Plan.
func NewManaged(plan *oscalTypes.AssessmentPlan, frameworkId string) *Managed {
	editor := &Managed{
		plan:         plan,
		framkeworkId: frameworkId,
	}
	return editor
}

// Parse the plan into populated EditableFields.
func (m *Managed) Parse() *EditableFields {
	fields := &EditableFields{
		Controls:                  make(map[string]ControlState),
		Rules:                     make(map[string]RuleState),
		OrganizationDefinedParams: make(map[string]string),
	}

	if m.plan.LocalDefinitions == nil || m.plan.LocalDefinitions.Activities == nil {
		return fields
	}

	for _, activity := range *m.plan.LocalDefinitions.Activities {
		fields.Rules[activity.Title] = Verify
		if activity.Props != nil {
			props := extensions.FindAllProps(*activity.Props, extensions.WithClass(extensions.TestParameterClass))
			for _, prop := range props {
				fields.OrganizationDefinedParams[prop.Name] = prop.Value
			}
		}
		if activity.RelatedControls != nil {
			for _, controls := range activity.RelatedControls.ControlSelections {
				if controls.IncludeControls == nil {
					continue
				}
				for _, include := range *controls.IncludeControls {
					fields.Controls[include.ControlId] = Assess
				}
			}
		}
	}
	return fields
}

func (m *Managed) Save(planLocation string, fields EditableFields) error {

	// Deleted activity to delete associated activities
	deletedActivies := make(map[string]struct{})

	var keptActivities []oscalTypes.Activity
	for _, activity := range *m.plan.LocalDefinitions.Activities {
		hasControls := updateActivity(&activity, fields)
		if hasControls {
			keptActivities = append(keptActivities, activity)
		} else {
			deletedActivies[activity.UUID] = struct{}{}
		}
	}

	var keptAssocActivities []oscalTypes.AssociatedActivity
	if m.plan.Tasks == nil || len(*m.plan.Tasks) != 0 {
		return errors.New("bug: tasks should be length of 1")
	}

	tasks := *m.plan.Tasks
	for _, assocActivity := range *tasks[0].AssociatedActivities {
		if _, deleted := deletedActivies[assocActivity.ActivityUuid]; !deleted {
			keptAssocActivities = append(keptAssocActivities, assocActivity)
		}
	}

	tasks[0].AssociatedActivities = &keptAssocActivities
	m.plan.LocalDefinitions.Activities = &keptActivities

	// Remove any activities with no assessed Controls
	// Remove associated activities

	empty := updateReviewedControls(&m.plan.ReviewedControls, fields)
	if empty {
		return errors.New("reviewed controls cannot be empty")
	}

	return WritePlan(m.plan, m.framkeworkId, planLocation)
}

// updateActivity updates the Activity with editable fields and returns a boolean value of
// whether the Activity has controls.
func updateActivity(activity *oscalTypes.Activity, fields EditableFields) bool {
	empty := updateReviewedControls(activity.RelatedControls, fields)
	return !empty
}

// updateReviewedControl updates the ReviewedControls with editable fields and returns a boolean value of
// whether the ReviewedControls are empty.
func updateReviewedControls(reviewedControls *oscalTypes.ReviewedControls, fields EditableFields) bool {
	var keptControls []oscalTypes.AssessedControlsSelectControlById
	for _, controls := range reviewedControls.ControlSelections {
		if controls.IncludeControls == nil {
			continue
		}
		for _, include := range *controls.IncludeControls {
			controlState, ok := fields.Controls[include.ControlId]
			if ok || controlState == Assess {
				keptControls = append(keptControls, include)
			}
		}
	}

	if len(keptControls) == 0 {
		return true
	}

	reviewedControls.ControlSelections = []oscalTypes.AssessedControls{
		{
			IncludeControls: &keptControls,
		},
	}

	return false
}
