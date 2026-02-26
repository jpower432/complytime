// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"time"

	"github.com/spf13/cobra"

	"github.com/complytime/complyctl/internal/cache"
	"github.com/complytime/complyctl/internal/complytime"
	"github.com/complytime/complyctl/internal/output"
	"github.com/complytime/complyctl/internal/policy"
	"github.com/complytime/complyctl/internal/terminal"
	"github.com/complytime/complyctl/pkg/plugin"
)

type scanOptions struct {
	*Common
	policyID  string
	format    string
	dryRun    bool
	timeout   time.Duration
	cacheDir  string
	pluginDir string
}

func scanCmd(common *Common) *cobra.Command {
	o := &scanOptions{
		Common: common,
	}
	cmd := &cobra.Command{
		Use:   "scan [flags]",
		Short: "Scan targets and produce compliance reports",
		Example: `complyctl scan --policy-id nist-800-53-r5
  complyctl scan --policy-id nist-800-53-r5 --format pretty
  complyctl scan --policy-id nist-800-53-r5 --format oscal
  complyctl scan --policy-id nist-800-53-r5 --format sarif`,
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := o.validate(); err != nil {
				return err
			}
			if err := o.complete(); err != nil {
				return err
			}
			return o.run(cmd.Context())
		},
	}
	cmd.Flags().StringVarP(&o.policyID, "policy-id", "p", "", "Policy ID to scan (required)")
	cmd.Flags().StringVarP(&o.format, "format", "f", "", "Output format: oscal, pretty, sarif")
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Generate artifacts and show execution plan without scanning")
	cmd.Flags().DurationVarP(&o.timeout, "timeout", "t", complytime.DefaultCommandTimeout, "Maximum time for the scan operation (e.g. 5m, 10m, 1h)")
	if err := cmd.MarkFlagRequired("policy-id"); err != nil {
		logger.Error("Failed to mark policy-id as required", "error", err)
	}
	return cmd
}

func (o *scanOptions) validate() error {
	if o.format != "" {
		switch o.format {
		case complytime.OutputFormatOSCAL, complytime.OutputFormatPretty, complytime.OutputFormatSARIF:
		default:
			return fmt.Errorf("invalid format %q: must be one of %s, %s, %s",
				o.format, complytime.OutputFormatOSCAL, complytime.OutputFormatPretty, complytime.OutputFormatSARIF)
		}
	}
	return nil
}

func (o *scanOptions) complete() error {
	var err error
	o.cacheDir, err = complytime.ResolveCacheDir()
	if err != nil {
		return fmt.Errorf("failed to resolve cache directory: %w", err)
	}
	o.pluginDir, err = complytime.ResolvePluginDir()
	if err != nil {
		return fmt.Errorf("failed to resolve plugin directory: %w", err)
	}
	return nil
}

func (o *scanOptions) run(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	ws := complytime.NewWorkspace()
	if err := ws.LoadAndValidate(); err != nil {
		return fmt.Errorf("failed to load complytime: %w", err)
	}

	cfg := ws.Config()

	targets := cfg.Targets
	if len(targets) == 0 {
		logger.Warn("No targets configured in complytime")
		return fmt.Errorf("no targets in complytime config (add targets with policies)")
	}

	cacheMgr := cache.NewCache(o.cacheDir)
	loader := policy.NewLoader(cacheMgr)
	resolver := policy.NewResolver(loader)

	entry, found := complytime.FindPolicy(cfg.Policies, o.policyID)
	if !found {
		return fmt.Errorf("policy %s not found in config", o.policyID)
	}
	ref := complytime.ParsePolicyRef(entry.URL)
	eid := entry.EffectiveID()

	version, err := loader.ResolveVersion(ref.Repository, ref.Version)
	if err != nil {
		return err
	}
	logger.Info("Resolved policy version", "policy", ref.Repository, "version", version)

	graph, err := resolver.ResolvePolicyGraph(ref.Repository, version)
	if err != nil {
		return fmt.Errorf("failed to resolve policy graph: %w", err)
	}

	assessmentConfigs := policy.ExtractAssessmentConfigs(ref.Repository, graph)
	groups := policy.GroupByEvaluator(assessmentConfigs, graph)

	globalVars := cfg.Variables
	if err := policy.ValidateGlobalVars(groups, globalVars, ws.Path()); err != nil {
		return err
	}

	mgr, err := plugin.NewManager(o.pluginDir, logFile)
	if err != nil {
		return fmt.Errorf("plugin manager init failed: %w", err)
	}
	defer mgr.Cleanup()

	if err := mgr.LoadPlugins(); err != nil {
		return fmt.Errorf("plugin discovery failed: %w", err)
	}

	plugins := mgr.ListPlugins()
	if len(plugins) == 0 {
		return fmt.Errorf("no plugins found in %s (HealthCheck may have failed)", o.pluginDir)
	}

	// Freshness check: determine whether generation state needs updating.
	// Generate is always called because the plugin process is ephemeral and
	// needs RouteGenerate to initialize its config (file paths, profile, etc.)
	// before Scan can run. Generate is idempotent — calling it when artifacts
	// are fresh recreates the same output.
	// See R37: specs/001-gemara-native-workflow/research.md
	stateStale := false
	cacheState, err := cache.LoadState(o.cacheDir)
	if err != nil {
		return fmt.Errorf("failed to load cache state: %w", err)
	}
	policyState, _ := cacheState.GetPolicyState(ref.Repository)

	genState, err := policy.LoadGenerationState(".", ref.Repository)
	if err != nil {
		return fmt.Errorf("failed to load generation state: %w", err)
	}

	switch {
	case genState == nil:
		logger.Info("No generation state found, generating", "policy", ref.Repository)
		stateStale = true
	case !genState.IsFresh(policyState.Digest):
		logger.Warn("Policy cache updated since last generate — regenerating",
			"policy", ref.Repository, "cached_digest", policyState.Digest, "gen_digest", genState.PolicyDigest)
		stateStale = true
	default:
		logger.Info("Generation artifacts are fresh", "policy", ref.Repository)
	}

	var evaluatorIDs []string
	for evalID := range groups {
		evaluatorIDs = append(evaluatorIDs, evalID)
	}

	var policyTargets []complytime.TargetConfig
	for _, t := range targets {
		if slices.Contains(t.Policies, eid) {
			policyTargets = append(policyTargets, t)
		}
	}

	genSpin := terminal.NewSpinner("Generating policy artifacts...")
	genSpin.Start()

	for evalID, group := range groups {
		for _, target := range policyTargets {
			if err := mgr.RouteGenerate(ctx, evalID, globalVars, target.Variables, group.Configs); err != nil {
				genSpin.Stop()
				return err
			}
		}
	}

	genSpin.Stop()

	if stateStale {
		newGenState := policy.NewGenerationState(ref.Repository, policyState.Digest, evaluatorIDs)
		if err := policy.SaveGenerationState(".", ref.Repository, newGenState); err != nil {
			return fmt.Errorf("failed to save generation state: %w", err)
		}
	}

	// --dry-run: persist GenerationState and output execution plan, then exit.
	// "Dry" means "don't scan," not "don't generate."
	// See FR-033: specs/001-gemara-native-workflow/spec.md
	if o.dryRun {
		var routes []output.EvaluatorRoute
		for evalID, group := range groups {
			route := output.EvaluatorRoute{
				EvaluatorID:      evalID,
				RequirementCount: len(group.Configs),
				Status:           "healthy",
			}
			if lp, lookupErr := mgr.GetPlugin(evalID); lookupErr == nil {
				route.PluginPath = lp.Info.ExecutablePath
			} else {
				route.Status = "ERROR"
			}
			routes = append(routes, route)
		}

		var scopes []output.TargetScope
		for _, t := range targets {
			if slices.Contains(t.Policies, eid) {
				scopes = append(scopes, output.TargetScope{
					TargetID:     t.ID,
					PolicyID:     ref.Repository,
					EvaluatorIDs: evaluatorIDs,
				})
			}
		}

		fmt.Print(output.FormatExecutionPlan(ref.Repository, routes, scopes))
		return nil
	}

	// Pre-scan summary (FR-034)
	var targetIDs []string
	for _, t := range targets {
		if slices.Contains(t.Policies, eid) {
			targetIDs = append(targetIDs, t.ID)
		}
	}
	fmt.Println(output.FormatPreScanSummary(len(assessmentConfigs), evaluatorIDs, targetIDs))

	reqToControl := extractReqToControlMap(graph)
	eval := output.NewEvaluator(ref.Repository, reqToControl)
	outDir := filepath.Join(".", complytime.WorkspaceDir, complytime.ScanOutputDir)

	scanSpin := terminal.NewSpinner("Scanning targets...")
	scanSpin.Start()

	var allAssessments []plugin.AssessmentLog

	for _, target := range targets {
		if !slices.Contains(target.Policies, eid) {
			continue
		}

		pluginTargets := []plugin.Target{{
			TargetID:  target.ID,
			Variables: target.Variables,
		}}

		var assessments []plugin.AssessmentLog

		for evalID := range groups {
			results, routeErr := mgr.RouteScan(ctx, evalID, pluginTargets)
			if routeErr != nil {
				scanSpin.Stop()
				return routeErr
			}
			assessments = append(assessments, results...)
		}

		if len(groups) == 0 {
			results, routeErr := mgr.RouteScan(ctx, "", pluginTargets)
			if routeErr != nil {
				scanSpin.Stop()
				return routeErr
			}
			assessments = append(assessments, results...)
		}

		eval.AddTarget(assessments)
		allAssessments = append(allAssessments, assessments...)
	}

	scanSpin.Stop()

	logPath, err := eval.Write(outDir)
	if err != nil {
		return fmt.Errorf("failed to write evaluation log: %w", err)
	}
	fmt.Printf("Evaluation log written: %s\n", logPath)

	fmt.Println(output.FormatScanSummary(allAssessments))

	gemaraLog := eval.GemaraLog()

	switch o.format {
	case complytime.OutputFormatPretty:
		md := output.NewMarkdown(ref.Repository, gemaraLog)
		md.SetEmbedEvaluationLog(logPath)
		mdPath, err := md.Write(outDir)
		if err != nil {
			return fmt.Errorf("failed to write markdown report: %w", err)
		}
		fmt.Printf("Markdown report written: %s\n", mdPath)
	case complytime.OutputFormatSARIF:
		sarifPath, err := output.ToSARIF(gemaraLog, "file:///scan", outDir)
		if err != nil {
			return fmt.Errorf("failed to export SARIF: %w", err)
		}
		fmt.Printf("SARIF report written: %s\n", sarifPath)
	case complytime.OutputFormatOSCAL:
		oscalPath, err := output.ToOSCAL(gemaraLog, outDir)
		if err != nil {
			return fmt.Errorf("failed to export OSCAL: %w", err)
		}
		fmt.Printf("OSCAL report written: %s\n", oscalPath)
	}

	fmt.Println("\nScan completed.")
	return nil
}

// extractReqToControlMap builds a requirement-ID → control-ID mapping
// from the parsed control catalogs in the dependency graph.
func extractReqToControlMap(graph *policy.DependencyGraph) map[string]string {
	m := make(map[string]string)
	for _, ctrl := range graph.Controls {
		if ctrl.Parsed == nil {
			continue
		}
		for _, c := range ctrl.Parsed.Controls {
			for _, ar := range c.AssessmentRequirements {
				m[ar.Id] = c.Id
			}
		}
	}
	return m
}
