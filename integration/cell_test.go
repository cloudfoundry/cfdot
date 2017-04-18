package integration_test

import (
	"encoding/json"
	"os/exec"

	"code.cloudfoundry.org/bbs/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("cell", func() {
	itValidatesBBSFlags("cell")

	Context("when cell command is called", func() {
		var presence *models.CellPresence

		BeforeEach(func() {
			presence = &models.CellPresence{
				CellId:     "cell-1",
				RepAddress: "rep-1",
			}

			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/cells/list.r1"),
					ghttp.RespondWithProto(200, &models.CellsResponse{
						Cells: []*models.CellPresence{presence},
					}),
				),
			)
		})

		It("returns the json encoding of the cell presences", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "cell", "cell-1")
			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(gexec.Exit(0))

			jsonData, err := json.Marshal(presence)
			Expect(err).NotTo(HaveOccurred())
			Expect(sess.Out).To(gbytes.Say(string(jsonData)))
		})

		Context("when the cell does not exist", func() {
			It("exits with status code of 5", func() {
				cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "cell", "cell-id-dsafasdklfjasdlkf")

				sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(gexec.Exit(5))
			})
		})
	})

	Context("when cell command is called with extra arguments", func() {
		It("exits with status code of 3", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "cell", "cell-id", "extra-argument")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(gexec.Exit(3))
		})
	})
})
