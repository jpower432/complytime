// SPDX-License-Identifier: Apache-2.0

package terminal

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))
	_ tea.Model = (*Model)(nil)
)

type Model struct {
	Table table.Model
	// displayed above table
	HeaderMsg string
	//displayed below table
	HelpMsg string
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func (m Model) View() string {

	output := baseStyle.Render(m.Table.View()) + "\n"

	if len(m.HeaderMsg) > 0 {
		output = m.HeaderMsg + "\n" + output
	}

	if len(m.HelpMsg) > 0 {
		output = output + m.HelpMsg + "\n"
	}
	return output
}

func WrapText(text string, lineWidth int) string {
	words := strings.Fields(text)
	var wrapped strings.Builder
	currentLineLength := 0

	for _, word := range words {
		if currentLineLength+len(word) > lineWidth {
			wrapped.WriteString("\n")
			currentLineLength = 0
		}
		wrapped.WriteString(word + " ")
		currentLineLength += len(word) + 1
	}
	return strings.TrimSpace(wrapped.String())
}

// AutoColumnWidths returns a copy of columns with widths set to the maximum
// display width (header or cell) plus padding. Uses lipgloss.Width to
// correctly measure multi-byte characters like emoji.
func AutoColumnWidths(columns []table.Column, rows []table.Row, padding int) []table.Column {
	out := make([]table.Column, len(columns))
	for i, col := range columns {
		maxW := lipgloss.Width(col.Title)
		for _, row := range rows {
			if i < len(row) {
				if w := lipgloss.Width(row[i]); w > maxW {
					maxW = w
				}
			}
		}
		out[i] = table.Column{Title: col.Title, Width: maxW + padding}
	}
	return out
}

// ShowPlainTable renders a plain text formatted table to writer.
func ShowPlainTable(writer io.Writer, columns []table.Column, rows []table.Row) {
	for _, col := range columns {
		_, _ = fmt.Fprintf(writer, "%-*s", col.Width, col.Title)
	}
	_, _ = fmt.Fprintln(writer)
	for _, row := range rows {
		for i, cell := range row {
			_, _ = fmt.Fprintf(writer, "%-*s", columns[i].Width, cell)
		}
		_, _ = fmt.Fprintln(writer)
	}
}
