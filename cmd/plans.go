package cmd

import (
	"github.com/lade-io/go-lade"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var plansCmd = &cobra.Command{
	Use:   "plans",
	Short: "List available plans",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient()
		if err != nil {
			return err
		}
		return plansRun(client)
	},
}

func plansRun(client *lade.Client) error {
	plans, err := client.Plan.List()
	if err != nil {
		return err
	}
	t := table.New("ID", "MEMORY", "CPUS", "DISK", "PRICE HOURLY", "PRICE MONTHLY")
	for _, plan := range plans {
		priceHourly := printPrice(plan.PriceHourly, -1)
		priceMonthly := printPrice(plan.PriceMonthly, 2)
		t.AddRow(plan.ID, plan.Ram, plan.Cpu, plan.Disk, priceHourly, priceMonthly)
	}
	t.Print()
	return nil
}
