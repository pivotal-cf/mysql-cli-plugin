package internal

import (
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/mysql-cli-plugin/test_helpers/poll_cf/commandreporter"
)

func Cf(cmdStarter Starter, args ...string) *gexec.Session {
	reporter := commandreporter.NewCommandReporter()
	request, err := cmdStarter.Start(reporter, "cf", args...)
	if err != nil {
		panic(err)
	}
	return request
}
