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

var _ = Describe("retire-actual-lrp", func() {
	itValidatesBBSFlags("retire-actual-lrp", "test-guid", "1")

	Context("when the bbs returns everything successfully", func() {
		var (
			cfdotArgs     []string
			serverTimeout int
			session       *gexec.Session
		)
		BeforeEach(func() {
			serverTimeout = 0
			cfdotArgs = []string{"--bbsURL", bbsServer.URL()}
		})

		JustBeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/desired_lrps/get_by_process_guid.r2"),
					func(w http.ResponseWriter, req *http.Request) {
						time.Sleep(time.Duration(serverTimeout) * time.Second)
					},
					ghttp.RespondWithProto(200, &models.DesiredLRPResponse{
						DesiredLrp: &models.DesiredLRP{
							Domain: "test-domain",
						},
					}),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/actual_lrps/retire"),
					ghttp.RespondWithProto(200, &models.ActualLRPLifecycleResponse{}),
				),
			)
			execArgs := append(
				cfdotArgs,
				"retire-actual-lrp",
				"test-process-guid",
				"1",
			)

			cfdotCmd := exec.Command(
				cfdotPath,
				execArgs...,
			)
			var err error
			session, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
		})

		It("exits with exit code 0", func() {
			Eventually(session).Should(gexec.Exit(0))
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
					Eventually(session, 2).Should(gexec.Exit(4))
					Expect(session.Err).To(gbytes.Say(`Timeout exceeded`))
				})
			})

			Context("when request is within the timeout", func() {
				It("exits with status code of 0", func() {
					Eventually(session).Should(gexec.Exit(0))
				})
			})
		})
	})

	Context("when the bbs returns an error", func() {
		BeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/desired_lrps/get_by_process_guid.r2"),
					ghttp.RespondWithProto(200, &models.DesiredLRPResponse{
						Error: &models.Error{
							Type:    models.Error_Deadlock,
							Message: "the request failed due to deadlock",
						},
					}),
				),
			)
		})

		It("exits with exit code 4", func() {
			cfdotCmd := exec.Command(
				cfdotPath,
				"retire-actual-lrp",
				"--bbsURL", bbsServer.URL(),
				"test-process-guid",
				"1",
			)

			session, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(4))
		})
	})

	Context("when invalid arguments are passed", func() {
		It("exits with exit code 3", func() {
			cfdotCmd := exec.Command(
				cfdotPath,
				"retire-actual-lrp",
				"--bbsURL", bbsServer.URL(),
				"test-process-guid",
				"a",
			)

			session, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(3))
		})
	})
})
