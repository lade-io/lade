package cmd

import (
	"fmt"

	"github.com/lade-io/go-lade"
	"github.com/spf13/cobra"
)

var deployCmd = func() *cobra.Command {
	var appName string
	opts := &lade.ReleaseCreateOpts{}
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy an app",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			return deployRun(client, opts, appName)
		},
	}
	cmd.Flags().StringVarP(&appName, "app", "a", "", "App Name")
	return cmd
}()

func deployRun(client *lade.Client, opts *lade.ReleaseCreateOpts, appName string) error {
	err := askSelect("App Name:", getAppName, client, getAppOptions, &appName)
	if err != nil {
		return err
	}
	opts.Source, err = lade.GetTarFile()
	if err != nil {
		return err
	}
	release, err := client.Release.Create(appName, opts)
	if err != nil {
		return err
	}
	logOpts := &lade.LogStreamOpts{Follow: true}
	err = client.Log.ReleaseStream(release, logOpts, printDeployLog)
	if err != nil {
		return err
	}
	fmt.Printf("Build finished use \"%s logs -a %s -f\" to view app logs\n", RootCmd.Use, appName)
	return nil
}
