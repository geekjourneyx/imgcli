package cmd

import (
	"github.com/spf13/cobra"

	"github.com/geekjourneyx/imgcli/pkg/apperr"
	"github.com/geekjourneyx/imgcli/pkg/cli"
	"github.com/geekjourneyx/imgcli/pkg/compose"
)

func newComposeCommand() *cobra.Command {
	var opts compose.Options

	cmd := &cobra.Command{
		Use:   "compose",
		Short: "Render a fixed-layout creator card from an existing image",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cli.ValidateJSONEnabled(rootOpts.JSON); err != nil {
				return apperr.Wrap("INVALID_ARGUMENT", 2, err, "invalid output mode")
			}
			result, err := compose.Run(opts)
			if err != nil {
				return err
			}
			return cli.PrintSuccess(cmd.OutOrStdout(), "compose", result)
		},
	}

	cmd.Flags().StringVar(&opts.Input, "input", "", "input image path")
	cmd.Flags().StringVar(&opts.Output, "output", "", "output image path (.jpg/.jpeg/.png)")
	cmd.Flags().IntVar(&opts.Width, "width", 0, "output width")
	cmd.Flags().IntVar(&opts.Height, "height", 0, "output height")
	cmd.Flags().StringVar(&opts.BackgroundColor, "background-color", "", "background color hex")
	cmd.Flags().StringVar(&opts.BackgroundImage, "background-image", "", "background image path")
	cmd.Flags().StringVar((*string)(&opts.Layout), "layout", string(compose.LayoutPoster), "layout family: poster|cover|quote-card|product-card")
	cmd.Flags().StringVar(&opts.Title, "title", "", "title text")
	cmd.Flags().StringVar(&opts.Subtitle, "subtitle", "", "subtitle text")
	cmd.Flags().Float64Var(&opts.TitleSize, "title-size", 0, "title font size")
	cmd.Flags().Float64Var(&opts.SubtitleSize, "subtitle-size", 0, "subtitle font size")
	cmd.Flags().StringVar(&opts.TitleColor, "title-color", "", "title color hex")
	cmd.Flags().StringVar(&opts.SubtitleColor, "subtitle-color", "", "subtitle color hex")
	cmd.Flags().StringVar(&opts.FontPath, "font", "", "font path")
	cmd.Flags().StringVar(&opts.Logo, "logo", "", "logo image path")
	cmd.Flags().StringVar(&opts.BannerBadge, "badge", "", "badge text")
	cmd.Flags().IntVar(&opts.Padding, "padding", 48, "base padding in pixels")
	cmd.Flags().Float64Var(&opts.Radius, "radius", 28, "foreground image corner radius")
	cmd.Flags().StringVar(&opts.SafeArea, "safe-area", "", "safe area top,right,bottom,left")
	cmd.Flags().IntVar(&opts.Quality, "quality", 85, "JPEG quality")
	return cmd
}
