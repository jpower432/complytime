// SPDX-License-Identifier: Apache-2.0

package terminal

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

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
