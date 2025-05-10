// SPDX-License-Identifier: Apache-2.0
package resource

import (
	"context"

	"github.com/revanite-io/sci/layer4"
)

type Resource struct {
	ID      string
	Payload any
}

// AuditArtifact represents a type of artifact that would be review as part
// of an audit.
type AuditArtifact interface {
	Attach(resource Resource, eval layer4.Layer4) error
}

// Exportable define methods to export the artifact to a centralized location.
type Exportable interface {
	Export(ctx context.Context) error
}

// ExportableArtifact that can be exported to a centralized location.
type ExportableArtifact interface {
	AuditArtifact
	Exportable
}
