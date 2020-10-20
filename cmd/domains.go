package cmd

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/lade-io/go-lade"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var domainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "Manage domains",
}

var domainsAddCmd = func() *cobra.Command {
	var appName string
	opts := &lade.DomainCreateOpts{}
	cmd := &cobra.Command{
		Use:   "add <domain-name>",
		Short: "Add domain to an app",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			if len(args) > 0 {
				opts.Hostname = args[0]
			}
			return domainsAddRun(client, opts, appName)
		},
	}
	cmd.Flags().StringVarP(&appName, "app", "a", "", "App Name")
	return cmd
}()

var domainsListCmd = func() *cobra.Command {
	var appName string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List domains of an app",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			return domainsListRun(client, appName)
		},
	}
	cmd.Flags().StringVarP(&appName, "app", "a", "", "App Name")
	return cmd
}()

var domainsRemoveCmd = func() *cobra.Command {
	var appName string
	cmd := &cobra.Command{
		Use:   "remove <domain-name>",
		Short: "Remove domain from an app",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			var hostname string
			if len(args) > 0 {
				hostname = args[0]
			}
			return domainsRemoveRun(client, hostname, appName)
		},
	}
	cmd.Flags().StringVarP(&appName, "app", "a", "", "App Name")
	return cmd
}()

func init() {
	domainsCmd.AddCommand(domainsAddCmd)
	domainsCmd.AddCommand(domainsListCmd)
	domainsCmd.AddCommand(domainsRemoveCmd)
}

func domainsAddRun(client *lade.Client, opts *lade.DomainCreateOpts, appName string) error {
	if err := askSelect("App Name:", getAppName, client, getAppOptions, &appName); err != nil {
		return err
	}
	if err := askInput("Domain Name:", nil, &opts.Hostname, validateDomainName(client, appName)); err != nil {
		return err
	}
	_, err := client.Domain.Create(appName, opts)
	return err
}

func domainsListRun(client *lade.Client, appName string) error {
	if err := askSelect("App Name:", getAppName, client, getAppOptions, &appName); err != nil {
		return err
	}
	app, err := client.App.Get(appName)
	if err != nil {
		return err
	}
	domains, err := client.Domain.List(appName)
	if err != nil {
		return err
	}
	t := table.New("NAME", "TYPE", "TARGET")
	for _, domain := range domains {
		t.AddRow(domain.Hostname, "CNAME", app.Hostname)
	}
	t.Print()
	return nil
}

func domainsRemoveRun(client *lade.Client, hostname, appName string) error {
	if err := askSelect("App Name:", getAppName, client, getAppOptions, &appName); err != nil {
		return err
	}
	if err := askSelect("Domain Name:", "", client, getDomainOptions(appName), &hostname); err != nil {
		return err
	}
	domain, err := client.Domain.Get(appName, hostname)
	if err != nil {
		return err
	}
	prompt := &survey.Confirm{
		Message: "Do you really want to delete " + domain.Hostname + "?",
	}
	delete := false
	survey.AskOne(prompt, &delete, nil)
	if delete {
		err = client.Domain.Delete(domain)
	}
	return err
}
