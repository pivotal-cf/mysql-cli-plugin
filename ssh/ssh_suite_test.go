package ssh_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSsh(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SSH Suite")
}
