package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/lade-io/go-lade"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage app environment",
}

var envEditCmd = func() *cobra.Command {
	var appName string
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit env variables of an app",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			return envEditRun(client, appName)
		},
	}
	cmd.Flags().StringVarP(&appName, "app", "a", "", "App Name")
	return cmd
}()

var envListCmd = func() *cobra.Command {
	var appName string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List env variables of an app",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			return envListRun(client, appName)
		},
	}
	cmd.Flags().StringVarP(&appName, "app", "a", "", "App Name")
	return cmd
}()

var envSetCmd = func() *cobra.Command {
	var appName string
	cmd := &cobra.Command{
		Use:   "set <key>=<val>...",
		Short: "Set env variables of an app",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			opts, err := parseEnvSetArgs(args)
			if err != nil {
				return err
			}
			return envSetRun(client, appName, opts)
		},
	}
	cmd.Flags().StringVarP(&appName, "app", "a", "", "App Name")
	return cmd
}()

var envUnsetCmd = func() *cobra.Command {
	var appName string
	cmd := &cobra.Command{
		Use:   "unset <key>...",
		Short: "Unset env variables of an app",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getClient()
			if err != nil {
				return err
			}
			opts, err := parseEnvUnsetArgs(args)
			if err != nil {
				return err
			}
			return envUnsetRun(client, appName, opts)
		},
	}
	cmd.Flags().StringVarP(&appName, "app", "a", "", "App Name")
	return cmd
}()

func init() {
	envCmd.AddCommand(envEditCmd)
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envUnsetCmd)
}

func envEditRun(client *lade.Client, appName string) error {
	if err := askSelect("App Name:", getAppName, client, getAppOptions, &appName); err != nil {
		return err
	}
	envs, err := client.Env.List(appName)
	if err != nil {
		return err
	}
	editor := &survey.Editor{Message: "Env Variables:", HideDefault: true, AppendDefault: true}
	envMap := map[string]string{}
	for _, env := range envs {
		editor.Default += env.Name + "=" + env.Value + "\n"
		envMap[env.Name] = env.Value
	}
	var answer string
	survey.AskOne(editor, &answer, nil)
	answerMap, err := parseEnvAnswer(answer)
	if err != nil {
		return err
	}
	opts, names := mergeEnvMaps(envMap, answerMap)
	if len(opts.Envs) == 0 {
		return errors.New("No edits to env variables")
	}
	prompt := &survey.Confirm{
		Message: "Do you really want to edit " + strings.Join(names, ", ") + "?",
	}
	confirm := false
	survey.AskOne(prompt, &confirm, nil)
	if confirm {
		_, err = client.Env.Set(appName, opts)
	}
	return err
}

func envListRun(client *lade.Client, appName string) error {
	if err := askSelect("App Name:", getAppName, client, getAppOptions, &appName); err != nil {
		return err
	}
	envs, err := client.Env.List(appName)
	if err != nil {
		return err
	}
	for _, env := range envs {
		fmt.Println(env.Name + "=" + env.Value)
	}
	return nil
}

func envSetRun(client *lade.Client, appName string, opts *lade.EnvSetOpts) error {
	err := askSelect("App Name:", getAppName, client, getAppOptions, &appName)
	if err != nil {
		return err
	}
	_, err = client.Env.Set(appName, opts)
	return err
}

func envUnsetRun(client *lade.Client, appName string, opts *lade.EnvUnsetOpts) error {
	err := askSelect("App Name:", getAppName, client, getAppOptions, &appName)
	if err != nil {
		return err
	}
	err = askMultiSelect("Env Keys:", client, getKeyOptions(appName), &opts.Names, survey.Required)
	if err != nil {
		return err
	}
	envs, err := client.Env.List(appName)
	if err != nil {
		return err
	}
	envMap := map[string]bool{}
	for _, env := range envs {
		envMap[env.Name] = true
	}
	for _, name := range opts.Names {
		if !envMap[name] {
			return fmt.Errorf("Name not found %s", name)
		}
	}
	prompt := &survey.Confirm{
		Message: "Do you really want to unset " + strings.Join(opts.Names, ", ") + "?",
	}
	confirm := false
	survey.AskOne(prompt, &confirm, nil)
	if confirm {
		err = client.Env.Unset(appName, opts)
	}
	return err
}

func mergeEnvMaps(envMap, answerMap map[string]string) (*lade.EnvSetOpts, []string) {
	opts := new(lade.EnvSetOpts)
	names := []string{}
	for name, value := range answerMap {
		if envMap[name] == value {
			continue
		}
		opts.AddEnv(name, value)
		names = append(names, name)
	}
	for name := range envMap {
		if _, ok := answerMap[name]; ok {
			continue
		}
		opts.AddEnv(name, "")
		names = append(names, name)
	}
	return opts, names
}

func parseEnvAnswer(answer string) (map[string]string, error) {
	envMap := map[string]string{}
	scanner := bufio.NewScanner(strings.NewReader(answer))
	for scanner.Scan() {
		arg := scanner.Text()
		name, value, err := splitEnvArg(arg)
		if err != nil {
			return nil, err
		}
		if err = validateEnvName(name); err != nil {
			return nil, err
		}
		envMap[name] = value
	}
	return envMap, nil
}

func parseEnvSetArgs(args []string) (*lade.EnvSetOpts, error) {
	opts := new(lade.EnvSetOpts)
	for _, arg := range args {
		name, value, err := splitEnvArg(arg)
		if err != nil {
			return nil, err
		}
		if err = validateEnvName(name); err != nil {
			return nil, err
		}
		opts.AddEnv(name, value)
	}
	return opts, nil
}

func parseEnvUnsetArgs(args []string) (*lade.EnvUnsetOpts, error) {
	opts := new(lade.EnvUnsetOpts)
	for _, name := range args {
		if err := validateEnvName(name); err != nil {
			return nil, err
		}
		opts.Names = append(opts.Names, name)
	}
	return opts, nil
}
