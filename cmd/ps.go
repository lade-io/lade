package cmd

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/lade-io/go-lade"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var psCmd = func() *cobra.Command {
	var appName string
	cmd := &cobra.Command{
		Use:   "ps",
		Short: "Display running tasks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			return psRun(client, appName)
		},
	}
	cmd.Flags().StringVarP(&appName, "app", "a", "", "App Name")
	return cmd
}()

func psRun(client *lade.Client, appName string) error {
	if err := askSelect("App Name:", getAppName, client, getAppOptions, &appName); err != nil {
		return err
	}
	containers, err := client.Container.List(appName)
	if err != nil {
		return err
	}
	t := table.New("NAME", "PLAN", "STARTED", "COMMAND")
	for _, container := range containers {
		if container.Process == nil {
			continue
		}
		number := container.Process.Number
		if number == 0 {
			number = container.Number
		}
		name := fmt.Sprintf("%s.%d", container.Process.Type, number)
		t.AddRow(name, container.PlanID, humanize.Time(container.CreatedAt), container.Process.Command)
	}
	t.Print()
	return nil
}
