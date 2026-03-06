package main

import (
	_ "embed"

	"github.com/complytime/complyctl/cmd/ampel-plugin/server"
	"github.com/complytime/complyctl/pkg/plugin"
)

//go:embed templates/command.md
var commandTemplate string

//go:embed templates/skill.md
var skillTemplate string

func main() {
	// Register init command handler
	if plugin.RegisterInit(plugin.InitTemplates{
		CommandTemplate: commandTemplate,
		SkillTemplate:   skillTemplate,
		CommandName:     "ampel-create-policy",
		SkillName:       "ampel-create-policy",
	}) {
		return
	}

	// Default behavior: serve the plugin
	ampelPlugin := server.New()
	plugin.Serve(ampelPlugin)
}
