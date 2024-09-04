package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/core"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/iancoleman/orderedmap"
	"github.com/lade-io/go-lade"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/mgutz/ansi"
	"github.com/olekukonko/ts"
)

var (
	validEnvName = regexp.MustCompile(`^[A-Z0-9-_]+$`)
	validName    = regexp.MustCompile(`^[a-z][a-z0-9-_]*$`)
	validPath    = regexp.MustCompile(`^(/[a-zA-Z0-9-_]+)+$`)
)

type optionsFunc func(*lade.Client) (*orderedmap.OrderedMap, error)

func askError(err error) error {
	if err == terminal.InterruptErr {
		os.Exit(1)
	}
	return err
}

func askConfirm(msg string, choice bool, result interface{}) error {
	prompt := &survey.Confirm{Message: msg, Default: choice}
	return askError(survey.AskOne(prompt, result, nil))
}

func askInput(msg string, value, result interface{}, validator survey.Validator) error {
	if !isZero(result) {
		if validator != nil {
			return validator(toString(result))
		}
		return nil
	}
	var choice string
	switch v := value.(type) {
	case func() string:
		choice = v()
	case string:
		choice = v
	case int:
		choice = strconv.Itoa(v)
	}
	prompt := &survey.Input{Message: msg, Default: choice}
	return askError(survey.AskOne(prompt, result, survey.WithValidator(validator)))
}

func askSelect(msg string, value interface{}, client *lade.Client, fn optionsFunc, result interface{}) error {
	if !isZero(result) {
		return nil
	}
	var wg sync.WaitGroup
	var choice string
	switch v := value.(type) {
	case func(client *lade.Client) string:
		wg.Add(1)
		go func() {
			defer wg.Done()
			choice = v(client)
		}()
	case func() string:
		choice = v()
	case string:
		choice = v
	}
	options, err := fn(client)
	if err != nil {
		return err
	}
	wg.Wait()
	prompt := &survey.Select{Message: msg, Options: options.Keys(), PageSize: getPageSize()}
	if choice != "" {
		key, ok := options.GetKey(choice)
		if ok {
			prompt.Default = key
		}
	}
	var answer string
	err = survey.AskOne(prompt, &answer, nil)
	if err != nil {
		return askError(err)
	}
	value, ok := options.Get(answer)
	if ok {
		core.WriteAnswer(result, "", value)
	}
	return nil
}

func askMultiSelect(msg string, client *lade.Client, fn optionsFunc, result interface{}, validator survey.Validator) error {
	if !isZero(result) {
		return nil
	}
	options, err := fn(client)
	if err != nil {
		return err
	}
	var answers []string
	prompt := &survey.MultiSelect{Message: msg, Options: options.Keys(), PageSize: getPageSize()}
	err = survey.AskOne(prompt, &answers, survey.WithValidator(validator))
	if err != nil {
		return askError(err)
	}
	var values []string
	for _, answer := range answers {
		value, ok := options.Get(answer)
		if ok {
			values = append(values, value.(string))
		}
	}
	core.WriteAnswer(result, "", values)
	return nil
}

func getAppOptions(client *lade.Client) (*orderedmap.OrderedMap, error) {
	apps, err := client.App.List()
	if err != nil {
		return nil, err
	}
	if len(apps) == 0 {
		return nil, errors.New("You have not created any apps")
	}
	options := orderedmap.New()
	for _, app := range apps {
		options.Set(app.Name, app.Name)
	}
	options.SortKeys(sort.Strings)
	return options, nil
}

func getAddonOptions(client *lade.Client) (*orderedmap.OrderedMap, error) {
	addons, err := client.Addon.List()
	if err != nil {
		return nil, err
	}
	if len(addons) == 0 {
		return nil, errors.New("You have not created any addons")
	}
	options := orderedmap.New()
	for _, addon := range addons {
		options.Set(addon.Name, addon.Name)
	}
	options.SortKeys(sort.Strings)
	return options, nil
}

func getDiskOptions(appName string) optionsFunc {
	return func(client *lade.Client) (*orderedmap.OrderedMap, error) {
		disks, err := client.Disk.List(appName)
		if err != nil {
			return nil, err
		}
		if len(disks) == 0 {
			return nil, errors.New("There are no disks available")
		}
		options := orderedmap.New()
		for _, disk := range disks {
			options.Set(disk.Name, disk.Name)
		}
		options.SortKeys(sort.Strings)
		return options, nil
	}
}

func getDiskPlanOptions(id string) optionsFunc {
	return func(client *lade.Client) (*orderedmap.OrderedMap, error) {
		plans, err := client.Plan.User(id, "disk")
		if err != nil {
			return nil, err
		}
		if len(plans) == 0 {
			return nil, errors.New("There are no plans available")
		}
		options := orderedmap.New()
		for _, plan := range plans {
			options.Set(plan.ID, plan.ID)
		}
		return options, nil
	}
}

func getDomainOptions(appName string) optionsFunc {
	return func(client *lade.Client) (*orderedmap.OrderedMap, error) {
		domains, err := client.Domain.List(appName)
		if err != nil {
			return nil, err
		}
		if len(domains) == 0 {
			return nil, errors.New("There are no domains available")
		}
		options := orderedmap.New()
		for _, domain := range domains {
			options.Set(domain.Hostname, domain.Hostname)
		}
		return options, nil
	}
}

func getKeyOptions(appName string) optionsFunc {
	return func(client *lade.Client) (*orderedmap.OrderedMap, error) {
		envs, err := client.Env.List(appName)
		if err != nil {
			return nil, err
		}
		if len(envs) == 0 {
			return nil, errors.New("There are no keys available")
		}
		options := orderedmap.New()
		for _, env := range envs {
			options.Set(env.Name, env.Name)
		}
		return options, nil
	}
}

func getPlanOptions(id string) optionsFunc {
	return func(client *lade.Client) (*orderedmap.OrderedMap, error) {
		plans, err := client.Plan.User(id, "")
		if err != nil {
			return nil, err
		}
		if len(plans) == 0 {
			return nil, errors.New("There are no plans available")
		}
		options := orderedmap.New()
		for _, plan := range plans {
			options.Set(plan.ID, plan.ID)
		}
		return options, nil
	}
}

func getProcessOptions(appName string) optionsFunc {
	return func(client *lade.Client) (*orderedmap.OrderedMap, error) {
		processes, err := client.Process.List(appName)
		if err != nil {
			return nil, err
		}
		if len(processes) == 0 {
			return nil, errors.New("There are no processes available")
		}
		options := orderedmap.New()
		for _, process := range processes {
			options.Set(process.Type, process.Type)
		}
		return options, nil
	}
}

func getRegionOptions(client *lade.Client) (*orderedmap.OrderedMap, error) {
	regions, err := client.Region.List()
	if err != nil {
		return nil, err
	}
	if len(regions) == 0 {
		return nil, errors.New("There are no regions available")
	}
	options := orderedmap.New()
	for _, region := range regions {
		options.Set(region.Name, region.ID)
	}
	return options, nil
}

func getServiceOptions(client *lade.Client) (*orderedmap.OrderedMap, error) {
	services, err := client.Service.List()
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, errors.New("There are no services available")
	}
	options := orderedmap.New()
	for _, service := range services {
		options.Set(service.Title, service.Name)
	}
	return options, nil
}

func getVersionOptions(serviceName string) optionsFunc {
	return func(client *lade.Client) (*orderedmap.OrderedMap, error) {
		service, err := client.Service.Get(serviceName)
		if err != nil {
			return nil, err
		}
		if service.Repo == nil || len(service.Repo.Tags) == 0 {
			return nil, errors.New("There are no versions available")
		}
		options := orderedmap.New()
		for _, version := range service.Repo.Tags {
			options.Set(version, version)
		}
		return options, nil
	}
}

func getPageSize() int {
	size, err := ts.GetSize()
	if err != nil {
		return 20
	}
	return size.Row() * 4 / 5
}

func getValueOf(val interface{}) reflect.Value {
	v := reflect.ValueOf(val)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}

func isZero(val interface{}) bool {
	v := getValueOf(val)
	if v.Kind() == reflect.Slice {
		return v.IsNil()
	}
	z := reflect.Zero(v.Type())
	return v.Interface() == z.Interface()
}

func toString(val interface{}) interface{} {
	return fmt.Sprint(getValueOf(val).Interface())
}

func printBool(val bool) string {
	if val {
		return "Yes"
	}
	return "No"
}

func printDeployLog(cancel context.CancelFunc, entry *lade.LogEntry) {
	if entry.Source == "stderr" {
		if entry.Line == io.EOF.Error() {
			cancel()
			return
		}
		log.Fatal(entry.Line)
	}
	fmt.Println(entry.Line)
}

func printLog(cancel context.CancelFunc, entry *lade.LogEntry) {
	fmt.Println(entry.Line)
}

func printNameLog(width int) lade.LogHandler {
	out := colorable.NewColorableStdout()
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		ansi.DisableColors(true)
	}
	colors := []string{"yellow", "green", "cyan", "blue", "magenta", "red"}
	names := map[string]string{}
	return func(cancel context.CancelFunc, entry *lade.LogEntry) {
		name, ok := names[entry.Name]
		if !ok {
			color := colors[len(names)%len(colors)]
			name = ansi.Color(fmt.Sprintf("%-*s | ", width+2, entry.Name), color)
			names[entry.Name] = name
		}
		fmt.Fprintln(out, name+entry.Line)
	}
}

func printPrice(val float64, prec int) string {
	if val == 0 {
		return "Free"
	}
	return strconv.FormatFloat(val, 'f', prec, 64)
}

func processInfo(processes []*lade.Process) string {
	results := []string{}
	for _, process := range processes {
		if process.Replicas == 0 {
			continue
		}
		result := fmt.Sprintf("%s: %d/%d", process.Type, process.Count, process.Replicas)
		results = append(results, result)
	}
	return strings.Join(results, ", ")
}

func splitEnvArg(arg string) (string, string, error) {
	args := strings.SplitN(arg, "=", 2)
	if len(args) < 2 || args[0] == "" || args[1] == "" {
		return "", "", errors.New("Argument must be declared <key>=<val>")
	}
	return args[0], args[1], nil
}

func splitProcessArg(arg string) (string, string, string, error) {
	args := strings.SplitN(arg, "=", 2)
	if len(args) < 2 || args[0] == "" || args[1] == "" {
		return "", "", "", errors.New("Argument must be declared <type>=<count>[:<plan>]")
	}
	spec := strings.SplitN(args[1], ":", 2)
	if len(spec) < 2 {
		return args[0], args[1], "", nil
	}
	return args[0], spec[0], spec[1], nil
}

func validateAddonName(client *lade.Client) survey.Validator {
	return survey.ComposeValidators(validateName, validateUniqueName(func(name string) error {
		return client.Addon.Head(name)
	}))
}

func validateAppName(client *lade.Client) survey.Validator {
	return survey.ComposeValidators(validateName, validateUniqueName(func(name string) error {
		return client.App.Head(name)
	}))
}

func validateDiskName(client *lade.Client, appName string) survey.Validator {
	return survey.ComposeValidators(survey.Required, validateUniqueName(func(name string) error {
		return client.Disk.Head(appName, name)
	}))
}

func validateDomainName(client *lade.Client, appName string) survey.Validator {
	return survey.ComposeValidators(survey.Required, validateUniqueName(func(name string) error {
		return client.Domain.Head(appName, name)
	}))
}

func validateCount(min, max int) func(interface{}) error {
	return func(val interface{}) error {
		num, err := strconv.Atoi(val.(string))
		if err != nil {
			return errors.New("Count must be a number")
		}
		if num < min {
			return fmt.Errorf("Count must be at least %d", min)
		}
		if num > max {
			return fmt.Errorf("Count must be at most %d", max)
		}
		return nil
	}
}

func validateEnvName(val interface{}) error {
	if !validEnvName.MatchString(val.(string)) {
		return errors.New("Name must only contain A-Z, 0-9, dash (-) or underscore (_)")
	}
	return nil
}

func validateName(val interface{}) error {
	if !validName.MatchString(val.(string)) {
		return errors.New("Name must start with a-z followed by a-z, 0-9, dash (-) or underscore (_)")
	}
	return nil
}

func validatePath(val interface{}) error {
	if !validPath.MatchString(val.(string)) {
		return errors.New("Path must be valid absolute directory")
	}
	return nil
}

func validateUniqueName(fn func(string) error) survey.Validator {
	return func(val interface{}) error {
		err := fn(val.(string))
		if err == nil {
			return errors.New("Name is already taken")
		}
		var e *lade.APIError
		if !errors.As(err, &e) || e.Status != http.StatusNotFound {
			return err
		}
		return nil
	}
}
