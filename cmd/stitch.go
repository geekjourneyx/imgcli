package cmd

import (
	"image/color"

	"github.com/spf13/cobra"

	"github.com/geekjourneyx/imgcli/pkg/apperr"
	"github.com/geekjourneyx/imgcli/pkg/cli"
	"github.com/geekjourneyx/imgcli/pkg/stitch"
)

func newStitchCommand() *cobra.Command {
	var inputs []string
	var inputDir string
	var output string
	var width int
	var quality int
	var partHeightLimit int

	cmd := &cobra.Command{
		Use:   "stitch",
		Short: "Stitch images vertically into one or more long images",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cli.ValidateJSONEnabled(rootOpts.JSON); err != nil {
				return apperr.Wrap("INVALID_ARGUMENT", 2, err, "invalid output mode")
			}
			if output == "" || width <= 0 {
				return apperr.New("INVALID_ARGUMENT", "--output and --width are required", 2)
			}
			result, err := stitch.Run(stitch.Options{
				Inputs:          inputs,
				InputDir:        inputDir,
				Output:          output,
				Width:           width,
				Quality:         quality,
				PartHeightLimit: partHeightLimit,
				Background:      color.White,
			})
			if err != nil {
				return err
			}
			return cli.PrintSuccess(cmd.OutOrStdout(), "stitch", result)
		},
	}

	cmd.Flags().StringArrayVar(&inputs, "input", nil, "input image path (repeatable)")
	cmd.Flags().StringVar(&inputDir, "input-dir", "", "directory of input images")
	cmd.Flags().StringVar(&output, "output", "", "output image path (.jpg/.jpeg/.png)")
	cmd.Flags().IntVar(&width, "width", 0, "target output width")
	cmd.Flags().IntVar(&quality, "quality", 85, "JPEG quality")
	cmd.Flags().IntVar(&partHeightLimit, "part-height-limit", stitch.DefaultPartHeightLimit, "maximum height per output part")
	return cmd
}
