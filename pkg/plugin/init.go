// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// InitTemplates defines the templates a plugin provides for agent integration.
type InitTemplates struct {
	// CommandTemplate is the markdown template for the command definition.
	// Required for Cursor agent.
	CommandTemplate string

	// SkillTemplate is the markdown template for the skill definition.
	// Required for Cursor agent.
	SkillTemplate string

	// CommandName is the kebab-case name of the command (e.g., "ampel-create-policy").
	// Used to generate file paths and command references.
	CommandName string

	// SkillName is the kebab-case name of the skill directory (e.g., "ampel-create-policy").
	// If empty, defaults to CommandName.
	SkillName string
}

// RegisterInit registers the init command handler for a plugin.
// Plugins should call this from main() before calling Serve().
// If os.Args[1] == "init", it handles the command and exits.
// Otherwise, it returns false and the plugin should continue to Serve().
func RegisterInit(templates InitTemplates) bool {
	if len(os.Args) < 2 || os.Args[1] != "init" {
		return false
	}

	if err := runInit(os.Args[2:], templates); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	return true
}

// runInit executes the init command with the given arguments and templates.
func runInit(args []string, templates InitTemplates) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	tool := fs.String("tool", "", "AI agent tool (cursor, opencode, claude-code)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *tool == "" {
		fmt.Fprintf(os.Stderr, "Error: --tool flag is required\n")
		fmt.Fprintf(os.Stderr, "Usage: init --tool <cursor|opencode|claude-code>\n")
		fmt.Fprintf(os.Stderr, "\nSupported tools:\n")
		fmt.Fprintf(os.Stderr, "  - cursor\n")
		fmt.Fprintf(os.Stderr, "  - opencode\n")
		fmt.Fprintf(os.Stderr, "  - claude-code\n")
		return fmt.Errorf("--tool flag required")
	}

	// Validate templates
	if templates.CommandTemplate == "" {
		return fmt.Errorf("CommandTemplate is required")
	}
	if templates.CommandName == "" {
		return fmt.Errorf("CommandName is required")
	}
	if templates.SkillName == "" {
		templates.SkillName = templates.CommandName
	}

	// Find workspace root
	workspaceRoot, err := findWorkspaceRoot()
	if err != nil {
		return fmt.Errorf("failed to find workspace root: %w", err)
	}

	// Route to agent-specific handler
	switch *tool {
	case "cursor":
		return installCursor(workspaceRoot, templates)
	case "opencode":
		return installOpenCode(workspaceRoot, templates)
	case "claude-code":
		return installClaudeCode(workspaceRoot, templates)
	default:
		return fmt.Errorf("unsupported tool: %s", *tool)
	}
}

// findWorkspaceRoot finds the workspace root directory.
// It looks for .git directory or uses current directory.
func findWorkspaceRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up to find .git directory
	for {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root, use current directory
			return dir, nil
		}
		dir = parent
	}
}
