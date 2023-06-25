package cmd

import (
	"time"

	"github.com/lade-io/go-lade"
	"github.com/spf13/cobra"
)

var logsCmd = func() *cobra.Command {
	var appName string
	var since time.Duration
	opts := &lade.LogStreamOpts{}
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Show logs from an app",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			if since > 0 {
				opts.Since = time.Now().UTC().Add(-since)
			}
			return logsRun(client, opts, appName)
		},
	}
	cmd.Flags().StringVarP(&appName, "app", "a", "", "App Name")
	cmd.Flags().BoolVarP(&opts.Follow, "follow", "f", false, "Follow Log")
	cmd.Flags().DurationVarP(&since, "since", "s", 0, "Show Logs Since")
	cmd.Flags().IntVarP(&opts.Tail, "tail", "t", 0, "Number of Lines")
	return cmd
}()

func logsRun(client *lade.Client, opts *lade.LogStreamOpts, appName string) error {
	err := askSelect("App Name:", getAppName, client, getAppOptions, &appName)
	if err != nil {
		return err
	}
	processes, err := client.Process.List(appName)
	if err != nil {
		return err
	}
	var width int
	for _, process := range processes {
		if len(process.Type) > width {
			width = len(process.Type)
		}
	}
	return client.Log.AppStream(appName, opts, printNameLog(width))
}
