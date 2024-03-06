package helpers

import (
	"os"
	"strings"

	. "github.com/onsi/gomega"
)

// CheckForRequiredEnvVars asserts that environment variables in envs must be
// set to a non-empty string. If any environment variable in the slice is not
// set, an error will be returned denoting which variable is unset.
func CheckForRequiredEnvVars(envs []string) {
	var missingEnvs []string

	for _, v := range envs {
		if os.Getenv(v) == "" {
			missingEnvs = append(missingEnvs, v)
		}
	}

	Expect(missingEnvs).To(BeEmpty(), "Missing environment variables: %s", strings.Join(missingEnvs, ", "))
}
