// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/oscal-compass/oscal-sdk-go/settings"
	"github.com/oscal-compass/oscal-sdk-go/transformers"
	"github.com/spf13/cobra"

	"github.com/complytime/complytime/cmd/complytime/option"
	"github.com/complytime/complytime/internal/complytime"
)

const assessmentPlanLocation = "assessment-plan.json"

// PlanOptions defines options for the "plan" subcommand
type planOptions struct {
	*option.Common
	complyTimeOpts *option.ComplyTime
	frameworkID    string
}

func setOptsPlanFromArgs(args []string, opts *planOptions) {
	if len(args) == 1 {
		opts.frameworkID = filepath.Clean(args[0])
	}
}

// planCmd creates a new cobra.Command for the "plan" subcommand
func planCmd(common *option.Common) *cobra.Command {
	planOpts := &planOptions{
		Common:         common,
		complyTimeOpts: &option.ComplyTime{},
	}
	cmd := &cobra.Command{
		Use:     "plan [flags] id",
		Short:   "Generate a new assessment plan for a given compliance framework id.",
		Example: "complytime plan myframework",
		Args:    cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			enableDebug(logger, common)
			setOptsPlanFromArgs(args, planOpts)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runPlan(cmd, planOpts)
		},
	}
	planOpts.complyTimeOpts.BindFlags(cmd.Flags())
	return cmd
}

func runPlan(cmd *cobra.Command, opts *planOptions) error {
	// Create the application directory if it does not exist
	appDir, err := complytime.NewApplicationDirectory(true)
	if err != nil {
		return err
	}
	logger.Debug(fmt.Sprintf("Using application directory: %s", appDir.AppDir()))
	componentDefs, err := complytime.FindComponentDefinitions(appDir.BundleDir())
	if err != nil {
		return err
	}
	logger.Debug(fmt.Sprintf("Using bundle directory: %s for component definitions.", appDir.BundleDir()))
	assessmentPlan, err := transformers.ComponentDefinitionsToAssessmentPlan(cmd.Context(), componentDefs, opts.frameworkID)
	if err != nil {
		return err
	}

	filePath := filepath.Join(opts.complyTimeOpts.UserWorkspace, assessmentPlanLocation)
	cleanedPath := filepath.Clean(filePath)

	if err := complytime.WritePlan(assessmentPlan, opts.frameworkID, cleanedPath); err != nil {
		return fmt.Errorf("error writing assessment plan to %s: %w", cleanedPath, err)
	}
	logger.Info(fmt.Sprintf("Assessment plan written to %s\n", cleanedPath))
	return nil
}

// getPlanSettingsForWorkspace loads the assessment plan for the workspace and create new
// Settings from that data.
func getPlanSettingsForWorkspace(opts *option.ComplyTime) (settings.Settings, error) {
	// Load settings from assessment plan
	filePath := filepath.Join(opts.UserWorkspace, assessmentPlanLocation)
	cleanedPath := filepath.Clean(filePath)

	planSettings, err := complytime.PlanSettings(cleanedPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return planSettings, fmt.Errorf("error: assessment plan does exist in workspace %s: %w\n\nDid you run the plan command?",
				opts.UserWorkspace,
				err)
		}
		if errors.Is(err, complytime.ErrNoActivities) {
			return planSettings, fmt.Errorf("assessment plan %s does not have associated activities: %w", cleanedPath, err)
		}
		return planSettings, err
	}
	return planSettings, nil
}
