// SPDX-License-Identifier: Apache-2.0

package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/hashicorp/go-hclog"

	"github.com/complytime/complyctl/cmd/openscap-plugin/config"
	"github.com/complytime/complyctl/cmd/openscap-plugin/oscap"
	"github.com/complytime/complyctl/cmd/openscap-plugin/scan"
	"github.com/complytime/complyctl/cmd/openscap-plugin/xccdf"
	"github.com/complytime/complyctl/internal/complytime"
	"github.com/complytime/complyctl/pkg/plugin"
)

var (
	_ plugin.Plugin = (*PluginServer)(nil)
	// ovalRegex is a regular expression for capturing the check short name
	// in an OVAL check definition identifier.
	ovalRegex = regexp.MustCompile(`^[^:]*?:[^-]*?-(.*?):.*?$`)
)

const ovalCheckType = "http://oval.mitre.org/XMLSchema/oval-definitions-5"

type PluginServer struct {
	Config *config.Config
	// requirementIDs is populated during Generate and consumed during Scan.
	// See R47: providers evaluate all requirements from Generate-time state.
	requirementIDs []string
}

func New() *PluginServer {
	return &PluginServer{
		Config: config.NewConfig(),
	}
}

func (s *PluginServer) HealthCheck(_ context.Context, _ *plugin.HealthCheckRequest) (*plugin.HealthCheckResponse, error) {
	return &plugin.HealthCheckResponse{
		Healthy:                 true,
		Version:                 "0.1.0",
		RequiredTargetVariables: []string{"profile"},
	}, nil
}

func (s *PluginServer) Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	if len(req.Configuration) == 0 {
		return &plugin.GenerateResponse{
			Success:      false,
			ErrorMessage: "no assessment configurations provided",
		}, nil
	}

	evalConfig := mergeVariables(req.GlobalVariables, req.TargetVariables)
	if err := s.Config.LoadSettings(evalConfig); err != nil {
		return &plugin.GenerateResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("configuration error: %v", err),
		}, nil
	}

	hclog.Default().Info("Generating a tailoring file")
	tailoringXML, err := xccdf.PolicyToXML(req.Configuration, s.Config)
	if err != nil {
		return &plugin.GenerateResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("tailoring generation failed: %v", err),
		}, nil
	}

	policyPath := s.Config.Files.Policy
	dst, err := os.Create(policyPath)
	if err != nil {
		return &plugin.GenerateResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to create policy file: %v", err),
		}, nil
	}
	defer dst.Close()
	if _, err := dst.WriteString(tailoringXML); err != nil {
		return &plugin.GenerateResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("failed to write policy file: %v", err),
		}, nil
	}

	hclog.Default().Info("Generating remediation files")
	pluginDir := filepath.Join(complytime.WorkspaceDir, config.PluginDir)
	err = oscap.OscapGenerateFix(
		ctx,
		pluginDir,
		s.Config.Parameters.Profile,
		s.Config.Files.Policy,
		s.Config.Files.Datastream,
	)
	if err != nil {
		return &plugin.GenerateResponse{
			Success:      false,
			ErrorMessage: fmt.Sprintf("remediation generation failed: %v", err),
		}, nil
	}

	s.requirementIDs = make([]string, 0, len(req.Configuration))
	for _, cfg := range req.Configuration {
		s.requirementIDs = append(s.requirementIDs, cfg.RequirementID)
	}

	return &plugin.GenerateResponse{Success: true}, nil
}

func (s *PluginServer) Scan(ctx context.Context, _ *plugin.ScanRequest) (*plugin.ScanResponse, error) {
	if len(s.requirementIDs) == 0 {
		return nil, fmt.Errorf("no requirements loaded — call Generate before Scan")
	}

	policyChecks := newChecks()

	_, err := scan.ScanSystem(ctx, s.Config, s.Config.Parameters.Profile)
	if err != nil {
		return nil, fmt.Errorf("scan failed: %w", err)
	}

	requirementRules := buildRequirementRuleMap(s.requirementIDs)
	policyChecks.LoadRequirements(requirementRules)

	file, err := os.Open(filepath.Clean(s.Config.Files.ARF))
	if err != nil {
		return nil, fmt.Errorf("failed to open ARF: %w", err)
	}
	defer file.Close()

	xmlnode, err := xmlquery.Parse(bufio.NewReader(file))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ARF: %w", err)
	}

	targetEl := xmlnode.SelectElement("//target")
	if targetEl == nil {
		return nil, errors.New("result has no 'target' attribute")
	}
	target := targetEl.InnerText()
	hclog.Default().Debug(fmt.Sprintf("hostname from results target is %s", target))

	ruleTable := xccdf.NewRuleHashTable(xmlnode)
	results := xmlnode.SelectElements("//rule-result")

	var assessments []plugin.AssessmentLog

	for i := range results {
		result := results[i]
		ruleIDRef := result.SelectAttr("idref")

		rule, ok := ruleTable[ruleIDRef]
		if !ok {
			continue
		}

		var ovalRefEl *xmlquery.Node
		for _, check := range rule.SelectElements("//xccdf-1.2:check") {
			if check.SelectAttr("system") == ovalCheckType {
				ovalRefEl = check.SelectElement("xccdf-1.2:check-content-ref")
				break
			}
		}
		if ovalRefEl == nil {
			continue
		}
		ovalCheck, err := parseCheck(ovalRefEl)
		if err != nil {
			return nil, err
		}
		if reqID, found := policyChecks.Match(ovalCheck); found {
			mappedResult, err := mapResultStatus(result)
			if err != nil {
				return nil, err
			}
			resultText := ""
			if el := result.SelectElement("result"); el != nil {
				resultText = el.InnerText()
			}
			assessments = append(assessments, plugin.AssessmentLog{
				RequirementID: reqID,
				Steps: []plugin.Step{
					{
						Name:    ruleIDRef,
						Result:  mappedResult,
						Message: fmt.Sprintf("openscap rule-result is %s", resultText),
					},
				},
				Message:    fmt.Sprintf("Host %s evaluated", target),
				Confidence: plugin.ConfidenceLevelHigh,
			})
		}
	}

	// Emit error assessments for requirements with no matching result.
	covered := make(map[string]bool)
	for _, a := range assessments {
		covered[a.RequirementID] = true
	}
	for _, reqID := range s.requirementIDs {
		if !covered[reqID] {
			assessments = append(assessments, plugin.AssessmentLog{
				RequirementID: reqID,
				Steps: []plugin.Step{{
					Name:    "no-result",
					Result:  plugin.ResultSkipped,
					Message: "no matching rule-result found in ARF",
				}},
				Message:    "skipped — rule not evaluated by OpenSCAP",
				Confidence: plugin.ConfidenceLevelNotSet,
			})
		}
	}

	return &plugin.ScanResponse{Assessments: assessments}, nil
}

// checks is a Set implementation for comparing OSCAL and OVAL check IDs.
// In the new workflow it maps OVAL check short names to requirement IDs.
type checks map[string]string

func newChecks() checks {
	return make(checks)
}

// LoadRequirements populates the set from requirement-to-checks mapping
// built during Scan.
func (c checks) LoadRequirements(ruleMap map[string][]string) {
	for reqID, checkIDs := range ruleMap {
		for _, checkID := range checkIDs {
			c[checkID] = reqID
		}
	}
}

// Match returns the requirement ID if the check is in the set.
func (c checks) Match(check string) (string, bool) {
	reqID, ok := c[check]
	return reqID, ok
}

// Has returns true if the check ID is tracked.
func (c checks) Has(check string) bool {
	_, ok := c[check]
	return ok
}

// mergeVariables combines global and target variable maps into a single
// config map. Target variables override global ones for the same key.
func mergeVariables(global, target map[string]string) map[string]string {
	merged := make(map[string]string, len(global)+len(target))
	for k, v := range global {
		merged[k] = v
	}
	for k, v := range target {
		merged[k] = v
	}
	return merged
}

// buildRequirementRuleMap creates a mapping from requirement IDs to their
// OVAL check short names. In the OpenSCAP model, the requirement ID is
// treated as the XCCDF rule short name, which is also the OVAL check short
// name used in the ARF results.
func buildRequirementRuleMap(requirementIDs []string) map[string][]string {
	m := make(map[string][]string, len(requirementIDs))
	for _, reqID := range requirementIDs {
		m[reqID] = []string{reqID}
	}
	return m
}

func parseCheck(check *xmlquery.Node) (string, error) {
	ovalCheckName := strings.TrimSpace(check.SelectAttr("name"))
	if ovalCheckName == "" {
		return "", errors.New("check-content-ref node has no 'name' attribute")
	}
	matches := ovalRegex.FindStringSubmatch(ovalCheckName)

	minimumPart, shortNameLoc := 2, 1
	if len(matches) < minimumPart {
		return "", fmt.Errorf("check id %q is in unexpected format", ovalCheckName)
	}
	trimmedCheckName := matches[shortNameLoc]
	return trimmedCheckName, nil
}

func mapResultStatus(result *xmlquery.Node) (plugin.Result, error) {
	resultEl := result.SelectElement("result")
	if resultEl == nil {
		return plugin.ResultError, errors.New("result node has no 'result' attribute")
	}
	switch resultEl.InnerText() {
	case "pass", "fixed":
		return plugin.ResultPassed, nil
	case "fail":
		return plugin.ResultFailed, nil
	case "notselected", "notapplicable":
		return plugin.ResultSkipped, nil
	case "error", "unknown":
		return plugin.ResultError, nil
	}

	return plugin.ResultError, fmt.Errorf("couldn't match %s", resultEl.InnerText())
}
