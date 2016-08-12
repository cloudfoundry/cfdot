package commands_test

import (
	"code.cloudfoundry.org/cfdot/commands"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/spf13/cobra"
)

var _ = Describe("Set Domain Flags", func() {
	var validFlags map[string]string
	var dummyCmd *cobra.Command
	var err error
	var output *gbytes.Buffer
	var dummyArgs []string

	BeforeEach(func() {
		dummyCmd = &cobra.Command{
			Use: "dummy",
			Run: func(cmd *cobra.Command, args []string) {},
		}
		commands.AddSetDomainFlags(dummyCmd)
		dummyCmd.PreRunE = commands.SetDomainPrehook

		output = gbytes.NewBuffer()
		dummyCmd.SetOutput(output)

		validFlags = map[string]string{
			"--ttl": "100",
			"-t":    "40000",
		}
	})

	JustBeforeEach(func() {
		err = dummyCmd.PreRunE(dummyCmd, dummyArgs)
	})

	Describe("ttl", func() {

		itReturnsErrorMessage := func(message string) {
			It("returns an error message", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(Equal(message))
			})
		}

		itExitsWithErrorCode := func(errorCode int) {
			It("exits with error code", func() {
				argsErr, ok := err.(commands.CFDotError)
				Expect(ok).To(BeTrue())
				Expect(argsErr.ExitCode()).To(Equal(errorCode))
			})
		}

		Context("when a domain is given", func() {
			BeforeEach(func() {
				dummyArgs = []string{"dummy-domain"}
			})

			Context("when just --ttl is given", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validFlags, "-t"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				Context("when --ttl isn't numeric", func() {
					BeforeEach(func() {
						parseFlagsErr := dummyCmd.ParseFlags(
							replaceFlagValue(validFlags, "--ttl", "notnumeric"),
						)
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					itReturnsErrorMessage("ttl is non-numeric")
					itExitsWithErrorCode(3)
				})

				Context("when --ttl is negative", func() {
					BeforeEach(func() {
						parseFlagsErr := dummyCmd.ParseFlags(
							replaceFlagValue(validFlags, "--ttl", "-1"),
						)
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					itReturnsErrorMessage("ttl is negative")
					itExitsWithErrorCode(3)
				})

				Context("when --ttl is valid", func() {
					It("does not error", func() {
						Expect(err).NotTo(HaveOccurred())
						Expect(commands.TTLAsInt()).To(Equal(100))
					})
				})

			})

			Context("when just -t is given", func() {
				BeforeEach(func() {
					parseFlagsErr := dummyCmd.ParseFlags(removeFlag(validFlags, "--ttl"))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})

				Context("when -t isn't numeric", func() {
					BeforeEach(func() {
						parseFlagsErr := dummyCmd.ParseFlags(
							replaceFlagValue(validFlags, "-t", "notnumeric"),
						)
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					itReturnsErrorMessage("ttl is non-numeric")
					itExitsWithErrorCode(3)
				})

				Context("when -t is negative", func() {
					BeforeEach(func() {
						parseFlagsErr := dummyCmd.ParseFlags(
							replaceFlagValue(validFlags, "-t", "-1"),
						)
						Expect(parseFlagsErr).NotTo(HaveOccurred())
					})

					itReturnsErrorMessage("ttl is negative")
					itExitsWithErrorCode(3)
				})

				Context("when -t is valid", func() {
					It("does not error", func() {
						Expect(err).NotTo(HaveOccurred())
						Expect(commands.TTLAsInt()).To(Equal(40000))
					})
				})

			})

			Context("when -t and --ttl aren't given", func() {
				BeforeEach(func() {
					delete(validFlags, "--ttl")
					delete(validFlags, "-t")
					parseFlagsErr := dummyCmd.ParseFlags(buildArgList(validFlags))
					Expect(parseFlagsErr).NotTo(HaveOccurred())
				})
				It("does not error", func() {
					Expect(err).NotTo(HaveOccurred())
				})
				It("sets the ttl to 0", func() {
					Expect(commands.TTLAsInt()).To(Equal(0))
				})
			})

		})

		Context("when a domain is not given", func() {
			BeforeEach(func() {
				dummyArgs = []string{}
			})
			itReturnsErrorMessage("No domain given")
			itExitsWithErrorCode(3)
		})
	})
})
