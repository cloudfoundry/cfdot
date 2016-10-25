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

var _ = Describe("task", func() {
	itValidatesBBSFlags("task", "task-guid")

	Context("when the server responds for task", func() {
		var task = &models.Task{
			TaskGuid: "task-guid",
		}

		BeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/tasks/get_by_task_guid.r2"),
					ghttp.VerifyProtoRepresenting(&models.TaskByGuidRequest{TaskGuid: "task-guid"}),
					ghttp.RespondWithProto(200, &models.TaskResponse{Task: task}),
				),
			)
		})

		It("task prints a json representation of the task", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "task", "task-guid")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess).Should(gexec.Exit(0))

			taskJSON, err := json.Marshal(task)
			Expect(err).NotTo(HaveOccurred())
			Expect(sess.Out.Contents()).To(MatchJSON(taskJSON))
		})
	})

	Context("when the server responds with error", func() {
		It("exits with status code 4", func() {
			bbsServer.RouteToHandler(
				"POST",
				"/v1/tasks/get_by_task_guid.r2",
				ghttp.RespondWithProto(200, &models.TaskResponse{
					Error: models.ErrUnknownError,
				}),
			)

			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "task", "task-guid")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess).Should(gexec.Exit(4))

			Expect(sess.Err).To(gbytes.Say("UnknownError"))
		})
	})

	Context("validates that exactly one guid is passed in", func() {
		It("fails with no arugments", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "task")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess).Should(gexec.Exit(3))
			Expect(sess.Err).To(gbytes.Say("Error: Missing arguments"))
		})

		It("fails with 2+ arguments", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "task", "arg1", "arg2")

			sess, err := gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess).Should(gexec.Exit(3))
			Expect(sess.Err).To(gbytes.Say("Error: Too many arguments specified"))
		})
	})
})
