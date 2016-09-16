package integration_test

import (
	"encoding/json"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/bbs/models"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("desired-lrp", func() {
	var (
		sess       *gexec.Session
		desiredLRP *models.DesiredLRP
	)

	BeforeEach(func() {
		desiredLRP = &models.DesiredLRP{
			ProcessGuid: "test-guid",
			Instances:   2,
		}

		bbsServer.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/v1/desired_lrps/get_by_process_guid.r2"),
				ghttp.VerifyProtoRepresenting(&models.DesiredLRPByProcessGuidRequest{
					ProcessGuid: "test-guid",
				}),
				ghttp.RespondWithProto(200, &models.DesiredLRPResponse{
					DesiredLrp: desiredLRP,
					Error:      nil,
				}),
			),
		)
	})

	JustBeforeEach(func() {
		cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "desired-lrp", "test-guid")

		var err error
		sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(sess.Exited).Should(BeClosed())
	})

	It("exits with status code of 0", func() {
		Expect(sess.ExitCode()).To(Equal(0))
	})

	It("returns the json encoding of the desired lrp scheduling info", func() {
		jsonData, err := json.Marshal(desiredLRP)
		Expect(err).NotTo(HaveOccurred())

		Expect(sess.Out).To(gbytes.Say(string(jsonData)))
	})
})
