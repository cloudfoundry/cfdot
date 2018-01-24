package integration_test

import (
	"net/http"
	"os/exec"
	"time"

	"code.cloudfoundry.org/bbs/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("domains", func() {
	var sess *gexec.Session

	itValidatesBBSFlags("domains")
	itHasNoArgs("domains", false)

	Context("when the server responds with domains", func() {
		var (
			serverTimeout int
			cfdotArgs     []string
		)
		BeforeEach(func() {
			cfdotArgs = []string{"--bbsURL", bbsServer.URL()}
			serverTimeout = 0
		})

		JustBeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/domains/list"),
					func(w http.ResponseWriter, req *http.Request) {
						time.Sleep(time.Duration(serverTimeout) * time.Second)
					},
					ghttp.RespondWithProto(200, &models.DomainsResponse{
						Error:   nil,
						Domains: []string{"domain-1", "domain-2"},
					}),
				),
			)

			execArgs := append(cfdotArgs, "domains")
			cfdotCmd := exec.Command(
				cfdotPath,
				execArgs...,
			)

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
		})

		It("domains prints a json stream of all the domains", func() {
			Eventually(sess).Should(gexec.Exit(0))
			Expect(sess.Out).To(gbytes.Say(`"domain-1"\n"domain-2"\n`))
		})

		Context("when timeout flag is present", func() {
			BeforeEach(func() {
				cfdotArgs = append(cfdotArgs, "--timeout", "1")
			})

			Context("when request exceeds timeout", func() {
				BeforeEach(func() {
					serverTimeout = 2
				})

				It("exits with code 4 and a timeout message", func() {
					Eventually(sess, 2).Should(gexec.Exit(4))
					Expect(sess.Err).To(gbytes.Say(`Timeout exceeded`))
				})
			})

			Context("when request is within the timeout", func() {
				It("exits with status code of 0", func() {
					Eventually(sess).Should(gexec.Exit(0))
					Expect(sess.Out).To(gbytes.Say(`"domain-1"\n"domain-2"\n`))
				})
			})
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

			Eventually(sess, 2*time.Second).Should(gexec.Exit(4))
			Expect(sess.Err).To(gbytes.Say("Invalid Response with status code: 500"))
		})
	})

	Context("when the server returns an error", func() {
		BeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/domains/list"),
					ghttp.RespondWithProto(200, &models.DomainsResponse{
						Error: &models.Error{
							Type:    models.Error_Deadlock,
							Message: "the request failed due to deadlock",
						},
						Domains: nil,
					}),
				),
			)
		})

		JustBeforeEach(func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "domains")

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
		})

		It("exits with status code 4 and should print the type and message of the error", func() {
			Eventually(sess).Should(gexec.Exit(4))
			Expect(sess.Err).To(gbytes.Say("BBS error"))
			Expect(sess.Err).To(gbytes.Say("Type 28: Deadlock"))
			Expect(sess.Err).To(gbytes.Say("Message: the request failed due to deadlock"))
		})

		It("should not print the usage", func() {
			Expect(sess.Err).NotTo(gbytes.Say("Usage:"))
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

				Eventually(sess).Should(gexec.Exit(0))
			})

			It("works with a --bbsURL flag specified after domains", func() {
				cfdotCmd := exec.Command(cfdotPath, "domains", "--bbsURL", bbsServer.URL())

				sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())

				Eventually(sess).Should(gexec.Exit(0))
			})
		})
	})
})
