package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/geekjourneyx/imgcli/pkg/apperr"
	"github.com/geekjourneyx/imgcli/pkg/cli"
	"github.com/geekjourneyx/imgcli/pkg/version"
)

type rootOptions struct {
	JSON     bool
	Config   string
	LogLevel string
}

var rootOpts rootOptions

func Execute() int {
	rootCmd := newRootCommand()
	if err := rootCmd.Execute(); err != nil {
		cli.PrintError(os.Stderr, err)
		return apperr.From(err).ExitCode
	}
	return 0
}

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           version.Name,
		Short:         "Agent-native image processing CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.Version,
	}
	cmd.PersistentFlags().BoolVar(&rootOpts.JSON, "json", true, "emit JSON output")
	cmd.PersistentFlags().StringVar(&rootOpts.Config, "config", "", "config file path (reserved)")
	cmd.PersistentFlags().StringVar(&rootOpts.LogLevel, "log-level", "info", "log level (reserved)")
	cmd.AddCommand(newInspectCommand())
	cmd.AddCommand(newComposeCommand())
	cmd.AddCommand(newConvertCommand())
	cmd.AddCommand(newVariantsCommand())
	cmd.AddCommand(newRunCommand())
	cmd.AddCommand(newSmartpadCommand())
	cmd.AddCommand(newTopDFCommand())
	cmd.AddCommand(newStitchCommand())
	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cli.ValidateJSONEnabled(rootOpts.JSON); err != nil {
				return apperr.Wrap("INVALID_ARGUMENT", 2, err, "invalid output mode")
			}
			return cli.PrintSuccess(cmd.OutOrStdout(), "version", map[string]string{
				"name":    version.Name,
				"version": version.Version,
			})
		},
	})
	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)
	cmd.SetVersionTemplate(fmt.Sprintf("{\"ok\":true,\"command\":\"version\",\"data\":{\"name\":\"%s\",\"version\":\"%s\"}}\n", version.Name, version.Version))
	return cmd
}
