package poll_cf

import (
	"time"

	"github.com/onsi/gomega/gexec"

	"github.com/pivotal-cf/mysql-v2-cli-plugin/test_helpers/poll_cf/commandreporter"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/test_helpers/poll_cf/commandstarter"
	"github.com/pivotal-cf/mysql-v2-cli-plugin/test_helpers/poll_cf/internal"
)

func PollCf(args ...string) *gexec.Session {
	cmdStarter := commandstarter.NewCommandStarter()
	return internal.Cf(cmdStarter, args...)
}

func ReportPoll(command string) {
	reporter := commandreporter.NewCommandReporter()
	reporter.Report(time.Now(), command)
}
