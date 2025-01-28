// SPDX-License-Identifier: Apache-2.0

package option

import (
	"io"

	"github.com/spf13/pflag"
)

// Common options for the ComplyTime CLI.
type Common struct {
	Debug bool
	Output
}

// Output options for
type Output struct {
	// Out think, os.Stdout
	Out io.Writer
	// ErrOut think, os.Stderr
	ErrOut io.Writer
}

// BindFlags populate Common options from user-specified flags.
func (o *Common) BindFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.Debug, "debug", "d", false, "output debug logs")
}

// ComplyTime options are configurations needed for the ComplyTime CLI to run.
// They are less generic the Common options and would only be used in a subset of
// commands.
type ComplyTime struct {
	UserWorkspace string
}

// BindFlags populate ComplyTime options from user-specified flags.
func (o *ComplyTime) BindFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.UserWorkspace, "workspace", "w", ".", "workspace to use for artifact generation")
}
