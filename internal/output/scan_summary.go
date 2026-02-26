// SPDX-License-Identifier: Apache-2.0

package output

import (
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/gemaraproj/go-gemara"

	"github.com/complytime/complyctl/internal/complytime"
	"github.com/complytime/complyctl/internal/terminal"
	"github.com/complytime/complyctl/pkg/plugin"
)

type nonPassingEntry struct {
	result  gemara.Result
	emoji   string
	message string
}

func nonPassingSortPriority(r gemara.Result) int {
	switch r {
	case gemara.Failed:
		return 1
	case gemara.Unknown:
		return 2
	case gemara.NeedsReview:
		return 3
	case gemara.NotApplicable, gemara.NotRun:
		return 4
	default:
		return 5
	}
}

func resultToEmoji(r gemara.Result) string {
	switch r {
	case gemara.Passed:
		return complytime.StatusPassed
	case gemara.Failed:
		return complytime.StatusFailed
	case gemara.NotApplicable, gemara.NotRun:
		return complytime.StatusSkipped
	case gemara.Unknown:
		return complytime.StatusError
	default:
		return complytime.StatusError
	}
}

// matchingStepMessage returns the message from the first step whose result
// matches the aggregated outcome. Falls back to the first step's message.
// See R45: scanning provider authors control the failure text.
func matchingStepMessage(steps []plugin.Step, target gemara.Result) string {
	for _, s := range steps {
		if pluginResultToGemara(s.Result) == target {
			return s.Message
		}
	}
	if len(steps) > 0 {
		return steps[0].Message
	}
	return ""
}

// FormatScanSummary builds ActionError-style post-scan output per R45/FR-037.
// Only non-passing results are listed individually (emoji + message per line).
// Passed results appear in the single-row totals table only.
// Non-passing lines are ordered by severity: failed -> error -> skipped.
func FormatScanSummary(assessments []plugin.AssessmentLog) string {
	var passCount, failCount, skipCount, errCount int
	var entries []nonPassingEntry

	for i := range assessments {
		a := &assessments[i]
		result := aggregateResultFromSteps(a.Steps)

		switch result {
		case gemara.Passed:
			passCount++
		case gemara.Failed:
			failCount++
			entries = append(entries, nonPassingEntry{
				result:  result,
				emoji:   complytime.StatusFailed,
				message: matchingStepMessage(a.Steps, result),
			})
		case gemara.NotApplicable, gemara.NotRun:
			skipCount++
			entries = append(entries, nonPassingEntry{
				result:  result,
				emoji:   complytime.StatusSkipped,
				message: matchingStepMessage(a.Steps, result),
			})
		default:
			errCount++
			entries = append(entries, nonPassingEntry{
				result:  result,
				emoji:   complytime.StatusError,
				message: matchingStepMessage(a.Steps, result),
			})
		}
	}

	slices.SortStableFunc(entries, func(a, b nonPassingEntry) int {
		return nonPassingSortPriority(a.result) - nonPassingSortPriority(b.result)
	})

	var b strings.Builder
	for _, e := range entries {
		fmt.Fprintf(&b, "%s %s\n", e.emoji, e.message)
	}

	if len(entries) > 0 {
		b.WriteString("\n")
	}

	totalsRow := table.Row{
		fmt.Sprintf("%d %s", passCount, complytime.StatusPassed),
		fmt.Sprintf("%d %s", failCount, complytime.StatusFailed),
		fmt.Sprintf("%d %s", skipCount, complytime.StatusSkipped),
		fmt.Sprintf("%d %s", errCount, complytime.StatusError),
	}
	cols := terminal.AutoColumnWidths([]table.Column{
		{Title: "Passed"}, {Title: "Failed"}, {Title: "Skipped"}, {Title: "Error"},
	}, []table.Row{totalsRow}, 2)

	b.WriteString(renderStyledTable(cols, []table.Row{totalsRow}))
	b.WriteString("\n")
	return b.String()
}
