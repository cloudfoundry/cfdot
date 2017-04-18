package integration_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/bbs/models"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("create-task", func() {
	var sess *gexec.Session

	itValidatesBBSFlags("create-task")

	Context("when no spec is passed", func() {
		JustBeforeEach(func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "create-task")

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess.Exited).Should(BeClosed())
		})

		It("exits with status code of 3 and prints the error and usage", func() {
			Eventually(sess).Should(gexec.Exit(3))
			Expect(sess.Err).To(gbytes.Say(`missing spec`))
			Expect(sess.Err).To(gbytes.Say("cfdot create-task \\(SPEC\\|@FILE\\) .*"))
		})
	})

	Context("when bbs responds with 200 status code", func() {
		var (
			task *models.Task
			args []string
		)

		BeforeEach(func() {
			task = &models.Task{
				TaskGuid: "some-task-guid",
			}
		})

		JustBeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/tasks/desire.r2"),
					ghttp.VerifyProtoRepresenting(&models.DesireTaskRequest{
						TaskGuid:       task.TaskGuid,
						Domain:         task.Domain,
						TaskDefinition: task.TaskDefinition,
					}),
					ghttp.RespondWithProto(200, &models.TaskLifecycleResponse{
						Error: nil,
					}),
				),
			)

			args = append([]string{"--bbsURL", bbsServer.URL(), "create-task"}, args...)
			cfdotCmd := exec.Command(
				cfdotPath,
				args...,
			)
			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("as json", func() {
			BeforeEach(func() {
				spec, err := json.Marshal(task)
				Expect(err).NotTo(HaveOccurred())
				args = []string{string(spec)}
			})

			It("exits with status code of 0", func() {
				Eventually(sess).Should(gexec.Exit(0))
			})
		})

		Context("as a file", func() {
			BeforeEach(func() {
				spec, err := json.Marshal(task)
				Expect(err).NotTo(HaveOccurred())
				f, err := ioutil.TempFile(os.TempDir(), "desired_task_spec")
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()
				_, err = f.Write(spec)
				Expect(err).NotTo(HaveOccurred())
				args = []string{"@" + f.Name()}
			})

			It("exits with status code 0", func() {
				Eventually(sess).Should(gexec.Exit(0))
			})
		})

		Context("empty spec", func() {
			BeforeEach(func() {
				args = nil
			})

			It("exits with status code of 3", func() {
				Eventually(sess).Should(gexec.Exit(3))
			})
		})

		Context("invalid spec", func() {
			BeforeEach(func() {
				args = []string{"foo"}
			})

			It("exits with status code of 3 and prints the error", func() {
				Eventually(sess).Should(gexec.Exit(3))
				Expect(sess.Err).To(gbytes.Say("Invalid JSON:"))
			})
		})

		Context("non-existing spec file", func() {
			BeforeEach(func() {
				args = []string{"@/path/to/non/existing/file"}
			})

			It("exits with status 3 and prints the error", func() {
				Eventually(sess).Should(gexec.Exit(3))
				Expect(sess.Err).To(gbytes.Say("no such file"))
			})
		})
	})

	Context("when bbs responds with non-200 status code", func() {
		JustBeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/tasks/desire.r2"),
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
				"--bbsURL", bbsServer.URL(), "create-task", "{}",
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
