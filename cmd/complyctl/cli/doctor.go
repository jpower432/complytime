// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/complytime/complyctl/internal/cache"
	"github.com/complytime/complyctl/internal/complytime"
	"github.com/complytime/complyctl/internal/doctor"
	"github.com/complytime/complyctl/internal/policy"
	"github.com/complytime/complyctl/internal/registry"
)

func doctorCmd(common *Common) *cobra.Command {
	_ = common
	var verbose bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run pre-flight diagnostics on the workspace",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runDoctor(verbose)
		},
	}
	cmd.Flags().BoolVar(&verbose, "verbose", false, "expand per-provider variable detail")
	return cmd
}

// registryVersionResolver adapts registry.Client to doctor.VersionResolver.
// See R55: specs/001-gemara-native-workflow/spec.md
type registryVersionResolver struct {
	timeout time.Duration
}

func (r *registryVersionResolver) ResolveLatestVersion(registryURL, repository string) (string, error) {
	credFunc, err := registry.NewCredentialFunc()
	if err != nil {
		credFunc = nil
	}
	client := registry.NewClient(registryURL, credFunc)
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	_, version, err := client.DefinitionVersion(ctx, repository)
	if err != nil {
		return "", err
	}
	return version, nil
}

// See FR-039, R44, R51, R52, R55: specs/001-gemara-native-workflow/spec.md
func runDoctor(verbose bool) error {
	pluginDir, err := complytime.ResolvePluginDir()
	if err != nil {
		return fmt.Errorf("failed to resolve plugin directory: %w", err)
	}

	cacheBaseDir, err := complytime.ResolveCacheDir()
	if err != nil {
		return fmt.Errorf("failed to resolve cache directory: %w", err)
	}
	policiesCacheDir := filepath.Join(cacheBaseDir, complytime.PoliciesSubdir)

	configPath := complytime.WorkspaceConfigFile
	var cfg *complytime.WorkspaceConfig

	loaded, loadErr := complytime.LoadFrom(configPath)
	if loadErr == nil {
		cfg = loaded
	}

	var resolver doctor.PolicyGraphResolver
	cacheMgr := cache.NewCache(policiesCacheDir)
	loader := policy.NewLoader(cacheMgr)
	resolver = policy.NewResolver(loader)

	versionResolver := &registryVersionResolver{timeout: 5 * time.Second}

	results := doctor.Run(cfg, configPath, pluginDir, cacheBaseDir, policiesCacheDir, resolver, versionResolver, verbose, logFile)

	hasBlockingFailure := false
	for _, r := range results {
		var emoji string
		switch r.Status {
		case doctor.StatusPass:
			emoji = complytime.StatusPassed
		case doctor.StatusFail:
			emoji = complytime.StatusFailed
		case doctor.StatusWarn:
			emoji = complytime.StatusSkipped
		}
		fmt.Printf("%s %s: %s\n", emoji, r.Name, r.Message)
		if r.Blocking && r.Status == doctor.StatusFail {
			hasBlockingFailure = true
		}
	}

	if hasBlockingFailure {
		return fmt.Errorf("one or more blocking checks failed")
	}
	return nil
}
