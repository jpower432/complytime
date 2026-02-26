// SPDX-License-Identifier: Apache-2.0

package output_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gemaraproj/go-gemara"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/complytime/complyctl/internal/output"
)

func mockGemaraEvalLog() *gemara.EvaluationLog {
	return &gemara.EvaluationLog{
		Metadata: gemara.Metadata{
			Id:          "test-policy",
			Description: "Test evaluation log",
			Author: gemara.Actor{
				Id:   "complytime",
				Name: "complytime",
				Type: gemara.Software,
			},
		},
		Evaluations: []*gemara.ControlEvaluation{
			{
				Name:    "ctrl-1",
				Result:  gemara.Passed,
				Message: "ok",
				Control: gemara.EntryMapping{ReferenceId: "test-policy", EntryId: "ctrl-1"},
				AssessmentLogs: []*gemara.AssessmentLog{
					{
						Requirement:     gemara.EntryMapping{ReferenceId: "test-policy", EntryId: "req-1"},
						Description:     "passed",
						Result:          gemara.Passed,
						Message:         "passed",
						Applicability:   []string{"default"},
						Start:           "2026-01-01T00:00:00Z",
						StepsExecuted:   1,
						ConfidenceLevel: gemara.High,
					},
				},
			},
			{
				Name:    "ctrl-2",
				Result:  gemara.NotApplicable,
				Message: "tailored",
				Control: gemara.EntryMapping{ReferenceId: "test-policy", EntryId: "ctrl-2"},
				AssessmentLogs: []*gemara.AssessmentLog{
					{
						Requirement:     gemara.EntryMapping{ReferenceId: "test-policy", EntryId: "req-2"},
						Description:     "tailored",
						Result:          gemara.NotApplicable,
						Message:         "tailored",
						Applicability:   []string{"default"},
						Start:           "2026-01-01T00:00:00Z",
						StepsExecuted:   1,
						ConfidenceLevel: gemara.NotSet,
					},
				},
			},
		},
	}
}

func TestToOSCAL_ProducesValidJSON(t *testing.T) {
	outDir := t.TempDir()
	log := mockGemaraEvalLog()

	path, err := output.ToOSCAL(log, outDir)
	require.NoError(t, err)
	assert.FileExists(t, path)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var doc output.OSCALAssessmentResults
	require.NoError(t, json.Unmarshal(data, &doc))

	require.NotNil(t, doc.AssessmentResults, "assessment-results must be present")
	assert.Equal(t, "1.2.0", doc.AssessmentResults.Metadata.OscalVersion)
	assert.NotEmpty(t, doc.AssessmentResults.UUID)
	assert.NotEmpty(t, doc.AssessmentResults.Results)

	assert.Contains(t, doc.AssessmentResults.ImportAp.Href, "test-policy")
}

func TestToOSCAL_SkippedMapsToNotApplicable(t *testing.T) {
	outDir := t.TempDir()
	log := mockGemaraEvalLog()

	path, err := output.ToOSCAL(log, outDir)
	require.NoError(t, err)

	data, _ := os.ReadFile(path)
	var doc output.OSCALAssessmentResults
	require.NoError(t, json.Unmarshal(data, &doc))

	require.NotNil(t, doc.AssessmentResults)
	require.NotNil(t, doc.AssessmentResults.Results[0].Findings)
	findings := *doc.AssessmentResults.Results[0].Findings
	require.Len(t, findings, 2)

	assert.Equal(t, "pass", findings[0].Target.Status.State)
	assert.Equal(t, "not-applicable", findings[1].Target.Status.State)
	assert.Contains(t, findings[1].Remarks, "Skipped due to Tailoring")
}

func TestToOSCAL_ProperUUIDs(t *testing.T) {
	outDir := t.TempDir()
	log := mockGemaraEvalLog()

	path, err := output.ToOSCAL(log, outDir)
	require.NoError(t, err)

	data, _ := os.ReadFile(path)
	var doc output.OSCALAssessmentResults
	require.NoError(t, json.Unmarshal(data, &doc))

	require.NotNil(t, doc.AssessmentResults)
	assert.Len(t, doc.AssessmentResults.UUID, 36)
	for _, result := range doc.AssessmentResults.Results {
		assert.Len(t, result.UUID, 36)
		if result.Findings != nil {
			for _, finding := range *result.Findings {
				assert.Len(t, finding.UUID, 36)
			}
		}
	}
}

func TestToOSCAL_OutputFileNaming(t *testing.T) {
	outDir := t.TempDir()
	log := mockGemaraEvalLog()

	path, err := output.ToOSCAL(log, outDir)
	require.NoError(t, err)

	filename := filepath.Base(path)
	assert.Contains(t, filename, "assessment-results-test-policy-")
	assert.Contains(t, filename, ".json")
}

// TestToOSCAL_SchemaStructureValidation validates the generated OSCAL document
// conforms to key structural requirements of the OSCAL assessment-results model (SC-007, T038).
func TestToOSCAL_SchemaStructureValidation(t *testing.T) {
	outDir := t.TempDir()
	log := mockGemaraEvalLog()

	path, err := output.ToOSCAL(log, outDir)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var raw map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &raw))

	ar, ok := raw["assessment-results"].(map[string]interface{})
	require.True(t, ok, "document must have 'assessment-results' root key")

	assert.Contains(t, ar, "uuid", "must have uuid")
	assert.Contains(t, ar, "metadata", "must have metadata")
	assert.Contains(t, ar, "results", "must have results")

	meta, ok := ar["metadata"].(map[string]interface{})
	require.True(t, ok, "metadata must be an object")
	assert.Contains(t, meta, "title")
	assert.Contains(t, meta, "last-modified")
	assert.Contains(t, meta, "version")
	assert.Contains(t, meta, "oscal-version")
	assert.Equal(t, "1.2.0", meta["oscal-version"])

	importAP, ok := ar["import-ap"].(map[string]interface{})
	require.True(t, ok, "import-ap must be present")
	assert.Contains(t, importAP, "href")

	results, ok := ar["results"].([]interface{})
	require.True(t, ok, "results must be an array")
	assert.NotEmpty(t, results)

	for i, r := range results {
		result, ok := r.(map[string]interface{})
		require.True(t, ok, "result[%d] must be an object", i)
		assert.Contains(t, result, "uuid", "result[%d] must have uuid", i)
		assert.Contains(t, result, "title", "result[%d] must have title", i)
		assert.Contains(t, result, "description", "result[%d] must have description", i)
		assert.Contains(t, result, "start", "result[%d] must have start", i)
		assert.Contains(t, result, "reviewed-controls", "result[%d] must have reviewed-controls", i)

		findings, ok := result["findings"].([]interface{})
		require.True(t, ok, "result[%d] findings must be an array", i)
		for j, f := range findings {
			finding, ok := f.(map[string]interface{})
			require.True(t, ok, "finding[%d][%d] must be an object", i, j)
			assert.Contains(t, finding, "uuid", "finding must have uuid")
			assert.Contains(t, finding, "title", "finding must have title")
			assert.Contains(t, finding, "description", "finding must have description")
			assert.Contains(t, finding, "target", "finding must have target")

			target, ok := finding["target"].(map[string]interface{})
			require.True(t, ok, "finding target must be an object")
			assert.Contains(t, target, "type")
			assert.Contains(t, target, "target-id")
			assert.Contains(t, target, "status")

			status, ok := target["status"].(map[string]interface{})
			require.True(t, ok, "finding target status must be an object")
			stateVal, _ := status["state"].(string)
			validStates := []string{"pass", "fail", "not-applicable", "error", "unknown"}
			assert.Contains(t, validStates, stateVal,
				"finding state '%s' must be valid OSCAL value", stateVal)
		}
	}
}
