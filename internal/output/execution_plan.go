// SPDX-License-Identifier: Apache-2.0

package output

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"github.com/complytime/complyctl/internal/terminal"
)

// EvaluatorRoute describes one evaluator's role in the execution plan.
type EvaluatorRoute struct {
	EvaluatorID      string
	RequirementCount int
	PluginPath       string
	Status           string
}

// TargetScope describes one target's relationship to a policy and evaluators.
type TargetScope struct {
	TargetID     string
	PolicyID     string
	EvaluatorIDs []string
}

var (
	headerLabelStyle = lipgloss.NewStyle().Bold(true).Padding(0, 0, 1, 0)
	tableBorderStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240"))
)

// FormatExecutionPlan produces the two-table execution plan output using
// charmbracelet styled tables.
// See R36, R38: specs/001-gemara-native-workflow/research.md
func FormatExecutionPlan(policyID string, routes []EvaluatorRoute, scopes []TargetScope) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Execution Plan: %s\n\n", policyID)

	routeRows := make([]table.Row, 0, len(routes))
	for _, r := range routes {
		pluginPath := r.PluginPath
		if pluginPath == "" {
			pluginPath = "(not found)"
		}
		routeRows = append(routeRows, table.Row{
			r.EvaluatorID, strconv.Itoa(r.RequirementCount), pluginPath, r.Status,
		})
	}

	routeCols := terminal.AutoColumnWidths([]table.Column{
		{Title: "Evaluator"},
		{Title: "Requirements"},
		{Title: "Plugin"},
		{Title: "Status"},
	}, routeRows, 2)

	b.WriteString(headerLabelStyle.Render("Evaluator Routing:"))
	b.WriteString("\n")
	b.WriteString(renderStyledTable(routeCols, routeRows))
	b.WriteString("\n")

	scopeRows := make([]table.Row, 0, len(scopes))
	for _, s := range scopes {
		scopeRows = append(scopeRows, table.Row{
			s.TargetID, s.PolicyID, strings.Join(s.EvaluatorIDs, ", "),
		})
	}

	scopeCols := terminal.AutoColumnWidths([]table.Column{
		{Title: "Target"},
		{Title: "Policy"},
		{Title: "Evaluators"},
	}, scopeRows, 2)

	b.WriteString(headerLabelStyle.Render("Target Scope:"))
	b.WriteString("\n")
	b.WriteString(renderStyledTable(scopeCols, scopeRows))

	return b.String()
}

func renderStyledTable(columns []table.Column, rows []table.Row) string {
	tbl := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(len(rows)+1),
	)
	tableStyle := table.DefaultStyles()
	tableStyle.Header = tableStyle.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	tbl.SetStyles(tableStyle)
	return tableBorderStyle.Render(tbl.View())
}

// FormatPreScanSummary produces a brief one-line summary for normal scan mode.
// See FR-034: specs/001-gemara-native-workflow/spec.md
func FormatPreScanSummary(requirementCount int, evaluatorIDs []string, targetIDs []string) string {
	evals := strings.Join(evaluatorIDs, ", ")
	targets := strings.Join(targetIDs, ", ")
	return fmt.Sprintf("Scanning %d requirements via %s for target(s): %s...",
		requirementCount, evals, targets)
}
