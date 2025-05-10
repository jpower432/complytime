// SPDX-License-Identifier: Apache-2.0
package cli

import (
	"github.com/spf13/cobra"

	"github.com/complytime/complytime/cmd/complytime/option"
)

type auditOptions struct {
	*option.Common
	complyTimeOpts *option.ComplyTime
}

func auditCmd(common *option.Common) *cobra.Command {
	auditOpts := &auditOptions{
		Common:         common,
		complyTimeOpts: &option.ComplyTime{},
	}
	cmd := &cobra.Command{
		Use:          "audit [flags]",
		Short:        "Create package",
		Example:      "complytime audit",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runAudit(cmd, auditOpts)
		},
	}

	auditOpts.complyTimeOpts.BindFlags(cmd.Flags())
	return cmd
}

func runAudit(cmd *cobra.Command, opts *auditOptions) error {
	// Create an assessment result here for aggregated reporting by querying Archivista
	return nil
}
