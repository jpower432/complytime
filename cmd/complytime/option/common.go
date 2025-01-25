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
