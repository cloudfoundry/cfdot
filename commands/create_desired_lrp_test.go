package commands_test

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/bbs/fake_bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/cfdot/commands"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("CreateDesiredLRP", func() {
	var (
		fakeBBSClient      *fake_bbs.FakeClient
		stdout, stderr     *gbytes.Buffer
		expectedDesiredLRP *models.DesiredLRP
		spec               []byte
	)

	BeforeEach(func() {
		fakeBBSClient = &fake_bbs.FakeClient{}
		stdout = gbytes.NewBuffer()
		stderr = gbytes.NewBuffer()

		expectedDesiredLRP = &models.DesiredLRP{
			ProcessGuid: "some-desired-lrp",
		}
		var err error
		spec, err = json.Marshal(expectedDesiredLRP)
		Expect(err).NotTo(HaveOccurred())
	})

	It("creates the desired lrp", func() {
		err := commands.CreateDesiredLRP(stdout, stderr, fakeBBSClient, spec)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeBBSClient.DesireLRPCallCount()).To(Equal(1))
		_, lrp := fakeBBSClient.DesireLRPArgsForCall(0)
		Expect(lrp).To(Equal(expectedDesiredLRP))
	})

	Context("when a file is passed as an argument", func() {
		var filename string

		BeforeEach(func() {
			f, err := ioutil.TempFile(os.TempDir(), "spec_file")
			Expect(err).NotTo(HaveOccurred())
			defer f.Close()
			_, err = f.Write(spec)
			Expect(err).NotTo(HaveOccurred())
			filename = f.Name()
		})

		It("validates the input file successfully", func() {
			args := []string{"@" + filename}
			actualSpec, err := commands.ValidateCreateDesiredLRPArguments(args)
			Expect(err).NotTo(HaveOccurred())
			Expect(actualSpec).To(Equal(spec))
		})

	})

	Context("when the bbs errors", func() {
		BeforeEach(func() {
			fakeBBSClient.DesireLRPReturns(models.ErrUnknownError)
		})

		It("fails with a relevant error", func() {
			err := commands.CreateDesiredLRP(stdout, stderr, fakeBBSClient, []byte("{}"))
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(models.ErrUnknownError))
		})
	})
})
