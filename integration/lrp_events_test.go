package integration_test

import (
	"os/exec"

	"code.cloudfoundry.org/bbs/events"
	"code.cloudfoundry.org/bbs/models"

	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("lrp-events", func() {
	itValidatesBBSFlags("lrp-events")
	itHasNoArgs("lrp-events", false)

	Context("events filtering by cell id", func() {
		var (
			session *gexec.Session
			cmd     *exec.Cmd
		)

		BeforeEach(func() {
			var err error
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			var err error
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Eventually(session).Should(gexec.Exit())
		})

		Context("when the cell id is specified", func() {
			BeforeEach(func() {
				expectedRequest := &models.EventsByCellId{CellId: "some-cell-id"}
				expectedBody, err := proto.Marshal(expectedRequest)
				Expect(err).NotTo(HaveOccurred())
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v1/events"),
						ghttp.VerifyBody(expectedBody),
					),
				)
				cmd = exec.Command(cfdotPath, "lrp-events", "--bbsURL", bbsServer.URL(), "-c", "some-cell-id")
			})

			It("passes the cell id to the bbs client", func() {
				Eventually(bbsServer.ReceivedRequests).Should(HaveLen(1))
			})
		})

		Context("when the cell id is not specified", func() {
			BeforeEach(func() {
				expectedRequest := &models.EventsByCellId{CellId: ""}
				expectedBody, err := proto.Marshal(expectedRequest)
				Expect(err).NotTo(HaveOccurred())
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/v1/events"),
						ghttp.VerifyBody(expectedBody),
					),
				)
				cmd = exec.Command(cfdotPath, "lrp-events", "--bbsURL", bbsServer.URL())
			})

			It("passes empty cell id to the bbs client", func() {
				Eventually(bbsServer.ReceivedRequests).Should(HaveLen(1))
			})
		})

	})

	Context("when the server responds with events", func() {
		BeforeEach(func() {
			lrp := models.DesiredLRP{ProcessGuid: "some-guid"}
			desiredLRPEvent := models.NewDesiredLRPRemovedEvent(&lrp)
			sseEvent, err := events.NewEventFromModelEvent(1, desiredLRPEvent)
			Expect(err).ToNot(HaveOccurred())

			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v1/events"),
					ghttp.RespondWith(200, sseEvent.Encode()),
				),
			)
		})

		It("prints out the event stream", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "lrp-events")

			var err error
			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess).Should(gexec.Exit(0))
			Expect(sess.Out).To(gbytes.Say("some-guid"))
		})
	})

	Context("when there is a BBS error", func() {
		BeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/v1/events"),
					ghttp.RespondWith(418, ""),
				),
			)
		})

		It("responds with a status code 4", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "lrp-events")

			var err error
			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess).Should(gexec.Exit(4))
		})
	})
})
