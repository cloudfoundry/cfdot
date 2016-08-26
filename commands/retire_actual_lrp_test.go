package commands_test

import (
	"code.cloudfoundry.org/bbs/fake_bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/cfdot/commands"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("RetireActualLrp", func() {

	var (
		fakeBBSClient  *fake_bbs.FakeClient
		desiredLRP     *models.DesiredLRP
		stdout, stderr *gbytes.Buffer
	)

	BeforeEach(func() {
		stdout = gbytes.NewBuffer()
		stderr = gbytes.NewBuffer()
		fakeBBSClient = &fake_bbs.FakeClient{}

		desiredLRP = &models.DesiredLRP{
			Domain: "test-domain.com",
		}

		fakeBBSClient.RetireActualLRPReturns(nil)
		fakeBBSClient.DesiredLRPByProcessGuidReturns(desiredLRP, nil)
	})

	It("retires the actual lrp", func() {
		err := commands.RetireActualLRP(stdout, stderr, fakeBBSClient, []string{"process-guid", "1"}, "process-guid", 1)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeBBSClient.DesiredLRPByProcessGuidCallCount()).To(Equal(1))
		_, processGuid := fakeBBSClient.DesiredLRPByProcessGuidArgsForCall(0)
		Expect(processGuid).To(Equal("process-guid"))

		Expect(fakeBBSClient.RetireActualLRPCallCount()).To(Equal(1))
		_, actualLRPKey := fakeBBSClient.RetireActualLRPArgsForCall(0)
		Expect(actualLRPKey).To(Equal(&models.ActualLRPKey{
			ProcessGuid: "process-guid",
			Index:       1,
			Domain:      "test-domain.com",
		}))
	})

	Context("when retiring the actual lrp fails", func() {
		BeforeEach(func() {
			fakeBBSClient.RetireActualLRPReturns(models.ErrUnknownError)
		})

		It("fails with a relevant error ", func() {
			err := commands.RetireActualLRP(stdout, stderr, fakeBBSClient, nil, "process-guid", 2)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(models.ErrUnknownError))
		})
	})

	Context("when fetching the desired lrp fails", func() {
		BeforeEach(func() {
			fakeBBSClient.DesiredLRPByProcessGuidReturns(nil, models.ErrUnknownError)
		})

		It("fails with a relevant error ", func() {
			err := commands.RetireActualLRP(stdout, stderr, fakeBBSClient, nil, "process-guid", 2)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(models.ErrUnknownError))

			Expect(fakeBBSClient.DesiredLRPByProcessGuidCallCount()).To(Equal(1))
			Expect(fakeBBSClient.RetireActualLRPCallCount()).To(Equal(0))
		})
	})
})
