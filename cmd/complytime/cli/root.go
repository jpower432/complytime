// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"github.com/spf13/cobra"

	"github.com/complytime/complytime/cmd/complytime/option"
)

// New creates a new cobra.Command root for ComplyTime
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "complytime [command]",
		SilenceErrors: true,
		SilenceUsage:  false,
	}

	opts := option.Common{
		Output: option.Output{
			Out:    cmd.OutOrStdout(),
			ErrOut: cmd.ErrOrStderr(),
		},
	}
	opts.BindFlags(cmd.PersistentFlags())

	cmd.AddCommand(
		versionCmd(&opts),
		scanCmd(&opts),
		generateCmd(&opts),
		planCmd(&opts),
		listCmd(&opts),
	)

	return cmd
}
