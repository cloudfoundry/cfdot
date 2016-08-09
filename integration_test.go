package main_test

import (
	"os/exec"

	"code.cloudfoundry.org/bbs/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("cfdot Integration", func() {
	Context("when the server responds with domains", func() {
		BeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/domains/list"),
					ghttp.RespondWithProto(200, &models.DomainsResponse{
						Error:   nil,
						Domains: []string{"domain-1", "domain-2"},
					}),
				),
			)
		})

		It("prints a json stream of all the domains", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "domains")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(0))

			Expect(sess.Out).To(gbytes.Say(`"domain-1"\n"domain-2"\n`))
		})
	})

	Context("when the server doesn't respond", func() {
		BeforeEach(func() {
			bbsServer.RouteToHandler("POST", "/v1/domains/list",
				ghttp.RespondWith(500, []byte{}))
		})

		It("fails with a relevant error message", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "domains")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(1))

			Expect(sess.Err).To(gbytes.Say("Invalid Response with status code: 500"))
		})
	})

	Describe("flag parsing", func() {
		BeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/domains/list"),
					ghttp.RespondWithProto(200, &models.DomainsResponse{}),
				),
			)
		})

		It("works with a --bbsURL flag specified before domains", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "domains")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(0))
		})

		It("works with a --bbsURL flag specified after domains", func() {
			cfdotCmd := exec.Command(cfdotPath, "domains", "--bbsURL", bbsServer.URL())

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(0))
		})
	})
})
