// SPDX-License-Identifier: Apache-2.0

package complytime

import (
	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/oscal-compass/oscal-sdk-go/generators"
)

// The assessment results md file needs the catalog title information
// LoadCatalogSource returns an OSCAL catalogs from a given application directory.

func LoadCatalogSource(appDir ApplicationDirectory, catalogSource string) (*oscalTypes.Catalog, error) {
	sourceFile, err := findControlSource(appDir, catalogSource)
	if err != nil {
		return nil, err
	}
	defer sourceFile.Close()
	return generators.NewCatalog(sourceFile)
}
