// SPDX-License-Identifier: Apache-2.0

package resource

import (
	"context"

	"github.com/revanite-io/sci/layer2"
	"github.com/revanite-io/sci/layer4"
)

// Collector collects evaluations and resource data
type Collector interface {
	Collect(ctx context.Context) (Resource, error)
	Plan(ctx context.Context, catalogID string) (layer4.Layer4, error)
	AddCatalog(catalog layer2.Layer2) error
}

var _ Collector = (*NoopCollector)(nil)

type NoopCollector struct{}

func (n NoopCollector) Plan(ctx context.Context, catalogID string) (layer4.Layer4, error) {
	//TODO implement me
	panic("implement me")
}

func (n NoopCollector) AddCatalog(catalog layer2.Layer2) error {
	//TODO implement me
	panic("implement me")
}

func (n NoopCollector) Collect(ctx context.Context) (Resource, error) {
	//TODO implement me
	panic("implement me")
}
