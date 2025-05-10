// SPDX-License-Identifier: Apache-2.0
package transformer

import (
	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/revanite-io/sci/layer2"
)

type Layer2 struct {
	catalog layer2.Layer2
}

func (l *Layer2) ToComponentDefinition() (oscalTypes.ComponentDefinition, error) {
	return oscalTypes.ComponentDefinition{}, nil
}
