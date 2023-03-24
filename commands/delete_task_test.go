package commands_test

import (
	"code.cloudfoundry.org/bbs/fake_bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/cfdot/commands"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("DeleteTask", func() {
	var (
		fakeBBSClient  *fake_bbs.FakeClient
		returnedError  error
		stdout, stderr *gbytes.Buffer
		taskGuid       string
	)

	BeforeEach(func() {
		fakeBBSClient = &fake_bbs.FakeClient{}
		stdout = gbytes.NewBuffer()
		stderr = gbytes.NewBuffer()

		fakeBBSClient.DeleteTaskReturns(returnedError)
	})

	It("deletes the task", func() {
		err := commands.DeleteTask(stdout, stderr, fakeBBSClient, taskGuid)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeBBSClient.ResolvingTaskCallCount()).To(Equal(1))
		_, task := fakeBBSClient.ResolvingTaskArgsForCall(0)
		Expect(task).To(Equal(taskGuid))

		Expect(fakeBBSClient.DeleteTaskCallCount()).To(Equal(1))
		_, task = fakeBBSClient.DeleteTaskArgsForCall(0)
		Expect(task).To(Equal(taskGuid))
	})

	Context("when the bbs errors", func() {
		BeforeEach(func() {
			fakeBBSClient.DeleteTaskReturns(models.ErrUnknownError)
		})

		It("fails with a relevant error", func() {
			err := commands.DeleteTask(stdout, stderr, fakeBBSClient, "the-task-guid")
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(models.ErrUnknownError))
		})
	})
})
