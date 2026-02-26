// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSanitizeInput tests the SanitizeInput function with various valid and invalid inputs.
func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		input       string
		expected    string
		expectError bool
	}{
		// Valid inputs
		{"valid-input", "valid-input", false},
		{"another_valid.input", "another_valid.input", false},
		{"CAPS_and_numbers123", "CAPS_and_numbers123", false},
		{"mixed-123.UP_case", "mixed-123.UP_case", false},

		// Invalid inputs
		{"invalid/input", "", true},     // contains /
		{"input with spaces", "", true}, // contains spaces
		{"invalid@input", "", true},     // contains @
		{"<invalid>", "", true},         // contains < >
		{";ls", "", true},               // contains ;
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := SanitizeInput(tt.input)
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}
			if result != tt.expected {
				t.Errorf("Expected result: %s, got: %s", tt.expected, result)
			}
		})
	}
}

// TestSanitizePath tests the SanitizePath function with various inputs.
func TestSanitizePath(t *testing.T) {
	usr, _ := user.Current()
	homeDir := usr.HomeDir

	tests := []struct {
		input       string
		expected    string
		expectError bool
	}{
		// Normalizing paths
		{"/foo/bar/../baz", "/foo/baz", false},
		{"./foo/bar", "foo/bar", false},
		{"foo/./bar", "foo/bar", false},
		{"foo/bar/..", "foo", false},
		{"/foo//bar", "/foo/bar", false},
		{"foo//bar//baz", "foo/bar/baz", false},
		{"foo/bar/../../baz", "baz", false},
		{"./../foo", "../foo", false},

		// Expanding paths
		{"~/foo/bar", filepath.Join(homeDir, "foo", "bar"), false},
		{"~", homeDir, false},

		// Weird but valid cases
		{"~weird", "~weird", false}, // not common but possible
		{"", ".", false},            // empty path is updated to the current directory
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := SanitizePath(tt.input)
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}
			if result != tt.expected {
				t.Errorf("Expected result: %s, got: %s", tt.expected, result)
			}
		})
	}
}

func setupTestFiles() error {
	if err := os.MkdirAll("testdata", os.ModePerm); err != nil {
		return err
	}

	if err := os.WriteFile("testdata/valid.xml", []byte(`<root></root>`), 0600); err != nil {
		return err
	}
	if err := os.WriteFile("testdata/invalid.xml", []byte(`<root>`), 0600); err != nil {
		return err
	}
	return nil
}

func teardownTestFiles() {
	os.RemoveAll("testdata")
}

func TestIsXMLFile(t *testing.T) {
	if err := setupTestFiles(); err != nil {
		t.Fatalf("Failed to setup test files: %v", err)
	}
	defer teardownTestFiles()

	tests := []struct {
		name      string
		filePath  string
		want      bool
		expectErr bool
	}{
		{
			name:      "Valid XML file",
			filePath:  "testdata/valid.xml",
			want:      true,
			expectErr: false,
		},
		{
			name:      "Invalid XML file",
			filePath:  "testdata/invalid.xml",
			want:      false,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isXML, err := IsXMLFile(tt.filePath)
			if (err != nil) != tt.expectErr {
				t.Errorf("IsXMLFile(%s) error = %v, expectErr %v", tt.filePath, err, tt.expectErr)
				return
			}
			if isXML != tt.want {
				t.Errorf("IsXMLFile() = %v, want %v", isXML, tt.want)
			}
		})
	}
}

// TestEnsureDirectory tests the ensureDirectory function with various cases.
func TestEnsureDirectory(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		path        string
		expectError bool
	}{
		// Valid cases
		{filepath.Join(tempDir, "absent_dir"), false},   // directory does not exist, should be created
		{filepath.Join(tempDir, "existing_dir"), false}, // directory already exists

		// Invalid cases
		{tempDir + "/invalid\000dir", true}, // invalid directory name
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if tt.path == filepath.Join(tempDir, "existing_dir") {
				// Create directory for existing_dir test
				if err := os.MkdirAll(tt.path, 0750); err != nil {
					t.Fatalf("Failed to create directory: %v", err)
				}
			}

			err := ensureDirectory(tt.path)
			if (err != nil) != tt.expectError {
				t.Errorf("Expected error: %v, got: %v", tt.expectError, err)
			}

			// Check if directory was created
			if !tt.expectError {
				if _, err := os.Stat(tt.path); os.IsNotExist(err) {
					t.Errorf("Expected directory to be created: %s", tt.path)
				}
			}
		})
	}
}

// TestEnsureWorkspace tests the ensureWorkspace function.
func TestEnsureWorkspace(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	t.Cleanup(func() { os.Chdir(originalDir) })

	directories, err := ensureWorkspace()
	require.NoError(t, err)

	for _, dir := range directories {
		if _, statErr := os.Stat(dir); os.IsNotExist(statErr) {
			t.Errorf("Expected directory to be created: %s", dir)
		}
	}
}

// TestDefineFilesPaths tests the defineFilesPaths function.
func TestDefineFilesPaths(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	t.Cleanup(func() { os.Chdir(originalDir) })

	cfg := &Config{
		Files: struct {
			Datastream string `config:"datastream"`
			Results    string
			ARF        string
			Policy     string
		}{
			Datastream: filepath.Join(tempDir, "datastream.xml"),
			Results:    "results.xml",
			ARF:        "arf.xml",
			Policy:     "policy.yaml",
		},
	}

	require.NoError(t, defineFilesPaths(cfg))

	wsDir := ".complytime"
	expectedPolicyPath := filepath.Join(wsDir, PluginDir, "policy", "policy.yaml")
	expectedResultsPath := filepath.Join(wsDir, PluginDir, "results", "results.xml")
	expectedARFPath := filepath.Join(wsDir, PluginDir, "results", "arf.xml")

	if cfg.Files.Policy != expectedPolicyPath {
		t.Errorf("Expected policy path: %s, got: %s", expectedPolicyPath, cfg.Files.Policy)
	}
	if cfg.Files.Results != expectedResultsPath {
		t.Errorf("Expected results path: %s, got: %s", expectedResultsPath, cfg.Files.Results)
	}
	if cfg.Files.ARF != expectedARFPath {
		t.Errorf("Expected ARF path: %s, got: %s", expectedARFPath, cfg.Files.ARF)
	}
}

func TestConfig_LoadSettings(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	t.Cleanup(func() { os.Chdir(originalDir) })

	tempDataStream := filepath.Join(tempDir, "datastream.xml")
	err = os.WriteFile(tempDataStream, []byte("example"), 0400)
	require.NoError(t, err)

	wsDir := ".complytime"

	tests := []struct {
		name          string
		inputSettings map[string]string
		expectError   string
		wantCfg       Config
	}{
		{
			name: "Valid/AllSettingsSupplied",
			inputSettings: map[string]string{
				"datastream": tempDataStream,
				"profile":    "test",
			},
			wantCfg: Config{
				Files: struct {
					Datastream string `config:"datastream"`
					Results    string
					ARF        string
					Policy     string
				}{
					Datastream: tempDataStream,
					Results:    filepath.Join(wsDir, "openscap", "results", DefaultResultsFile),
					ARF:        filepath.Join(wsDir, "openscap", "results", DefaultARFFile),
					Policy:     filepath.Join(wsDir, "openscap", "policy", DefaultPolicyFile),
				},
				Parameters: struct {
					Profile string `config:"profile"`
				}{Profile: "test"},
			},
			expectError: "",
		},
		{
			name: "Invalid/MissingSettings",
			inputSettings: map[string]string{
				"datastream": tempDataStream,
			},
			expectError: "missing configuration value for option \"profile\" (field: Profile)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConfig := NewConfig()
			err := gotConfig.LoadSettings(tt.inputSettings)

			if tt.expectError != "" {
				require.EqualError(t, err, tt.expectError)
			} else {
				require.Equal(t, tt.wantCfg, *gotConfig)
				require.NoError(t, err)
			}
		})
	}
}
