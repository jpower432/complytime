// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/complytime/complyctl/internal/complytime"
	"github.com/complytime/complyctl/internal/terminal"
	"github.com/complytime/complyctl/pkg/plugin"
)

type providersOptions struct {
	*Common
	pluginDir string
	plain     bool
}

// See R46: specs/001-gemara-native-workflow/research.md
func providersCmd(common *Common) *cobra.Command {
	o := &providersOptions{
		Common: common,
	}
	cmd := &cobra.Command{
		Use:   "providers",
		Short: "List discovered scanning providers and their health status",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := o.complete(); err != nil {
				return err
			}
			return o.run(cmd.Context())
		},
	}
	cmd.Flags().BoolVarP(&o.plain, "plain", "", false, "print table with minimal formatting")
	return cmd
}

func (o *providersOptions) complete() error {
	var err error
	o.pluginDir, err = complytime.ResolvePluginDir()
	if err != nil {
		return fmt.Errorf("failed to resolve plugin directory: %w", err)
	}
	return nil
}

// See FR-032, R38, R46: specs/001-gemara-native-workflow/spec.md
func (o *providersOptions) run(ctx context.Context) error {
	mgr, err := plugin.NewManager(o.pluginDir, logFile)
	if err != nil {
		return fmt.Errorf("plugin manager init failed: %w", err)
	}
	defer mgr.Cleanup()

	if err := mgr.LoadPlugins(); err != nil {
		return fmt.Errorf("plugin discovery failed: %w", err)
	}

	plugins := mgr.ListPlugins()
	if len(plugins) == 0 {
		fmt.Fprintf(os.Stderr, "No scanning providers found in %s\n", o.pluginDir)
		return nil
	}

	rows := make([]table.Row, 0, len(plugins))
	for _, lp := range plugins {
		status := "healthy"
		version := ""
		resp, err := lp.Client.HealthCheck(ctx, &plugin.HealthCheckRequest{})
		if err != nil {
			status = "ERROR"
		} else if !resp.Healthy {
			status = "unhealthy"
		} else {
			version = resp.Version
		}
		relPath, relErr := filepath.Rel(o.pluginDir, lp.Info.ExecutablePath)
		if relErr != nil {
			relPath = lp.Info.ExecutablePath
		}
		rows = append(rows, table.Row{
			lp.Info.EvaluatorID, relPath, status, version,
		})
	}

	columns := terminal.AutoColumnWidths([]table.Column{
		{Title: "Evaluator ID"},
		{Title: "Path"},
		{Title: "Status"},
		{Title: "Version"},
	}, rows, 2)

	if o.plain {
		terminal.ShowPlainTable(os.Stdout, columns, rows)
		return nil
	}

	tbl := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(7, len(rows)+2)),
	)
	tableStyle := table.DefaultStyles()
	tableStyle.Header = tableStyle.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	tbl.SetStyles(tableStyle)
	fmt.Fprintln(os.Stdout, tbl.View())
	return nil
}
