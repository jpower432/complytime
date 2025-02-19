// SPDX-License-Identifier: Apache-2.0
// The code here is still under development and minimally functional for testing purposes.

package server

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ComplianceAsCode/compliance-operator/pkg/utils"
	"github.com/antchfx/xmlquery"
	"github.com/oscal-compass/compliance-to-policy-go/v2/policy"

	"github.com/complytime/complytime/cmd/openscap-plugin/config"
	"github.com/complytime/complytime/cmd/openscap-plugin/scan"
	"github.com/complytime/complytime/cmd/openscap-plugin/xccdf"
)

var _ policy.Provider = (*PluginServer)(nil)

const ovalCheckType = "http://oval.mitre.org/XMLSchema/oval-definitions-5"

type PluginServer struct {
	Config *config.Config
}

func New(cfg *config.Config) PluginServer {
	return PluginServer{Config: cfg}
}

func (s PluginServer) Configure(m map[string]string) error {
	//TODO implement me
	panic("implement me")
}

func (s PluginServer) Generate(policy policy.Policy) error {
	fmt.Println("Generating a tailoring file")
	tailoringXML, err := xccdf.PolicyToXML(policy, s.Config)
	if err != nil {
		return err
	}

	policyPath := s.Config.Files.Policy
	dst, err := os.Create(policyPath)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err := dst.WriteString(tailoringXML); err != nil {
		return err
	}
	return nil
}

func (s PluginServer) GetResults(_ policy.Policy) (policy.PVPResult, error) {
	fmt.Println("I am being scanned by OpenSCAP")
	pvpResults := policy.PVPResult{}
	_, err := scan.ScanSystem(s.Config, s.Config.Parameters.Profile)
	if err != nil {
		return policy.PVPResult{}, err
	}

	// get some results here
	file, err := os.Open(filepath.Clean(s.Config.Files.ARF))
	if err != nil {
		return policy.PVPResult{}, err
	}
	defer file.Close()

	xmlnode, err := utils.ParseContent(bufio.NewReader(file))
	if err != nil {
		return policy.PVPResult{}, err
	}

	ruleTable := xccdf.NewRuleHashTable(xmlnode)
	results := xmlnode.SelectElements("//rule-result")
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
		ovalCheckName := strings.TrimSpace(ovalRefEl.SelectAttr("name"))

		mappedResult, err := mapResultStatus(result)
		if err != nil {
			return policy.PVPResult{}, err
		}
		observation := policy.ObservationByCheck{
			Title:     ruleIDRef,
			Methods:   []string{"AUTOMATED"},
			Collected: time.Now(),
			CheckID:   ovalCheckName,
			Subjects: []policy.Subject{
				{
					Title:       "My Comp",
					Type:        "component",
					ResourceID:  ruleIDRef,
					EvaluatedOn: time.Now(),
					Result:      mappedResult,
					Reason:      "my reason",
				},
			},
		}
		pvpResults.ObservationsByCheck = append(pvpResults.ObservationsByCheck, observation)
	}

	return pvpResults, nil
}

func mapResultStatus(result *xmlquery.Node) (policy.Result, error) {
	resultEl := result.SelectElement("result")
	if resultEl == nil {
		return policy.ResultInvalid, errors.New("result node has no 'result' attribute")
	}
	switch resultEl.InnerText() {
	case "pass", "fixed":
		return policy.ResultPass, nil
	case "fail":
		return policy.ResultFail, nil
	case "notselected", "notapplicable":
		return policy.ResultError, nil
	case "error", "unknown":
		return policy.ResultError, nil
	}

	return policy.ResultInvalid, fmt.Errorf("couldn't match %s", resultEl.InnerText())
}
