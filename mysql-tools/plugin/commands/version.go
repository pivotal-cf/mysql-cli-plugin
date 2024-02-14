package commands

import (
	"fmt"

	"github.com/pivotal-cf/mysql-cli-plugin/version"
)

func Version() error {
	fmt.Printf("%s (%s)\n", version.Version, version.GitSHA)
	return nil
}
