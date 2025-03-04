// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"

	"github.com/oscal-compass/compliance-to-policy-go/v2/framework"
	"github.com/spf13/cobra"

	"github.com/complytime/complytime/cmd/complytime/option"
	"github.com/complytime/complytime/internal/complytime"
)

// generateOptions defines options for the "generate" subcommand
type generateOptions struct {
	*option.Common
	complyTimeOpts *option.ComplyTime
}

// generateCmd creates a new cobra.Command for the "generate" subcommand
func generateCmd(common *option.Common) *cobra.Command {
	generateOpts := &generateOptions{
		Common:         common,
		complyTimeOpts: &option.ComplyTime{},
	}
	cmd := &cobra.Command{
		Use:     "generate [flags]",
		Short:   "Generate PVP policy from an assessment plan",
		Example: "complytime generate",
		Args:    cobra.NoArgs,
		PreRun: func(cmd *cobra.Command, _ []string) {
			enableDebug(logger, common)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runGenerate(cmd, generateOpts)
		},
	}
	generateOpts.complyTimeOpts.BindFlags(cmd.Flags())
	return cmd
}

func runGenerate(cmd *cobra.Command, opts *generateOptions) error {

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
	logger.Debug("The configuration from the C2PConfig was successfully loaded.")

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

	// Ensure all the plugins launch above are cleaned up
	defer manager.Clean()

	err = manager.GeneratePolicy(cmd.Context(), plugins, planSettings)
	if err != nil {
		return err
	}
	logger.Info("Policy generation completed successfully.")
	return nil
}
