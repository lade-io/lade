package cmd

import (
	"fmt"
	"strconv"

	"github.com/lade-io/go-lade"
	"github.com/spf13/cobra"
)

var scaleCmd = func() *cobra.Command {
	var appName string
	cmd := &cobra.Command{
		Use:   "scale <type>=<count>[:<plan>]...",
		Short: "Scale an app",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			max, err := client.Quota.Max()
			if err != nil {
				return err
			}
			opts, err := parseProcessUpdateArgs(args, max.Quota)
			if err != nil {
				return err
			}
			return scaleRun(client, appName, max.Quota, opts)
		},
	}
	cmd.Flags().StringVarP(&appName, "app", "a", "", "App Name")
	return cmd
}()

func scaleRun(client *lade.Client, appName string, maxQuota int, opts *lade.ProcessUpdateOpts) error {
	err := askSelect("App Name:", getAppName, client, getAppOptions, &appName)
	if err != nil {
		return err
	}
	processes, err := client.Process.List(appName)
	if err != nil {
		return err
	}
	procMap := map[string]*lade.Process{}
	for _, process := range processes {
		procMap[process.Type] = process
	}
	for _, process := range opts.Processes {
		if _, ok := procMap[process.Type]; !ok {
			return fmt.Errorf("Process type not found %s", process.Type)
		}
	}
	if len(opts.Processes) == 0 {
		opt := &lade.Process{}
		err = askSelect("Process Type:", "", client, getProcessOptions(appName), &opt.Type)
		if err != nil {
			return err
		}
		process := procMap[opt.Type]
		err = askInput("Count:", process.Replicas, &opt.Replicas, validateCount(0, maxQuota))
		if err != nil {
			return err
		}
		err = askSelect("Plan:", process.PlanID, client, getPlanOptions, &opt.PlanID)
		if err != nil {
			return err
		}
		opts.Processes = append(opts.Processes, opt)
	}
	_, err = client.Process.Update(appName, opts)
	return err
}

func parseProcessUpdateArgs(args []string, maxQuota int) (*lade.ProcessUpdateOpts, error) {
	opts := new(lade.ProcessUpdateOpts)
	for _, arg := range args {
		ptype, count, planID, err := splitProcessArg(arg)
		if err != nil {
			return nil, err
		}
		if err = validateCount(0, maxQuota)(count); err != nil {
			return nil, err
		}
		replicas, _ := strconv.Atoi(count)
		opts.AddProcess(ptype, planID, replicas)
	}
	return opts, nil
}
