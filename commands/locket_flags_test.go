package commands_test

import (
	"os"

	"code.cloudfoundry.org/cfdot/commands"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/spf13/cobra"
)

var _ = Describe("Locket Flags", func() {
	var validTLSFlags map[string]string
	var dummyCmd *cobra.Command
	var err error
	var output *gbytes.Buffer

	BeforeEach(func() {
		dummyCmd = &cobra.Command{
			Use: "dummy",
			Run: func(cmd *cobra.Command, args []string) {},
		}
		commands.AddLocketFlags(dummyCmd)
		output = gbytes.NewBuffer()
		dummyCmd.SetOutput(output)

		validTLSFlags = map[string]string{
			"--locketSkipCertVerify": "false",
			"--locketAPILocation":    "127.0.0.1:9802",
			"--locketCACertFile":     "fixtures/bbsCACert.crt",
			"--locketCertFile":       "fixtures/bbsClient.crt",
			"--locketKeyFile":        "fixtures/bbsClient.key",
		}
	})

	JustBeforeEach(func() {
		err = dummyCmd.PreRunE(dummyCmd, dummyCmd.Flags().Args())
	})

	Describe("locketAPILocation", func() {
		Context("when the --locketAPILocation isn't given", func() {
			BeforeEach(func() {
				parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--locketAPILocation"))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("returns an error message", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(Equal(
					"Locket API Location not set. Please specify one with the '--locketAPILocation' flag or the 'LOCKET_API_LOCATION' environment variable."))
			})

			It("exits with code 3", func() {
				argsErr, ok := err.(commands.CFDotError)
				Expect(ok).To(BeTrue())
				Expect(argsErr.ExitCode()).To(Equal(3))
			})
		})

		Context("when a LOCKET_API_LOCATION environment variable is specified", func() {
			AfterEach(func() {
				os.Unsetenv("LOCKET_API_LOCATION")
			})

			Context("when the LOCKET_API_LOCATION is valid", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--locketAPILocation"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
					os.Setenv("LOCKET_API_LOCATION", "example.com:9083")
				})

				It("does not error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the --locketAPILocation flag is also specified", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
					os.Setenv("LOCKET_API_LOCATION", "broken url")
				})

				It("uses the value from the flag instead of the environment variable", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Describe("locketSkipCertVerify", func() {
		Context("when locketSkipCertVerify is true", func() {
			Context("when the CA cert file is absent", func() {
				BeforeEach(func() {
					validTLSFlags["--locketSkipCertVerify"] = "true"
					delete(validTLSFlags, "--locketCACertFile")
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("ignores the missing CA cert", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("when a LOCKET_SKIP_CERT_VERIFY environment variable is specified", func() {
			AfterEach(func() {
				os.Unsetenv("LOCKET_SKIP_CERT_VERIFY")
			})

			Context("when the LOCKET_SKIP_CERT_VERIFY is valid", func() {
				BeforeEach(func() {
					os.Setenv("LOCKET_SKIP_CERT_VERIFY", "true")
				})

				Context("when the flag is not present", func() {
					BeforeEach(func() {
						delete(validTLSFlags, "--locketSkipCertVerify")
						delete(validTLSFlags, "--locketCACert")
						parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					It("does not error", func() {
						Expect(err).NotTo(HaveOccurred())
					})
				})

				Context("when the flag is set to false", func() {
					BeforeEach(func() {
						parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--locketCACertFile"))
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					It("uses the 'false' value from the flag", func() {
						Expect(err).To(MatchError("--caCertFile must be specified if using HTTPS and --skipCertVerify is not set"))
					})
				})
			})

			Context("when the LOCKET_SKIP_CERT_VERIFY is not valid", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--locketSkipCertVerify"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
					os.Setenv("LOCKET_SKIP_CERT_VERIFY", "sponge")
				})

				It("returns an error", func() {
					Expect(err).To(MatchError("The value 'sponge' is not a valid value for LOCKET_SKIP_CERT_VERIFY. Please specify one of the following valid boolean values: 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False"))
				})
			})

			Context("when the --locketSkipCertVerify flag is also specified", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validTLSFlags, "--locketSkipCertVerify", "true"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
					os.Setenv("LOCKET_SKIP_CERT_VERIFY", "false")
				})

				It("uses the value from the flag instead of the environment variable", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Describe("locketCert/KeyFile", func() {
		Context("when a cert file is specified, but a key file is not", func() {
			BeforeEach(func() {
				parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--locketKeyFile"))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("returns a validation error", func() {
				Expect(err).To(MatchError("--clientCertFile and --clientKeyFile must both be specified for TLS connections."))
			})

			It("exits with code 3", func() {
				cfdotError, ok := err.(commands.CFDotError)
				Expect(ok).To(BeTrue())
				Expect(cfdotError.ExitCode()).To(Equal(3))
			})
		})

		Context("when a key file is specified, but a cert file is not", func() {
			BeforeEach(func() {
				parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--locketCertFile"))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("returns a validation error", func() {
				Expect(err).To(MatchError("--clientCertFile and --clientKeyFile must both be specified for TLS connections."))
			})

			It("exits with code 3", func() {
				cfdotError, ok := err.(commands.CFDotError)
				Expect(ok).To(BeTrue())
				Expect(cfdotError.ExitCode()).To(Equal(3))
			})
		})

		Context("when the key file flag points to a nonexistent file", func() {
			BeforeEach(func() {
				parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validTLSFlags, "--locketKeyFile", "sandwich.key"))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("returns a validation error", func() {
				keyfile := validTLSFlags["--locketKeyFile"]
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
				replaceFlagValue(validTLSFlags, "--locketKeyFile", "fixtures/locketClientBadPermissions.key")
				parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("returns a validation error", func() {
				keyfile := validTLSFlags["--locketKeyFile"]
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
				parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validTLSFlags, "--locketCertFile", "sandwich.crt"))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("returns a validation error", func() {
				certfile := validTLSFlags["--locketCertFile"]
				Expect(err).To(MatchError(MatchRegexp("^cert file '" + certfile + "' doesn't exist or is not readable: .*")))
			})

			It("exits with code 3", func() {
				cfdotError, ok := err.(commands.CFDotError)
				Expect(ok).To(BeTrue())
				Expect(cfdotError.ExitCode()).To(Equal(3))
			})
		})

		Context("when a LOCKET_CERT_FILE environment variable is specified", func() {
			AfterEach(func() {
				os.Unsetenv("LOCKET_CERT_FILE")
			})

			Context("when the LOCKET_CERT_FILE is valid", func() {
				BeforeEach(func() {
					os.Setenv("LOCKET_CERT_FILE", validTLSFlags["--locketCertFile"])
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--locketCertFile"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("does not error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the LOCKET_CERT_FILE points to a nonexistent file", func() {
				BeforeEach(func() {
					os.Setenv("LOCKET_CERT_FILE", "sponge")
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--locketCertFile"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					Expect(err).To(MatchError(MatchRegexp("^cert file 'sponge' doesn't exist or is not readable: .*")))
				})
			})

			Context("when the --locketCertFile flag is also specified", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
					os.Setenv("LOCKET_CERT_FILE", "not a cert file")
				})

				It("uses the value from the flag instead of the environment variable", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("when a LOCKET_KEY_FILE environment variable is specified", func() {
			AfterEach(func() {
				os.Unsetenv("LOCKET_KEY_FILE")
			})

			Context("when the LOCKET_KEY_FILE is valid", func() {
				BeforeEach(func() {
					os.Setenv("LOCKET_KEY_FILE", validTLSFlags["--locketKeyFile"])
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--locketKeyFile"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("does not error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the LOCKET_KEY_FILE points to a nonexistent file", func() {
				BeforeEach(func() {
					os.Setenv("LOCKET_KEY_FILE", "sponge")
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--locketKeyFile"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					Expect(err).To(MatchError(MatchRegexp("^key file 'sponge' doesn't exist or is not readable: .*")))
				})
			})

			Context("when the --locketKeyFile flag is also specified", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
					os.Setenv("LOCKET_KEY_FILE", "not a key file")
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
				parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--locketCACertFile"))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("returns a validation error", func() {
				Expect(err).To(MatchError("--caCertFile must be specified if using HTTPS and --skipCertVerify is not set"))
			})

			It("exits with code 3", func() {
				cfdotError, ok := err.(commands.CFDotError)
				Expect(ok).To(BeTrue())
				Expect(cfdotError.ExitCode()).To(Equal(3))
			})
		})

		Context("when the CA cert file flag points to a nonexistent file", func() {
			BeforeEach(func() {
				parseFlagsErr := dummyCmd.ParseFlags(replaceFlagValue(validTLSFlags, "--locketCACertFile", "notreal.cacrt"))
				Expect(parseFlagsErr).NotTo(HaveOccurred())
			})

			It("returns a validation error", func() {
				certfile := validTLSFlags["--locketCACertFile"]
				Expect(err).To(MatchError(MatchRegexp("^CA cert file '" + certfile + "' doesn't exist or is not readable: .*")))
			})

			It("exits with code 3", func() {
				cfdotError, ok := err.(commands.CFDotError)
				Expect(ok).To(BeTrue())
				Expect(cfdotError.ExitCode()).To(Equal(3))
			})
		})

		Context("when a LOCKET_CA_CERT_FILE environment variable is specified", func() {
			AfterEach(func() {
				os.Unsetenv("LOCKET_CA_CERT_FILE")
			})

			Context("when the LOCKET_CA_CERT_FILE is valid", func() {
				BeforeEach(func() {
					os.Setenv("LOCKET_CA_CERT_FILE", validTLSFlags["--locketCACertFile"])
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--locketCACertFile"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("does not error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the LOCKET_CA_CERT_FILE points to a nonexistent file", func() {
				BeforeEach(func() {
					os.Setenv("LOCKET_CA_CERT_FILE", "sponge")
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validTLSFlags, "--locketCACertFile"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					Expect(err).To(MatchError(MatchRegexp("^CA cert file 'sponge' doesn't exist or is not readable: .*")))
				})
			})

			Context("when the --locketCACertFile flag is also specified", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validTLSFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
					os.Setenv("LOCKET_CA_CERT_FILE", "not a key file")
				})

				It("uses the value from the flag instead of the environment variable", func() {
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
