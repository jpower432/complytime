// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/complytime/complytime/cmd/complytime/option"
)

func TestGetPlanSettingsForWorkspace(t *testing.T) {
	// Test that the proper error messages are thrown when
	// collecting setting information
	testOpts := &option.ComplyTime{
		UserWorkspace: "doesnotexist",
	}
	wantErr := "error: assessment plan does exist in workspace doesnotexist: o" +
		"pen doesnotexist/assessment-plan.json: no such file or directory\n\nDid you run the plan command?"
	_, gotErr := getPlanSettingsForWorkspace(testOpts)
	require.EqualError(t, gotErr, wantErr)

	testOpts.UserWorkspace = "testdata"
	wantErr = "assessment plan testdata/assessment-plan.json does not have associated activities: no local activities detected"
	_, gotErr = getPlanSettingsForWorkspace(testOpts)
	require.EqualError(t, gotErr, wantErr)
}
