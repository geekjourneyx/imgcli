package cmd

import (
	"github.com/spf13/cobra"

	"github.com/geekjourneyx/imgcli/pkg/apperr"
	"github.com/geekjourneyx/imgcli/pkg/cli"
	"github.com/geekjourneyx/imgcli/pkg/convert"
)

func newConvertCommand() *cobra.Command {
	var opts convert.Options

	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Normalize image format, size, and delivery settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cli.ValidateJSONEnabled(rootOpts.JSON); err != nil {
				return apperr.Wrap("INVALID_ARGUMENT", 2, err, "invalid output mode")
			}
			result, err := convert.Run(opts)
			if err != nil {
				return err
			}
			return cli.PrintSuccess(cmd.OutOrStdout(), "convert", result)
		},
	}

	cmd.Flags().StringVar(&opts.Input, "input", "", "input image path")
	cmd.Flags().StringVar(&opts.Output, "output", "", "output image path (.jpg/.jpeg/.png)")
	cmd.Flags().IntVar(&opts.Quality, "quality", 85, "JPEG quality")
	cmd.Flags().BoolVar(&opts.StripMetadata, "strip-metadata", false, "re-encode without preserving original metadata")
	cmd.Flags().StringVar(&opts.FlattenBackground, "flatten-background", "", "hex background color for alpha flattening")
	cmd.Flags().IntVar(&opts.MaxWidth, "max-width", 0, "maximum output width")
	cmd.Flags().IntVar(&opts.MaxHeight, "max-height", 0, "maximum output height")
	return cmd
}
