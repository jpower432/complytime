// SPDX-License-Identifier: Apache-2.0

package complytime

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
)

var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// WorkspaceConfig is the top-level YAML configuration for a complytime complytime.
// See R48, R49: three-tier variable model.
type WorkspaceConfig struct {
	Registry   RegistryConfig      `yaml:"registry"`
	Policies   []PolicyConfig      `yaml:"policies"`
	Targets    []TargetConfig      `yaml:"targets"`
	Parameters []ParameterOverride `yaml:"parameters,omitempty"`
	Variables  map[string]string   `yaml:"variables,omitempty"`
}

// ParameterOverride allows local parameter selection from policy-defined accepted values.
type ParameterOverride struct {
	PolicyID    string `yaml:"policy_id"`
	ParameterID string `yaml:"parameter_id"`
	Value       string `yaml:"value"`
}

// RegistryConfig holds OCI registry connection settings.
type RegistryConfig struct {
	URL string `yaml:"url"`
}

// PolicyConfig identifies a policy and an optional pinned version.
type PolicyConfig struct {
	ID      string `yaml:"id"`
	Version string `yaml:"version,omitempty"`
}

// TargetConfig binds a scan target to one or more policies with optional variables.
type TargetConfig struct {
	ID        string            `yaml:"id"`
	PolicyIDs []string          `yaml:"policy_ids"`
	Variables map[string]string `yaml:"variables,omitempty"`
}

// Load reads, parses, and validates the complytime configuration from the
// default complytime.yaml path.
func Load() (*WorkspaceConfig, error) {
	return LoadFrom(WorkspaceConfigFile)
}

// LoadFrom reads, parses, and validates the complytime configuration from the
// given path.
func LoadFrom(configPath string) (*WorkspaceConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf(
				"complytime config not found: %s (run 'complyctl init' to create)",
				configPath,
			)
		}
		return nil, fmt.Errorf("failed to read complytime file %s: %w", configPath, err)
	}

	var config WorkspaceConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf(
			"corrupted complytime file %s: invalid YAML: %w",
			configPath, err,
		)
	}

	if err := resolveEnvVars(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// resolveEnvVars expands ${VAR} references in target variable values from the
// process environment. Returns an error if a referenced variable is not set.
func resolveEnvVars(config *WorkspaceConfig) error {
	for i, target := range config.Targets {
		for key, val := range target.Variables {
			resolved, err := expandEnvRef(val)
			if err != nil {
				return fmt.Errorf("targets[%s].variables.%s: %w", target.ID, key, err)
			}
			config.Targets[i].Variables[key] = resolved
		}
	}
	return nil
}

// expandEnvRef replaces all ${VAR} occurrences in s with their environment
// values. Returns an error if any referenced variable is unset.
func expandEnvRef(s string) (string, error) {
	var missing []string
	result := envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		varName := envVarPattern.FindStringSubmatch(match)[1]
		if val, ok := os.LookupEnv(varName); ok {
			return val
		}
		missing = append(missing, varName)
		return match
	})
	if len(missing) > 0 {
		return "", fmt.Errorf("unset environment variable(s): %s", strings.Join(missing, ", "))
	}
	return result, nil
}

// Save writes complytime configuration to the default complytime.yaml path.
func Save(config *WorkspaceConfig) error {
	return SaveTo(config, WorkspaceConfigFile)
}

// SaveTo writes complytime configuration to the given path.
func SaveTo(config *WorkspaceConfig, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal complytime: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write complytime file %s: %w", configPath, err)
	}

	return nil
}

// Validate checks required fields, uniqueness constraints, and URL validity.
func Validate(config *WorkspaceConfig) error {
	if config.Registry.URL == "" {
		return fmt.Errorf("registry.url is required")
	}

	registryURL := strings.TrimSpace(config.Registry.URL)
	if _, err := url.Parse(registryURL); err != nil {
		return fmt.Errorf("registry.url is invalid: %w", err)
	}

	if len(config.Policies) == 0 {
		return fmt.Errorf("policies: at least one policy is required")
	}

	policyIDs := make(map[string]bool)
	for _, policy := range config.Policies {
		if policy.ID == "" {
			return fmt.Errorf("policies[].id cannot be empty")
		}
		if policyIDs[policy.ID] {
			return fmt.Errorf("policies[].id: duplicate %s", policy.ID)
		}
		policyIDs[policy.ID] = true
	}

	targetIDs := make(map[string]bool)
	for _, target := range config.Targets {
		if target.ID == "" {
			return fmt.Errorf("targets[].id cannot be empty")
		}
		if targetIDs[target.ID] {
			return fmt.Errorf("targets[].id: duplicate %s", target.ID)
		}
		targetIDs[target.ID] = true
		if len(target.PolicyIDs) == 0 {
			return fmt.Errorf("targets[%s].policy_ids: at least one required", target.ID)
		}
	}
	return nil
}

// ParsePolicyID splits "nist-800-53-r5@v1.2.3" into PolicyConfig{ID, Version}.
// If no @ separator is present, Version is empty (resolved at sync time).
func ParsePolicyID(raw string) PolicyConfig {
	raw = strings.TrimSpace(raw)
	if idx := strings.LastIndex(raw, "@"); idx > 0 && idx < len(raw)-1 {
		return PolicyConfig{
			ID:      raw[:idx],
			Version: raw[idx+1:],
		}
	}
	return PolicyConfig{ID: raw}
}

// ValidateTargetPolicyVersions ensures every target references policies that
// exist in the complytime policies list with no duplicates.
func ValidateTargetPolicyVersions(config *WorkspaceConfig) error {
	policySet := make(map[string]string)
	for _, p := range config.Policies {
		policySet[p.ID] = p.Version
	}
	for _, target := range config.Targets {
		seen := make(map[string]bool)
		for _, pid := range target.PolicyIDs {
			if _, exists := policySet[pid]; !exists {
				return fmt.Errorf("target %s: policy %s not in policies list", target.ID, pid)
			}
			if seen[pid] {
				return fmt.Errorf("target %s: duplicate policy ID %s", target.ID, pid)
			}
			seen[pid] = true
		}
	}
	return nil
}
