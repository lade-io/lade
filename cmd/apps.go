package cmd

import (
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/dustin/go-humanize"
	"github.com/lade-io/go-lade"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "Manage apps",
}

var appsCreateCmd = func() *cobra.Command {
	opts := &lade.AppCreateOpts{}
	cmd := &cobra.Command{
		Use:   "create <app-name>",
		Short: "Create an app",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			if len(args) > 0 {
				opts.Name = args[0]
			}
			return appsCreateRun(client, opts)
		},
	}
	cmd.Flags().StringVarP(&opts.PlanID, "plan", "p", "", "Plan")
	cmd.Flags().StringVarP(&opts.RegionID, "region", "r", "", "Region")
	return cmd
}()

var appsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List apps",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient()
		if err != nil {
			return err
		}
		return appsListRun(client)
	},
}

var appsRemoveCmd = &cobra.Command{
	Use:   "remove <app-name>",
	Short: "Remove an app",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient()
		if err != nil {
			return err
		}
		var name string
		if len(args) > 0 {
			name = args[0]
		}
		return appsRemoveRun(client, name)
	},
}

var appsShowCmd = &cobra.Command{
	Use:   "show <app-name>",
	Short: "Show app info",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient()
		if err != nil {
			return err
		}
		var name string
		if len(args) > 0 {
			name = args[0]
		}
		return appsShowRun(client, name)
	},
}

func init() {
	appsCmd.AddCommand(appsCreateCmd)
	appsCmd.AddCommand(appsListCmd)
	appsCmd.AddCommand(appsRemoveCmd)
	appsCmd.AddCommand(appsShowCmd)
}

func appsCreateRun(client *lade.Client, opts *lade.AppCreateOpts) error {
	if err := askInput("App Name:", getAppName, &opts.Name, validateAppName(client)); err != nil {
		return err
	}
	if err := askSelect("Plan:", getPlan, client, getPlanOptions(""), &opts.PlanID); err != nil {
		return err
	}
	if err := askSelect("Region:", getRegion, client, getRegionOptions, &opts.RegionID); err != nil {
		return err
	}
	_, err := client.App.Create(opts)
	return err
}

func appsListRun(client *lade.Client) error {
	apps, err := client.App.List()
	if err != nil {
		return err
	}
	t := table.New("NAME", "PLAN", "CREATED", "STATUS")
	for _, app := range apps {
		t.AddRow(app.Name, app.PlanID, humanize.Time(app.CreatedAt), app.Status)
	}
	t.Print()
	return nil
}

func appsRemoveRun(client *lade.Client, name string) error {
	err := askSelect("App Name:", getAppName, client, getAppOptions, &name)
	if err != nil {
		return err
	}
	app, err := client.App.Get(name)
	if err != nil {
		return err
	}
	prompt := &survey.Confirm{
		Message: "Do you really want to remove " + app.Name + "?",
	}
	confirm := false
	survey.AskOne(prompt, &confirm, nil)
	if confirm {
		err = client.App.Delete(app)
	}
	return err
}

func appsShowRun(client *lade.Client, name string) error {
	err := askSelect("App Name:", getAppName, client, getAppOptions, &name)
	if err != nil {
		return err
	}
	app, err := client.App.Get(name)
	if err != nil {
		return err
	}
	processes, err := client.Process.List(name)
	if err != nil {
		return err
	}
	t := table.New("Owner:", app.Owner.Email)
	if len(processes) > 0 {
		t.AddRow("Processes:", processInfo(processes))
	}
	t.AddRow("Plan:", app.PlanID)
	t.AddRow("Region:", app.Region.Name)
	t.AddRow("Status:", app.Status)
	t.AddRow("Web URL:", app.Hostname)
	t.Print()
	return nil
}

func getAppName() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Base(cwd)
}

func getDiskPlan(client *lade.Client) string {
	plan, err := client.Plan.Default("disk")
	if err != nil {
		return ""
	}
	return plan.ID
}

func getPlan(client *lade.Client) string {
	plan, err := client.Plan.Default("")
	if err != nil {
		return ""
	}
	return plan.ID
}

func getRegion(client *lade.Client) string {
	user, err := client.User.Me()
	if err != nil {
		return ""
	}
	return user.RegionID
}
