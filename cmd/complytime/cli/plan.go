// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/oscal-compass/oscal-sdk-go/transformers"

	"github.com/complytime/complytime/cmd/complytime/option"
	"github.com/complytime/complytime/internal/complytime"
)

const assessmentPlanLocation = "assessment-plan.json"

// PlanOptions defines options for the "plan" subcommand
type planOptions struct {
	*option.Common
	userWorkspace string
	frameworkID   string
}

func setOptsPlanFromArgs(args []string, opts *planOptions) {
	if len(args) == 1 {
		opts.frameworkID = filepath.Clean(args[0])
	}
}

// planCmd creates a new cobra.Command for the "plan" subcommand
func planCmd(common *option.Common) *cobra.Command {
	planOpts := &planOptions{Common: common}
	cmd := &cobra.Command{
		Use:     "plan [flags] id",
		Short:   "Generate a new assessment plan for a given compliance framework id.",
		Example: "complytime plan myframework",
		Args:    cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			setOptsPlanFromArgs(args, planOpts)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runPlan(cmd, planOpts)
		},
	}

	cmd.Flags().StringVarP(&planOpts.userWorkspace, "workspace", "w", ".", "workspace to use for artifact generation")
	return cmd
}

func runPlan(cmd *cobra.Command, opts *planOptions) error {
	// Create the application directory if it does not exist
	appDir, err := complytime.NewApplicationDirectory(true)
	if err != nil {
		return err
	}
	componentDefs, err := complytime.FindComponentDefinitions(appDir.BundleDir())
	if err != nil {
		return err
	}
	assessmentPlan, err := transformers.ComponentDefinitionsToAssessmentPlan(cmd.Context(), componentDefs, opts.frameworkID)
	if err != nil {
		return err
	}

	assessmentPlanData, err := json.MarshalIndent(assessmentPlan, "", " ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(opts.userWorkspace, assessmentPlanLocation)
	cleanedPath := filepath.Clean(filePath)

	if err := os.WriteFile(cleanedPath, assessmentPlanData, 0640); err != nil {
		return err
	}

	return nil
}
