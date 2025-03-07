// SPDX-License-Identifier: Apache-2.0

package plan

type RuleState uint

const (
	Verify RuleState = iota
	Waived
)

// String prints a string representation of the RuleState.
func (r RuleState) String() string {
	return [...]string{"verify", "waived"}[r]
}

type ControlState uint

const (
	Assess ControlState = iota
	Remove
)

// String prints a string representation of the ControlState.
func (r ControlState) String() string {
	return [...]string{"assess", "remove"}[r]
}

// EditableFields represents editable attributes of an OSCAL Assessment Plan.
type EditableFields struct {
	// controls to keep in the assessment
	// boolean is whether to exclude from the assessment
	Controls map[string]ControlState
	// rules to verify as part of the assessment plan
	// boolean is whether to waive
	Rules map[string]RuleState
	// ODPs with default values
	OrganizationDefinedParams map[string]string
}
