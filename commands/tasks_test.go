package commands_test

import (
	"encoding/json"
	"errors"

	"code.cloudfoundry.org/bbs/fake_bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/cfdot/commands"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("tasks", func() {
	Context("Tasks", func() {
		var (
			stdout               *gbytes.Buffer
			bbsClient            *fake_bbs.FakeClient
			testTask1, testTask2 *models.Task
			testData             []*models.Task
			testError            error
		)

		BeforeEach(func() {
			stdout = gbytes.NewBuffer()
			bbsClient = &fake_bbs.FakeClient{}

			testTask1 = &models.Task{TaskGuid: "task-guid"}
			testTask2 = &models.Task{TaskGuid: "another-task-guid"}

			testData = []*models.Task{testTask1, testTask2}
		})

		JustBeforeEach(func() {
			bbsClient.TasksReturns(testData, testError)
		})

		It("fetches tasks from BBS", func() {
			err := commands.Tasks(stdout, nil, bbsClient)
			Expect(err).NotTo(HaveOccurred())
			Expect(bbsClient.TasksCallCount()).To(Equal(1))
		})

		It("outputs some JSON tasks", func() {
			bbsClient.TasksReturns(testData, nil)

			err := commands.Tasks(stdout, nil, bbsClient)
			Expect(err).NotTo(HaveOccurred())

			expectedOutput1, err := json.Marshal(&testTask1)
			Expect(err).NotTo(HaveOccurred())
			expectedOutput2, err := json.Marshal(&testTask2)
			Expect(err).NotTo(HaveOccurred())

			Expect(stdout).To(gbytes.Say(string(expectedOutput1) + "\n" + string(expectedOutput2)))
		})

		Context("when there are no tasks", func() {
			BeforeEach(func() {
				testData = []*models.Task{}
			})

			It("outputs nothing", func() {
				err := commands.Tasks(stdout, nil, bbsClient)
				Expect(err).NotTo(HaveOccurred())
				Expect(stdout.Contents()).To(BeEmpty())
			})
		})

		Context("when BBS returns an error", func() {
			It("should return the error", func() {
				testError := errors.New("barf")
				bbsClient.TasksReturns(nil, testError)
				err := commands.Tasks(stdout, nil, bbsClient)
				Expect(err).To(Equal(testError))
			})
		})

		Context("when Encoder fails", func() {
			It("should return the error", func() {
				err := stdout.Close()
				Expect(err).NotTo(HaveOccurred())
				err = commands.Tasks(stdout, nil, bbsClient)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("ValidateTaskArgs", func() {
		It("succeeds with no arguments", func() {
			Expect(commands.ValidateTasksArgs([]string{})).To(Succeed())
		})

		It("fails with any arguments", func() {
			err := commands.ValidateTasksArgs([]string{"foo"})
			Expect(err).To(Equal(errors.New("Too many arguments specified")))
			err = commands.ValidateTasksArgs([]string{"foo", "bar"})
			Expect(err).To(Equal(errors.New("Too many arguments specified")))
		})
	})
})
