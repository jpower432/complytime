// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"os"

	"github.com/goccy/go-yaml"
	"github.com/revanite-io/sci/layer2"
	"github.com/spf13/cobra"

	"github.com/complytime/complytime/cmd/complytime/option"
	"github.com/complytime/complytime/pkg/agentkit"
	"github.com/complytime/complytime/pkg/agentkit/resource"
)

// agentOptions defined options for the agent subcommand.
type agentOptions struct {
	*option.Common
	archivistaURL string
	catalogPath   string
}

// agentCmd creates a new cobra.Command for the agent subcommand.
func agentCmd(common *option.Common) *cobra.Command {
	agentOpts := &agentOptions{
		Common: common,
	}
	cmd := &cobra.Command{
		Use:          "agent [flags]",
		Short:        "Collect and export Layer 4 evaluations",
		Long:         "An agent for deploying with assessment tools to collect and export Layer 4 evaluations",
		Example:      "complytime agent",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			catalog, err := getCatalog(agentOpts.catalogPath)
			if err != nil {
				return err
			}
			collector := resource.NoopCollector{}

			// Needed?
			err = collector.AddCatalog(catalog)
			if err != nil {
				return err
			}

			agent := agentkit.NewAgent(collector, catalog.Metadata.Id)
			err = agent.Run(cmd.Context(), agentkit.RunWithExporterURL(agentOpts.archivistaURL))
			if err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&agentOpts.archivistaURL, "archivista-url", "a", "localhost:8081", "URL to archivista instance")
	cmd.Flags().StringVarP(&agentOpts.catalogPath, "catalogPath", "p", "", "Path to layer 2 catalog")
	agentOpts.BindFlags(cmd.Flags())
	return cmd
}

func getCatalog(filepath string) (layer2.Layer2, error) {
	var catalog layer2.Layer2
	yamlFile, err := os.ReadFile(filepath)
	if err != nil {
		return catalog, err
	}
	err = yaml.Unmarshal(yamlFile, &catalog)
	if err != nil {
		return catalog, err
	}
	return catalog, nil
}
