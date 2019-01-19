package command

import (
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/migrate"
	"log"
)

type VersionExecutor struct {
	version, gitSHA string
}

func (v *VersionExecutor) Execute(client migrate.Client, args []string) error {
	if v.version == "" {
		v.version = version
	}

	if v.gitSHA == "" {
		v.gitSHA = gitSHA
	}
	log.Printf("%s (%s)\n", v.version, v.gitSHA)
	return nil
}
