package commands_test

import (
	"code.cloudfoundry.org/bbs/fake_bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/cfdot/commands"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Set Domain", func() {
	var (
		fakeBBSClient  *fake_bbs.FakeClient
		stdout, stderr *gbytes.Buffer
	)

	BeforeEach(func() {
		stdout = gbytes.NewBuffer()
		stderr = gbytes.NewBuffer()
		fakeBBSClient = &fake_bbs.FakeClient{}
	})

	Context("when the bbs does not respond with an error", func() {
		BeforeEach(func() {
			fakeBBSClient.UpsertDomainReturns(nil)
		})

		It("prints a success message when a domain is given", func() {
			err := commands.SetDomain(stdout, stderr, fakeBBSClient, []string{"anything"}, 40)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when the bbs errors", func() {
		BeforeEach(func() {
			fakeBBSClient.UpsertDomainReturns(models.ErrUnknownError)
		})

		It("fails with a relevant error", func() {
			err := commands.SetDomain(stdout, stderr, fakeBBSClient, []string{"anything"}, 0)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(models.ErrUnknownError))
		})
	})
})
