package cmd

import (
	"github.com/spf13/cobra"

	"github.com/geekjourneyx/imgcli/pkg/apperr"
	"github.com/geekjourneyx/imgcli/pkg/cli"
	"github.com/geekjourneyx/imgcli/pkg/variants"
)

func newVariantsCommand() *cobra.Command {
	var opts variants.Options

	cmd := &cobra.Command{
		Use:   "variants",
		Short: "Export multiple platform variants from one source image",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cli.ValidateJSONEnabled(rootOpts.JSON); err != nil {
				return apperr.Wrap("INVALID_ARGUMENT", 2, err, "invalid output mode")
			}
			if err := variants.ValidateTemplate(opts.FilenameTemplate); err != nil {
				return apperr.New("INVALID_ARGUMENT", err.Error(), 2)
			}
			result, err := variants.Run(opts)
			if err != nil {
				return err
			}
			return cli.PrintSuccess(cmd.OutOrStdout(), "variants", result)
		},
	}

	cmd.Flags().StringVar(&opts.Input, "input", "", "input image path")
	cmd.Flags().StringVar(&opts.OutputDir, "output-dir", "", "output directory")
	cmd.Flags().StringVar(&opts.PresetSet, "preset-set", "", "built-in preset set name")
	cmd.Flags().StringArrayVar(&opts.Presets, "preset", nil, "preset name (repeatable)")
	cmd.Flags().StringVar((*string)(&opts.Background), "background", "blur", "background mode: blur|solid")
	cmd.Flags().StringVar(&opts.FilenameTemplate, "filename-template", "{base}_{preset}{ext}", "output filename template")
	cmd.Flags().Float64Var(&opts.BlurSigma, "blur-sigma", 5.0, "blur sigma for blur background")
	cmd.Flags().IntVar(&opts.Quality, "quality", 85, "JPEG quality")
	return cmd
}
