package commands

import (
	"fmt"
	"os"
	"strings"

	"code.cloudfoundry.org/cli/cf/configuration/confighelpers"
	"github.com/jessevdk/go-flags"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite"
)

const (
	SaveTargetUsage   = `cf mysql-tools save-target <target-name>`
	RemoveTargetUsage = `Usage: cf mysql-tools remove-target <target-name>`
)

//counterfeiter:generate -o fakes/fake_multisite_config.go . MultisiteConfig
type MultisiteConfig interface {
	ListConfigs() ([]multisite.Target, error)
	SaveConfig(configPath, targetName string) (multisite.Target, error)
	RemoveConfig(targetName string) error
	ConfigDir(targetName string) (path string)
}

func ListTargets(cfg MultisiteConfig) error {

	configs, err := cfg.ListConfigs()

	if len(configs) == 0 {
		fmt.Println("No saved targets")
		return nil
	}

	fmt.Println("Targets:")
	for _, config := range configs {
		fmt.Printf(" Name: %s\n", config.Name)
		fmt.Printf("  API: %s\n", config.API)
		fmt.Printf("  Org: %s\n", config.Organization)
		fmt.Printf("Space: %s\n", config.Space)
		fmt.Println()
	}
	if err != nil {
		return fmt.Errorf("error listing multisite targets: %v", err)
	}
	return nil
}

func SaveTarget(args []string, cfg MultisiteConfig) error {
	var opts struct {
		Args struct {
			TargetName string `positional-arg-name:"<target-name>"`
		} `positional-args:"yes" required:"yes"`
	}
	parser := flags.NewParser(&opts, flags.None)
	parser.Name = "cf mysql-tools save-target"
	args, err := parser.ParseArgs(args)
	if err != nil || len(args) != 0 {
		msg := fmt.Sprintf("unexpected arguments: %s", strings.Join(args, " "))
		if err != nil {
			msg = err.Error()
		}
		return fmt.Errorf("Usage: %s\n\n%s", SaveTargetUsage, msg)
	}

	targetConfigName := opts.Args.TargetName

	cfConfigFilename, err := confighelpers.DefaultFilePath()
	if err != nil {
		return fmt.Errorf("error saving target %s: %w", targetConfigName, err)
	}
	if _, err := os.Stat(cfConfigFilename); os.IsNotExist(err) {
		return fmt.Errorf("error saving target %s: %w", targetConfigName, err)
	}

	result, err := cfg.SaveConfig(cfConfigFilename, targetConfigName)
	if err != nil {
		return fmt.Errorf("error saving target %s: %w", targetConfigName, err)
	}

	fmt.Println("Success")
	fmt.Printf(" Name: %s\n", targetConfigName)
	fmt.Printf("  API: %s\n", result.API)
	fmt.Printf("  Org: %s\n", result.Organization)
	fmt.Printf("Space: %s\n", result.Space)

	return nil
}

func RemoveTarget(args []string, cfg MultisiteConfig) error {
	var opts struct {
		Args struct {
			TargetName string `positional-arg-name:"<target-name>"`
		} `positional-args:"yes" required:"yes"`
	}

	parser := flags.NewParser(&opts, flags.None)
	parser.Name = "cf mysql-tools remove-target"
	args, err := parser.ParseArgs(args)

	if err != nil || len(args) != 0 {
		msg := fmt.Sprintf("unexpected arguments: %s", strings.Join(args, " "))
		if err != nil {
			msg = err.Error()
		}
		return fmt.Errorf("Usage: %s\n\n%s", RemoveTargetUsage, msg)
	}

	removeConfigName := opts.Args.TargetName
	err = cfg.RemoveConfig(removeConfigName)

	if err != nil {
		return fmt.Errorf("error trying to remove the target config: %v", err)
	}

	return nil
}
