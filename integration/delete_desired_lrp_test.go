package integration_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/bbs/models"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("delete-desired-lrp", func() {
	itValidatesBBSFlags("delete-desired-lrp")

	Context("when a set of invalid arguments is passed", func() {
		Context("when two arguments are passed", func() {
			It("exits with status 3 and prints the usage and the error", func() {
				sess := RunCFDot("delete-desired-lrp", "arg1", "arg2")
				Eventually(sess).Should(gexec.Exit(3))
				Expect(sess.Err).To(gbytes.Say(`Error:`))
				Expect(sess.Err).To(gbytes.Say("cfdot delete-desired-lrp PROCESS_GUID .*"))
			})
		})

		Context("when no arguments are passed", func() {
			It("exits with status 3 and prints the usage and the error", func() {
				sess := RunCFDot("delete-desired-lrp")
				Eventually(sess).Should(gexec.Exit(3))
				Expect(sess.Err).To(gbytes.Say(`Error:`))
				Expect(sess.Err).To(gbytes.Say("cfdot delete-desired-lrp PROCESS_GUID .*"))
			})
		})
	})

	Context("when bbs responds with 200 status code", func() {
		const processGuid = "process-guid"
		var (
			serverTimeout int
		)

		BeforeEach(func() {
			serverTimeout = 0
		})

		JustBeforeEach(func() {
			request := &models.RemoveDesiredLRPRequest{
				ProcessGuid: processGuid,
			}
			response := &models.DesiredLRPLifecycleResponse{}
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/desired_lrp/remove"),
					func(w http.ResponseWriter, req *http.Request) {
						time.Sleep(time.Duration(serverTimeout) * time.Second)
					},
					ghttp.VerifyProtoRepresenting(request.ToProto()),
					ghttp.RespondWithProto(200, response.ToProto()),
				),
			)
		})

		It("exits with status code of 0", func() {
			sess := RunCFDot("delete-desired-lrp", processGuid)
			Eventually(sess).Should(gexec.Exit(0))
		})

		Context("when timeout flag is present", func() {
			var sess *gexec.Session

			BeforeEach(func() {
				sess = RunCFDot("delete-desired-lrp", "--timeout", "1", processGuid)
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
				})
			})
		})
	})

	Context("when bbs responds with non-200 status code", func() {
		BeforeEach(func() {
			response := &models.DesiredLRPLifecycleResponse{
				Error: &models.Error{
					Type:    models.Error_Deadlock,
					Message: "deadlock detected",
				},
			}
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/desired_lrp/remove"),
					ghttp.RespondWithProto(500, response.ToProto()),
				),
			)
		})

		It("exits with status code 4 and prints the error", func() {
			sess := RunCFDot("delete-desired-lrp", "any-process-guid")
			Eventually(sess).Should(gexec.Exit(4))
			Expect(sess.Err).To(gbytes.Say("deadlock"))
		})
	})
})
