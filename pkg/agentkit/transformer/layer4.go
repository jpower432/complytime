// SPDX-License-Identifier: Apache-2.0
package transformer

import (
	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/revanite-io/sci/layer4"
)

func ToAssessmentPlan(evals []layer4.Layer4) (oscalTypes.AssessmentPlan, error) {
	return oscalTypes.AssessmentPlan{}, nil
}

func ToAssessmentResults(plan oscalTypes.AssessmentPlan, evals []layer4.Layer4) (oscalTypes.AssessmentResults, error) {
	return oscalTypes.AssessmentResults{}, nil
}
