package commands_test

import (
	"os"

	"code.cloudfoundry.org/cfdot/commands"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/spf13/cobra"
)

var _ = Describe("BBS Flags", func() {
	var validFlags, validTLSFlags map[string]string
	var dummyCmd *cobra.Command
	var err error
	var output *gbytes.Buffer

	BeforeEach(func() {
		dummyCmd = &cobra.Command{
			Use: "dummy",
			Run: func(cmd *cobra.Command, args []string) {},
		}
		commands.AddBBSFlags(dummyCmd)
		output = gbytes.NewBuffer()
		dummyCmd.SetOutput(output)

		validFlags = map[string]string{"--bbsURL": "http://example.com"}

		validTLSFlags = map[string]string{
			"--bbsSkipCertVerify": "false",
			"--bbsURL":            "https://example.com",
			"--bbsCACertFile":     "fixtures/bbsCACert.crt",
			"--bbsCertFile":       "fixtures/bbsClient.crt",
			"--bbsKeyFile":        "fixtures/bbsClient.key",
		}
	})

	JustBeforeEach(func() {
		err = dummyCmd.PreRunE(dummyCmd, dummyCmd.Flags().Args())
	})

	Describe("bbsURL", func() {
		Context("when the --bbsURL isn't given", func() {
			BeforeEach(func() {
				parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validFlags, "--bbsURL"))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("returns an error message", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(Equal(
					"BBS URL not set. Please specify one with the '--bbsURL' flag or the 'BBS_URL' environment variable."))
			})

			It("exits with code 3", func() {
				argsErr, ok := err.(commands.CFDotError)
				Expect(ok).To(BeTrue())
				Expect(argsErr.ExitCode()).To(Equal(3))
			})
		})

		Context("when the --bbsURL isn't a valid URL", func() {
			BeforeEach(func() {
				parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validFlags, "--bbsURL", ":"))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("returns an error message", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(Equal(
					"The value ':' is not a valid BBS URL. Please specify one with the '--bbsURL' flag or the 'BBS_URL' environment variable."))
			})

			It("exits with code 3", func() {
				argsErr, ok := err.(commands.CFDotError)
				Expect(ok).To(BeTrue())
				Expect(argsErr.ExitCode()).To(Equal(3))
			})
		})

		Context("when the --bbsURL is not http or https", func() {
			BeforeEach(func() {
				parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validFlags, "--bbsURL", "nohttp.com"))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("returns an error message", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(Equal(
					"The URL 'nohttp.com' does not have an 'http' or 'https' scheme. Please specify one with the '--bbsURL' flag or the 'BBS_URL' environment variable."))
			})

			It("exits with code 3", func() {
				argsErr, ok := err.(commands.CFDotError)
				Expect(ok).To(BeTrue())
				Expect(argsErr.ExitCode()).To(Equal(3))
			})
		})

		Context("when a BBS_URL environment variable is specified", func() {
			AfterEach(func() {
				os.Unsetenv("BBS_URL")
			})

			Context("when the BBS_URL is valid", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validFlags, "--bbsURL"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
					os.Setenv("BBS_URL", "http://example.com")
				})

				It("does not error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the BBS_URL is not valid", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validFlags, "--bbsURL"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
					os.Setenv("BBS_URL", ":")
				})

				It("returns an err", func() {
					Expect(err).To(MatchError("The value ':' is not a valid BBS URL. Please specify one with the '--bbsURL' flag or the 'BBS_URL' environment variable."))
				})
			})

			Context("when the --bbsURL flag is also specified", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
					os.Setenv("BBS_URL", "broken url")
				})

				It("uses the value from the flag instead of the environment variable", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Describe("bbsSkipCertVerify", func() {
		Context("when the URL does not start with HTTPS", func() {
			BeforeEach(func() {
				parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validTLSFlags, "--bbsURL", "http://example.com"))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("ignores the flag", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when bbsSkipCertVerify is true", func() {
			Context("when the CA cert file is absent", func() {
				BeforeEach(func() {
					validTLSFlags["--bbsSkipCertVerify"] = "true"
					delete(validTLSFlags, "--bbsCACertFile")
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("ignores the missing CA cert", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("when a BBS_SKIP_CERT_VERIFY environment variable is specified", func() {
			AfterEach(func() {
				os.Unsetenv("BBS_SKIP_CERT_VERIFY")
			})

			Context("when the BBS_SKIP_CERT_VERIFY is valid", func() {
				BeforeEach(func() {
					os.Setenv("BBS_SKIP_CERT_VERIFY", "true")
				})

				Context("when the flag is not present", func() {
					BeforeEach(func() {
						delete(validTLSFlags, "--bbsSkipCertVerify")
						delete(validTLSFlags, "--bbsCACert")
						parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					It("does not error", func() {
						Expect(err).NotTo(HaveOccurred())
					})
				})

				Context("when the flag is set to false", func() {
					BeforeEach(func() {
						parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--bbsCACertFile"))
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					It("uses the 'false' value from the flag", func() {
						Expect(err).To(MatchError("--bbsCACertFile must be specified if using HTTPS and --bbsSkipCertVerify is not set"))
					})
				})
			})

			Context("when the BBS_SKIP_CERT_VERIFY is not valid", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--bbsSkipCertVerify"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
					os.Setenv("BBS_SKIP_CERT_VERIFY", "sponge")
				})

				It("returns an error", func() {
					Expect(err).To(MatchError("The value 'sponge' is not a valid value for BBS_SKIP_CERT_VERIFY. Please specify one of the following valid boolean values: 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False"))
				})
			})

			Context("when the --bbsSkipCertVerify flag is also specified", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validTLSFlags, "--bbsSkipCertVerify", "true"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
					os.Setenv("BBS_SKIP_CERT_VERIFY", "false")
				})

				It("uses the value from the flag instead of the environment variable", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Describe("bbsCert/KeyFile", func() {
		Context("when a cert file is specified, but a key file is not", func() {
			BeforeEach(func() {
				parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--bbsKeyFile"))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("returns a validation error", func() {
				Expect(err).To(MatchError("--bbsCertFile and --bbsKeyFile must both be specified for mutual TLS connections"))
			})

			It("exits with code 3", func() {
				cfdotError, ok := err.(commands.CFDotError)
				Expect(ok).To(BeTrue())
				Expect(cfdotError.ExitCode()).To(Equal(3))
			})
		})

		Context("when a key file is specified, but a cert file is not", func() {
			BeforeEach(func() {
				parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--bbsCertFile"))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("returns a validation error", func() {
				Expect(err).To(MatchError("--bbsCertFile and --bbsKeyFile must both be specified for mutual TLS connections"))
			})

			It("exits with code 3", func() {
				cfdotError, ok := err.(commands.CFDotError)
				Expect(ok).To(BeTrue())
				Expect(cfdotError.ExitCode()).To(Equal(3))
			})
		})

		Context("when both the key file and cert file flags are missing", func() {
			BeforeEach(func() {
				delete(validTLSFlags, "--bbsCertFile")
				delete(validTLSFlags, "--bbsKeyFile")

				parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("does not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the key file flag points to a nonexistent file", func() {
			BeforeEach(func() {
				parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validTLSFlags, "--bbsKeyFile", "sandwich.key"))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("returns a validation error", func() {
				keyfile := validTLSFlags["--bbsKeyFile"]
				Expect(err).To(MatchError(MatchRegexp("^key file '" + keyfile + "' doesn't exist or is not readable: .*")))
			})

			It("exits with code 3", func() {
				cfdotError, ok := err.(commands.CFDotError)
				Expect(ok).To(BeTrue())
				Expect(cfdotError.ExitCode()).To(Equal(3))
			})
		})

		Context("when the key file flag points to a file without read permissions", func() {
			BeforeEach(func() {
				chmodErr := os.Chmod("fixtures/bbsClient.key", 0300)
				Expect(chmodErr).NotTo(HaveOccurred())
				parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				chmodErr := os.Chmod("fixtures/bbsClient.key", 0644)
				Expect(chmodErr).NotTo(HaveOccurred())
			})

			It("returns a validation error", func() {
				keyfile := validTLSFlags["--bbsKeyFile"]
				Expect(err).To(MatchError(MatchRegexp("^key file '" + keyfile + "' doesn't exist or is not readable: .*")))
			})

			It("exits with code 3", func() {
				cfdotError, ok := err.(commands.CFDotError)
				Expect(ok).To(BeTrue())
				Expect(cfdotError.ExitCode()).To(Equal(3))
			})
		})

		Context("when the cert file flag points to a nonexistent file", func() {
			BeforeEach(func() {
				parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validTLSFlags, "--bbsCertFile", "sandwich.crt"))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("returns a validation error", func() {
				certfile := validTLSFlags["--bbsCertFile"]
				Expect(err).To(MatchError(MatchRegexp("^cert file '" + certfile + "' doesn't exist or is not readable: .*")))
			})

			It("exits with code 3", func() {
				cfdotError, ok := err.(commands.CFDotError)
				Expect(ok).To(BeTrue())
				Expect(cfdotError.ExitCode()).To(Equal(3))
			})
		})

		Context("when a BBS_CERT_FILE environment variable is specified", func() {
			AfterEach(func() {
				os.Unsetenv("BBS_CERT_FILE")
			})

			Context("when the BBS_CERT_FILE is valid", func() {
				BeforeEach(func() {
					os.Setenv("BBS_CERT_FILE", validTLSFlags["--bbsCertFile"])
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--bbsCertFile"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("does not error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the BBS_CERT_FILE points to a nonexistent file", func() {
				BeforeEach(func() {
					os.Setenv("BBS_CERT_FILE", "sponge")
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--bbsCertFile"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					Expect(err).To(MatchError(MatchRegexp("^cert file 'sponge' doesn't exist or is not readable: .*")))
				})
			})

			Context("when the --bbsCertFile flag is also specified", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
					os.Setenv("BBS_CERT_FILE", "not a cert file")
				})

				It("uses the value from the flag instead of the environment variable", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("when a BBS_KEY_FILE environment variable is specified", func() {
			AfterEach(func() {
				os.Unsetenv("BBS_KEY_FILE")
			})

			Context("when the BBS_KEY_FILE is valid", func() {
				BeforeEach(func() {
					os.Setenv("BBS_KEY_FILE", validTLSFlags["--bbsKeyFile"])
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--bbsKeyFile"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("does not error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the BBS_KEY_FILE points to a nonexistent file", func() {
				BeforeEach(func() {
					os.Setenv("BBS_KEY_FILE", "sponge")
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--bbsKeyFile"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					Expect(err).To(MatchError(MatchRegexp("^key file 'sponge' doesn't exist or is not readable: .*")))
				})
			})

			Context("when the --bbsKeyFile flag is also specified", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
					os.Setenv("BBS_KEY_FILE", "not a key file")
				})

				It("uses the value from the flag instead of the environment variable", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Context("CA Cert file", func() {
		Context("when CA cert is not specified", func() {
			BeforeEach(func() {
				parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--bbsCACertFile"))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("returns a validation error", func() {
				Expect(err).To(MatchError("--bbsCACertFile must be specified if using HTTPS and --bbsSkipCertVerify is not set"))
			})

			It("exits with code 3", func() {
				cfdotError, ok := err.(commands.CFDotError)
				Expect(ok).To(BeTrue())
				Expect(cfdotError.ExitCode()).To(Equal(3))
			})
		})

		Context("when the CA cert file flag points to a nonexistent file", func() {
			BeforeEach(func() {
				parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validTLSFlags, "--bbsCACertFile", "notreal.cacrt"))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("returns a validation error", func() {
				certfile := validTLSFlags["--bbsCACertFile"]
				Expect(err).To(MatchError(MatchRegexp("^CA cert file '" + certfile + "' doesn't exist or is not readable: .*")))
			})

			It("exits with code 3", func() {
				cfdotError, ok := err.(commands.CFDotError)
				Expect(ok).To(BeTrue())
				Expect(cfdotError.ExitCode()).To(Equal(3))
			})
		})

		Context("when a BBS_CA_CERT_FILE environment variable is specified", func() {
			AfterEach(func() {
				os.Unsetenv("BBS_CA_CERT_FILE")
			})

			Context("when the BBS_CA_CERT_FILE is valid", func() {
				BeforeEach(func() {
					os.Setenv("BBS_CA_CERT_FILE", validTLSFlags["--bbsCACertFile"])
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--bbsCACertFile"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("does not error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the BBS_CA_CERT_FILE points to a nonexistent file", func() {
				BeforeEach(func() {
					os.Setenv("BBS_CA_CERT_FILE", "sponge")
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--bbsCACertFile"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					Expect(err).To(MatchError(MatchRegexp("^CA cert file 'sponge' doesn't exist or is not readable: .*")))
				})
			})

			Context("when the --bbsCACertFile flag is also specified", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
					os.Setenv("BBS_CA_CERT_FILE", "not a key file")
				})

				It("uses the value from the flag instead of the environment variable", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})

func removeFlag(flags map[string]string, toRemove string) []string {
	delete(flags, toRemove)

	return buildArgList(flags)
}

func replaceFlagValue(flags map[string]string, key string, newValue string) []string {
	flags[key] = newValue

	return buildArgList(flags)
}

func buildArgList(flags map[string]string) []string {
	list := []string{}

	for key, value := range flags {
		list = append(list, key+"="+value)
	}

	return list
}
