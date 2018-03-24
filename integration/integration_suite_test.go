package integration

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var (
	donorServiceName     = "donor_service"
	donorPassword        = "donor-password"
	donorPort            = "3307"
	recipientServiceName = "recipient_service"
	recipientPassword    = "recipient-password"
	recipientPort        = "3308"
)
