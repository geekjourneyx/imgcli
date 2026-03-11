package cmd

import (
	"github.com/spf13/cobra"

	"github.com/geekjourneyx/imgcli/pkg/apperr"
	"github.com/geekjourneyx/imgcli/pkg/cli"
	"github.com/geekjourneyx/imgcli/pkg/topdf"
)

func newTopDFCommand() *cobra.Command {
	var inputs []string
	var inputDir string
	var output string
	var watermarkText string
	var watermarkOpacity float64
	var watermarkSize float64
	var watermarkPosition string
	var quality int

	cmd := &cobra.Command{
		Use:   "topdf",
		Short: "Pack images into a PDF with optional visible watermark",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cli.ValidateJSONEnabled(rootOpts.JSON); err != nil {
				return apperr.Wrap("INVALID_ARGUMENT", 2, err, "invalid output mode")
			}
			if output == "" {
				return apperr.New("INVALID_ARGUMENT", "--output is required", 2)
			}
			result, err := topdf.Run(topdf.Options{
				Inputs:            inputs,
				InputDir:          inputDir,
				Output:            output,
				WatermarkText:     watermarkText,
				WatermarkOpacity:  watermarkOpacity,
				WatermarkSize:     watermarkSize,
				WatermarkPosition: topdf.WatermarkPosition(watermarkPosition),
				Quality:           quality,
			})
			if err != nil {
				return err
			}
			return cli.PrintSuccess(cmd.OutOrStdout(), "topdf", result)
		},
	}

	cmd.Flags().StringArrayVar(&inputs, "input", nil, "input image path (repeatable)")
	cmd.Flags().StringVar(&inputDir, "input-dir", "", "directory of input images")
	cmd.Flags().StringVar(&output, "output", "", "output PDF path")
	cmd.Flags().StringVar(&watermarkText, "watermark-text", "", "watermark text")
	cmd.Flags().Float64Var(&watermarkOpacity, "watermark-opacity", 0.25, "watermark opacity")
	cmd.Flags().Float64Var(&watermarkSize, "watermark-size", 42, "watermark font size")
	cmd.Flags().StringVar(&watermarkPosition, "watermark-position", "br", "watermark position: br|center|tile")
	cmd.Flags().IntVar(&quality, "quality", 85, "JPEG quality for embedded pages")
	return cmd
}
