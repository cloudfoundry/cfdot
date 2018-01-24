package integration_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/bbs/models"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("create-desired-lrp", func() {
	var sess *gexec.Session

	itValidatesBBSFlags("create-desired-lrp")

	Context("when no spec is passed", func() {
		JustBeforeEach(func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "create-desired-lrp")

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess.Exited).Should(BeClosed())
		})

		It("exits with status code of 3 and prints the error and usage", func() {
			Eventually(sess).Should(gexec.Exit(3))
			Expect(sess.Err).To(gbytes.Say(`missing spec`))
			Expect(sess.Err).To(gbytes.Say("cfdot create-desired-lrp \\(SPEC\\|@FILE\\) .*"))
		})
	})

	Context("when bbs responds with 200 status code", func() {
		var (
			lrp           *models.DesiredLRP
			cfdotArgs     []string
			cmdArgs       []string
			serverTimeout int
		)

		BeforeEach(func() {
			lrp = &models.DesiredLRP{
				ProcessGuid: "some-process-guid",
			}
			cfdotArgs = []string{"--bbsURL", bbsServer.URL()}
			serverTimeout = 0
		})

		JustBeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/desired_lrp/desire.r2"),
					func(w http.ResponseWriter, req *http.Request) {
						time.Sleep(time.Duration(serverTimeout) * time.Second)
					},
					ghttp.VerifyProtoRepresenting(&models.DesireLRPRequest{
						DesiredLrp: lrp,
					}),
					ghttp.RespondWithProto(200, &models.DesiredLRPLifecycleResponse{
						Error: nil,
					}),
				),
			)

			execArgs := append(append(cfdotArgs, "create-desired-lrp"), cmdArgs...)

			cfdotCmd := exec.Command(
				cfdotPath,
				execArgs...,
			)
			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("as json", func() {
			BeforeEach(func() {
				spec, err := json.Marshal(lrp)
				Expect(err).NotTo(HaveOccurred())
				cmdArgs = []string{string(spec)}
			})

			It("exits with status code of 0", func() {
				Eventually(sess).Should(gexec.Exit(0))
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
					})
				})
			})
		})

		Context("as a file", func() {
			BeforeEach(func() {
				spec, err := json.Marshal(lrp)
				Expect(err).NotTo(HaveOccurred())
				f, err := ioutil.TempFile(os.TempDir(), "desired_lrp_spec")
				Expect(err).NotTo(HaveOccurred())
				defer f.Close()
				_, err = f.Write(spec)
				Expect(err).NotTo(HaveOccurred())
				cmdArgs = []string{"@" + f.Name()}
			})

			It("exits with status code 0", func() {
				Eventually(sess).Should(gexec.Exit(0))
			})
		})

		Context("empty spec", func() {
			BeforeEach(func() {
				cmdArgs = nil
			})

			It("exits with status code of 3", func() {
				Eventually(sess).Should(gexec.Exit(3))
			})
		})

		Context("invalid spec", func() {
			BeforeEach(func() {
				cmdArgs = []string{"foo"}
			})

			It("exits with status code of 3 and prints the error", func() {
				Eventually(sess).Should(gexec.Exit(3))
				Expect(sess.Err).To(gbytes.Say("Invalid JSON:"))
			})
		})

		Context("non-existing spec file", func() {
			BeforeEach(func() {
				cmdArgs = []string{"@/path/to/non/existing/file"}
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
					ghttp.VerifyRequest("POST", "/v1/desired_lrp/desire.r2"),
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
				"--bbsURL", bbsServer.URL(), "create-desired-lrp", "{}",
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
