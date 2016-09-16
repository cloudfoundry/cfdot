package integration_test

import (
	"os/exec"

	"code.cloudfoundry.org/bbs/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("actual-lrp-groups-for-guid", func() {
	var sess *gexec.Session

	itValidatesBBSFlags("actual-lrp-groups-for-guid", "test-guid")

	Context("when there is no filter", func() {
		JustBeforeEach(func() {
			cfdotCmd := exec.Command(
				cfdotPath,
				"--bbsURL", bbsServer.URL(),
				"actual-lrp-groups-for-guid", "random-guid",
			)

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
		})

		Context("when the server returns a valid response", func() {
			BeforeEach(func() {
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrp_groups/list_by_process_guid"),
						ghttp.VerifyProtoRepresenting(&models.ActualLRPGroupsByProcessGuidRequest{
							ProcessGuid: "random-guid",
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
				Expect(sess.ExitCode()).To(Equal(0))
			})

			It("returns the json encoding of the actual lrp", func() {
				Expect(sess.Out).To(gbytes.Say(`"state":"running"`))
			})
		})

		Context("when the server returns an error", func() {
			BeforeEach(func() {
				bbsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/v1/actual_lrp_groups/list_by_process_guid"),
						ghttp.RespondWithProto(500, &models.DomainsResponse{
							Error: &models.Error{
								Type:    models.Error_Deadlock,
								Message: "the request failed due to deadlock",
							},
							Domains: nil,
						}),
					),
				)
			})

			It("exits with status code 4", func() {
				Expect(sess.ExitCode()).To(Equal(4))
			})

			It("should print the type and message of the error", func() {
				Expect(sess.Err).To(gbytes.Say("BBS error"))
				Expect(sess.Err).To(gbytes.Say("Type 28: Deadlock"))
				Expect(sess.Err).To(gbytes.Say("Message: the request failed due to deadlock"))
			})

			It("should not print the usage", func() {
				Expect(sess.Err).NotTo(gbytes.Say("Usage:"))
			})
		})
	})

	Context("when passing index as filter", func() {
		JustBeforeEach(func() {
			cfdotCmd := exec.Command(
				cfdotPath,
				"--bbsURL", bbsServer.URL(),
				"actual-lrp-groups-for-guid", "test-process-guid",
				"-i", "1",
			)

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

			<-sess.Exited
			Expect(sess.ExitCode()).To(Equal(0))
		})

		BeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/actual_lrp_groups/get_by_process_guid_and_index"),
					ghttp.VerifyProtoRepresenting(&models.ActualLRPGroupByProcessGuidAndIndexRequest{
						ProcessGuid: "test-process-guid",
						Index:       1,
					}),
					ghttp.RespondWithProto(200, &models.ActualLRPGroupsResponse{
						ActualLrpGroups: []*models.ActualLRPGroup{
							{
								Instance: &models.ActualLRP{
									ActualLRPKey: models.ActualLRPKey{
										Index: 1,
									},
									State: "running",
								},
							},
						},
					}),
				),
			)
		})

		It("returns the json encoding of the actual lrp", func() {
			Expect(sess.Out).To(gbytes.Say(`"state":"running"`))
		})
	})
})
