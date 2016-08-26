package integration_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("help", func() {

	var cfdotCmd *exec.Cmd

	itPrintsHelp := func() {
		It("prints help", func() {

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(0))

			Expect(sess.Out).To(gbytes.Say("A command-line tool to interact with a Cloud Foundry Diego deployment"))
			Expect(sess.Out).To(gbytes.Say("Usage:"))
			Expect(sess.Out).To(gbytes.Say("cfdot"))
			Expect(sess.Out).To(gbytes.Say("Available Commands:"))
			Expect(sess.Out).To(gbytes.Say("domains"))
			Expect(sess.Out).To(gbytes.Say("set-domain"))
		})
	}

	Context("called with no command", func() {
		BeforeEach(func() {
			cfdotCmd = exec.Command(cfdotPath, "help")
		})
		itPrintsHelp()
	})

	Context("called with help option", func() {
		BeforeEach(func() {
			cfdotCmd = exec.Command(cfdotPath, "help")
		})
		itPrintsHelp()
	})

	Context("called with -h", func() {
		BeforeEach(func() {
			cfdotCmd = exec.Command(cfdotPath, "help")
		})
		itPrintsHelp()
	})

	Context("called with --help", func() {
		BeforeEach(func() {
			cfdotCmd = exec.Command(cfdotPath, "help")
		})
		itPrintsHelp()
	})

	Context("called `cfdot help set-domain`", func() {
		BeforeEach(func() {
			cfdotCmd = exec.Command(cfdotPath, "help", "set-domain")
		})
		It("displays the set-domain usage message", func() {
			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(0))

			Expect(sess.Out).To(gbytes.Say("Mark a domain as fresh"))
			Expect(sess.Out).To(gbytes.Say("Usage:"))
			Expect(sess.Out).To(gbytes.Say("cfdot set-domain <domain>"))
			Expect(sess.Out).To(gbytes.Say("Flags:"))
			Expect(sess.Out).To(gbytes.Say("--bbsCACertFile"))
			Expect(sess.Out).To(gbytes.Say("--bbsCertFile"))
			Expect(sess.Out).To(gbytes.Say("--bbsSkipCertVerify"))
			Expect(sess.Out).To(gbytes.Say("--bbsURL"))
			Expect(sess.Out).To(gbytes.Say("-t"))
			Expect(sess.Out).To(gbytes.Say("--ttl"))

		})
	})

})
