package commandstarter

import (
	"os/exec"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf/mysql-cli-plugin/test_helpers/poll_cf/internal"
)

type CommandStarter struct {
}

func NewCommandStarter() *CommandStarter {
	return &CommandStarter{}
}

func (r *CommandStarter) Start(reporter internal.Reporter, executable string, args ...string) (*gexec.Session, error) {
	cmd := exec.Command(executable, args...)
	reporter.Polling()
	return gexec.Start(cmd, nil, ginkgo.GinkgoWriter)
}