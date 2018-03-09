package main_test

import (
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestMysqlV2CliPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MysqlV2CliPlugin Suite")
}