# Migration from gemara2ampel to init --tool

This document describes how to migrate from using the manual `gemara2ampel` script to using AI-assisted policy creation with the `init --tool` command.

## Overview

Previously, creating AMPEL policies from Gemara artifacts required:
1. Running the `gemara2ampel` script manually
2. Converting Gemara artifacts to AMPEL format
3. Manually editing and validating policies

With the new `init --tool` command, you can:
1. Generate AI agent skill definitions for your preferred tool (OpenCode, Cursor, Claude Code)
2. Use AI assistants to help create AMPEL policies interactively
3. Validate policies against the AMPEL schema automatically

## Migration Steps

### Step 1: Install Skill Definitions

Choose your AI coding assistant and install the appropriate skill definition:

```bash
# For OpenCode (installs to .cursorrules)
complyctl-provider-ampel init --tool opencode

# For Cursor (installs to .cursor/commands/ampel-create-policy.md and .cursor/skills/ampel-create-policy/SKILL.md)
complyctl-provider-ampel init --tool cursor

# For Claude Code (installs to .claude/ampel-policy-tool.json)
complyctl-provider-ampel init --tool claude-code
```

The `init` command automatically creates the necessary directories and installs the skill definitions in the correct locations.

### Step 3: Use AI Assistant to Create Policies

Instead of running `gemara2ampel` manually, use your AI assistant:

- **OpenCode**: Ask it to create an AMPEL policy from a Gemara requirement
- **Cursor**: Use the `/ampel-create-policy` slash command
- **Claude Code**: Invoke the `create_ampel_policy` tool

The AI assistant will:
- Understand the AMPEL policy format
- Convert Gemara requirements to AMPEL policies
- Validate against the schema
- Provide guidance on policy structure

### Step 4: Validate Policies

Policies created with AI assistance should follow the AMPEL schema. The skill definitions include the complete JSON Schema for validation.

## Benefits

- **Interactive**: Get help from AI as you create policies
- **Validated**: Schema validation ensures correct structure
- **Documented**: Skill definitions include conversion guidelines
- **Consistent**: AI follows the same patterns every time

## gemara2ampel Script Status

The `gemara2ampel` script remains available in the [complytime-demos](https://github.com/complytime/complytime-demos) repository for batch conversion workflows. For interactive policy creation, prefer using `init --tool` with your AI coding assistant.
