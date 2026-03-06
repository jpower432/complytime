# Plugin Command Template

This template shows how to implement the default serve behavior and optional `init` command in a complyctl plugin.

## Basic Structure

```go
package main

import (
	_ "embed"

	"github.com/complytime/complyctl/pkg/plugin"
	// Import your plugin implementation
)

//go:embed templates/command.md
var commandTemplate string

//go:embed templates/skill.md
var skillTemplate string

func main() {
	// Register init command handler (optional)
	if plugin.RegisterInit(plugin.InitTemplates{
		CommandTemplate: commandTemplate,
		SkillTemplate:   skillTemplate,
		CommandName:     "my-command",
		SkillName:       "my-skill",  // Optional: defaults to CommandName
	}) {
		return  // Command handled, exit
	}

	// Default behavior: serve the plugin
	myPlugin := NewMyPlugin()
	plugin.Serve(myPlugin)
}
```

## Default Behavior (Serve)

The plugin serves by default, starting the gRPC server that handles runtime policy execution:

```go
// Default behavior: serve the plugin
myPlugin := NewMyPlugin()
plugin.Serve(myPlugin)
```

This is called automatically by complyctl when the plugin is discovered and loaded, or when the plugin is run without arguments.

## Init Command

The `init` command generates AI agent integration artifacts. It is handled automatically by `plugin.RegisterInit()`.

### Template Files

Create two template files in your plugin:

1. **`templates/command.md`**: Command definition template (Cursor format)
   - Contains the command metadata and instructions
   - See `cmd/ampel-plugin/templates/command.md` for an example

2. **`templates/skill.md`**: Skill definition template
   - Contains the skill documentation and guidelines
   - See `cmd/ampel-plugin/templates/skill.md` for an example

### Usage

Users run:
```bash
complyctl-provider-myplugin init --tool cursor
```

The library handles:
- Parsing the `--tool` flag
- Finding the workspace root
- Creating necessary directories
- Writing files to agent-specific locations
- Providing user feedback

### Supported Agents

- **cursor**: Installs to `.cursor/commands/` and `.cursor/skills/`
- **opencode**: Appends to `.cursorrules`
- **claude-code**: Writes to `.claude/`

## Reference Implementation

See `cmd/ampel-plugin/main.go` and `cmd/ampel-plugin/templates/` for a complete working example.
