package multisite_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite"
)

var _ = Describe("Config", func() {
	var subject multisite.Config

	BeforeEach(func() {
		t, err := os.MkdirTemp("", "multisite_config_test_")
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() {
			Expect(os.RemoveAll(t)).To(Succeed())
		})

		subject = multisite.Config{Dir: t}

		format.TruncatedDiff = false
	})

	Context("NewConfig", func() {
		When("CF_HOME and CF_PLUGIN_HOME are unset", func() {
			BeforeEach(func() {
				Expect(os.Unsetenv("CF_HOME")).To(Succeed())
				Expect(os.Unsetenv("CF_PLUGIN_HOME")).To(Succeed())
			})

			It("uses the expected cf cli config directory", func() {
				cfg := multisite.NewConfig()

				homedir, err := os.UserHomeDir()
				Expect(err).NotTo(HaveOccurred())

				Expect(cfg.Dir).To(Equal(filepath.Join(homedir, ".cf", ".mysql-tools")))
			})
		})

		When("CF_HOME is set, but CF_PLUGIN_HOME is not", func() {
			var cfHome string
			BeforeEach(func() {
				var err error
				cfHome, err = os.MkdirTemp("", "cf_home_")
				Expect(err).NotTo(HaveOccurred())

				Expect(os.Setenv("CF_HOME", cfHome)).To(Succeed())
				Expect(os.Unsetenv("CF_PLUGIN_HOME")).To(Succeed())
				DeferCleanup(func() {
					Expect(os.Unsetenv("CF_HOME")).To(Succeed())
				})
			})

			It("uses the expected cf cli config directory", func() {
				cfg := multisite.NewConfig()
				Expect(cfg.Dir).To(Equal(filepath.Join(cfHome, ".cf", ".mysql-tools")))
			})

			When("CF_PLUGIN_HOME is also set", func() {
				var cfPluginHome string
				BeforeEach(func() {
					var err error
					cfPluginHome, err = os.MkdirTemp("", "cf_plugin_home_")
					Expect(err).NotTo(HaveOccurred())

					Expect(os.Setenv("CF_PLUGIN_HOME", cfPluginHome)).To(Succeed())
					DeferCleanup(func() {
						Expect(os.Unsetenv("CF_PLUGIN_HOME")).To(Succeed())
					})
				})

				It("uses the expected cf cli config directory", func() {
					cfg := multisite.NewConfig()
					Expect(cfg.Dir).To(Equal(filepath.Join(cfPluginHome, ".cf", ".mysql-tools")))
				})
			})
		})

		When("CF_PLUGIN_HOME is set without CF_HOME", func() {
			var cfPluginHome string
			BeforeEach(func() {
				var err error
				cfPluginHome, err = os.MkdirTemp("", "cf_plugin_home_")
				Expect(err).NotTo(HaveOccurred())

				Expect(os.Setenv("CF_PLUGIN_HOME", cfPluginHome)).To(Succeed())
				Expect(os.Unsetenv("CF_HOME")).To(Succeed())
				DeferCleanup(func() {
					Expect(os.Unsetenv("CF_PLUGIN_HOME")).To(Succeed())
				})
			})

			It("uses the expected cf cli config directory", func() {
				cfg := multisite.NewConfig()
				Expect(cfg.Dir).To(Equal(filepath.Join(cfPluginHome, ".cf", ".mysql-tools")))
			})
		})
	})

	Context("ConfigDir", func() {
		It("provides the path to the .cf home dir for a given target", func() {
			path := subject.ConfigDir("some-target")

			Expect(path).To(Equal(subject.Dir + "/some-target"))

			otherPath := subject.ConfigDir("some-other-target")
			Expect(otherPath).To(Equal(subject.Dir + "/some-other-target"))
		})
	})

	Context("ListConfigs", func() {
		It("lists saved multi-site foundation configs", func() {
			_, _ = subject.SaveConfig("fixtures/sample-config.json", "testFoundation1")
			_, _ = subject.SaveConfig("fixtures/sample-alt-config.json", "testFoundation2")

			targets, err := subject.ListConfigs()
			Expect(err).NotTo(HaveOccurred())
			Expect(targets).To(ConsistOf([]multisite.Target{
				{
					Name:         "testFoundation1",
					Organization: "sample-org",
					Space:        "sample-space",
					API:          "https://api.sample.foundation.tld",
				},
				{
					Name:         "testFoundation2",
					Organization: "alt-sample-org",
					Space:        "alt-sample-space",
					API:          "https://api.sample.alt.foundation.tld",
				},
			}))
		})

		When("the cf root directory does not exist", func() {
			It("does not return an error", func() {
				_, err := subject.ListConfigs()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("one of the config file paths does not exist", func() {
			copyFile := func(srcPath, dstPath string) error {
				contents, err := os.ReadFile(srcPath)
				if err != nil {
					return err
				}

				return os.WriteFile(dstPath, contents, 0644)
			}

			BeforeEach(func() {
				Expect(os.MkdirAll(subject.Dir+"/testfoundation1/.cf", 0700)).To(Succeed())
				testConfig := filepath.Join(subject.Dir, "testfoundation1", ".cf", "config.json")

				Expect(copyFile("fixtures/sample-config.json", testConfig)).To(Succeed())

				Expect(os.MkdirAll(subject.Dir+"/testfoundation2/.cf", 0700)).To(Succeed())
			})

			It("returns configs that were found", func() {
				configs, err := subject.ListConfigs()
				Expect(err).NotTo(HaveOccurred())
				Expect(configs).To(ContainElement(
					multisite.Target{
						Name:         "testfoundation1",
						Organization: "sample-org",
						Space:        "sample-space",
						API:          "https://api.sample.foundation.tld",
					},
				))
			})
		})
	})

	Context("SaveConfig", func() {
		When("saving a valid config", func() {
			It("returns a JSON subset of relevant config info", func() {
				savedTarget, err := subject.SaveConfig("fixtures/sample-config.json", "testFoundation")
				Expect(err).NotTo(HaveOccurred())
				Expect(savedTarget).NotTo(BeNil())
				Expect(savedTarget.Organization).To(Equal("sample-org"))
				Expect(savedTarget.Space).To(Equal("sample-space"))
				Expect(savedTarget.API).To(Equal("https://api.sample.foundation.tld"))
			})
		})

		When("saving two configs to the same name", func() {
			It("the second save successfully overwrites the first", func() {
				_, err := subject.SaveConfig("fixtures/sample-config.json", "reusedName")
				Expect(err).NotTo(HaveOccurred())

				savedTarget, err := subject.SaveConfig("fixtures/sample-alt-config.json", "reusedName")
				Expect(err).NotTo(HaveOccurred())
				Expect(savedTarget).NotTo(BeNil())

				Expect(savedTarget.Organization).To(Equal("alt-sample-org"))
				Expect(savedTarget.Space).To(Equal("alt-sample-space"))
				Expect(savedTarget.API).To(Equal("https://api.sample.alt.foundation.tld"))

				originalContents, err := os.ReadFile("fixtures/sample-alt-config.json")
				Expect(err).NotTo(HaveOccurred())
				savedContents, err := os.ReadFile(filepath.Join(subject.Dir, "reusedName", ".cf", "config.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(savedContents)).To(Equal(string(originalContents)), `blah, contents don't match in %q`, subject.Dir)
			})
		})

		When("saving a non-existent config file", func() {
			It("returns a descriptive error", func() {
				result, err := subject.SaveConfig("/path/does/not/exists", "reusedFoundation")
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				Expect(result).To(BeZero())
			})
		})

		When("saving a config with corrupted JSON", func() {
			It("returns a descriptive error", func() {

				result, err := subject.SaveConfig("fixtures/sample-config-with-invalid.json", "reusedFoundation")
				Expect(err).To(MatchError(ContainSubstring("invalid character")))
				Expect(result).To(BeZero())
			})
		})

		When("saving a config missing required values", func() {
			It("returns a descriptive error", func() {
				result, err := subject.SaveConfig("fixtures/sample-config-with-missing-fields.json", "reusedFoundation")
				Expect(err).To(MatchError(ContainSubstring("saved configuration must target Cloudfoundry: missing fields: [API endpoint,Organization,Space]")))
				Expect(result).To(BeZero())
			})
		})
	})

	Context("RemoveConfig", func() {
		It("removes a previously saved configuration", func() {
			_, err := subject.SaveConfig("fixtures/sample-config.json", "target-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(filepath.Join(subject.Dir, "target-name", ".cf", "config.json")).To(BeAnExistingFile())

			err = subject.RemoveConfig("target-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(filepath.Join(subject.Dir, "target-name", ".cf", "config.json")).ToNot(BeAnExistingFile())
		})

		It("does not remove the root directory even if a bad name is provided", func() {
			err := subject.RemoveConfig(".")
			Expect(err).To(MatchError(`invalid target name "."`))
			Expect(subject.Dir).To(BeADirectory())

			err = subject.RemoveConfig("../../../etc/shadow")
			Expect(err).To(MatchError(`invalid target name "../../../etc/shadow"`))
		})
	})
})
