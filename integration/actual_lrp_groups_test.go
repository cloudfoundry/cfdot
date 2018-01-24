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

var _ = Describe("actual-lrp-groups", func() {
	var sess *gexec.Session

	itValidatesBBSFlags("actual-lrp-groups")
	itHasNoArgs("actual-lrp-groups", false)

	Context("when no filters are passed", func() {
		var (
			serverTimeout int
			cfdotArgs     []string
		)
		BeforeEach(func() {
			cfdotArgs = []string{"--bbsURL", bbsServer.URL()}
			serverTimeout = 0
		})

		JustBeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/actual_lrp_groups/list"),
					func(w http.ResponseWriter, req *http.Request) {
						time.Sleep(time.Duration(serverTimeout) * time.Second)
					},
					ghttp.RespondWithProto(200, &models.ActualLRPGroupsResponse{
						ActualLrpGroups: []*models.ActualLRPGroup{
							{
								Instance: &models.ActualLRP{
									State: "running",
								},
							},
						},
					}),
				),
			)
			execArgs := append(cfdotArgs, "actual-lrp-groups")
			cfdotCmd := exec.Command(
				cfdotPath,
				execArgs...,
			)

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

		})

		It("returns the json encoding of the actual lrp", func() {
			Eventually(sess).Should(gexec.Exit(0))
			Expect(sess.Out).To(gbytes.Say(`"state":"running"`))
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
					Expect(sess.Out).To(gbytes.Say(`"state":"running"`))
				})
			})
		})
	})

	Context("when passing filters", func() {
		BeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/actual_lrp_groups/list"),
					ghttp.VerifyProtoRepresenting(&models.ActualLRPGroupsRequest{
						Domain: "cf-apps",
						CellId: "cell_z1-0",
					}),
					ghttp.RespondWithProto(200, &models.ActualLRPGroupsResponse{
						ActualLrpGroups: []*models.ActualLRPGroup{
							{
								Instance: &models.ActualLRP{
									State: "running",
								},
							},
						},
					}),
				),
			)
		})

		It("exits with status code of 0", func() {
			cfdotCmd := exec.Command(
				cfdotPath,
				"--bbsURL", bbsServer.URL(),
				"actual-lrp-groups",
				"-d", "cf-apps",
				"-c", "cell_z1-0",
			)

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess).Should(gexec.Exit(0))
		})

		It("exits with status code of 0", func() {
			cfdotCmd := exec.Command(
				cfdotPath,
				"--bbsURL", bbsServer.URL(),
				"actual-lrp-groups",
				"--domain", "cf-apps",
				"--cell-id", "cell_z1-0",
			)

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			Eventually(sess).Should(gexec.Exit(0))
		})
	})
})
