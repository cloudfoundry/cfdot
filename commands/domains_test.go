package commands_test

import (
	"code.cloudfoundry.org/bbs/fake_bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/cfdot/commands"
	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("DomainsCommand", func() {
	var (
		fakeBBSClient *fake_bbs.FakeClient
		command       commands.DomainsCommand
		buffer        *gbytes.Buffer
	)

	BeforeEach(func() {
		buffer = gbytes.NewBuffer()
		fakeBBSClient = &fake_bbs.FakeClient{}
		logger := lagertest.NewTestLogger("test")
		commands.Configure(logger, buffer, fakeBBSClient)
	})

	Context("when the bbs responds with domains", func() {
		BeforeEach(func() {
			fakeBBSClient.DomainsReturns([]string{"domain-1", "domain-2"}, nil)
		})

		It("prints a json stream of all the domains", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(buffer).To(gbytes.Say(`"domain-1"\n"domain-2"\n`))
		})
	})

	Context("when the bbs responds with no domains", func() {
		BeforeEach(func() {
			fakeBBSClient.DomainsReturns([]string{}, nil)
		})

		It("returns an empty response", func() {
			err := command.Execute(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(buffer.Contents()).To(BeEmpty())
		})
	})

	Context("when the bbs errors", func() {
		BeforeEach(func() {
			fakeBBSClient.DomainsReturns(nil, models.ErrUnknownError)
		})

		It("fails with a relevant error", func() {
			err := command.Execute(nil)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(models.ErrUnknownError))
		})
	})
})
