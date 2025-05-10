// SPDX-License-Identifier: Apache-2.0
package resource

import (
	"context"
	"time"

	gowitness "github.com/in-toto/go-witness"
	"github.com/in-toto/go-witness/archivista"
	"github.com/in-toto/go-witness/attestation"
	"github.com/invopop/jsonschema"
	"github.com/revanite-io/sci/layer4"
)

var _ ExportableArtifact = (*Attestation)(nil)

type Attestation struct {
	archivistaURL string
	attestors     []attestation.Attestor
}

func NewAttestation(archivistaURL string) Attestation {
	return Attestation{archivistaURL: archivistaURL}
}

func (a *Attestation) Attach(resource Resource, eval layer4.Layer4) error {
	NewL4Attestor(eval, resource)
	return nil
}

func (a *Attestation) Export(ctx context.Context) error {
	// using this purposefully for not because I just want the one envelope
	runResults, err := gowitness.Run("step", gowitness.RunWithAttestors(a.attestors))
	if err != nil {
		return err
	}

	// export attestations to Archivista
	client := archivista.New(a.archivistaURL)
	_, err = client.Store(ctx, runResults.SignedEnvelope)
	if err != nil {
		return err
	}
	return nil
}

const Name = "layer4"
const Type = "https://github.com/revanite-io/sci/blob/main/schemas/layer-4.cue"
const RunType = attestation.VerifyRunType

type Layer4Attestor struct {
	eval     layer4.Layer4
	resource Resource
}

func NewL4Attestor(eval layer4.Layer4, resource Resource) Layer4Attestor {
	return Layer4Attestor{eval: eval, resource: resource}
}

func (l Layer4Attestor) Name() string {
	return Name
}

func (l Layer4Attestor) Type() string {
	return Type
}

func (l Layer4Attestor) RunType() attestation.RunType {
	return RunType
}

func (l Layer4Attestor) Attest(ctx *attestation.AttestationContext) error {
	l.eval.StartTime = time.Now()
	for _, controlEval := range l.eval.ControlEvaluations {
		for _, assessment := range controlEval.Assessments {
			for _, method := range assessment.Methods {
				method.Run = true
				// TODO: Setup changes
				result, err := method.RunMethod(l.resource, nil)
				if err != nil {
					return err
				}
				method.Result = &result
			}
		}
	}
	l.eval.EndTime = time.Now()
	return nil
}

func (l Layer4Attestor) Schema() *jsonschema.Schema {
	return jsonschema.Reflect(l.eval)
}
