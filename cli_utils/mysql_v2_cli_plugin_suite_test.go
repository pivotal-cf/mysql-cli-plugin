package cli_utils_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMysqlV2CliPlugin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MysqlV2CliPlugin Suite")
}
