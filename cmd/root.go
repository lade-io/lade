package cmd

import (
	"log"

	"github.com/lade-io/lade/config"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var (
	conf    = &config.Config{}
	RootCmd = &cobra.Command{
		Use:   "lade",
		Short: "Manage your Lade resources",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmd.SilenceUsage = true
		},
	}
)

func SetVersion(version string) {
	RootCmd.Version = version
}

func init() {
	cobra.OnInitialize(initLogger, initConfig, initPrompt)
	RootCmd.SetHelpTemplate(helpTemplate)
	RootCmd.SetUsageTemplate(usageTemplate)

	RootCmd.PersistentFlags().BoolP("help", "h", false, "Print help message")
	RootCmd.Flags().BoolP("version", "v", false, "Print version and exit")

	RootCmd.AddCommand(addonsCmd)
	RootCmd.AddCommand(appsCmd)
	RootCmd.AddCommand(deployCmd)
	RootCmd.AddCommand(disksCmd)
	RootCmd.AddCommand(domainsCmd)
	RootCmd.AddCommand(envCmd)
	RootCmd.AddCommand(loginCmd)
	RootCmd.AddCommand(logoutCmd)
	RootCmd.AddCommand(logsCmd)
	RootCmd.AddCommand(plansCmd)
	RootCmd.AddCommand(psCmd)
	RootCmd.AddCommand(regionsCmd)
	RootCmd.AddCommand(runCmd)
	RootCmd.AddCommand(scaleCmd)
	RootCmd.AddCommand(versionCmd)
	disableFlagsUsage(RootCmd)
}

func initConfig() {
	if err := config.Load(conf); err != nil {
		log.Fatal(err)
	}
}

func initLogger() {
	log.SetFlags(0)
}

func initPrompt() {
	table.DefaultPadding = 4
}

func disableFlagsUsage(cmd *cobra.Command) {
	cmd.DisableFlagsInUseLine = true
	for _, sub := range cmd.Commands() {
		disableFlagsUsage(sub)
	}
}
