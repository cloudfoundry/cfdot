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

var _ = Describe("tasks", func() {
	itValidatesBBSFlags("tasks")

	Context("when the server responds for tasks", func() {
		task := models.Task{
			TaskGuid: "task-guid",
		}
		var (
			sess          *gexec.Session
			cfdotArgs     []string
			serverTimeout int
		)
		BeforeEach(func() {
			cfdotArgs = []string{"--bbsURL", bbsServer.URL()}
			serverTimeout = 0
		})
		JustBeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/tasks/list.r2"),
					func(w http.ResponseWriter, req *http.Request) {
						time.Sleep(time.Duration(serverTimeout) * time.Second)
					},
					ghttp.RespondWithProto(200, &models.TasksResponse{Tasks: []*models.Task{&task}}),
				),
			)

			execArgs := append(cfdotArgs, "tasks")
			cfdotCmd := exec.Command(
				cfdotPath,
				execArgs...,
			)

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the tasks as a stream of json objects", func() {
			Eventually(sess).Should(gexec.Exit(0))

			Expect(bbsServer.ReceivedRequests()).To(HaveLen(1))

			expectedOutput, err := json.Marshal(&task)
			Expect(err).NotTo(HaveOccurred())

			Expect(sess.Out).To(gbytes.Say(string(expectedOutput)))
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

					Expect(bbsServer.ReceivedRequests()).To(HaveLen(1))

					expectedOutput, err := json.Marshal(&task)
					Expect(err).NotTo(HaveOccurred())

					Expect(sess.Out).To(gbytes.Say(string(expectedOutput)))
				})
			})
		})
	})

	Context("when there are filters for tasks", func() {
		Context("with domain filters", func() {
			BeforeEach(func() {
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/tasks/list.r2"),
						ghttp.VerifyProtoRepresenting(&models.TasksRequest{
							Domain: "domain",
						}),
						ghttp.RespondWithProto(200, &models.TaskResponse{
							Task: &models.Task{
								TaskGuid: "task-guid",
							},
						}),
					),
				)
			})

			Context("with -d", func() {
				It("should exit with status code 0", func() {
					cmd := exec.Command(cfdotPath, "tasks", "--bbsURL", bbsServer.URL(), "-d", "domain")
					sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(gexec.Exit(0))
				})
			})

			Context("with --domain", func() {
				It("should exit with status code 0", func() {
					cmd := exec.Command(cfdotPath, "tasks", "--bbsURL", bbsServer.URL(), "--domain", "domain")
					sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(gexec.Exit(0))
				})
			})

			Context("with both --domain and -d", func() {
				It("should exit with a status code of 3", func() {
					cmd := exec.Command(cfdotPath, "tasks", "--bbsURL", bbsServer.URL(), "--domain", "domain", "-d", "domain")
					sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(gexec.Exit(3))
				})
			})
		})

		Context("with cell filters", func() {
			BeforeEach(func() {
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/tasks/list.r2"),
						ghttp.VerifyProtoRepresenting(&models.TasksRequest{
							CellId: "cell-id",
						}),
						ghttp.RespondWithProto(200, &models.TaskResponse{
							Task: &models.Task{
								TaskGuid: "task-guid",
							},
						}),
					),
				)
			})

			Context("with -c", func() {
				It("should exit with status code 0", func() {
					cmd := exec.Command(cfdotPath, "tasks", "--bbsURL", bbsServer.URL(), "-c", "cell-id")
					sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(gexec.Exit(0))
				})
			})

			Context("with --cell-id", func() {
				It("should exit with status code 0", func() {
					cmd := exec.Command(cfdotPath, "tasks", "--bbsURL", bbsServer.URL(), "--cell-id", "cell-id")
					sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(gexec.Exit(0))
				})
			})

			Context("with both --cell-id and -c", func() {
				It("should exit with a status code of 3", func() {
					cmd := exec.Command(cfdotPath, "tasks", "--bbsURL", bbsServer.URL(), "--cell-id", "cell-id", "-c", "cell-id")
					sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(gexec.Exit(3))
				})
			})
		})

		Context("with cell and domain filters", func() {
			BeforeEach(func() {
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/tasks/list.r2"),
						ghttp.VerifyProtoRepresenting(&models.TasksRequest{
							Domain: "domain",
							CellId: "cell-id",
						}),
						ghttp.RespondWithProto(200, &models.TaskResponse{
							Task: &models.Task{
								TaskGuid: "task-guid",
							},
						}),
					),
				)
			})

			It("should exit with status code 0", func() {
				cmd := exec.Command(cfdotPath, "tasks", "--bbsURL", bbsServer.URL(), "-c", "cell-id", "-d", "domain")
				sess, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(gexec.Exit(0))
			})
		})
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
