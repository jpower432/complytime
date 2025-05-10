// SPDX-License-Identifier: Apache-2.0
package resource

import (
	"context"
	"encoding/json"
	"os"

	"github.com/revanite-io/sci/layer4"

	"github.com/complytime/complytime/pkg/agentkit/transformer"
)

var _ ExportableArtifact = (*OSCALDocument)(nil)

type OSCALDocument struct {
	filePath string
	evals    []layer4.Layer4
}

func NewOSCALArtifact() *OSCALDocument {
	return &OSCALDocument{}
}

func (o *OSCALDocument) Attach(resource Resource, eval layer4.Layer4) error {
	//TODO: Handle resource
	// For OSCAL, this would be the resource subject type
	o.evals = append(o.evals, eval)
	return nil
}

func (o *OSCALDocument) Export(_ context.Context) error {
	plan, err := transformer.ToAssessmentPlan(o.evals)
	if err != nil {
		return err
	}

	result, err := transformer.ToAssessmentResults(plan, o.evals)
	if err != nil {
		return err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return os.WriteFile(o.filePath, data, os.ModeAppend)
}
