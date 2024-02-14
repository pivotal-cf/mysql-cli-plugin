package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"

	findbindings "github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/find-bindings"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/presentation"
)

//counterfeiter:generate -o fakes/fake_binding_finder.go . BindingFinder
type BindingFinder interface {
	FindBindings(serviceLabel string) ([]findbindings.Binding, error)
}

func FindBindings(args []string, bf BindingFinder) error {
	const (
		findUsage = `cf mysql-tools find-bindings [-h] <mysql-v1-service-name>`
	)

	var opts struct {
		Args struct {
			ServiceName string `positional-arg-name:"<mysql-v1-service-name>"`
		} `positional-args:"yes" required:"yes"`
	}

	parser := flags.NewParser(&opts, flags.None)
	parser.Name = "cf mysql-tools find-bindings"
	args, err := parser.ParseArgs(args)
	if err != nil || len(args) != 0 {
		msg := fmt.Sprintf("unexpected arguments: %s", strings.Join(args, " "))
		if err != nil {
			msg = err.Error()
		}
		return fmt.Errorf("Usage: %s\n\n%s", findUsage, msg)
	}

	serviceName := opts.Args.ServiceName
	binding, err := bf.FindBindings(serviceName)
	if err != nil {
		return err
	}

	presentation.Report(os.Stdout, binding)

	return nil
}
