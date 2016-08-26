package commands_test

import (
	"code.cloudfoundry.org/bbs/fake_bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/cfdot/commands"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("ActualLrpGroupsGuid", func() {
	var (
		fakeBBSClient   *fake_bbs.FakeClient
		actualLRPGroups []*models.ActualLRPGroup
		returnedError   error
		stdout, stderr  *gbytes.Buffer
	)

	BeforeEach(func() {
		actualLRPGroups = nil
		returnedError = nil
		stdout = gbytes.NewBuffer()
		stderr = gbytes.NewBuffer()
		fakeBBSClient = &fake_bbs.FakeClient{}
	})

	JustBeforeEach(func() {
		fakeBBSClient.ActualLRPGroupsByProcessGuidReturns(actualLRPGroups, returnedError)
	})

	Context("when the bbs responds with actual lrp groups for a process id", func() {
		BeforeEach(func() {
			actualLRPGroups = []*models.ActualLRPGroup{
				{
					Instance: &models.ActualLRP{
						State: "running",
					},
				},
			}
		})

		It("prints a json stream of the actual lrp for a process id", func() {
			err := commands.ActualLRPGroupsByProcessGuid(stdout, stderr, fakeBBSClient, []string{"test-id"})
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(gbytes.Say(`"state":"running"`))
		})
	})

	Context("when the bbs errors", func() {
		BeforeEach(func() {
			returnedError = models.ErrUnknownError
		})

		It("fails with a relevant error", func() {
			err := commands.ActualLRPGroupsByProcessGuid(stdout, stderr, fakeBBSClient, []string{"testid"})
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(models.ErrUnknownError))
		})
	})

})
