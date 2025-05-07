// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"

	gowitness "github.com/in-toto/go-witness"
	"github.com/in-toto/go-witness/archivista"
	"github.com/in-toto/go-witness/attestation"
	"github.com/invopop/jsonschema"
	"github.com/oscal-compass/compliance-to-policy-go/v2/framework"
	"github.com/oscal-compass/compliance-to-policy-go/v2/framework/actions"
	"github.com/oscal-compass/oscal-sdk-go/extensions"
	"github.com/oscal-compass/oscal-sdk-go/validation"
	"github.com/revanite-io/sci/layer2"
	"github.com/revanite-io/sci/layer4"
	"github.com/spf13/cobra"

	"github.com/complytime/complytime/cmd/complytime/option"
	"github.com/complytime/complytime/internal/complytime"
)

// evaluateOptions defined options for the scan subcommand.
type evaluateOptions struct {
	*option.Common
	complyTimeOpts *option.ComplyTime
	archivistaURL  string
}

// scanCmd creates a new cobra.Command for the version subcommand.
func scanCmd(common *option.Common) *cobra.Command {
	scanOpts := &evaluateOptions{
		Common:         common,
		complyTimeOpts: &option.ComplyTime{},
	}
	cmd := &cobra.Command{
		Use:          "evaluate [flags]",
		Short:        "Scan environment with assessment plan",
		Example:      "complytime scan",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runEvaluation(cmd, scanOpts)
		},
	}
	cmd.Flags().StringVarP(&scanOpts.archivistaURL, "archivista-url", "a", "localhost:8081", "URL to archivista instance")
	scanOpts.complyTimeOpts.BindFlags(cmd.Flags())
	return cmd
}

func runEvaluation(cmd *cobra.Command, opts *evaluateOptions) error {
	validator := validation.NewSchemaValidator()
	// Load settings from assessment plan
	ap, _, err := loadPlan(opts.complyTimeOpts, validator)
	if err != nil {
		return err
	}

	inputContext, err := complytime.ActionsContextFromPlan(ap)
	if err != nil {
		return err
	}

	// Create the application directory if it does not exist
	appDir, err := complytime.NewApplicationDirectory(true)
	if err != nil {
		return err
	}
	logger.Debug(fmt.Sprintf("Using application directory: %s", appDir.AppDir()))

	cfg, err := complytime.Config(appDir)
	if err != nil {
		return err
	}

	// set config logger to CLI charm logger
	cfg.Logger = logger

	manager, err := framework.NewPluginManager(cfg)
	if err != nil {
		return fmt.Errorf("error initializing plugin manager: %w", err)
	}

	// Determine what profile to load from framework information captured
	// from state (assessment plan). This is required to populate complyTime required plugin options.
	frameworkProp, valid := extensions.GetTrestleProp(extensions.FrameworkProp, *ap.Metadata.Props)
	if !valid {
		return fmt.Errorf("error reading framework property from assessment plan")
	}
	opts.complyTimeOpts.FrameworkID = frameworkProp.Value
	logger.Debug(fmt.Sprintf("Framework property was successfully read from the assessment plan: %v.", frameworkProp))

	pluginOptions := opts.complyTimeOpts.ToPluginOptions()
	plugins, cleanup, err := complytime.Plugins(manager, inputContext, pluginOptions)
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		return fmt.Errorf("errors launching plugins: %w", err)
	}
	logger.Info(fmt.Sprintf("Successfully loaded %v plugin(s).", len(plugins)))

	allResults, err := actions.AggregateResults(cmd.Context(), inputContext, plugins)
	if err != nil {
		return err
	}

	var attestors []attestation.Attestor
	for _, result := range allResults {
		evaluation, err := actions.Evaluate(cmd.Context(), inputContext, []layer2.Control{}, result)
		if err != nil {
			return err
		}

		// Make an at testator
		l4Attestor := Layer4{
			eval: evaluation,
		}
		attestors = append(attestors, l4Attestor)
	}

	// using this purposefully for not because I just want the one envelope
	runResults, err := gowitness.Run("step", gowitness.RunWithAttestors(attestors))
	if err != nil {
		return err
	}

	// export attestations to Archivista
	client := archivista.New(opts.archivistaURL)
	_, err = client.Store(cmd.Context(), runResults.SignedEnvelope)
	if err != nil {
		return err
	}

	return nil
}

var _ attestation.Attestor = (*Layer4)(nil)

type Layer4 struct {
	eval layer4.Layer4
}

func (l Layer4) Name() string {
	//TODO implement me
	panic("implement me")
}

func (l Layer4) Type() string {
	//TODO implement me
	panic("implement me")
}

func (l Layer4) RunType() attestation.RunType {
	//TODO implement me
	panic("implement me")
}

func (l Layer4) Attest(ctx *attestation.AttestationContext) error {
	//TODO implement me
	panic("implement me")
}

func (l Layer4) Schema() *jsonschema.Schema {
	//TODO implement me
	panic("implement me")
}
