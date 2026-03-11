package cmd

import (
	"github.com/spf13/cobra"

	"github.com/geekjourneyx/imgcli/pkg/apperr"
	"github.com/geekjourneyx/imgcli/pkg/cli"
	"github.com/geekjourneyx/imgcli/pkg/runbook"
)

func newRunCommand() *cobra.Command {
	var opts runbook.Options

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Execute a declarative image workflow recipe",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cli.ValidateJSONEnabled(rootOpts.JSON); err != nil {
				return apperr.Wrap("INVALID_ARGUMENT", 2, err, "invalid output mode")
			}
			result, err := runbook.Run(opts)
			if err != nil {
				return err
			}
			return cli.PrintSuccess(cmd.OutOrStdout(), "run", result)
		},
	}

	cmd.Flags().StringVar(&opts.RecipePath, "recipe", "", "recipe file path (.json/.yaml/.yml)")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "validate and print the resolved execution plan without writing outputs")
	return cmd
}
