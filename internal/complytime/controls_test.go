// SPDX-License-Identifier: Apache-2.0

package complytime

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadFrameworks(t *testing.T) {
	tests := []struct {
		name           string
		appDir         func() ApplicationDirectory
		wantFrameworks []Framework
		wantErr        error
	}{
		{
			name: "Valid/HappyPath",
			appDir: func() ApplicationDirectory {
				appDir, err := newApplicationDirectory("testdata", false)
				require.NoError(t, err)
				return appDir
			},
			wantFrameworks: []Framework{
				{
					Title:               "Example Profile (low)",
					ID:                  "example",
					SupportedComponents: []string{"My Software"},
				},
			},
		},
		{
			name: "Invalid/NoComponentDefinitions",
			appDir: func() ApplicationDirectory {
				tmpDir := t.TempDir()
				appDir, err := newApplicationDirectory(tmpDir, true)
				require.NoError(t, err)
				return appDir
			},
			wantErr: ErrNoComponentDefinitionsFound,
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			appDir := c.appDir()
			gotFrameworks, err := LoadFrameworks(appDir)
			if c.wantErr != nil {
				require.ErrorIs(t, err, c.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, c.wantFrameworks, gotFrameworks)
			}
		})
	}
}

func TestLoadControlSource(t *testing.T) {
	tmpDir := t.TempDir()
	appDir, err := newApplicationDirectory(tmpDir, true)
	require.NoError(t, err)

	tests := []struct {
		name    string
		source  string
		wantErr string
	}{
		{
			name:    "Invalid/Format",
			source:  "profile/profile.json",
			wantErr: "parse \"profile/profile.json\": invalid URI for request",
		},
		{
			name:    "Invalid/WrongDirectory",
			source:  "file://anotherdirectory/profile.json",
			wantErr: fmt.Sprintf("got path anotherdirectory/profile.json, control source is expected to be under path %s", appDir.ControlDir()),
		},
		{
			name:    "Invalid/FileDoesNotExist",
			source:  fmt.Sprintf("file://%s/profile.json", appDir.AppDir()),
			wantErr: fmt.Sprintf("open %s/profile.json: no such file or directory", appDir.AppDir()),
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			_, err := LoadControlSource(appDir, c.source)
			if c.wantErr != "" {
				require.EqualError(t, err, c.wantErr)
			}
		})
	}
}
