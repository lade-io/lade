package cmd

import (
	"strconv"

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
		priceHourly := strconv.FormatFloat(plan.PriceHourly, 'f', -1, 64)
		priceMonthly := strconv.FormatFloat(plan.PriceMonthly, 'f', 2, 64)
		t.AddRow(plan.ID, plan.Ram, plan.Cpu, plan.Disk, priceHourly, priceMonthly)
	}
	t.Print()
	return nil
}
