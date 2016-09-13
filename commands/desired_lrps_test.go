package commands_test

import (
	"code.cloudfoundry.org/bbs/fake_bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/cfdot/commands"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("DesiredLRPs", func() {
	var (
		fakeBBSClient  *fake_bbs.FakeClient
		desiredLrps    []*models.DesiredLRP
		returnedError  error
		stdout, stderr *gbytes.Buffer
	)

	BeforeEach(func() {
		desiredLrps = nil
		returnedError = nil
		stdout = gbytes.NewBuffer()
		stderr = gbytes.NewBuffer()
		fakeBBSClient = &fake_bbs.FakeClient{}
	})

	JustBeforeEach(func() {
		fakeBBSClient.DesiredLRPsReturns(desiredLrps, returnedError)
	})

	Context("when the bbs responds with desired lrps", func() {
		BeforeEach(func() {
			desiredLrps = []*models.DesiredLRP{
				{
					Instances: 1,
				},
			}
		})

		It("prints a json stream of all the desired lrps", func() {
			err := commands.DesiredLRPs(stdout, stderr, fakeBBSClient, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(stdout).To(gbytes.Say(`"instances":1`))
		})
	})

	Context("when the bbs errors", func() {
		BeforeEach(func() {
			returnedError = models.ErrUnknownError
		})

		It("fails with a relevant error", func() {
			err := commands.DesiredLRPs(stdout, stderr, fakeBBSClient, nil)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(models.ErrUnknownError))
		})
	})
})
