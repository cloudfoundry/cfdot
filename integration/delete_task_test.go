package integration_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/bbs/models"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("delete-task", func() {
	var sess *gexec.Session

	itValidatesBBSFlags("delete-task")

	itFailsWithValidationError := func() {
		It("exits with status 3 and prints the usage and the error", func() {
			Eventually(sess).Should(gexec.Exit(3))
			Expect(sess.Err).To(gbytes.Say(`Error:`))
			Expect(sess.Err).To(gbytes.Say("cfdot delete-task TASK_GUID .*"))
		})
	}

	Context("when a set of invalid arguments is passed", func() {
		var (
			args []string
		)

		JustBeforeEach(func() {
			args = append([]string{"--bbsURL", bbsServer.URL(), "delete-task"}, args...)

			cfdotCmd := exec.Command(cfdotPath, args...)

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when two arguments are passed", func() {
			BeforeEach(func() {
				args = []string{"arg-1", "arg-2"}
			})

			itFailsWithValidationError()
		})

		Context("when no arguments are passed", func() {
			BeforeEach(func() {
				args = []string{}
			})

			itFailsWithValidationError()
		})
	})

	Context("when bbs responds with 200 status code", func() {
		BeforeEach(func() {
			taskGuid := "task-guid"
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/tasks/resolving"),
					ghttp.VerifyProtoRepresenting(&models.TaskGuidRequest{
						TaskGuid: taskGuid,
					}),
					ghttp.RespondWithProto(200, &models.TaskLifecycleResponse{}),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/tasks/delete"),
					ghttp.VerifyProtoRepresenting(&models.TaskGuidRequest{
						TaskGuid: taskGuid,
					}),
					ghttp.RespondWithProto(200, &models.TaskLifecycleResponse{}),
				),
			)

			cfdotCmd := exec.Command(
				cfdotPath,
				[]string{"--bbsURL", bbsServer.URL(), "delete-task", taskGuid}...,
			)
			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
		})

		It("exits with status code of 0", func() {
			Eventually(sess).Should(gexec.Exit(0))
		})
	})

	Context("when bbs responds with non-200 status code", func() {
		BeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/tasks/resolving"),
					ghttp.RespondWithProto(500, &models.TaskLifecycleResponse{
						Error: &models.Error{
							Type:    models.Error_Deadlock,
							Message: "deadlock detected",
						},
					}),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/tasks/delete"),
					ghttp.RespondWithProto(500, &models.TaskLifecycleResponse{
						Error: &models.Error{
							Type:    models.Error_Deadlock,
							Message: "deadlock detected",
						},
					}),
				),
			)

			cfdotCmd := exec.Command(
				cfdotPath,
				"--bbsURL", bbsServer.URL(), "delete-task", "any-task-guid",
			)
			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
		})

		It("exits with status code 4 and prints the error", func() {
			Eventually(sess).Should(gexec.Exit(4))
			Expect(sess.Err).To(gbytes.Say("deadlock"))
		})
	})
})
