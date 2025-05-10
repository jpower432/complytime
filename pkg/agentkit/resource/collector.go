// SPDX-License-Identifier: Apache-2.0
package resource

import (
	"context"

	"github.com/revanite-io/sci/layer4"
)

// Collector collects evaluations and resource data
type Collector interface {
	Collect(ctx context.Context) (Resource, *layer4.Layer4, error)
}

var _ Collector = (*NoopCollector)(nil)

type NoopCollector struct{}

func (n NoopCollector) Collect(ctx context.Context) (Resource, *layer4.Layer4, error) {
	//TODO implement me
	panic("implement me")
}
