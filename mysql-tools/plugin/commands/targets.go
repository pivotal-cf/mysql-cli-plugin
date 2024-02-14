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

//counterfeiter:generate -o fakes/fake_multisite.go . MultiSite
type MultiSite interface {
	ListConfigs() ([]*multisite.ConfigCoreSubset, error)
	SaveConfig(cfConfig, targetName string) (*multisite.ConfigCoreSubset, error)
	RemoveConfig(targetName string) error
	SetupReplication(primaryFoundation, primaryInstance, secondaryFoundation, secondaryInstance string) error
}

func ListTargets(ms MultiSite) error {

	configs, err := ms.ListConfigs()
	fmt.Println("Targets:")
	for _, config := range configs {
		fmt.Printf("Target: %s\n", config.Name)
		fmt.Printf("   API: %s\n", config.Target)
		fmt.Printf("   Org: %s\n", config.OrganizationFields.Name)
		fmt.Printf(" Space: %s\n", config.SpaceFields.Name)
		fmt.Println()
	}
	if err != nil {
		return fmt.Errorf("error listing multisite targets: %v", err)
	}
	return nil
}

func SaveTarget(args []string, ms MultiSite) error {
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

	result, err := ms.SaveConfig(cfConfigFilename, targetConfigName)
	if err != nil {
		return fmt.Errorf("error saving target %s: %w", targetConfigName, err)
	}

	fmt.Println("Success")
	fmt.Printf("Target: %s\n", targetConfigName)
	fmt.Printf("   API: %s\n", result.Target)
	fmt.Printf("   Org: %s\n", result.OrganizationFields.Name)
	fmt.Printf(" Space: %s\n", result.SpaceFields.Name)

	return nil
}

func RemoveTarget(args []string, ms MultiSite) error {
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
	err = ms.RemoveConfig(removeConfigName)

	if err != nil {
		return fmt.Errorf("error trying to remove the target config: %v", err)
	}

	return nil
}
