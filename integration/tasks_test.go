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

var _ = Describe("tasks", func() {
	itValidatesBBSFlags("tasks")

	It("should return the tasks as a stream of json objects", func() {
		task := models.Task{
			TaskGuid: "task-guid",
		}

		bbsServer.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("POST", "/v1/tasks/list.r2"),
				ghttp.RespondWithProto(200, &models.TasksResponse{Tasks: []*models.Task{&task}}),
			),
		)

		cmd := exec.Command(cfdotPath, "tasks", "--bbsURL", bbsServer.URL())
		sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(sess).Should(gexec.Exit(0))

		Expect(bbsServer.ReceivedRequests()).To(HaveLen(1))

		expectedOutput, err := json.Marshal(&task)
		Expect(err).NotTo(HaveOccurred())

		Expect(sess.Out).To(gbytes.Say(string(expectedOutput)))
	})

	Context("when extra args are given", func() {
		It("returns an error and exits with status 3", func() {
			cmd := exec.Command(cfdotPath, "tasks", "garbage", "--bbsURL", bbsServer.URL())
			sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess).Should(gexec.Exit(3))
			Expect(sess.Err).To(gbytes.Say("Too many arguments specified"))
		})
	})

	Context("when the bbs returns an error", func() {
		It("returns an error and exits with status 4", func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/tasks/list.r2"),
					ghttp.RespondWithProto(200, &models.TasksResponse{Error: models.ErrUnknownError}),
				),
			)

			cmd := exec.Command(cfdotPath, "tasks", "--bbsURL", bbsServer.URL())
			sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(gexec.Exit(4))
			Expect(sess.Err).To(gbytes.Say("the request failed for an unknown reason"))
		})
	})
})
