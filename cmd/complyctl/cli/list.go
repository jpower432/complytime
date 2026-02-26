// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/complytime/complyctl/internal/cache"
	"github.com/complytime/complyctl/internal/complytime"
	"github.com/complytime/complyctl/internal/policy"
	"github.com/complytime/complyctl/internal/terminal"
)

type listOptions struct {
	*Common
	plain    bool
	policyID string
	cacheDir string
}

func listCmd(common *Common) *cobra.Command {
	o := &listOptions{
		Common: common,
	}
	cmd := &cobra.Command{
		Use:          "list [flags]",
		Short:        "List cached Gemara policies",
		SilenceUsage: true,
		Example:      "complyctl list\n  complyctl list --plain\n  complyctl list --policy-id nist-800-53-r5",
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := o.validate(); err != nil {
				return err
			}
			if err := o.complete(); err != nil {
				return err
			}
			return o.run(cmd.Context())
		},
	}
	cmd.Flags().BoolVarP(&o.plain, "plain", "", false, "print table with minimal formatting")
	cmd.Flags().StringVar(&o.policyID, "policy-id", "", "Filter by policy ID")
	return cmd
}

func (o *listOptions) validate() error {
	return nil
}

func (o *listOptions) complete() error {
	var err error
	o.cacheDir, err = complytime.ResolveCacheDir()
	if err != nil {
		return fmt.Errorf("failed to resolve cache directory: %w", err)
	}
	return nil
}

func (o *listOptions) run(_ context.Context) error {
	ws := complytime.NewWorkspace()
	if err := ws.LoadAndValidate(); err != nil {
		return fmt.Errorf("failed to load complytime: %w", err)
	}

	cfg := ws.Config()

	cacheMgr := cache.NewCache(o.cacheDir)
	loader := policy.NewLoader(cacheMgr)

	cached, err := loader.ListCachedPolicies()
	if err != nil {
		logger.Error("List cached policies failed", "error", err)
		return fmt.Errorf("failed to list cached policies: %w", err)
	}

	for _, p := range cfg.Policies {
		ref := complytime.ParsePolicyRef(p.URL)
		if _, ok := cached[ref.Repository]; !ok {
			cached[ref.Repository] = []string{"(not cached — " + p.EffectiveID() + ")"}
		}
	}

	if o.policyID != "" {
		if versions, ok := cached[o.policyID]; ok {
			cached = map[string][]string{o.policyID: versions}
		} else {
			cached = map[string][]string{}
		}
	}

	return printGemaraPolicyTable(o.Out, cached, o.plain)
}

func printGemaraPolicyTable(w io.Writer, policies map[string][]string, plain bool) error {
	rows := make([]table.Row, 0, len(policies))
	for id, vers := range policies {
		sort.Strings(vers)
		rows = append(rows, table.Row{id, strings.Join(vers, ", ")})
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i][0] < rows[j][0] })

	columns := terminal.AutoColumnWidths([]table.Column{
		{Title: "Policy ID"},
		{Title: "Versions"},
	}, rows, 2)

	if plain {
		terminal.ShowPlainTable(w, columns, rows)
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
	fmt.Fprintln(w, tbl.View())
	return nil
}
