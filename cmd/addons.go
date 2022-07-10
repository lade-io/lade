package cmd

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/dustin/go-humanize"
	"github.com/lade-io/go-lade"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var addonsCmd = &cobra.Command{
	Use:   "addons",
	Short: "Manage addons",
}

var addonsAttachCmd = func() *cobra.Command {
	var appName string
	var addonName string
	opts := &lade.AttachmentCreateOpts{}
	cmd := &cobra.Command{
		Use:   "attach <addon-name>",
		Short: "Attach addon to an app",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			if len(args) > 0 {
				addonName = args[0]
			}
			return addonsAttachRun(client, opts, appName, addonName)
		},
	}
	cmd.Flags().StringVarP(&appName, "app", "a", "", "App Name")
	cmd.Flags().StringVarP(&opts.Name, "env", "e", "", "Env Name")
	return cmd
}()

var addonsCreateCmd = func() *cobra.Command {
	opts := &lade.AddonCreateOpts{}
	cmd := &cobra.Command{
		Use:   "create <service-name>",
		Short: "Create an addon",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			if len(args) > 0 {
				opts.Service = args[0]
			}
			return addonsCreateRun(client, opts)
		},
	}
	cmd.Flags().StringVarP(&opts.Name, "name", "n", "", "Name")
	cmd.Flags().StringVarP(&opts.PlanID, "plan", "p", "", "Plan")
	cmd.Flags().StringVarP(&opts.RegionID, "region", "r", "", "Region")
	return cmd
}()

var addonsDetachCmd = func() *cobra.Command {
	var appName string
	var addonName string
	cmd := &cobra.Command{
		Use:   "detach <addon-name>",
		Short: "Detach addon from an app",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			if len(args) > 0 {
				addonName = args[0]
			}
			return addonsDetachRun(client, appName, addonName)
		},
	}
	cmd.Flags().StringVarP(&appName, "app", "a", "", "App Name")
	return cmd
}()

var addonsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List addons",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient()
		if err != nil {
			return err
		}
		return addonsListRun(client)
	},
}

var addonsLogsCmd = func() *cobra.Command {
	var addonName string
	var since time.Duration
	opts := &lade.LogStreamOpts{}
	cmd := &cobra.Command{
		Use:   "logs <addon-name>",
		Short: "Show logs from addon",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			if len(args) > 0 {
				addonName = args[0]
			}
			if since > 0 {
				opts.Since = time.Now().UTC().Add(-since)
			}
			return addonsLogsRun(client, opts, addonName)
		},
	}
	cmd.Flags().BoolVarP(&opts.Follow, "follow", "f", false, "Follow Log")
	cmd.Flags().DurationVarP(&since, "since", "s", 0, "Show Logs Since")
	cmd.Flags().IntVarP(&opts.Tail, "tail", "t", 0, "Number of Lines")
	return cmd
}()

var addonsRemoveCmd = &cobra.Command{
	Use:   "remove <addon-name>",
	Short: "Remove an addon",
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
		return addonsRemoveRun(client, name)
	},
}

var addonsServicesCmd = &cobra.Command{
	Use:   "services",
	Short: "List available services",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := getClient()
		if err != nil {
			return err
		}
		return addonsServicesRun(client)
	},
}

var addonsShowCmd = &cobra.Command{
	Use:   "show <addon-name>",
	Short: "Show addon info",
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
		return addonsShowRun(client, name)
	},
}

var addonsUpdateCmd = func() *cobra.Command {
	opts := &lade.AddonUpdateOpts{}
	cmd := &cobra.Command{
		Use:   "update <addon-name>",
		Short: "Update an addon",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			var name string
			if len(args) > 0 {
				name = args[0]
			}
			return addonsUpdateRun(client, opts, name)
		},
	}
	cmd.Flags().StringVarP(&opts.PlanID, "plan", "p", "", "Plan")
	return cmd
}()

func init() {
	addonsCmd.AddCommand(addonsAttachCmd)
	addonsCmd.AddCommand(addonsCreateCmd)
	addonsCmd.AddCommand(addonsDetachCmd)
	addonsCmd.AddCommand(addonsListCmd)
	addonsCmd.AddCommand(addonsLogsCmd)
	addonsCmd.AddCommand(addonsRemoveCmd)
	addonsCmd.AddCommand(addonsServicesCmd)
	addonsCmd.AddCommand(addonsShowCmd)
	addonsCmd.AddCommand(addonsUpdateCmd)
}

func addonsAttachRun(client *lade.Client, opts *lade.AttachmentCreateOpts, appName, addonName string) error {
	if err := askSelect("Addon Name:", "", client, getAddonOptions, &addonName); err != nil {
		return err
	}
	if err := askSelect("App Name:", getAppName, client, getAppOptions, &appName); err != nil {
		return err
	}
	addon, err := client.Addon.Get(addonName)
	if err != nil {
		return err
	}
	envName := strings.ToUpper(addon.Service.Name) + "_URL"
	if err = askInput("Env Name:", envName, &opts.Name, validateEnvName); err != nil {
		return err
	}
	_, err = client.Attachment.Create(appName, addonName, opts)
	return err
}

func addonsCreateRun(client *lade.Client, opts *lade.AddonCreateOpts) error {
	if err := askSelect("Service:", "", client, getServiceOptions, &opts.Service); err != nil {
		return err
	}
	if err := askInput("Addon Name:", opts.Service, &opts.Name, validateAddonName(client)); err != nil {
		return err
	}
	if err := askSelect("Plan:", getPlan, client, getPlanOptions(""), &opts.PlanID); err != nil {
		return err
	}
	if err := askSelect("Region:", getRegion, client, getRegionOptions, &opts.RegionID); err != nil {
		return err
	}
	if err := askSelect("Version:", "", client, getVersionOptions(opts.Service), &opts.Release); err != nil {
		return err
	}
	if err := askConfirm("Public:", true, &opts.Public); err != nil {
		return err
	}
	_, err := client.Addon.Create(opts)
	return err
}

func addonsDetachRun(client *lade.Client, appName, addonName string) error {
	if err := askSelect("Addon Name:", "", client, getAddonOptions, &addonName); err != nil {
		return err
	}
	if err := askSelect("App Name:", getAppName, client, getAppOptions, &appName); err != nil {
		return err
	}
	attachments, err := client.Attachment.List(appName, addonName)
	if err != nil {
		return err
	}
	names := []string{}
	for _, attachment := range attachments {
		names = append(names, attachment.Name)
	}
	sort.Strings(names)
	prompt := &survey.Confirm{
		Message: "Do you really want to detach " + strings.Join(names, ", ") + "?",
	}
	confirm := false
	survey.AskOne(prompt, &confirm, nil)
	if confirm {
		err = client.Attachment.Delete(appName, addonName)
	}
	return err
}

func addonsListRun(client *lade.Client) error {
	addons, err := client.Addon.List()
	if err != nil {
		return err
	}
	t := table.New("NAME", "SERVICE", "PLAN", "CREATED", "STATUS")
	for _, addon := range addons {
		t.AddRow(addon.Name, addon.Service.Name, addon.PlanID, humanize.Time(addon.CreatedAt), addon.Status)
	}
	t.Print()
	return nil
}

func addonsLogsRun(client *lade.Client, opts *lade.LogStreamOpts, addonName string) error {
	err := askSelect("Addon Name:", "", client, getAddonOptions, &addonName)
	if err != nil {
		return err
	}
	_, err = client.Addon.Get(addonName)
	if err != nil {
		return err
	}
	return client.Log.AddonStream(addonName, opts, printLog)
}

func addonsRemoveRun(client *lade.Client, name string) error {
	err := askSelect("Addon Name:", "", client, getAddonOptions, &name)
	if err != nil {
		return err
	}
	addon, err := client.Addon.Get(name)
	if err != nil {
		return err
	}
	prompt := &survey.Confirm{
		Message: "Do you really want to remove " + addon.Name + "?",
	}
	confirm := false
	survey.AskOne(prompt, &confirm, nil)
	if confirm {
		err = client.Addon.Delete(addon)
	}
	return err
}

func addonsServicesRun(client *lade.Client) error {
	services, err := client.Service.List()
	if err != nil {
		return err
	}
	t := table.New("NAME", "TITLE", "VERSIONS")
	for _, service := range services {
		serviceTags := strings.Join(service.Repo.Tags, ", ")
		t.AddRow(service.Name, service.Title, serviceTags)
	}
	t.Print()
	return nil
}

func addonsShowRun(client *lade.Client, name string) error {
	err := askSelect("Addon Name:", "", client, getAddonOptions, &name)
	if err != nil {
		return err
	}
	addon, err := client.Addon.Get(name)
	if err != nil {
		return err
	}
	t := table.New("Owner:", addon.Owner.Email)
	t.AddRow("Service:", addon.Service.Title)
	t.AddRow("Plan:", addon.PlanID)
	t.AddRow("Region:", addon.Region.Name)
	t.AddRow("Version:", addon.Release)
	t.AddRow("Public:", printBool(addon.Public))
	t.AddRow("Status:", addon.Status)
	t.AddRow("Addon URI:", getAddonURI(addon))
	t.Print()
	return nil
}

func addonsUpdateRun(client *lade.Client, opts *lade.AddonUpdateOpts, name string) error {
	if err := askSelect("Addon Name:", "", client, getAddonOptions, &name); err != nil {
		return err
	}
	addon, err := client.Addon.Get(name)
	if err != nil {
		return err
	}
	if err = askSelect("Plan:", addon.PlanID, client, getPlanOptions(addon.PlanID), &opts.PlanID); err != nil {
		return err
	}
	if err = askSelect("Version:", addon.Release, client, getVersionOptions(addon.Service.Name), &opts.Release); err != nil {
		return err
	}
	if err = askConfirm("Public:", addon.Public, &opts.Public); err != nil {
		return err
	}
	_, err = client.Addon.Update(strconv.Itoa(addon.ID), opts)
	return err
}

func getAddonURI(addon *lade.Addon) string {
	u := url.URL{
		Scheme:   addon.Service.Connector,
		User:     url.UserPassword(addon.Username, addon.Password),
		Host:     fmt.Sprintf("%s:%d", addon.Hostname, addon.Port),
		Path:     addon.Database,
		RawQuery: addon.Service.Query,
	}
	return u.String()
}
