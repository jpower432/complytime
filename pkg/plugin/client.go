// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"context"
	"fmt"

	pluginv2 "github.com/complytime/complyctl/api/plugin"
	goplugin "github.com/hashicorp/go-plugin"
)

var _ Plugin = (*Client)(nil)

// GenerateRequest carries assessment plan configuration to a plugin.
// See R48: three-tier variable model.
type GenerateRequest struct {
	GlobalVariables map[string]string
	Configuration   []AssessmentConfiguration
	TargetVariables map[string]string
}

// AssessmentConfiguration binds a requirement ID to its plan and parameters.
type AssessmentConfiguration struct {
	PlanID        string
	RequirementID string
	Parameters    map[string]string
	// EvaluatorID is used for routing to the correct plugin. It is not
	// serialized over gRPC — routing is handled by the plugin manager.
	EvaluatorID string
}

// GenerateResponse confirms whether policy preparation succeeded.
type GenerateResponse struct {
	Success      bool
	ErrorMessage string
}

// ScanRequest carries targets to evaluate.
// The scanning provider evaluates all requirements from Generate-time state.
// See R47: specs/001-gemara-native-workflow/research.md
type ScanRequest struct {
	Targets []Target
}

// Target identifies a system or environment to scan, with plugin-specific variables.
type Target struct {
	TargetID  string
	Variables map[string]string
}

// ScanResponse carries assessment results from a plugin scan.
type ScanResponse struct {
	Assessments []AssessmentLog
}

// AssessmentLog holds the evaluation result for a single requirement.
type AssessmentLog struct {
	RequirementID string
	Steps         []Step
	Message       string
	Confidence    ConfidenceLevel
}

// Step is one discrete check within an assessment.
type Step struct {
	Name    string
	Result  Result
	Message string
}

// Result is the outcome of a single assessment step.
type Result int32

const (
	ResultUnspecified Result = 0
	ResultPassed      Result = 1
	ResultFailed      Result = 2
	ResultSkipped     Result = 3
	ResultError       Result = 4
)

// ConfidenceLevel indicates the evaluator's confidence in an assessment result.
// Mirrors go-gemara ConfidenceLevel enum values (1:1 mapping).
type ConfidenceLevel int32

const (
	ConfidenceLevelNotSet       ConfidenceLevel = 0
	ConfidenceLevelUndetermined ConfidenceLevel = 1
	ConfidenceLevelLow          ConfidenceLevel = 2
	ConfidenceLevelMedium       ConfidenceLevel = 3
	ConfidenceLevelHigh         ConfidenceLevel = 4
)

// HealthCheckRequest is sent to verify a plugin is alive and compatible.
type HealthCheckRequest struct{}

// HealthCheckResponse reports plugin health, version, any error, and
// required variable names for doctor validation (R51).
type HealthCheckResponse struct {
	Healthy                 bool
	Version                 string
	ErrorMessage            string
	RequiredGlobalVariables []string
	RequiredTargetVariables []string
}

// Client provides gRPC communication with a plugin subprocess managed by
// hashicorp/go-plugin.
type Client struct {
	executablePath string
	pluginClient   *goplugin.Client
	grpcClient     pluginv2.PluginClient
}

func (c *Client) Close() {
	if c.pluginClient != nil {
		c.pluginClient.Kill()
	}
}

func (c *Client) HealthCheck(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error) {
	_ = req

	protoResp, err := c.grpcClient.HealthCheck(ctx, &pluginv2.HealthCheckRequest{})
	if err != nil {
		return nil, fmt.Errorf("HealthCheck RPC failed: %w", err)
	}

	return &HealthCheckResponse{
		Healthy:                 protoResp.GetHealthy(),
		Version:                 protoResp.GetVersion(),
		ErrorMessage:            protoResp.GetErrorMessage(),
		RequiredGlobalVariables: protoResp.GetRequiredGlobalVariables(),
		RequiredTargetVariables: protoResp.GetRequiredTargetVariables(),
	}, nil
}

func (c *Client) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	protoConfigs := make([]*pluginv2.AssessmentConfiguration, 0, len(req.Configuration))
	for _, cfg := range req.Configuration {
		protoConfigs = append(protoConfigs, &pluginv2.AssessmentConfiguration{
			PlanId:        cfg.PlanID,
			RequirementId: cfg.RequirementID,
			Parameters:    cfg.Parameters,
		})
	}

	protoResp, err := c.grpcClient.Generate(ctx, &pluginv2.GenerateRequest{
		GlobalVariables: req.GlobalVariables,
		Configurations:  protoConfigs,
		TargetVariables: req.TargetVariables,
	})
	if err != nil {
		return nil, fmt.Errorf("Generate RPC failed: %w", err)
	}

	return &GenerateResponse{
		Success:      protoResp.GetSuccess(),
		ErrorMessage: protoResp.GetErrorMessage(),
	}, nil
}

func (c *Client) Scan(ctx context.Context, req *ScanRequest) (*ScanResponse, error) {
	protoTargets := make([]*pluginv2.Target, 0, len(req.Targets))
	for _, t := range req.Targets {
		protoTargets = append(protoTargets, &pluginv2.Target{
			TargetId:  t.TargetID,
			Variables: t.Variables,
		})
	}

	protoResp, err := c.grpcClient.Scan(ctx, &pluginv2.ScanRequest{
		Targets: protoTargets,
	})
	if err != nil {
		return nil, fmt.Errorf("Scan RPC failed: %w", err)
	}

	assessments := make([]AssessmentLog, 0, len(protoResp.GetAssessments()))
	for _, pa := range protoResp.GetAssessments() {
		steps := make([]Step, 0, len(pa.GetSteps()))
		for _, ps := range pa.GetSteps() {
			steps = append(steps, Step{
				Name:    ps.GetName(),
				Result:  protoResultToInternal(ps.GetResult()),
				Message: ps.GetMessage(),
			})
		}
		assessments = append(assessments, AssessmentLog{
			RequirementID: pa.GetRequirementId(),
			Steps:         steps,
			Message:       pa.GetMessage(),
			Confidence:    protoConfidenceToInternal(pa.GetConfidence()),
		})
	}

	return &ScanResponse{Assessments: assessments}, nil
}

func protoResultToInternal(r pluginv2.Result) Result {
	switch r {
	case pluginv2.Result_RESULT_PASSED:
		return ResultPassed
	case pluginv2.Result_RESULT_FAILED:
		return ResultFailed
	case pluginv2.Result_RESULT_SKIPPED:
		return ResultSkipped
	case pluginv2.Result_RESULT_ERROR:
		return ResultError
	default:
		return ResultUnspecified
	}
}

func protoConfidenceToInternal(c pluginv2.ConfidenceLevel) ConfidenceLevel {
	switch c {
	case pluginv2.ConfidenceLevel_CONFIDENCE_LEVEL_UNDETERMINED:
		return ConfidenceLevelUndetermined
	case pluginv2.ConfidenceLevel_CONFIDENCE_LEVEL_LOW:
		return ConfidenceLevelLow
	case pluginv2.ConfidenceLevel_CONFIDENCE_LEVEL_MEDIUM:
		return ConfidenceLevelMedium
	case pluginv2.ConfidenceLevel_CONFIDENCE_LEVEL_HIGH:
		return ConfidenceLevelHigh
	default:
		return ConfidenceLevelNotSet
	}
}
