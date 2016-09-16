package integration_test

import (
	"os/exec"

	"code.cloudfoundry.org/bbs/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("retire-actual-lrp", func() {
	itValidatesBBSFlags("retire-actual-lrp", "test-guid", "1")

	Context("when the bbs returns everything successfully", func() {
		BeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/desired_lrps/get_by_process_guid.r2"),
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
		})

		It("exits with exit code 0", func() {
			cfdotCmd := exec.Command(
				cfdotPath,
				"retire-actual-lrp",
				"--bbsURL", bbsServer.URL(),
				"test-process-guid",
				"1",
			)

			session, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(session.Exited).Should(BeClosed())
			Expect(session.ExitCode()).To(Equal(0))
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

			Eventually(session.Exited).Should(BeClosed())
			Expect(session.ExitCode()).To(Equal(4))
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

			Eventually(session.Exited).Should(BeClosed())
			Expect(session.ExitCode()).To(Equal(3))
		})
	})
})
