// SPDX-License-Identifier: Apache-2.0

package registry_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/complytime/complyctl/internal/registry"
)

func TestClient_GetDefinitions_WithMockFetcher(t *testing.T) {
	mock := registry.NewMockFetcher()
	mock.SeedTestPolicy("test-policy")

	client := registry.NewClientWithFetcher("mock-registry", nil, mock)

	data, err := client.GetDefinitions(context.Background(), "test-policy", "v1.0.0")
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestClient_GetDefinitions_EmptyPath(t *testing.T) {
	client := registry.NewClientWithFetcher("mock-registry", nil, registry.NewMockFetcher())

	_, err := client.GetDefinitions(context.Background(), "", "v1.0.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "module path cannot be empty")
}

func TestClient_GetDefinitions_EmptyVersion(t *testing.T) {
	client := registry.NewClientWithFetcher("mock-registry", nil, registry.NewMockFetcher())

	_, err := client.GetDefinitions(context.Background(), "test-policy", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version cannot be empty")
}

func TestClient_DefinitionVersion_WithMockFetcher(t *testing.T) {
	mock := registry.NewMockFetcher()
	mock.SeedTestPolicy("test-policy")

	client := registry.NewClientWithFetcher("mock-registry", nil, mock)

	digest, version, err := client.DefinitionVersion(context.Background(), "test-policy")
	require.NoError(t, err)
	assert.NotEmpty(t, digest)
	assert.NotEmpty(t, version)
}

func TestClient_DefinitionVersion_EmptyPath(t *testing.T) {
	client := registry.NewClientWithFetcher("mock-registry", nil, registry.NewMockFetcher())

	_, _, err := client.DefinitionVersion(context.Background(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "module path cannot be empty")
}

func TestClient_DefinitionVersion_NotFound(t *testing.T) {
	mock := registry.NewMockFetcher()
	client := registry.NewClientWithFetcher("mock-registry", nil, mock)

	_, _, err := client.DefinitionVersion(context.Background(), "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestClient_GetDefinitions_PolicyNotSeeded(t *testing.T) {
	mock := registry.NewMockFetcher()
	client := registry.NewClientWithFetcher("mock-registry", nil, mock)

	_, err := client.GetDefinitions(context.Background(), "missing-policy", "v1.0.0")
	require.Error(t, err)
}
