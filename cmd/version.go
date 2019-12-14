package cmd

import (
	"html/template"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:    "version",
	Short:  "Print current version",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		versionRun(RootCmd)
	},
}

func versionRun(cmd *cobra.Command) {
	t := template.Must(template.New("version").Parse(cmd.VersionTemplate()))
	t.Execute(cmd.OutOrStdout(), cmd)
}
