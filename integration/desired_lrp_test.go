package integration_test

import (
	"encoding/json"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/bbs/models"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("desired-lrp", func() {
	var (
		sess *gexec.Session
		args []string
	)

	itValidatesBBSFlags("desired-lrp", "test-guid")

	JustBeforeEach(func() {
		args = append([]string{"--bbsURL", bbsServer.URL(), "desired-lrp"}, args...)
		cfdotCmd := exec.Command(cfdotPath, args...)

		var err error
		sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(sess.Exited).Should(BeClosed())
	})

	Context("when no arguments are provided", func() {
		BeforeEach(func() {
			args = []string{}
		})

		It("fails with exit code 3", func() {
			Expect(sess.ExitCode()).To(Equal(3))
		})
		It("prints usage to stderr", func() {
			Expect(sess.Err).To(gbytes.Say("Missing arguments"))
			Expect(sess.Err).To(gbytes.Say("cfdot desired-lrp PROCESS_GUID \\[flags\\]"))
		})
	})

	Context("when two arguments are provided", func() {
		BeforeEach(func() {
			args = []string{"arg1", "arg2"}
		})

		It("fails with exit code 3", func() {
			Expect(sess.ExitCode()).To(Equal(3))
		})
		It("prints usage to stderr", func() {
			Expect(sess.Err).To(gbytes.Say("Too many arguments specified"))
			Expect(sess.Err).To(gbytes.Say("cfdot desired-lrp PROCESS_GUID \\[flags\\]"))
		})
	})

	Context("when an empty argument is provided", func() {
		BeforeEach(func() {
			args = []string{""}
		})

		It("fails with exit code 3", func() {
			Expect(sess.ExitCode()).To(Equal(3))
		})
		It("prints usage to stderr", func() {
			Expect(sess.Err).To(gbytes.Say("Process guid should be non empty string"))
			Expect(sess.Err).To(gbytes.Say("cfdot desired-lrp PROCESS_GUID \\[flags\\]"))
		})
	})

	Context("when a desired-lrp process_guid is provided", func() {
		var (
			desiredLRP *models.DesiredLRP
		)
		BeforeEach(func() {
			args = []string{"test-guid"}
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

			It("exits with status code of 0", func() {
				Expect(sess.ExitCode()).To(Equal(0))
			})

			It("returns the json encoding of the desired lrp scheduling info", func() {
				jsonData, err := json.Marshal(desiredLRP)
				Expect(err).NotTo(HaveOccurred())

				Expect(sess.Out).To(gbytes.Say(string(jsonData)))
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

			It("exits with status code of 4", func() {
				Expect(sess.ExitCode()).To(Equal(4))
			})

			It("prints the error", func() {
				Expect(sess.Err).To(gbytes.Say("deadlock"))
			})
		})

	})
})
