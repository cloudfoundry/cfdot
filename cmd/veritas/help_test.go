package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("help", func() {

	var veritasCmd *exec.Cmd

	itPrintsHelp := func() {
		It("prints help", func() {

			sess, err := gexec.Start(veritasCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(0))

			Expect(sess.Out).To(gbytes.Say("veritas"))
			Expect(sess.Out).To(gbytes.Say("Help Options:"))
			Expect(sess.Out).To(gbytes.Say("Available commands:"))
		})
	}

	Context("called with no command", func() {
		BeforeEach(func() {
			veritasCmd = exec.Command(veritasPath, "help")
		})
		itPrintsHelp()
	})

	Context("called with help option", func() {
		BeforeEach(func() {
			veritasCmd = exec.Command(veritasPath, "help")
		})
		itPrintsHelp()
	})

	Context("called with -h", func() {
		BeforeEach(func() {
			veritasCmd = exec.Command(veritasPath, "help")
		})
		itPrintsHelp()
	})

	Context("called with --help", func() {
		BeforeEach(func() {
			veritasCmd = exec.Command(veritasPath, "help")
		})
		itPrintsHelp()
	})

})
