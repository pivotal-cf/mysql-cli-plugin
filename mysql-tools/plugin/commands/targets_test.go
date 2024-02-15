package commands_test

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/multisite"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin/commands"
	"github.com/pivotal-cf/mysql-cli-plugin/mysql-tools/plugin/commands/fakes"
)

var _ = Describe("Environment Targeting Commands", func() {

	const (
		savedOrg      = "SavedOrgName"
		savedSpace    = "SavedSpaceName"
		savedEndpoint = "SavedAPIEndpoint"
		savedTarget   = "SavedTarget"
	)

	var fakeMultiSiteCfg *fakes.FakeMultisiteConfig

	var exampleTarget = multisite.Target{
		Name:         savedTarget,
		Organization: savedOrg,
		Space:        savedSpace,
		API:          savedEndpoint,
	}

	BeforeEach(func() {
		fakeMultiSiteCfg = new(fakes.FakeMultisiteConfig)
	})

	Context("ListTargets", func() {
		It("Prints a summary when successful", func() {
			fakeMultiSiteCfg.ListConfigsReturns([]multisite.Target{exampleTarget}, nil)

			r, w, _ := os.Pipe()
			tmp := os.Stdout
			defer func() {
				os.Stdout = tmp
			}()

			os.Stdout = w
			err := commands.ListTargets(fakeMultiSiteCfg)
			_ = w.Close()
			Expect(err).NotTo(HaveOccurred())

			By("showing a summary of the saved config")
			stdout, _ := io.ReadAll(r)
			Expect(string(stdout)).To(ContainSubstring("Targets:"))
			Expect(string(stdout)).To(ContainSubstring(savedEndpoint))
			Expect(string(stdout)).To(ContainSubstring(savedOrg))
			Expect(string(stdout)).To(ContainSubstring(savedOrg))
			Expect(string(stdout)).To(ContainSubstring(savedTarget))
			Expect(fakeMultiSiteCfg.ListConfigsCallCount()).To(Equal(1))
		})
		It("Prints a summary when there is an error", func() {
			fakeMultiSiteCfg.ListConfigsReturns([]multisite.Target{exampleTarget}, errors.New("some-error"))

			r, w, _ := os.Pipe()
			tmp := os.Stdout
			defer func() {
				os.Stdout = tmp
			}()

			os.Stdout = w
			err := commands.ListTargets(fakeMultiSiteCfg)
			_ = w.Close()
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("error listing multisite targets: some-error"))

			By("showing a summary of the saved config")
			stdout, _ := io.ReadAll(r)
			Expect(string(stdout)).To(ContainSubstring("Targets:"))
			Expect(string(stdout)).To(ContainSubstring(savedEndpoint))
			Expect(string(stdout)).To(ContainSubstring(savedOrg))
			Expect(string(stdout)).To(ContainSubstring(savedOrg))
			Expect(string(stdout)).To(ContainSubstring(savedTarget))
			Expect(fakeMultiSiteCfg.ListConfigsCallCount()).To(Equal(1))
		})
		It("prints nothing when there are no configs", func() {
			fakeMultiSiteCfg.ListConfigsReturns(nil, nil)

			r, w, _ := os.Pipe()
			tmp := os.Stdout
			defer func() {
				os.Stdout = tmp
			}()

			os.Stdout = w
			err := commands.ListTargets(fakeMultiSiteCfg)
			_ = w.Close()
			Expect(err).NotTo(HaveOccurred())

			By("showing a summary of the saved config")
			stdout, _ := io.ReadAll(r)
			Expect(string(stdout)).To(ContainSubstring("Targets:"))
			Expect(fakeMultiSiteCfg.ListConfigsCallCount()).To(Equal(1))
		})
	})

	Context("SaveTarget", func() {
		BeforeEach(func() {
			fakeMultiSiteCfg.SaveConfigReturns(exampleTarget, nil)
		})

		When("CF_HOME is set to a directory", func() {
			var (
				testCFHome, originalCFHome string
				hadCFHome                  bool
				err                        error
			)
			BeforeEach(func() {
				testCFHome, err = os.MkdirTemp("", "test_CFHOME_")
				if err != nil {
					panic("Failed to create temp CF_HOME: " + err.Error())
				}
				err = os.Mkdir(filepath.Join(testCFHome, ".cf"), 0750)
				if err != nil {
					panic("Failed to create temp CF_HOME/.cf/: " + err.Error())
				}
				originalCFHome, hadCFHome = os.LookupEnv("CF_HOME")
				err = os.Setenv("CF_HOME", testCFHome)
				if err != nil {
					panic("Failed to set new test CF_HOME environment variable: " + err.Error())
				}

				DeferCleanup(func() {
					if hadCFHome {
						err = os.Setenv("CF_HOME", originalCFHome)
						if err != nil {
							panic("Failed to restore original CF_HOME environment variable: " + err.Error())
						}
					}
					_ = os.RemoveAll(testCFHome)
				})
			})

			When("CF_HOME contains a config file", func() {
				BeforeEach(func() {
					Expect(os.WriteFile(filepath.Join(testCFHome, ".cf", "config.json"), nil, 0640)).To(Succeed())
				})

				It("saves a target without error", func() {
					var err error
					args := []string{"targetName"}

					r, w, _ := os.Pipe()
					tmp := os.Stdout
					defer func() {
						os.Stdout = tmp
					}()

					os.Stdout = w
					err = commands.SaveTarget(args, fakeMultiSiteCfg)
					_ = w.Close()
					Expect(err).To(Not(HaveOccurred()))

					By("constructing the path to the saved cf config file")
					inputConfigFile, target := fakeMultiSiteCfg.SaveConfigArgsForCall(0)
					Expect(inputConfigFile).To(Equal(filepath.Join(testCFHome, ".cf", "config.json")))
					Expect(target).To(Equal("targetName"))

					By("showing a summary of the saved config")
					stdout, _ := io.ReadAll(r)
					Expect(string(stdout)).To(ContainSubstring("Success"))
					Expect(string(stdout)).To(ContainSubstring(savedEndpoint))
					Expect(string(stdout)).To(ContainSubstring(savedOrg))
					Expect(string(stdout)).To(ContainSubstring(savedSpace))
				})
				It("surfaces any underlying multisite errors", func() {
					fakeMultiSiteCfg.SaveConfigReturns(multisite.Target{}, errors.New("low-level save error"))
					args := []string{"targetName"}

					err := commands.SaveTarget(args, fakeMultiSiteCfg)
					Expect(err).To(MatchError("error saving target targetName: low-level save error"))
				})
			})

		})
		When("CF_HOME is set to an invalid directory with no config", func() {
			var (
				testCFHome, originalCFHome string
				hadCFHome                  bool
				err                        error
			)
			BeforeEach(func() {
				testCFHome = os.TempDir()
				originalCFHome, hadCFHome = os.LookupEnv("CF_HOME")
				err = os.Setenv("CF_HOME", testCFHome)
				if err != nil {
					panic("Failed to set new test CF_HOME environment variable: " + err.Error())
				}
			})
			AfterEach(func() {
				if hadCFHome {
					err = os.Setenv("CF_HOME", originalCFHome)
					if err != nil {
						panic("Failed to restore original CF_HOME environment variable: " + err.Error())
					}
				}
			})
			It("errors without ever calling Multisite", func() {
				args := []string{"targetName"}

				err := commands.SaveTarget(args, fakeMultiSiteCfg)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
				Expect(fakeMultiSiteCfg.SaveConfigCallCount()).To(Equal(0))
			})
		})
		// When("no CF_HOME is set"), plugin uses default ${HOME}/.cf

		It("errors if not enough args are passed", func() {
			var args []string
			err := commands.SaveTarget(args, fakeMultiSiteCfg)
			Expect(err).To(MatchError("Usage: " + commands.SaveTargetUsage + "\n\nthe required argument `<target-name>` was not provided"))
		})

		It("errors if too many args are passed", func() {
			args := []string{"targetName", "extra-arg"}
			err := commands.SaveTarget(args, fakeMultiSiteCfg)
			Expect(err).To(MatchError("Usage: " + commands.SaveTargetUsage + "\n\nunexpected arguments: extra-arg"))
		})

		It("errors if an invalid flag is passed", func() {
			args := []string{"targetName", "--invalid-flag"}
			err := commands.SaveTarget(args, fakeMultiSiteCfg)
			Expect(err).To(MatchError("Usage: " + commands.SaveTargetUsage + "\n\nunknown flag `invalid-flag'"))
		})
	})

	Context("RemoveTarget", func() {
		It("able to remove target config without an error", func() {
			args := []string{"targetName"}
			err := commands.RemoveTarget(args, fakeMultiSiteCfg)
			Expect(err).To(Not(HaveOccurred()))
			Expect(fakeMultiSiteCfg.RemoveConfigCallCount()).To(Equal(1))

			target := fakeMultiSiteCfg.RemoveConfigArgsForCall(0)
			Expect(target).To(Equal("targetName"))
		})

		It("errors if the remove target fails", func() {
			fakeMultiSiteCfg.RemoveConfigReturns(errors.New("some-error"))
			args := []string{"targetName"}
			err := commands.RemoveTarget(args, fakeMultiSiteCfg)
			Expect(err).To(MatchError("error trying to remove the target config: some-error"))
		})

		It("errors if not enough args are passed", func() {
			var args []string
			err := commands.RemoveTarget(args, fakeMultiSiteCfg)
			Expect(err).To(MatchError("Usage: " + commands.RemoveTargetUsage + "\n\nthe required argument `<target-name>` was not provided"))
		})

		It("errors if too many args are passed", func() {
			args := []string{"targetName", "extra-arg"}
			err := commands.RemoveTarget(args, fakeMultiSiteCfg)
			Expect(err).To(MatchError("Usage: " + commands.RemoveTargetUsage + "\n\nunexpected arguments: extra-arg"))
		})

		It("errors if an invalid flag is passed", func() {
			args := []string{"targetName", "--invalid-flag"}
			err := commands.RemoveTarget(args, fakeMultiSiteCfg)
			Expect(err).To(MatchError("Usage: " + commands.RemoveTargetUsage + "\n\nunknown flag `invalid-flag'"))
		})

	})
})
