package integration_test

import (
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
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/events/lrp_instances"),
						ghttp.VerifyBody(expectedBody),
					),
				)
				sess := RunCFDot("lrp-events", "-c", "some-cell-id")
				Eventually(sess).Should(gexec.Exit(0))
			})

			It("passes the cell id to the bbs client", func() {
				Eventually(bbsServer.ReceivedRequests).Should(HaveLen(2))
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
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/events/lrp_instances"),
						ghttp.VerifyBody(expectedBody),
					),
				)
				sess := RunCFDot("lrp-events")
				Eventually(sess).Should(gexec.Exit(0))
			})

			It("passes empty cell id to the bbs client", func() {
				Eventually(bbsServer.ReceivedRequests).Should(HaveLen(2))
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
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/events/lrp_instances"),
					ghttp.RespondWith(200, sseEvent.Encode()),
				),
			)
		})

		It("prints out the event stream", func() {
			sess := RunCFDot("lrp-events")
			Eventually(sess).Should(gexec.Exit(0))
			Expect(sess.Out).To(gbytes.Say("some-guid"))
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
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/events/lrp_instances"),
					ghttp.RespondWith(418, ""),
				),
			)
		})

		It("responds with a status code 4", func() {
			sess := RunCFDot("lrp-events")
			Eventually(sess).Should(gexec.Exit(4))
		})
	})
})
