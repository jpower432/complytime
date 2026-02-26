// SPDX-License-Identifier: Apache-2.0

package complytime_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/complytime/complyctl/internal/complytime"
)

func validManifest() *complytime.PackManifest {
	return &complytime.PackManifest{
		ID:      "fedora-compliance",
		Version: "1.0.0",
		Policies: []complytime.PackPolicyEntry{
			{
				URL:     "registry.complytime.dev/policies/cis-fedora-l1-server@v1.0.0",
				ID:      "cis-fedora-l1-server",
				Profile: "cis_server_l1",
			},
		},
		Providers: []complytime.PackProviderEntry{
			{ID: "openscap", Binary: "complyctl-provider-openscap"},
		},
	}
}

func TestValidatePackManifest_Valid(t *testing.T) {
	err := complytime.ValidatePackManifest(validManifest())
	require.NoError(t, err)
}

func TestValidatePackManifest_MissingID(t *testing.T) {
	m := validManifest()
	m.ID = ""
	err := complytime.ValidatePackManifest(m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

func TestValidatePackManifest_MissingVersion(t *testing.T) {
	m := validManifest()
	m.Version = ""
	err := complytime.ValidatePackManifest(m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version is required")
}

func TestValidatePackManifest_MissingPolicyURL(t *testing.T) {
	m := validManifest()
	m.Policies = []complytime.PackPolicyEntry{{URL: "", ID: "test"}}
	err := complytime.ValidatePackManifest(m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "policies[].url cannot be empty")
}

func TestValidatePackManifest_DuplicatePolicyURL(t *testing.T) {
	m := validManifest()
	m.Policies = append(m.Policies, complytime.PackPolicyEntry{
		URL: "registry.complytime.dev/policies/cis-fedora-l1-server@v1.0.0",
		ID:  "different-id",
	})
	err := complytime.ValidatePackManifest(m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate policy url")
}

func TestValidatePackManifest_NoPolicies(t *testing.T) {
	m := validManifest()
	m.Policies = nil
	err := complytime.ValidatePackManifest(m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one policy")
}

func TestValidatePackManifest_NoProviders(t *testing.T) {
	m := validManifest()
	m.Providers = nil
	err := complytime.ValidatePackManifest(m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one provider")
}

func TestValidatePackManifest_DuplicatePolicy(t *testing.T) {
	m := validManifest()
	m.Policies = append(m.Policies, complytime.PackPolicyEntry{
		URL: "other-registry.io/policies/cis-fedora-l1-server@v2.0.0",
		ID:  "cis-fedora-l1-server",
	})
	err := complytime.ValidatePackManifest(m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate policy")
}

func TestValidatePackManifest_DuplicateProvider(t *testing.T) {
	m := validManifest()
	m.Providers = append(m.Providers, complytime.PackProviderEntry{
		ID: "openscap", Binary: "other-binary",
	})
	err := complytime.ValidatePackManifest(m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate provider")
}

func TestValidatePackManifest_EmptyPolicyID(t *testing.T) {
	m := validManifest()
	m.Policies = []complytime.PackPolicyEntry{{URL: "registry.io/p@v1", ID: ""}}
	err := complytime.ValidatePackManifest(m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "policies[].id cannot be empty")
}

func TestValidatePackManifest_EmptyProviderBinary(t *testing.T) {
	m := validManifest()
	m.Providers = []complytime.PackProviderEntry{{ID: "test", Binary: ""}}
	err := complytime.ValidatePackManifest(m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "binary is required")
}

func TestValidatePackManifest_SystemDepMissingKind(t *testing.T) {
	m := validManifest()
	m.SystemDependencies = []complytime.SystemDependency{{Name: "pkg", Kind: "", Value: "pkg"}}
	err := complytime.ValidatePackManifest(m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kind is required")
}

func TestValidatePackManifest_SystemDepInvalidKind(t *testing.T) {
	m := validManifest()
	m.SystemDependencies = []complytime.SystemDependency{{Name: "pkg", Kind: "shell", Value: "pkg"}}
	err := complytime.ValidatePackManifest(m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is not valid")
}

func TestValidatePackManifest_SystemDepMissingValue(t *testing.T) {
	m := validManifest()
	m.SystemDependencies = []complytime.SystemDependency{{Name: "pkg", Kind: complytime.CheckRPM, Value: ""}}
	err := complytime.ValidatePackManifest(m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "value is required")
}

func TestValidatePackManifest_SystemDepValidKinds(t *testing.T) {
	kinds := []complytime.DependencyCheckKind{
		complytime.CheckBinary, complytime.CheckRPM,
		complytime.CheckDEB, complytime.CheckPath,
	}
	for _, kind := range kinds {
		m := validManifest()
		m.SystemDependencies = []complytime.SystemDependency{
			{Name: "dep", Kind: kind, Value: "dep-value"},
		}
		err := complytime.ValidatePackManifest(m)
		require.NoError(t, err, "kind %q should be valid", kind)
	}
}

func TestValidatePackManifest_PlatformWithoutOS(t *testing.T) {
	m := validManifest()
	m.Platform = &complytime.PlatformConfig{OS: ""}
	err := complytime.ValidatePackManifest(m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "platform.os is required")
}

func TestValidatePackManifest_WithPlatform(t *testing.T) {
	m := validManifest()
	m.Platform = &complytime.PlatformConfig{
		OS:         "fedora",
		Datastream: "/usr/share/xml/scap/ssg/content/ssg-fedora-ds.xml",
	}
	err := complytime.ValidatePackManifest(m)
	require.NoError(t, err)
}

func TestLoadPackManifest_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "complypack.yaml")

	yaml := `id: test-pack
version: 0.1.0
policies:
  - url: registry.example.com/policies/test-policy@v1.0.0
    id: test-policy
  - url: other-registry.io/community/extra-policy@v2.0.0
    id: extra-policy
providers:
  - id: test-provider
    binary: complyctl-provider-test
system-dependencies:
  - name: test-dep
    kind: binary
    value: test-dep
    install: sudo dnf install -y test-dep
`
	require.NoError(t, os.WriteFile(path, []byte(yaml), 0600))

	m, err := complytime.LoadPackManifest(path)
	require.NoError(t, err)
	assert.Equal(t, "test-pack", m.ID)
	assert.Equal(t, "0.1.0", m.Version)
	assert.Len(t, m.Policies, 2)
	assert.Equal(t, "registry.example.com/policies/test-policy@v1.0.0", m.Policies[0].URL)
	assert.Equal(t, "test-policy", m.Policies[0].ID)
	assert.Equal(t, "other-registry.io/community/extra-policy@v2.0.0", m.Policies[1].URL)
	assert.Equal(t, "extra-policy", m.Policies[1].ID)
	assert.Len(t, m.Providers, 1)
	assert.Equal(t, "complyctl-provider-test", m.Providers[0].Binary)
	assert.Len(t, m.SystemDependencies, 1)
	assert.Equal(t, complytime.DependencyCheckKind("binary"), m.SystemDependencies[0].Kind)
	assert.Equal(t, "test-dep", m.SystemDependencies[0].Value)
}

func TestLoadPackManifest_NotFound(t *testing.T) {
	_, err := complytime.LoadPackManifest("/nonexistent/complypack.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pack manifest not found")
}

func TestLoadPackManifest_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "complypack.yaml")
	require.NoError(t, os.WriteFile(path, []byte("{{invalid"), 0600))

	_, err := complytime.LoadPackManifest(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid YAML")
}

func TestPackPolicyIDs(t *testing.T) {
	m := validManifest()
	m.Policies = append(m.Policies, complytime.PackPolicyEntry{
		URL: "other-registry.io/policies/second@v1.0.0",
		ID:  "second",
	})

	ids := complytime.PackPolicyIDs(m)
	assert.Equal(t, []string{"cis-fedora-l1-server", "second"}, ids)
}

func TestPackManifestFileConstant(t *testing.T) {
	assert.Equal(t, "complypack.yaml", complytime.PackManifestFile)
}

func TestValidatePackManifest_UnsupportedSchemaVersion(t *testing.T) {
	m := validManifest()
	m.SchemaVersion = 99
	err := complytime.ValidatePackManifest(m)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported schema-version")
}

func TestValidatePackManifest_CurrentSchemaVersion(t *testing.T) {
	m := validManifest()
	m.SchemaVersion = complytime.CurrentPackSchemaVersion
	err := complytime.ValidatePackManifest(m)
	require.NoError(t, err)
}

func TestValidatePackManifest_ZeroSchemaVersionAllowed(t *testing.T) {
	m := validManifest()
	m.SchemaVersion = 0
	err := complytime.ValidatePackManifest(m)
	require.NoError(t, err)
}

func TestToPolicyEntry(t *testing.T) {
	pack := complytime.PackPolicyEntry{
		URL:     "registry.com/policies/nist@v1.0",
		ID:      "nist",
		Profile: "cis_server_l1",
		Catalog: "cis",
		Source:  "upstream",
	}
	entry := pack.ToPolicyEntry()
	assert.Equal(t, "registry.com/policies/nist@v1.0", entry.URL)
	assert.Equal(t, "nist", entry.ID)
}

func TestPackToPolicyEntries(t *testing.T) {
	m := validManifest()
	m.Policies = append(m.Policies, complytime.PackPolicyEntry{
		URL: "other-registry.io/policies/cis@v2.0",
		ID:  "cis",
	})
	entries := complytime.PackToPolicyEntries(m)
	assert.Len(t, entries, 2)
	assert.Equal(t, "cis-fedora-l1-server", entries[0].ID)
	assert.Equal(t, "cis", entries[1].ID)
}
