package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/geekjourneyx/imgcli/pkg/apperr"
	"github.com/geekjourneyx/imgcli/pkg/cli"
	"github.com/geekjourneyx/imgcli/pkg/presets"
	"github.com/geekjourneyx/imgcli/pkg/smartpad"
)

func newSmartpadCommand() *cobra.Command {
	var input string
	var output string
	var preset string
	var background string
	var blurSigma float64
	var quality int

	cmd := &cobra.Command{
		Use:   "smartpad",
		Short: "Resize image into a social preset with safe padding",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cli.ValidateJSONEnabled(rootOpts.JSON); err != nil {
				return apperr.Wrap("INVALID_ARGUMENT", 2, err, "invalid output mode")
			}
			if input == "" || output == "" || preset == "" {
				return apperr.New("INVALID_ARGUMENT", "--input, --output, and --preset are required", 2)
			}
			target, err := presets.Get(preset)
			if err != nil {
				return apperr.New("PRESET_NOT_FOUND", fmt.Sprintf("preset %q not found", preset), 2)
			}
			result, err := smartpad.Run(smartpad.Options{
				Input:      input,
				Output:     output,
				Target:     target,
				Background: smartpad.BackgroundMode(background),
				BlurSigma:  blurSigma,
				Quality:    quality,
			})
			if err != nil {
				return err
			}
			return cli.PrintSuccess(cmd.OutOrStdout(), "smartpad", result)
		},
	}

	cmd.Flags().StringVar(&input, "input", "", "input image path")
	cmd.Flags().StringVar(&output, "output", "", "output image path (.jpg/.jpeg/.png)")
	cmd.Flags().StringVar(&preset, "preset", "", "target preset name")
	cmd.Flags().StringVar(&background, "background", "blur", "background mode: blur|solid")
	cmd.Flags().Float64Var(&blurSigma, "blur-sigma", 5.0, "blur sigma for blur background")
	cmd.Flags().IntVar(&quality, "quality", 85, "JPEG quality")
	return cmd
}
