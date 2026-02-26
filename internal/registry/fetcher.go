// SPDX-License-Identifier: Apache-2.0

package registry

import "context"

// Fetcher abstracts registry access for testing without a live OCI registry.
type Fetcher interface {
	GetDefinitions(ctx context.Context, modulePath, version string) ([]byte, error)
	DefinitionVersion(ctx context.Context, modulePath string) (string, string, error)
}
