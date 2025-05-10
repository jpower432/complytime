// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"

	"github.com/oscal-compass/compliance-to-policy-go/v2/framework"
	"github.com/oscal-compass/compliance-to-policy-go/v2/framework/actions"
	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"github.com/oscal-compass/oscal-sdk-go/validation"
	"github.com/revanite-io/sci/layer2"
	"github.com/spf13/cobra"

	"github.com/complytime/complytime/cmd/complytime/option"
	"github.com/complytime/complytime/internal/complytime"
	"github.com/complytime/complytime/pkg/agentkit/resource"
)

// evaluateOptions defined options for the scan subcommand.
type evaluateOptions struct {
	*option.Common
	complyTimeOpts *option.ComplyTime
	archivistaURL  string
}

// evaluateCmd creates a new cobra.Command for the evaluate subcommand.
func evaluateCmd(common *option.Common) *cobra.Command {
	scanOpts := &evaluateOptions{
		Common:         common,
		complyTimeOpts: &option.ComplyTime{},
	}
	cmd := &cobra.Command{
		Use:          "evaluate [flags]",
		Short:        "Scan environment with assessment plan",
		Example:      "complytime scan",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEvaluation(cmd, scanOpts)
		},
	}
	cmd.Flags().StringVarP(&scanOpts.archivistaURL, "archivista-url", "a", "localhost:8081", "URL to archivista instance")
	scanOpts.complyTimeOpts.BindFlags(cmd.Flags())
	return cmd
}

func runEvaluation(cmd *cobra.Command, opts *evaluateOptions) error {
	validator := validation.NewSchemaValidator()
	// Load settings from assessment plan
	ap, _, err := loadPlan(opts.complyTimeOpts, validator)
	if err != nil {
		return err
	}

	inputContext, err := complytime.ActionsContextFromPlan(ap)
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

	// Determine what profile to load from framework information captured
	// from state (assessment plan). This is required to populate complyTime required plugin options.
	frameworkProp, valid := extensions.GetTrestleProp(extensions.FrameworkProp, *ap.Metadata.Props)
	if !valid {
		return fmt.Errorf("error reading framework property from assessment plan")
	}
	opts.complyTimeOpts.FrameworkID = frameworkProp.Value
	logger.Debug(fmt.Sprintf("Framework property was successfully read from the assessment plan: %v.", frameworkProp))

	pluginOptions := opts.complyTimeOpts.ToPluginOptions()
	plugins, cleanup, err := complytime.Plugins(manager, inputContext, pluginOptions)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		return fmt.Errorf("errors launching plugins: %w", err)
	}
	logger.Info(fmt.Sprintf("Successfully loaded %v plugin(s).", len(plugins)))

	allResults, err := actions.AggregateResults(cmd.Context(), inputContext, plugins)
	if err != nil {
		return err
	}

	artifact := resource.NewAttestation(opts.archivistaURL)
	for _, result := range allResults {
		// TODO: ingest a real layer 2 catalog
		evaluation, err := actions.Evaluate(cmd.Context(), inputContext, []layer2.Control{}, result)
		if err != nil {
			return err
		}

		// TODO: Set resource here
		if err := artifact.Attach(resource.Resource{ID: "example"}, evaluation); err != nil {
			return err
		}
	}
	return artifact.Export(cmd.Context())
}
