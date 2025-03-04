// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"os"
	"path/filepath"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/oscal-compass/compliance-to-policy-go/v2/framework"
	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"github.com/spf13/cobra"

	"github.com/complytime/complytime/internal/complytime"

	"github.com/oscal-compass/oscal-sdk-go/generators"
	"github.com/oscal-compass/oscal-sdk-go/settings"

	"github.com/complytime/complytime/cmd/complytime/option"
)

const assessmentResultsLocation = "assessment-results.json"

// scanOptions defined options for the scan subcommand.
type scanOptions struct {
	*option.Common
	complyTimeOpts *option.ComplyTime
}

// scanCmd creates a new cobra.Command for the version subcommand.
func scanCmd(common *option.Common) *cobra.Command {
	scanOpts := &scanOptions{
		Common:         common,
		complyTimeOpts: &option.ComplyTime{},
	}
	cmd := &cobra.Command{
		Use:          "scan [flags]",
		Short:        "Scan environment with assessment plan",
		Example:      "complytime scan",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		PreRun: func(cmd *cobra.Command, _ []string) {
			enableDebug(logger, common)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runScan(cmd, scanOpts)
		},
	}
	scanOpts.complyTimeOpts.BindFlags(cmd.Flags())
	return cmd
}

func runScan(cmd *cobra.Command, opts *scanOptions) error {

	planSettings, err := getPlanSettingsForWorkspace(opts.complyTimeOpts)
	if err != nil {
		return err
	}
	// Create the application directory if it does not exist
	appDir, err := complytime.NewApplicationDirectory(true)
	if err != nil {
		return err
	}
	logger.Debug(fmt.Sprintf("Using application directory: %s", appDir.AppDir()))

	cfg, err := complytime.Config(appDir)
	if err != nil {
		return err
	}

	// set config logger to CLI charm logger
	cfg.Logger = logger

	manager, err := framework.NewPluginManager(cfg)
	if err != nil {
		return fmt.Errorf("error initializing plugin manager: %w", err)
	}
	plugins, err := manager.LaunchPolicyPlugins()
	if err != nil {
		return err
	}
	for _, plugin := range plugins {
		logger.Debug(fmt.Sprintf("Successfully loaded %v plugin.", plugin))
	}
	logger.Info("Information successfully retrieved from plugins.")
	// Ensure all the plugins launch above are cleaned up
	defer manager.Clean()

	allResults, err := manager.AggregateResults(cmd.Context(), plugins, planSettings)
	if err != nil {
		return err
	}

	apPath := filepath.Join(opts.complyTimeOpts.UserWorkspace, assessmentPlanLocation)
	apCleanedPath := filepath.Clean(apPath)
	assessmentPlan, err := loadAssessmentPlan(apCleanedPath)
	if err != nil {
		return err
	}

	frameworkProp, valid := extensions.GetTrestleProp(extensions.FrameworkProp, *assessmentPlan.Metadata.Props)
	if !valid {
		return fmt.Errorf("error reading framework property from assessment plan")
	}
	logger.Debug(fmt.Sprintf("Framework property was successfully read from the assessment plan: %v.", frameworkProp))
	r, err := framework.NewReporter(cfg)
	if err != nil {
		return err
	}

	var allImplementations []oscalTypes.ControlImplementationSet
	for _, compDef := range cfg.ComponentDefinitions {
		for _, component := range *compDef.Components {
			if component.ControlImplementations == nil {
				continue
			}
			allImplementations = append(allImplementations, *component.ControlImplementations...)

		}
	}

	implementationSettings, err := settings.Framework(frameworkProp.Value, allImplementations)
	if err != nil {
		return err
	}

	planHref := fmt.Sprintf("file://%s", apCleanedPath)
	assessmentResults, err := r.GenerateAssessmentResults(cmd.Context(), planHref, implementationSettings, allResults)
	if err != nil {
		return err
	}

	filePath := filepath.Join(opts.complyTimeOpts.UserWorkspace, assessmentResultsLocation)
	cleanedPath := filepath.Clean(filePath)

	err = complytime.WriteAssessmentResults(&assessmentResults, cleanedPath)
	if err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("The assessment results were successfully written to %v.", assessmentResultsLocation))
	return nil
}

// Load assessment plan from assessment-plan.json
func loadAssessmentPlan(filePath string) (*oscalTypes.AssessmentPlan, error) {

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	assessmentPlan, err := generators.NewAssessmentPlan(file)
	if err != nil {
		return nil, err
	}
	return assessmentPlan, nil
}
