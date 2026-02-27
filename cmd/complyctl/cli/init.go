// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/complytime/complyctl/internal/complytime"
)

type initOptions struct {
	*Common
}

func initCmd(common *Common) *cobra.Command {
	o := &initOptions{
		Common: common,
	}
	cmd := &cobra.Command{
		Use:               "init",
		Short:             "Create a workspace configuration file",
		SilenceUsage:      true,
		Example:           "complyctl init",
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
		RunE: func(_ *cobra.Command, _ []string) error {
			return o.run()
		},
	}
	return cmd
}

// See FR-003, R54: init is config-creation-only.
// Errors if complytime.yaml already exists (like go mod init).
// User runs `complyctl get` and `complyctl doctor` separately.
func (o *initOptions) run() error {
	workspace := complytime.NewWorkspace()

	if workspace.Exists() {
		return fmt.Errorf("%s already exists", workspace.Path())
	}

	if err := workspace.EnsureDir(); err != nil {
		return fmt.Errorf("failed to create .complytime directory: %w", err)
	}

	policies := promptPolicies()
	targets := promptTargets(policies)

	wsConfig := &complytime.WorkspaceConfig{
		Policies: policies,
		Targets:  targets,
	}

	workspace.SetConfig(wsConfig)
	if err := workspace.Save(); err != nil {
		logger.Error("Workspace creation failed", "error", err)
		return fmt.Errorf("failed to create workspace configuration: %w", err)
	}

	fmt.Fprintln(os.Stderr, "Workspace configuration created.")
	logger.Info("Workspace configuration created", "path", workspace.Path(),
		"policies", len(policies))

	return nil
}

func promptPolicies() []complytime.PolicyEntry {
	fmt.Println("Add policies (one per line). For each, enter the OCI URL and an optional short ID.")
	fmt.Println("Press Enter with an empty URL to finish.")

	var policies []complytime.PolicyEntry
	for i := 1; ; i++ {
		fmt.Printf("  Policy %d URL (e.g. registry.com/policies/nist-800-53-r5@v1.0): ", i)
		var url string
		if _, err := fmt.Scanln(&url); err != nil || strings.TrimSpace(url) == "" {
			break
		}
		url = strings.TrimSpace(url)

		fmt.Printf("  Policy %d ID (short name, or Enter to auto-derive): ", i)
		var id string
		if _, err := fmt.Scanln(&id); err != nil {
			id = ""
		}
		id = strings.TrimSpace(id)

		entry := complytime.PolicyEntry{URL: url, ID: id}
		fmt.Printf("  → %s (id: %s)\n", url, entry.EffectiveID())
		policies = append(policies, entry)
	}

	if len(policies) == 0 {
		return []complytime.PolicyEntry{
			{URL: "registry.example.com/example-policy@latest"},
		}
	}
	return policies
}

func promptTargets(policies []complytime.PolicyEntry) []complytime.TargetConfig {
	fmt.Print("Enter target IDs (comma-separated, or empty to skip): ")
	var input string
	if _, err := fmt.Scanln(&input); err != nil || input == "" {
		return []complytime.TargetConfig{}
	}

	var effectiveIDs []string
	for _, p := range policies {
		effectiveIDs = append(effectiveIDs, p.EffectiveID())
	}

	var targets []complytime.TargetConfig
	for _, targetID := range strings.Split(input, ",") {
		targetID = strings.TrimSpace(targetID)
		if targetID == "" {
			continue
		}

		fmt.Printf("Enter policy IDs for target '%s' (comma-separated, or empty for all): ", targetID)
		fmt.Printf("[available: %s] ", strings.Join(effectiveIDs, ", "))
		var policyInput string
		if _, err := fmt.Scanln(&policyInput); err != nil {
			policyInput = ""
		}

		var targetPolicies []string
		if policyInput == "" {
			targetPolicies = append(targetPolicies, effectiveIDs...)
		} else {
			for _, p := range strings.Split(policyInput, ",") {
				p = strings.TrimSpace(p)
				if p != "" {
					targetPolicies = append(targetPolicies, p)
				}
			}
		}

		targets = append(targets, complytime.TargetConfig{
			ID:       targetID,
			Policies: targetPolicies,
		})
	}

	return targets
}
