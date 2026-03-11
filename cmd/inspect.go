package cmd

import (
	"github.com/spf13/cobra"

	"github.com/geekjourneyx/imgcli/pkg/apperr"
	"github.com/geekjourneyx/imgcli/pkg/cli"
	"github.com/geekjourneyx/imgcli/pkg/inspect"
)

func newInspectCommand() *cobra.Command {
	var inputs []string
	var inputDir string
	var includeHash bool
	var includeColors bool
	var limit int

	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect image metadata for agent decision-making",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cli.ValidateJSONEnabled(rootOpts.JSON); err != nil {
				return apperr.Wrap("INVALID_ARGUMENT", 2, err, "invalid output mode")
			}
			result, err := inspect.Run(inspect.Options{
				Inputs:        inputs,
				InputDir:      inputDir,
				IncludeHash:   includeHash,
				IncludeColors: includeColors,
				Limit:         limit,
			})
			if err != nil {
				return err
			}
			return cli.PrintSuccess(cmd.OutOrStdout(), "inspect", result)
		},
	}

	cmd.Flags().StringArrayVar(&inputs, "input", nil, "input image path (repeatable)")
	cmd.Flags().StringVar(&inputDir, "input-dir", "", "directory of input images")
	cmd.Flags().BoolVar(&includeHash, "hash", false, "include sha256 and perceptual hash")
	cmd.Flags().BoolVar(&includeColors, "color-stats", false, "include average and dominant colors")
	cmd.Flags().IntVar(&limit, "limit", 0, "limit files when using --input-dir")
	return cmd
}
