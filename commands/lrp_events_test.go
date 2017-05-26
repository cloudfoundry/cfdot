package commands_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"code.cloudfoundry.org/bbs/events/eventfakes"
	"code.cloudfoundry.org/bbs/fake_bbs"
	"code.cloudfoundry.org/bbs/models"
	"code.cloudfoundry.org/cfdot/commands"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("LRP Events", func() {
	var (
		fakeBBSClient   *fake_bbs.FakeClient
		fakeEventSource *eventfakes.FakeEventSource
		stdout, stderr  *gbytes.Buffer
		lrp             *models.DesiredLRP
	)

	BeforeEach(func() {
		stdout = gbytes.NewBuffer()
		stderr = gbytes.NewBuffer()
		fakeBBSClient = &fake_bbs.FakeClient{}
		fakeEventSource = &eventfakes.FakeEventSource{}
		fakeBBSClient.SubscribeToEventsByCellIDReturns(fakeEventSource, nil)

		lrp = &models.DesiredLRP{
			ProcessGuid: "some-desired-lrp",
		}
		count := 0
		fakeEventSource.NextStub = func() (models.Event, error) {
			count += 1
			if count > 2 {
				return nil, io.EOF
			}
			return models.NewDesiredLRPCreatedEvent(lrp), nil
		}
	})

	It("prints a JSON object", func() {
		event := models.NewDesiredLRPCreatedEvent(lrp)

		lrpEvent := commands.LRPEvent{
			Type: event.EventType(),
			Data: event,
		}
		data, err := json.Marshal(lrpEvent)
		Expect(err).NotTo(HaveOccurred())

		expectedLines := []string{string(data), string(data)}

		err = commands.LRPEvents(stdout, stderr, fakeBBSClient, "")
		Expect(err).NotTo(HaveOccurred())

		stdoutData := stdout.Contents()
		lines := bytes.SplitN(stdoutData, []byte{'\n'}, 3)
		Expect(lines).To(HaveLen(3))

		for i := 0; i < 2; i++ {
			Expect(string(lines[i])).To(Equal(expectedLines[i]))
		}
	})

	It("closes the event stream", func() {
		err := commands.LRPEvents(stdout, stderr, fakeBBSClient, "")
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeEventSource.CloseCallCount()).To(Equal(1))
	})

	Context("when failing to subscribe to events", func() {
		BeforeEach(func() {
			fakeBBSClient.SubscribeToEventsByCellIDReturns(nil, errors.New("failed to connect"))
		})

		It("returns an error", func() {
			err := commands.LRPEvents(stdout, stderr, fakeBBSClient, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to connect"))
		})
	})

	Context("when failing to receive an event", func() {
		BeforeEach(func() {
			fakeEventSource.NextStub = nil
			fakeEventSource.NextReturns(nil, errors.New("boom"))
		})

		It("returns an error", func() {
			err := commands.LRPEvents(stdout, stderr, fakeBBSClient, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("boom"))
		})
	})
})
