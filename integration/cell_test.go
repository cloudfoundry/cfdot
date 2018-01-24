package integration_test

import (
	"encoding/json"
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

var _ = Describe("cell", func() {
	itValidatesBBSFlags("cell")

	Context("when cell command is called", func() {
		var (
			presence      *models.CellPresence
			serverTimeout int
			sess          *gexec.Session
			cfdotArgs     []string
			cmdArgs       []string
		)

		BeforeEach(func() {
			presence = &models.CellPresence{
				CellId:     "cell-1",
				RepAddress: "rep-1",
			}
			cfdotArgs = []string{"--bbsURL", bbsServer.URL()}
			cmdArgs = []string{"cell-1"}
			serverTimeout = 0
		})

		JustBeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/cells/list.r1"),
					func(w http.ResponseWriter, req *http.Request) {
						time.Sleep(time.Duration(serverTimeout) * time.Second)
					},
					ghttp.RespondWithProto(200, &models.CellsResponse{
						Cells: []*models.CellPresence{presence},
					}),
				),
			)
			execArgs := append(append(cfdotArgs, "cell"), cmdArgs...)
			cfdotCmd := exec.Command(
				cfdotPath,
				execArgs...,
			)

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the json encoding of the cell presences", func() {
			Eventually(sess).Should(gexec.Exit(0))

			jsonData, err := json.Marshal(presence)
			Expect(err).NotTo(HaveOccurred())
			Expect(sess.Out).To(gbytes.Say(string(jsonData)))
		})

		Context("when the cell does not exist", func() {
			BeforeEach(func() {
				cmdArgs = []string{"cell-id-dsafasdklfjasdlkf"}
			})
			It("exits with status code of 5", func() {
				Eventually(sess).Should(gexec.Exit(5))
			})
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

					jsonData, err := json.Marshal(presence)
					Expect(err).NotTo(HaveOccurred())
					Expect(sess.Out).To(gbytes.Say(string(jsonData)))
				})
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
