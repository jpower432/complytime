// SPDX-License-Identifier: Apache-2.0
package agentkit

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/complytime/complytime/pkg/agentkit/resource"
)

type Agent struct {
	collector     resource.Collector
	targetCatalog string
}

// NewAgent creates a new Agent.
func NewAgent(collector resource.Collector, targetCatalog string) *Agent {
	agent := &Agent{collector: collector}
	return agent
}

type runOptions struct {
	enableOtel bool
	exportURL  string
}

func (o *runOptions) defaults() {
	o.exportURL = "localhost:8080"
}

type RunOption func(ro *runOptions)

func RunWithInstrumentation(in bool) RunOption {
	return func(ro *runOptions) {
		ro.enableOtel = true
	}
}

// Perhaps set the exporter object instead?

func RunWithExporterURL(url string) RunOption {
	return func(ro *runOptions) {
		ro.exportURL = url
	}
}

func (a *Agent) Run(ctx context.Context, opts ...RunOption) error {
	options := runOptions{}
	options.defaults()
	for _, opt := range opts {
		opt(&options)
	}

	if options.enableOtel {
		metricExporter, err := otlpmetricgrpc.New(ctx)
		if err != nil {
			return err
		}

		meterProvider := metric.NewMeterProvider(
			metric.WithReader(metric.NewPeriodicReader(metricExporter,
				metric.WithInterval(3*time.Second))),
		)

		otel.SetMeterProvider(meterProvider)
		defer func() {
			err := meterProvider.Shutdown(ctx)
			if err != nil {
				fmt.Println(err)
			}
		}()
	}

	// Implement cancel if cancellation signal is passed.
	var wg sync.WaitGroup
	errs := make(chan error)
	errsDone := make(chan struct{})

	var resultErrs []error
	go func() {
		for err := range errs {
			resultErrs = append(resultErrs, err)
		}
		errsDone <- struct{}{}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		rs, err := a.collector.Collect(ctx)
		if err != nil {
			errs <- err
		}

		plan, err := a.collector.Plan(ctx, a.targetCatalog)
		if err != nil {
			errs <- err
		}

		artifact := resource.NewAttestation(options.exportURL)
		err = artifact.Attach(rs, plan)
		if err != nil {
			errs <- err
		}

		err = artifact.Export(ctx)
		if err != nil {
			errs <- err
		}
	}()

	wg.Wait()
	close(errs)

	<-errsDone

	return errors.Join(resultErrs...)
}
