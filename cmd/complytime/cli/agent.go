// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"

	"github.com/complytime/complytime/cmd/complytime/option"
	"github.com/complytime/complytime/pkg/agentkit"
	"github.com/complytime/complytime/pkg/agentkit/resource"
)

// agentOptions defined options for the agent subcommand.
type agentOptions struct {
	*option.Common
	archivistaURL string
}

// agentCmd creates a new cobra.Command for the agent subcommand.
func agentCmd(common *option.Common) *cobra.Command {
	agentOpts := &evaluateOptions{
		Common:         common,
		complyTimeOpts: &option.ComplyTime{},
	}
	cmd := &cobra.Command{
		Use:          "agent [flags]",
		Short:        "Collect and export Layer 4 evaluations",
		Long:         "An agent for deploying with assessment tools to collect and export Layer 4 evaluations",
		Example:      "complytime agent",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			agent := agentkit.NewAgent(resource.NoopCollector{})
			err := agent.Run(cmd.Context(), agentkit.RunWithExporterURL(agentOpts.archivistaURL))
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&agentOpts.archivistaURL, "archivista-url", "a", "localhost:8081", "URL to archivista instance")
	agentOpts.complyTimeOpts.BindFlags(cmd.Flags())
	return cmd
}
