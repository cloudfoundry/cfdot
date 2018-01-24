package integration_test

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/bbs/models"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("desired-lrp", func() {
	var sess *gexec.Session

	itValidatesBBSFlags("desired-lrp", "test-guid")

	Context("when BBS flags are valid", func() {
		var (
			cfdotArgs []string
			cmdArgs   []string
		)

		BeforeEach(func() {
			cfdotArgs = []string{"--bbsURL", bbsServer.URL()}
		})

		JustBeforeEach(func() {
			execArgs := append(append(cfdotArgs, "desired-lrp"), cmdArgs...)
			cfdotCmd := exec.Command(
				cfdotPath,
				execArgs...,
			)

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when no arguments are provided", func() {
			BeforeEach(func() {
				cmdArgs = []string{}
			})

			It("fails with exit code 3 and prints the usage to stderr", func() {
				Eventually(sess).Should(gexec.Exit(3))
				Expect(sess.Err).To(gbytes.Say("Missing arguments"))
				Expect(sess.Err).To(gbytes.Say("cfdot desired-lrp PROCESS_GUID \\[flags\\]"))
			})
		})

		Context("when two arguments are provided", func() {
			BeforeEach(func() {
				cmdArgs = []string{"arg1", "arg2"}
			})

			It("fails with exit code 3 and prints the usage to stderr", func() {
				Eventually(sess).Should(gexec.Exit(3))
				Expect(sess.Err).To(gbytes.Say("Too many arguments specified"))
				Expect(sess.Err).To(gbytes.Say("cfdot desired-lrp PROCESS_GUID \\[flags\\]"))
			})
		})

		Context("when an empty argument is provided", func() {
			BeforeEach(func() {
				cmdArgs = []string{""}
			})

			It("fails with exit code 3 and prints the usage to stderr", func() {
				Eventually(sess).Should(gexec.Exit(3))
				Expect(sess.Err).To(gbytes.Say("Process guid should be non empty string"))
				Expect(sess.Err).To(gbytes.Say("cfdot desired-lrp PROCESS_GUID \\[flags\\]"))
			})
		})

		Context("when a desired-lrp process_guid is provided", func() {
			var (
				desiredLRP    *models.DesiredLRP
				serverTimeout int
			)

			BeforeEach(func() {
				cmdArgs = []string{"test-guid"}
				serverTimeout = 0
			})

			Context("when bbs responds with 200 status code", func() {
				BeforeEach(func() {
					desiredLRP = &models.DesiredLRP{
						ProcessGuid: "test-guid",
						Instances:   2,
					}

					bbsServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/v1/desired_lrps/get_by_process_guid.r2"),
							func(w http.ResponseWriter, req *http.Request) {
								time.Sleep(time.Duration(serverTimeout) * time.Second)
							},
							ghttp.VerifyProtoRepresenting(&models.DesiredLRPByProcessGuidRequest{
								ProcessGuid: "test-guid",
							}),
							ghttp.RespondWithProto(200, &models.DesiredLRPResponse{
								DesiredLrp: desiredLRP,
								Error:      nil,
							}),
						),
					)

				})

				It("exits with status 0 and returns the json encoding of the desired lrp scheduling info", func() {
					Eventually(sess).Should(gexec.Exit(0))
					jsonData, err := json.Marshal(desiredLRP)
					Expect(err).NotTo(HaveOccurred())

					Expect(sess.Out).To(gbytes.Say(string(jsonData)))
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
							jsonData, err := json.Marshal(desiredLRP)
							Expect(err).NotTo(HaveOccurred())

							Expect(sess.Out).To(gbytes.Say(string(jsonData)))
						})
					})
				})
			})

			Context("when bbs responds with non-200 status code", func() {
				BeforeEach(func() {
					bbsServer.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", "/v1/desired_lrps/get_by_process_guid.r2"),
							ghttp.RespondWithProto(500, &models.DesiredLRPResponse{
								Error: &models.Error{
									Type:    models.Error_Deadlock,
									Message: "deadlock detected",
								},
							}),
						),
					)
				})

				It("exits with status 4 and prints the error", func() {
					Eventually(sess).Should(gexec.Exit(4))
					Expect(sess.Err).To(gbytes.Say("deadlock"))
				})
			})
		})
	})
})
