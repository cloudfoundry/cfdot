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

var _ = Describe("cancel-task", func() {
	itValidatesBBSFlags("cancel-task", "task-guid")

	Context("when the server responds ok for canceling task", func() {
		var (
			serverTimeout int
			sess          *gexec.Session
			cfdotArgs     []string
		)
		BeforeEach(func() {
			cfdotArgs = []string{"--bbsURL", bbsServer.URL()}
			serverTimeout = 0
		})
		JustBeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/tasks/cancel"),
					func(w http.ResponseWriter, req *http.Request) {
						time.Sleep(time.Duration(serverTimeout) * time.Second)
					},
					ghttp.VerifyProtoRepresenting(&models.TaskGuidRequest{TaskGuid: "task-guid"}),
					ghttp.RespondWith(200, nil),
				),
			)
			execArgs := append(cfdotArgs, "cancel-task", "task-guid")
			cfdotCmd := exec.Command(
				cfdotPath,
				execArgs...,
			)

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
		})

		It("exits 0", func() {
			Eventually(sess).Should(gexec.Exit(0))
			Expect(len(bbsServer.ReceivedRequests())).To(Equal(1))
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
					Expect(len(bbsServer.ReceivedRequests())).To(Equal(1))
				})
			})
		})
	})

	Context("when the server responds with error", func() {
		It("exits with status code 4", func() {
			bbsServer.RouteToHandler(
				"POST",
				"/v1/tasks/cancel",
				ghttp.RespondWithProto(200, &models.TaskResponse{
					Error: models.ErrUnknownError,
				}),
			)

			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "cancel-task", "task-guid")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess).Should(gexec.Exit(4))

			Expect(sess.Err).To(gbytes.Say("UnknownError"))
		})
	})

	Context("when passed invalid arguments", func() {
		It("fails with no arugments and prints the usage", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "cancel-task")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess).Should(gexec.Exit(3))
			Expect(sess.Err).To(gbytes.Say("Error: Missing arguments"))
			Expect(sess.Err).To(gbytes.Say("cfdot cancel-task TASK_GUID \\[flags\\]"))
		})

		It("fails with 2+ arguments", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "cancel-task", "arg1", "arg2")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess).Should(gexec.Exit(3))
			Expect(sess.Err).To(gbytes.Say("Error: Too many arguments specified"))
			Expect(sess.Err).To(gbytes.Say("cfdot cancel-task TASK_GUID \\[flags\\]"))
		})
	})
})
