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

var _ = Describe("update-desired-lrp", func() {
	var sess *gexec.Session

	itValidatesBBSFlags("update-desired-lrp")

	Context("when not enough args are provided", func() {
		JustBeforeEach(func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "update-desired-lrp")

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
		})

		It("exits with status 3 and prints an error on stderr", func() {
			Eventually(sess).Should(gexec.Exit(3))
			Expect(sess.Err).To(gbytes.Say(`Missing arguments`))
			Expect(sess.Err).To(gbytes.Say("cfdot update-desired-lrp process-guid \\(SPEC\\|@FILE\\) .*"))
		})
	})

	Context("when bbs responds with 200 status code", func() {
		var (
			lrpUpdate *models.DesiredLRPUpdate
			args      []string
		)

		BeforeEach(func() {
			updatedInstanceCount := int32(5)
			lrpUpdate = &models.DesiredLRPUpdate{
				Instances: &updatedInstanceCount,
			}
		})

		JustBeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/desired_lrp/update"),
					ghttp.VerifyProtoRepresenting(&models.UpdateDesiredLRPRequest{
						Update:      lrpUpdate,
						ProcessGuid: "process-guid",
					}),
					ghttp.RespondWithProto(200, &models.DesiredLRPLifecycleResponse{
						Error: nil,
					}),
				),
			)

			args = append([]string{"--bbsURL", bbsServer.URL(), "update-desired-lrp"}, args...)
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
				spec, err := json.Marshal(lrpUpdate)
				Expect(err).NotTo(HaveOccurred())
				args = []string{"process-guid", string(spec)}
			})

			It("exits with status code of 0", func() {
				Eventually(sess).Should(gexec.Exit(0))
			})
		})

		Context("as a file", func() {
			BeforeEach(func() {
				spec, err := json.Marshal(lrpUpdate)
				Expect(err).NotTo(HaveOccurred())
				f, err := ioutil.TempFile(os.TempDir(), "update_lrp_spec")
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()
				_, err = f.Write(spec)
				Expect(err).NotTo(HaveOccurred())
				args = []string{"process-guid", "@" + f.Name()}
			})

			It("exits with status code 0", func() {
				Eventually(sess).Should(gexec.Exit(0))
			})
		})

		Context("empty spec", func() {
			BeforeEach(func() {
				args = []string{"process-guid"}
			})

			It("exits with status code of 3", func() {
				Eventually(sess).Should(gexec.Exit(3))
			})
		})

		Context("invalid spec", func() {
			BeforeEach(func() {
				args = []string{"process-guid", "foo"}
			})

			It("exits with status code of 3 and prints the error", func() {
				Eventually(sess).Should(gexec.Exit(3))
				Expect(sess.Err).To(gbytes.Say("Invalid JSON:"))
			})
		})

		Context("non-existing spec file", func() {
			BeforeEach(func() {
				args = []string{"process-guid1", "@/path/to/non/existing/file"}
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
					ghttp.VerifyRequest("POST", "/v1/desired_lrp/update"),
					ghttp.RespondWithProto(500, &models.DesiredLRPLifecycleResponse{
						Error: &models.Error{
							Type:    models.Error_Deadlock,
							Message: "deadlock detected",
						},
					}),
				),
			)

			cfdotCmd := exec.Command(
				cfdotPath,
				"--bbsURL", bbsServer.URL(), "update-desired-lrp", "process-guid", "{}",
			)
			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess.Exited).Should(BeClosed())
		})

		It("exits with status 4 and prints the error", func() {
			Eventually(sess).Should(gexec.Exit(4))
			Expect(sess.Err).To(gbytes.Say("Deadlock"))
		})
	})
})
