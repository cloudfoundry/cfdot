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

var _ = Describe("cells", func() {
	var sess *gexec.Session

	itValidatesBBSFlags("cells")
	itHasNoArgs("cells")

	Context("when cells command is called", func() {
		BeforeEach(func() {
			bbsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/v1/cells/list.r1"),
					ghttp.RespondWithProto(200, &models.CellsResponse{
						Cells: []*models.CellPresence{
							{
								CellId:     "cell-1",
								RepAddress: "rep-1",
								Zone:       "zone1",
								Capacity: &models.CellCapacity{
									MemoryMb:   1024,
									DiskMb:     1024,
									Containers: 10,
								},
								RootfsProviders: []*models.Provider{
									{
										Name: "rootfs1",
									},
								},
								RepUrl: "http://rep1.com",
							},
						},
					}),
				),
			)
		})

		JustBeforeEach(func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "cells")

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())

		})

		It("returns the json encoding of the cell presences", func() {
			Eventually(sess).Should(gexec.Exit(0))
			Expect(sess.Out).To(gbytes.Say(`"rep_url":"http://rep1.com"`))
		})
	})
	Context("when cells command is called with extra arguments", func() {
		It("exits with status code of 3", func() {
			cfdotCmd := exec.Command(cfdotPath, "--bbsURL", bbsServer.URL(), "cells", "extra-args")

			var err error
			sess, err = gexec.Start(cfdotCmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(gexec.Exit(3))
		})
	})
})
