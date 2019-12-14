package cmd

import (
	"github.com/lade-io/go-lade"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var regionsCmd = &cobra.Command{
	Use:   "regions",
	Short: "List available regions",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient()
		if err != nil {
			return err
		}
		return regionsRun(client)
	},
}

func regionsRun(client *lade.Client) error {
	regions, err := client.Region.List()
	if err != nil {
		return err
	}
	t := table.New("ID", "NAME", "COUNTRY")
	for _, region := range regions {
		t.AddRow(region.ID, region.Name, region.Country)
	}
	t.Print()
	return nil
}
