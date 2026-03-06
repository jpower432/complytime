// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"fmt"
	"os"
	"path/filepath"
)

// installCursor installs Cursor agent integration files.
func installCursor(workspaceRoot string, templates InitTemplates) error {
	cursorDir := filepath.Join(workspaceRoot, ".cursor")
	commandsDir := filepath.Join(cursorDir, "commands")
	skillsDir := filepath.Join(cursorDir, "skills")
	skillDir := filepath.Join(skillsDir, templates.SkillName)

	// Create directories
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		return fmt.Errorf("failed to create commands directory: %w", err)
	}
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}

	// Create command file
	commandPath := filepath.Join(commandsDir, templates.CommandName+".md")
	if err := os.WriteFile(commandPath, []byte(templates.CommandTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write command file: %w", err)
	}

	// Create skill file
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(templates.SkillTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write skill file: %w", err)
	}

	fmt.Fprintf(os.Stdout, "✓ Installed Cursor command: %s\n", commandPath)
	fmt.Fprintf(os.Stdout, "✓ Installed Cursor skill: %s\n", skillPath)
	fmt.Fprintf(os.Stdout, "  Command: /%s\n", templates.CommandName)
	return nil
}

// installOpenCode installs OpenCode agent integration files.
func installOpenCode(workspaceRoot string, templates InitTemplates) error {
	// OpenCode uses .cursorrules file in workspace root
	cursorRulesPath := filepath.Join(workspaceRoot, ".cursorrules")

	// Append to existing .cursorrules or create new
	var content []byte
	if existing, err := os.ReadFile(cursorRulesPath); err == nil {
		content = existing
		// Add separator if file doesn't end with newline
		if len(content) > 0 && content[len(content)-1] != '\n' {
			content = append(content, '\n')
		}
		content = append(content, '\n')
	}

	// Append skill template
	content = append(content, []byte(templates.SkillTemplate)...)
	content = append(content, '\n')

	if err := os.WriteFile(cursorRulesPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write .cursorrules file: %w", err)
	}

	fmt.Fprintf(os.Stdout, "✓ Installed OpenCode integration: %s\n", cursorRulesPath)
	return nil
}

// installClaudeCode installs Claude Code agent integration files.
func installClaudeCode(workspaceRoot string, templates InitTemplates) error {
	claudeDir := filepath.Join(workspaceRoot, ".claude")

	// Create .claude directory if it doesn't exist
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	// Claude Code uses JSON tool definition
	// For now, we'll create a markdown file - plugins can extend this
	toolPath := filepath.Join(claudeDir, templates.CommandName+"-tool.md")
	if err := os.WriteFile(toolPath, []byte(templates.SkillTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write Claude Code tool file: %w", err)
	}

	fmt.Fprintf(os.Stdout, "✓ Installed Claude Code tool: %s\n", toolPath)
	fmt.Fprintf(os.Stdout, "  Note: Claude Code may require JSON format. Check Claude Code documentation.\n")
	return nil
}
