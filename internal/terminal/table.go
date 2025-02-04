package terminal

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-2"

	"github.com/oscal-compass/oscal-sdk-go/settings"
)

// ShowDefinitionTable returned a Model to be used with a `bubbletea` Program that
// renders a table with framework and component information from given component definitions.
func ShowDefinitionTable(componentDefinitions []oscalTypes.ComponentDefinition) (tea.Model, error) {
	columns := []table.Column{
		{Title: "Framework ID", Width: 30},
		{Title: "Supported Components", Width: 50},
	}

	rows, err := makeAllRows(componentDefinitions)
	if err != nil {
		return nil, fmt.Errorf("error processing component defintion rows: %w", err)
	}

	tbl := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	tableStyle := table.DefaultStyles()
	tableStyle.Header = tableStyle.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	tableStyle.Cell = tableStyle.Cell.
		Foreground(lipgloss.Color("250")).
		Bold(true)
	tbl.SetStyles(tableStyle)

	return model{tbl}, nil
}

func makeAllRows(definitions []oscalTypes.ComponentDefinition) ([]table.Row, error) {
	var rows []table.Row
	compsByProfiles := make(map[string][]string)

	for _, definition := range definitions {
		if definition.Components != nil {
			for _, comp := range *definition.Components {
				if comp.Type == "validation" {
					continue
				}
				profiles, err := processComponent(comp)
				if err != nil {
					return nil, err
				}
				for _, profile := range profiles {
					comps, ok := compsByProfiles[profile]
					if !ok {
						comps = []string{}
					}
					comps = append(comps, comp.Title)
					compsByProfiles[profile] = comps
				}
			}
		}
	}

	for profile, comps := range compsByProfiles {
		row := table.Row{profile, strings.Join(comps, ", ")}
		rows = append(rows, row)
	}

	return rows, nil
}

func processComponent(component oscalTypes.DefinedComponent) ([]string, error) {
	if component.ControlImplementations == nil {
		return nil, nil
	}
	var frameworkIDs []string
	for _, implementation := range *component.ControlImplementations {
		frameworkShortName, found := settings.GetFrameworkShortName(implementation)
		if !found {
			return nil, fmt.Errorf("no framework information found for implemenation %q", implementation.Description)
		}
		frameworkIDs = append(frameworkIDs, frameworkShortName)
	}
	return frameworkIDs, nil
}

var (
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))
	_ tea.Model = (*model)(nil)
)

type model struct {
	table table.Model
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}
