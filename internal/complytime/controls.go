// SPDX-License-Identifier: Apache-2.0

package complytime

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"
	"github.com/oscal-compass/oscal-sdk-go/generators"
	"github.com/oscal-compass/oscal-sdk-go/settings"
)

// Framework represents an implemented compliance framework across
// component sources.
type Framework struct {
	// ID is the short-name identifier that is used to consistently
	// represent a framework.
	ID string
	// Title is the human-readable name for a framework
	Title string
	// SupportedComponents define the component titles that implement the
	// framework.
	SupportedComponents []string
}

// LoadFrameworks returns all loaded framework information from a given application directory.
func LoadFrameworks(appDir ApplicationDirectory) ([]Framework, error) {
	definitions, err := FindComponentDefinitions(appDir.BundleDir())
	if err != nil {
		return nil, fmt.Errorf("error finding component defintions in %s: %w", appDir.BundleDir(), err)
	}

	byFramework := make(map[string]Framework)
	for _, definition := range definitions {
		if definition.Components != nil {
			for _, comp := range *definition.Components {
				// The goal is to only display the target component and abstract away the
				// validation implementation details, so skipping.
				if comp.Type == "validation" {
					continue
				}
				frameworks, err := processComponent(appDir, comp)
				if err != nil {
					return nil, err
				}
				for _, inputFramework := range frameworks {
					framework, ok := byFramework[inputFramework.ID]
					if !ok {
						framework = inputFramework
						framework.SupportedComponents = []string{}
					}
					framework.SupportedComponents = append(framework.SupportedComponents, comp.Title)
					byFramework[framework.ID] = framework
				}
			}
		}
	}

	var frameworks []Framework
	for _, framework := range byFramework {
		frameworks = append(frameworks, framework)
	}
	return frameworks, nil
}

func processComponent(appDir ApplicationDirectory, component oscalTypes.DefinedComponent) ([]Framework, error) {
	if component.ControlImplementations == nil {
		return nil, nil
	}
	var frameworks []Framework
	for _, implementation := range *component.ControlImplementations {
		frameworkShortName, found := settings.GetFrameworkShortName(implementation)
		if !found {
			return nil, fmt.Errorf("no framework information found for implemenation %q", implementation.Description)
		}

		// Load profile of local and get more information for the description
		profile, err := LoadControlSource(appDir, implementation.Source)
		if err != nil {
			return nil, fmt.Errorf("error loading control source %s for component %s: %w", frameworkShortName, component.Title, err)
		}

		newFramework := Framework{
			ID:    frameworkShortName,
			Title: profile.Metadata.Title,
		}
		frameworks = append(frameworks, newFramework)

	}
	return frameworks, nil
}

// LoadControlSource returns an OSCAL profiles from a given control source from a given application directory.
func LoadControlSource(appDir ApplicationDirectory, controlSource string) (*oscalTypes.Profile, error) {
	uri, err := url.ParseRequestURI(controlSource)
	if err != nil {
		return nil, err
	}

	path := uri.Host + uri.Path
	appDirPath := appDir.AppDir()

	// Handle the two supported cases:
	// An absolute path or
	// A path relative to the root of the complytime application directory
	if !filepath.IsAbs(path) {
		if !strings.HasPrefix(path, ControlsDir+string(os.PathSeparator)) {
			return nil, fmt.Errorf("got path %s, control source is expected to be under path %s", path, appDir.ControlDir())
		}
		path = filepath.Join(appDirPath, path)
	}
	cleanedPath := filepath.Clean(path)

	sourceFile, err := os.Open(cleanedPath)
	if err != nil {
		return nil, err
	}
	defer sourceFile.Close()

	return generators.NewProfile(sourceFile)
}
