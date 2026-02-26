// SPDX-License-Identifier: Apache-2.0

package output

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gemaraproj/go-gemara"

	oscalUUID "github.com/defenseunicorns/go-oscal/src/pkg/uuid"
	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"

	"github.com/complytime/complyctl/internal/complytime"
)

// FIXME(jpower432): This would probably make more sense in go-gemara/gemaraconv

// OSCALAssessmentResults is the go-oscal envelope for a standalone OSCAL document.
type OSCALAssessmentResults = oscalTypes.OscalCompleteSchema

const oscalVersion = "1.2.0"

// ToOSCAL converts a gemara.EvaluationLog to OSCAL assessment-results format.
func ToOSCAL(log *gemara.EvaluationLog, outDir string) (string, error) {
	now := time.Now().UTC()
	policyID := log.Metadata.Id

	findings := make([]oscalTypes.Finding, 0)
	for _, ce := range log.Evaluations {
		for _, al := range ce.AssessmentLogs {
			resultStr := gemaraResultToOSCAL(al.Result)
			remarks := al.Message
			if resultStr == "not-applicable" {
				remarks = fmt.Sprintf("Skipped due to Tailoring: %s", al.Message)
			}
			findings = append(findings, oscalTypes.Finding{
				UUID:        newUUID(),
				Title:       fmt.Sprintf("Assessment: %s", al.Requirement.EntryId),
				Description: al.Message,
				Target: oscalTypes.FindingTarget{
					Type:     "objective-id",
					TargetId: al.Requirement.EntryId,
					Status: oscalTypes.ObjectiveStatus{
						State: resultStr,
					},
				},
				Remarks: remarks,
			})
		}
	}

	result := oscalTypes.Result{
		UUID:        newUUID(),
		Title:       fmt.Sprintf("Policy: %s", policyID),
		Description: fmt.Sprintf("Assessment results for policy %s", policyID),
		Start:       now,
		ReviewedControls: oscalTypes.ReviewedControls{
			ControlSelections: []oscalTypes.AssessedControls{
				{Description: "All controls assessed for policy " + policyID},
			},
		},
	}
	if len(findings) > 0 {
		result.Findings = &findings
	}

	doc := OSCALAssessmentResults{
		AssessmentResults: &oscalTypes.AssessmentResults{
			UUID: newUUID(),
			Metadata: oscalTypes.Metadata{
				Title:        fmt.Sprintf("Compliance scan: %s", policyID),
				LastModified: now,
				Version:      "1.0.0",
				OscalVersion: oscalVersion,
			},
			ImportAp: oscalTypes.ImportAp{
				Href: fmt.Sprintf("file://assessment-plan-%s.json", policyID),
			},
			Results: []oscalTypes.Result{result},
		},
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal OSCAL: %w", err)
	}

	if outDir == "" {
		outDir = "."
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	filename := fmt.Sprintf("assessment-results-%s-%s.json",
		complytime.FilenameSafe(policyID), time.Now().Format("20060102-150405"))
	path := filepath.Join(outDir, filename)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return "", fmt.Errorf("failed to write OSCAL file: %w", err)
	}
	return path, nil
}

func newUUID() string {
	return oscalUUID.NewUUID()
}

func gemaraResultToOSCAL(r gemara.Result) string {
	switch r {
	case gemara.Passed:
		return "pass"
	case gemara.Failed:
		return "fail"
	case gemara.NotApplicable:
		return "not-applicable"
	case gemara.Unknown:
		return "error"
	default:
		return "unknown"
	}
}
