// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/complytime/complytime/cmd/complytime/option"
	"github.com/complytime/complytime/internal/complytime"
	"github.com/complytime/complytime/internal/terminal"
)

// listCmd creates a new cobra.Command for the "list" subcommand
func listCmd(common *option.Common) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list [flags]",
		Short:   "List information about supported frameworks and components.",
		Example: "complytime list",
		Args:    cobra.NoArgs,
		RunE:    func(_ *cobra.Command, _ []string) error { return list(common) },
	}
	return cmd
}

func list(opts *option.Common) error {
	appDir, err := complytime.NewApplicationDirectory(true)
	if err != nil {
		return err
	}
	compDefBundles, err := complytime.FindComponentDefinitions(appDir.BundleDir())
	if err != nil {
		return fmt.Errorf("error finding component defintions in %s: %w", appDir.BundleDir(), err)
	}
	model, err := terminal.ShowDefinitionTable(compDefBundles)
	if err != nil {
		return err
	}
	if _, err := tea.NewProgram(model, tea.WithOutput(opts.Out)).Run(); err != nil {
		return fmt.Errorf("failed to display component list: %w", err)
	}
	return nil
}
