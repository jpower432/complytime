// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"fmt"

	"github.com/complytime/complyctl/pkg/plugin"
)

// EvaluatorGroup bundles per-requirement configs for a single evaluator.
type EvaluatorGroup struct {
	EvaluatorID string
	Configs     []plugin.AssessmentConfiguration
}

// ValidateGlobalVars checks that global variables are present when evaluators
// are specified in the policy. Groups with an empty evaluator ID use broadcast
// mode and don't require explicit config.
// See FR-036, R48: specs/001-gemara-native-workflow/research.md
func ValidateGlobalVars(groups map[string]EvaluatorGroup, globalVars map[string]string, configPath string) error {
	if len(globalVars) > 0 {
		return nil
	}
	for evalID := range groups {
		if evalID != "" {
			return fmt.Errorf(
				"evaluator %q requires global variables in %s — add a 'variables' section with required fields (see provider documentation)",
				evalID, configPath,
			)
		}
	}
	return nil
}

// ExtractAssessmentConfigs converts a DependencyGraph into plugin-ready
// AssessmentConfiguration entries. EvaluatorID is set as a routing field
// on the struct — it is not injected into Parameters. Parameters should
// only carry per-requirement variable overrides for the plugin.
func ExtractAssessmentConfigs(policyID string, graph *DependencyGraph) []plugin.AssessmentConfiguration {
	configs := make([]plugin.AssessmentConfiguration, 0, len(graph.Assessments))

	for _, a := range graph.Assessments {
		configs = append(configs, plugin.AssessmentConfiguration{
			PlanID:        policyID,
			RequirementID: a.ID,
			Parameters:    a.Parameters,
			EvaluatorID:   a.EvaluatorID,
		})
	}

	return configs
}

// GroupByEvaluator groups assessment configs by EvaluatorID.
// See R32: specs/001-gemara-native-workflow/research.md
func GroupByEvaluator(configs []plugin.AssessmentConfiguration, graph *DependencyGraph) map[string]EvaluatorGroup {
	groups := make(map[string]EvaluatorGroup)

	if graph.EvaluatorID != "" {
		groups[graph.EvaluatorID] = EvaluatorGroup{
			EvaluatorID: graph.EvaluatorID,
			Configs:     configs,
		}
		return groups
	}

	for _, cfg := range configs {
		evalID := cfg.EvaluatorID
		g := groups[evalID]
		g.EvaluatorID = evalID
		g.Configs = append(g.Configs, cfg)
		groups[evalID] = g
	}

	return groups
}
