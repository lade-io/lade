package cmd

import (
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/dustin/go-humanize"
	"github.com/lade-io/go-lade"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var disksCmd = &cobra.Command{
	Use:   "disks",
	Short: "Manage disks",
}

var disksAddCmd = func() *cobra.Command {
	var appName string
	opts := &lade.DiskCreateOpts{}
	cmd := &cobra.Command{
		Use:   "add <disk-name>",
		Short: "Add a disk to an app",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			if len(args) > 0 {
				opts.Name = args[0]
			}
			return disksAddRun(client, opts, appName)
		},
	}
	cmd.Flags().StringVarP(&appName, "app", "a", "", "App Name")
	cmd.Flags().StringVar(&opts.Path, "path", "", "Path")
	cmd.Flags().StringVarP(&opts.PlanID, "plan", "p", "", "Plan")
	return cmd
}()

var disksListCmd = func() *cobra.Command {
	var appName string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List disks of an app",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			return disksListRun(client, appName)
		},
	}
	cmd.Flags().StringVarP(&appName, "app", "a", "", "App Name")
	return cmd
}()

var disksPlansCmd = &cobra.Command{
	Use:   "plans",
	Short: "List available plans",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient()
		if err != nil {
			return err
		}
		return disksPlansRun(client)
	},
}

var disksRemoveCmd = func() *cobra.Command {
	var appName string
	cmd := &cobra.Command{
		Use:   "remove <disk-name>",
		Short: "Remove a disk from an app",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			var diskName string
			if len(args) > 0 {
				diskName = args[0]
			}
			return disksRemoveRun(client, appName, diskName)
		},
	}
	cmd.Flags().StringVarP(&appName, "app", "a", "", "App Name")
	return cmd
}()

var disksUpdateCmd = func() *cobra.Command {
	var appName string
	opts := &lade.DiskUpdateOpts{}
	cmd := &cobra.Command{
		Use:   "update <disk-name>",
		Short: "Update a disk of an app",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			var diskName string
			if len(args) > 0 {
				diskName = args[0]
			}
			return disksUpdateRun(client, opts, appName, diskName)
		},
	}
	cmd.Flags().StringVarP(&appName, "app", "a", "", "App Name")
	cmd.Flags().StringVarP(&opts.PlanID, "plan", "p", "", "Plan")
	return cmd
}()

func init() {
	disksCmd.AddCommand(disksAddCmd)
	disksCmd.AddCommand(disksListCmd)
	disksCmd.AddCommand(disksPlansCmd)
	disksCmd.AddCommand(disksRemoveCmd)
	disksCmd.AddCommand(disksUpdateCmd)
}

func disksAddRun(client *lade.Client, opts *lade.DiskCreateOpts, appName string) error {
	if err := askSelect("App Name:", getAppName, client, getAppOptions, &appName); err != nil {
		return err
	}
	if err := askInput("Disk Name:", appName, &opts.Name, validateDiskName(client, appName)); err != nil {
		return err
	}
	if err := askSelect("Plan:", getDiskPlan, client, getDiskPlanOptions(""), &opts.PlanID); err != nil {
		return err
	}
	if err := askInput("Path:", "/data", &opts.Path, validatePath); err != nil {
		return err
	}
	_, err := client.Disk.Create(appName, opts)
	return err
}

func disksListRun(client *lade.Client, appName string) error {
	if err := askSelect("App Name:", getAppName, client, getAppOptions, &appName); err != nil {
		return err
	}
	disks, err := client.Disk.List(appName)
	if err != nil {
		return err
	}
	t := table.New("NAME", "PLAN", "PATH", "CREATED")
	for _, disk := range disks {
		t.AddRow(disk.Name, disk.PlanID, disk.Path, humanize.Time(disk.CreatedAt))
	}
	t.Print()
	return nil
}

func disksPlansRun(client *lade.Client) error {
	plans, err := client.Plan.List("disk")
	if err != nil {
		return err
	}
	t := table.New("ID", "DISK", "PRICE HOURLY", "PRICE MONTHLY")
	for _, plan := range plans {
		priceHourly := printPrice(plan.PriceHourly, -1)
		priceMonthly := printPrice(plan.PriceMonthly, 2)
		t.AddRow(plan.ID, plan.Disk, priceHourly, priceMonthly)
	}
	t.Print()
	return nil
}

func disksRemoveRun(client *lade.Client, appName, diskName string) error {
	if err := askSelect("App Name:", getAppName, client, getAppOptions, &appName); err != nil {
		return err
	}
	if err := askSelect("Disk Name:", "", client, getDiskOptions(appName), &diskName); err != nil {
		return err
	}
	disk, err := client.Disk.Get(appName, diskName)
	if err != nil {
		return err
	}
	prompt := &survey.Confirm{
		Message: "Do you really want to delete " + disk.Name + "?",
	}
	confirm := false
	survey.AskOne(prompt, &confirm, nil)
	if confirm {
		err = client.Disk.Delete(disk)
	}
	return err
}

func disksUpdateRun(client *lade.Client, opts *lade.DiskUpdateOpts, appName, diskName string) error {
	if err := askSelect("App Name:", getAppName, client, getAppOptions, &appName); err != nil {
		return err
	}
	if err := askSelect("Disk Name:", "", client, getDiskOptions(appName), &diskName); err != nil {
		return err
	}
	disk, err := client.Disk.Get(appName, diskName)
	if err != nil {
		return err
	}
	if err = askSelect("Plan:", disk.PlanID, client, getDiskPlanOptions(disk.PlanID), &opts.PlanID); err != nil {
		return err
	}
	_, err = client.Disk.Update(strconv.Itoa(disk.AppID), strconv.Itoa(disk.ID), opts)
	return err
}
