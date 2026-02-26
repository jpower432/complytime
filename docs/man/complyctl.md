% COMPLYCTL(1) Complyctl Manual
% Marcus Burghardt <maburgha@redhat.com>
% February 2026

# NAME

complyctl - Complyctl CLI performs compliance assessment activities using Gemara policies and evaluator plugins.

# SYNOPSIS

**complyctl** [command] [flags]

# DESCRIPTION

Complyctl CLI leverages Gemara policies from OCI registries to perform compliance assessment activities, using evaluator plugins for each stage of the lifecycle.

Complyctl can be extended to support desired policy engines (PVPs) by the use of plugins. The plugin acts as the integration between complyctl and the PVPs native interface. Each plugin is responsible for evaluating requirements from Gemara assessment plans and returning structured results.

Plugins communicate with complyctl via gRPC and can be authored using any preferred language. The plugin acts as the gRPC server while the complyctl CLI acts as the client. When a complyctl command is run, it invokes the appropriate method served by the plugin.

All commands operate on the **complytime.yaml** workspace configuration file in the current directory. In comply-pack environments, a **complypack.yaml** pack manifest coexists alongside the runtime config to declare bundled policies, providers, and system dependencies.

See more about authoring plugins at https://github.com/complytime/complyctl/blob/main/docs/PLUGIN_GUIDE.md.

Review the `openscap-plugin` that is shipped with complyctl at https://github.com/complytime/complyctl/tree/main/cmd/openscap-plugin/README.md.

Also check the "SEE ALSO" section for plugin specific man pages.

# COMMANDS

**init**
Initialize a Gemara-native workspace configuration file (complytime.yaml) by prompting for registry URL, target IDs, and policy ID mappings. Automatically runs **get** after creation.

**get**
Fetch or update Gemara policies from an OCI registry to the local cache (~/.complytime/policies/). Performs incremental sync with atomic updates and version resolution.

**generate --policy-id** *ID*
Resolve the dependency graph for a cached Gemara policy and invoke evaluator plugins to generate policy artifacts. Uses go-gemara parsed types for control catalogs and guidance documents.

**scan --policy-id** *ID* [**--format** *FORMAT*]
Execute compliance checks against a Gemara policy using gRPC evaluator plugins. Supported formats: **oscal** (OSCAL assessment-results JSON), **pretty** (Markdown report with embedded EvaluationLog), **sarif** (SARIF v2.1.0 JSON). Produces an EvaluationLog and formatted output.

**list** [**--plain**] [**--policy-id** *ID*]
List cached and remote Gemara policies from the workspace configuration. Remote-only policies are shown with a **(remote)** marker.

**version**
Display the complyctl version, build date, and git commit.

# OPTIONS

**-d**, **--debug**
Output debug logs.

**-h**, **--help**
Show help for complyctl.

Run **complyctl [command] --help** for more information about a specific command.

# EXAMPLES

## Gemara-Native Workflow

Initialize a workspace, fetch policies, run a scan, and view results:

```bash
$ complyctl init
# Prompts for registry URL, targets, and policy IDs → creates complytime.yaml

$ complyctl get
# Fetches policies from OCI registry to ~/.complytime/policies/

$ complyctl generate --policy-id nist-800-53-r5
# Resolves dependency graph and invokes evaluator plugins

$ complyctl scan --policy-id nist-800-53-r5 --format oscal
# Produces assessment-results-*.json in OSCAL format

$ complyctl scan --policy-id nist-800-53-r5 --format pretty
# Produces Markdown report with embedded EvaluationLog

$ complyctl scan --policy-id nist-800-53-r5 --format sarif
# Produces scan-*.sarif.json in SARIF v2.1.0 format

$ complyctl list --plain
# Lists cached and remote policies in plain table format
```

# SEE ALSO

complyctl-openscap-plugin(7)

See the Upstream project at https://github.com/complytime/complyctl for more detailed documentation.

# COPYRIGHT

© 2025 Red Hat, Inc. Complyctl is released under the terms of the Apache-2.0 license.
