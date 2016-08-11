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

		It("domains prints a json stream of all the domains", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "domains")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(0))

			Expect(sess.Out).To(gbytes.Say(`"domain-1"\n"domain-2"\n`))
		})
	})

	Context("when the server doesn't respond with domains", func() {
		BeforeEach(func() {
			bbsServer.RouteToHandler("POST", "/v1/domains/list",
				ghttp.RespondWith(500, []byte{}))
		})

		It("domains fails with a relevant error message", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "domains")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(1))

			Expect(sess.Err).To(gbytes.Say("Invalid Response with status code: 500"))
		})
	})

	Describe("flag parsing for bbsURL", func() {
		Context("when running domains", func() {
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

		Context("when running set-domain", func() {
			BeforeEach(func() {
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/domains/upsert"),
						ghttp.RespondWithProto(200, &models.UpsertDomainResponse{}),
					),
				)
			})

			It("works with a --bbsURL flag specified before set-domain", func() {
				cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "set-domain", "anything", "--ttl", "0")

				sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				<-sess.Exited
				Expect(sess.ExitCode()).To(Equal(0))
				Expect(sess.Out).To(gbytes.Say("Set domain \"anything\" with TTL at 0 seconds"))
			})

			It("works with a --bbsURL flag specified after set-domain", func() {
				cfdotCmd := exec.Command(cfdotPath, "set-domain", "--bbsURL", bbsServer.URL(), "anything", "--ttl", "40")

				sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				<-sess.Exited
				Expect(sess.ExitCode()).To(Equal(0))
				Expect(sess.Out).To(gbytes.Say("Set domain \"anything\" with TTL at 40 seconds"))
			})
		})
	})

	Context("when the server responds for set-domain", func() {
		BeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/domains/upsert"),
					ghttp.RespondWithProto(200, &models.UpsertDomainResponse{}),
				),
			)
		})

		It("set-domain works with a TTL not specified", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "set-domain", "any-domain")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(0))

			Expect(sess.Out).To(gbytes.Say(`Set domain "any-domain" with TTL at 0 seconds`))
		})

		It("set-domain works with a TTL specified", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "set-domain", "any-domain", "--ttl", "40")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(0))

			Expect(sess.Out).To(gbytes.Say(`Set domain "any-domain" with TTL at 40 seconds`))
		})

		It("set-domain prints to stderr when no domain specified", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "set-domain", "", "--ttl", "40")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(3))

			Expect(sess.Err).To(gbytes.Say(`No domain given`))
			Expect(sess.Err).To(gbytes.Say(`Usage`))
		})

		It("set-domain prints to stderr for negative TTL", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "set-domain", "any-domain", "--ttl", "-40")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(3))

			Expect(sess.Err).To(gbytes.Say(`ttl is negative`))
			Expect(sess.Err).To(gbytes.Say(`Usage:`))
		})

		It("set-domain prints to stderr for non-numeric TTL", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "set-domain", "any-domain", "-t", "asdf")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(3))

			Expect(sess.Err).To(gbytes.Say(`ttl is non-numeric`))
			Expect(sess.Err).To(gbytes.Say(`Usage:`))
		})
	})

	Context("when the server does not respond for set-domain", func() {
		BeforeEach(func() {
			bbsServer.RouteToHandler("POST", "/v1/domains/upsert",
				ghttp.RespondWith(500, []byte{}))
		})
		It("set-domain fails with a relevant error message", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "set-domain", "any-domain")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(1))

			Expect(sess.Err).To(gbytes.Say("Invalid Response with status code: 500"))
		})
	})
})
