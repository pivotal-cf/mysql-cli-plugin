package unpack_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUnpack(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unpack Suite")
}
